package integration

import (
	"log/slog"
	"sort"
	"strings"

	"github.com/quailyquaily/mistermorph/internal/toolsutil"
	"github.com/quailyquaily/mistermorph/secrets"
	"github.com/quailyquaily/mistermorph/tools"
	"github.com/quailyquaily/mistermorph/tools/builtin"
)

func (rt *Runtime) buildRegistry(cfg registrySnapshot, logger *slog.Logger) *tools.Registry {
	r := tools.NewRegistry()
	if rt == nil {
		return r
	}
	if logger == nil {
		logger = slog.Default()
	}

	selectedBuiltinTools := make(map[string]bool, len(rt.builtinToolNames))
	for _, name := range rt.builtinToolNames {
		selectedBuiltinTools[name] = true
		if !toolsutil.IsKnownBuiltinToolName(name) {
			logger.Warn("unknown_builtin_tool_name", "name", name)
		}
	}

	userAgent := strings.TrimSpace(cfg.UserAgent)
	secretsEnabled := cfg.SecretsEnabled
	secretsRequireSkillProfiles := cfg.SecretsRequireSkillProfiles

	allowProfiles := make(map[string]bool)
	for _, id := range cfg.SecretsAllowProfiles {
		id = strings.TrimSpace(id)
		if id == "" {
			continue
		}
		allowProfiles[id] = true
	}

	authProfiles := copyAuthProfilesMap(cfg.AuthProfiles)

	for _, p := range authProfiles {
		if err := p.Validate(); err != nil {
			logger.Warn("auth_profile_invalid", "profile", p.ID, "err", err)
			delete(authProfiles, p.ID)
		}
	}

	if secretsEnabled {
		logger.Info("secrets_enabled",
			"require_skill_profiles", secretsRequireSkillProfiles,
			"allow_profiles", keysSorted(allowProfiles),
			"auth_profiles", len(authProfiles),
		)
	} else {
		if len(allowProfiles) > 0 || len(authProfiles) > 0 {
			logger.Warn("secrets_disabled_but_configured",
				"allow_profiles", keysSorted(allowProfiles),
				"auth_profiles", len(authProfiles),
			)
		}
	}

	secretsAliases := copyStringMap(cfg.SecretsAliases)
	resolver := &secrets.EnvResolver{Aliases: secretsAliases}
	profileStore := secrets.NewProfileStore(authProfiles)

	toolsutil.RegisterStaticTools(r, toolsutil.StaticRegistryConfig{
		Common: toolsutil.StaticCommonConfig{
			UserAgent:      userAgent,
			FileCacheDir:   strings.TrimSpace(cfg.FileCacheDir),
			FileStateDir:   strings.TrimSpace(cfg.FileStateDir),
			SecretsEnabled: secretsEnabled,
		},
		ReadFile: toolsutil.StaticReadFileConfig{
			MaxBytes:  cfg.ToolsReadFileMaxBytes,
			DenyPaths: append([]string(nil), cfg.ToolsReadFileDenyPaths...),
		},
		WriteFile: toolsutil.StaticWriteFileConfig{
			Enabled:  cfg.ToolsWriteFileEnabled,
			MaxBytes: cfg.ToolsWriteFileMaxBytes,
		},
		Bash: toolsutil.StaticBashConfig{
			Enabled:        cfg.ToolsBashEnabled,
			Timeout:        cfg.ToolsBashTimeout,
			MaxOutputBytes: cfg.ToolsBashMaxOutputBytes,
			DenyPaths:      append([]string(nil), cfg.ToolsBashDenyPaths...),
		},
		URLFetch: toolsutil.StaticURLFetchConfig{
			Enabled:          cfg.ToolsURLFetchEnabled,
			Timeout:          cfg.ToolsURLFetchTimeout,
			MaxBytes:         cfg.ToolsURLFetchMaxBytes,
			MaxBytesDownload: cfg.ToolsURLFetchMaxBytesDownload,
			Auth: &builtin.URLFetchAuth{
				Enabled:       secretsEnabled,
				AllowProfiles: allowProfiles,
				Profiles:      profileStore,
				Resolver:      resolver,
			},
		},
		WebSearch: toolsutil.StaticWebSearchConfig{
			Enabled:    cfg.ToolsWebSearchEnabled,
			Timeout:    cfg.ToolsWebSearchTimeout,
			MaxResults: cfg.ToolsWebSearchMaxResults,
			BaseURL:    cfg.ToolsWebSearchBaseURL,
		},
		TodoUpdate: toolsutil.StaticTodoUpdateConfig{
			Enabled:      cfg.ToolsTodoUpdateEnabled,
			TODOPathWIP:  cfg.TODOPathWIP,
			TODOPathDone: cfg.TODOPathDone,
			ContactsDir:  cfg.ContactsDir,
		},
		ContactsSend: toolsutil.StaticContactsSendConfig{
			Enabled:          cfg.ToolsContactsSendEnabled,
			ContactsDir:      cfg.ContactsDir,
			TelegramBotToken: strings.TrimSpace(cfg.TelegramBotToken),
			TelegramBaseURL:  strings.TrimSpace(cfg.TelegramBaseURL),
			SlackBotToken:    strings.TrimSpace(cfg.SlackBotToken),
			SlackBaseURL:     strings.TrimSpace(cfg.SlackBaseURL),
			FailureCooldown:  cfg.ContactsFailureCooldown,
		},
	}, selectedBuiltinTools)

	return r
}

func keysSorted(m map[string]bool) []string {
	if len(m) == 0 {
		return nil
	}
	out := make([]string, 0, len(m))
	for k := range m {
		out = append(out, k)
	}
	sort.Strings(out)
	return out
}
