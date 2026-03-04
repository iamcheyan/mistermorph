package channelopts

import (
	"testing"
	"time"

	"github.com/quailyquaily/mistermorph/agent"
)

type stubConfigReader map[string]any

func (s stubConfigReader) GetStringSlice(key string) []string {
	if v, ok := s[key].([]string); ok {
		return append([]string(nil), v...)
	}
	return nil
}
func (s stubConfigReader) GetString(key string) string {
	if v, ok := s[key].(string); ok {
		return v
	}
	return ""
}
func (s stubConfigReader) GetFloat64(key string) float64 {
	if v, ok := s[key].(float64); ok {
		return v
	}
	return 0
}
func (s stubConfigReader) GetDuration(key string) time.Duration {
	if v, ok := s[key].(time.Duration); ok {
		return v
	}
	return 0
}
func (s stubConfigReader) GetInt(key string) int {
	if v, ok := s[key].(int); ok {
		return v
	}
	return 0
}
func (s stubConfigReader) GetInt64(key string) int64 {
	if v, ok := s[key].(int64); ok {
		return v
	}
	return 0
}
func (s stubConfigReader) GetBool(key string) bool {
	if v, ok := s[key].(bool); ok {
		return v
	}
	return false
}

func TestParseTelegramAllowedChatIDs(t *testing.T) {
	got, err := ParseTelegramAllowedChatIDs([]string{" 1 ", "", "-100", "1"})
	if err != nil {
		t.Fatalf("ParseTelegramAllowedChatIDs() error = %v", err)
	}
	if len(got) != 2 {
		t.Fatalf("len(got) = %d, want 2 (%#v)", len(got), got)
	}
	if got[0] != 1 || got[1] != -100 {
		t.Fatalf("got = %#v, want [1 -100]", got)
	}
}

func TestParseTelegramAllowedChatIDsInvalid(t *testing.T) {
	if _, err := ParseTelegramAllowedChatIDs([]string{"abc"}); err == nil {
		t.Fatalf("expected parse error")
	}
}

func TestBuildTelegramRunOptionsTaskTimeoutFallback(t *testing.T) {
	opts, err := BuildTelegramRunOptions(
		TelegramConfig{
			AllowedChatIDsRaw:                    []string{"100"},
			TaskTimeout:                          0,
			GlobalTaskTimeout:                    2 * time.Minute,
			PollTimeout:                          30 * time.Second,
			MaxConcurrency:                       3,
			AgentLimits:                          agent.Limits{ToolRepeatLimit: 9},
			DefaultGroupTriggerMode:              "smart",
			DefaultAddressingConfidenceThreshold: 0.6,
			DefaultAddressingInterjectThreshold:  0.6,
			MultimodalImageSources:               []string{"telegram"},
		},
		TelegramInput{
			BotToken:    "token",
			TaskTimeout: 0,
		},
	)
	if err != nil {
		t.Fatalf("BuildTelegramRunOptions() error = %v", err)
	}
	if opts.TaskTimeout != 2*time.Minute {
		t.Fatalf("task timeout = %v, want 2m", opts.TaskTimeout)
	}
	if len(opts.AllowedChatIDs) != 1 || opts.AllowedChatIDs[0] != 100 {
		t.Fatalf("allowed chat ids = %#v, want [100]", opts.AllowedChatIDs)
	}
	if opts.AgentLimits.ToolRepeatLimit != 9 {
		t.Fatalf("agent tool repeat limit = %d, want 9", opts.AgentLimits.ToolRepeatLimit)
	}
}

func TestBuildSlackRunOptionsTaskTimeoutFallback(t *testing.T) {
	opts := BuildSlackRunOptions(
		SlackConfig{
			TaskTimeout:                          0,
			GlobalTaskTimeout:                    3 * time.Minute,
			MaxConcurrency:                       3,
			AgentLimits:                          agent.Limits{ToolRepeatLimit: 11},
			DefaultGroupTriggerMode:              "smart",
			DefaultAddressingConfidenceThreshold: 0.6,
			DefaultAddressingInterjectThreshold:  0.6,
			MemoryEnabled:                        true,
			MemoryShortTermDays:                  9,
			MemoryInjectionEnabled:               true,
			MemoryInjectionMaxItems:              33,
		},
		SlackInput{
			BotToken:    "xoxb-1",
			AppToken:    "xapp-1",
			TaskTimeout: 0,
		},
	)
	if opts.TaskTimeout != 3*time.Minute {
		t.Fatalf("task timeout = %v, want 3m", opts.TaskTimeout)
	}
	if opts.AgentLimits.ToolRepeatLimit != 11 {
		t.Fatalf("agent tool repeat limit = %d, want 11", opts.AgentLimits.ToolRepeatLimit)
	}
	if !opts.MemoryEnabled || opts.MemoryShortTermDays != 9 || !opts.MemoryInjectionEnabled || opts.MemoryInjectionMaxItems != 33 {
		t.Fatalf("memory options mismatch: %#v", opts)
	}
}

func TestHeartbeatConfigFromReader(t *testing.T) {
	cfg := HeartbeatConfigFromReader(stubConfigReader{
		"heartbeat.enabled":  true,
		"heartbeat.interval": 15 * time.Minute,
	})
	if !cfg.Enabled {
		t.Fatalf("enabled = false, want true")
	}
	if cfg.Interval != 15*time.Minute {
		t.Fatalf("interval = %v, want 15m", cfg.Interval)
	}
}

func TestTelegramConfigFromReaderImageSources(t *testing.T) {
	cfg := TelegramConfigFromReader(stubConfigReader{
		"multimodal.image.sources": []string{" telegram ", "slack"},
	})
	if len(cfg.MultimodalImageSources) != 2 {
		t.Fatalf("MultimodalImageSources len = %d, want 2", len(cfg.MultimodalImageSources))
	}
}

func TestBuildTelegramRunOptionsImageRecognitionEnabledBySource(t *testing.T) {
	opts, err := BuildTelegramRunOptions(
		TelegramConfig{
			AllowedChatIDsRaw:      []string{"100"},
			MultimodalImageSources: []string{" TeLeGrAm "},
		},
		TelegramInput{BotToken: "token"},
	)
	if err != nil {
		t.Fatalf("BuildTelegramRunOptions() error = %v", err)
	}
	if !opts.ImageRecognitionEnabled {
		t.Fatalf("ImageRecognitionEnabled = false, want true when telegram is in sources")
	}
}

func TestBuildTelegramRunOptionsImageRecognitionDisabledWhenSourceMissing(t *testing.T) {
	opts, err := BuildTelegramRunOptions(
		TelegramConfig{
			AllowedChatIDsRaw:      []string{"100"},
			MultimodalImageSources: []string{"slack"},
		},
		TelegramInput{BotToken: "token"},
	)
	if err != nil {
		t.Fatalf("BuildTelegramRunOptions() error = %v", err)
	}
	if opts.ImageRecognitionEnabled {
		t.Fatalf("ImageRecognitionEnabled = true, want false when telegram is not in sources")
	}
}

func TestLineConfigFromReaderAllowedGroupIDs(t *testing.T) {
	cfg := LineConfigFromReader(stubConfigReader{
		"line.allowed_group_ids": []string{"g1", "g2"},
	})
	if len(cfg.AllowedGroupIDsRaw) != 2 {
		t.Fatalf("AllowedGroupIDsRaw len = %d, want 2", len(cfg.AllowedGroupIDsRaw))
	}
}

func TestBuildLineRunOptionsTaskTimeoutFallback(t *testing.T) {
	opts := BuildLineRunOptions(
		LineConfig{
			AllowedGroupIDsRaw:                   []string{"groupA"},
			TaskTimeout:                          0,
			GlobalTaskTimeout:                    4 * time.Minute,
			MaxConcurrency:                       3,
			DefaultGroupTriggerMode:              "smart",
			DefaultAddressingConfidenceThreshold: 0.6,
			DefaultAddressingInterjectThreshold:  0.6,
			AgentLimits:                          agent.Limits{ToolRepeatLimit: 7},
			MultimodalImageSources:               []string{"line"},
		},
		LineInput{
			ChannelAccessToken: "token",
			ChannelSecret:      "secret",
			TaskTimeout:        0,
		},
	)
	if opts.TaskTimeout != 4*time.Minute {
		t.Fatalf("task timeout = %v, want 4m", opts.TaskTimeout)
	}
	if len(opts.AllowedGroupIDs) != 1 || opts.AllowedGroupIDs[0] != "groupA" {
		t.Fatalf("allowed groups = %#v, want [groupA]", opts.AllowedGroupIDs)
	}
	if opts.AgentLimits.ToolRepeatLimit != 7 {
		t.Fatalf("agent tool repeat limit = %d, want 7", opts.AgentLimits.ToolRepeatLimit)
	}
	if !opts.ImageRecognitionEnabled {
		t.Fatalf("ImageRecognitionEnabled = false, want true when line is in sources")
	}
}

func TestBuildLineRunOptionsInputOverridesAndDedupesGroups(t *testing.T) {
	opts := BuildLineRunOptions(
		LineConfig{
			AllowedGroupIDsRaw: []string{"groupA"},
		},
		LineInput{
			AllowedGroupIDs: []string{" groupB ", "groupB", "groupC"},
		},
	)
	if len(opts.AllowedGroupIDs) != 2 || opts.AllowedGroupIDs[0] != "groupB" || opts.AllowedGroupIDs[1] != "groupC" {
		t.Fatalf("allowed groups = %#v, want [groupB groupC]", opts.AllowedGroupIDs)
	}
}
