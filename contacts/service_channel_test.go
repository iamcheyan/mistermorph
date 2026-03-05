package contacts

import (
	"strings"
	"testing"
)

func TestResolveDecisionChannel_LineChatHint(t *testing.T) {
	channel, err := ResolveDecisionChannel(Contact{
		ContactID: "tg:@alice",
		Channel:   ChannelTelegram,
	}, ShareDecision{
		ChatID: "line:Cgroup001",
	})
	if err != nil {
		t.Fatalf("ResolveDecisionChannel() error = %v", err)
	}
	if channel != ChannelLine {
		t.Fatalf("channel mismatch: got %q want %q", channel, ChannelLine)
	}
}

func TestResolveDecisionChannel_LineTargetFallback(t *testing.T) {
	channel, err := ResolveDecisionChannel(Contact{
		ContactID:   "line_user:U100",
		Channel:     ChannelLine,
		LineUserID:  "U100",
		LineChatIDs: []string{"Cgroup001"},
	}, ShareDecision{})
	if err != nil {
		t.Fatalf("ResolveDecisionChannel() error = %v", err)
	}
	if channel != ChannelLine {
		t.Fatalf("channel mismatch: got %q want %q", channel, ChannelLine)
	}
}

func TestResolveDecisionChannel_LineUserContactIDFallback(t *testing.T) {
	channel, err := ResolveDecisionChannel(Contact{
		ContactID: "line_user:U101",
	}, ShareDecision{})
	if err != nil {
		t.Fatalf("ResolveDecisionChannel() error = %v", err)
	}
	if channel != ChannelLine {
		t.Fatalf("channel mismatch: got %q want %q", channel, ChannelLine)
	}
}

func TestResolveDecisionChannel_InvalidProtocolHint(t *testing.T) {
	_, err := ResolveDecisionChannel(Contact{
		ContactID: "contact:test",
		Channel:   ChannelTelegram,
	}, ShareDecision{
		ChatID: "discord:123",
	})
	if err == nil {
		t.Fatalf("ResolveDecisionChannel() expected error")
	}
	if !strings.Contains(strings.ToLower(err.Error()), "invalid chat_id") {
		t.Fatalf("ResolveDecisionChannel() error mismatch: got %q", err.Error())
	}
}

func TestResolveDecisionChannel_MissingProtocolHint(t *testing.T) {
	_, err := ResolveDecisionChannel(Contact{
		ContactID: "contact:test",
		Channel:   ChannelTelegram,
	}, ShareDecision{
		ChatID: "-1001981343441",
	})
	if err == nil {
		t.Fatalf("ResolveDecisionChannel() expected error")
	}
	if !strings.Contains(strings.ToLower(err.Error()), "invalid chat_id") {
		t.Fatalf("ResolveDecisionChannel() error mismatch: got %q", err.Error())
	}
}
