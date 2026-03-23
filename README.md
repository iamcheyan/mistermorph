# Mister Morph

Unified Agent CLI + reusable Go agent core.

Other languages: [简体中文](docs/zh-CN/README.md) | [日本語](docs/ja-JP/README.md)

## Table of contents

- [Why Mister Morph](#why-mistermorph)
- [Quickstart](#quickstart)
- [Start Modes](#start-modes)
- [Supported Models](#supported-models)
- [Embedding](#embedding-to-other-projects)
- [Built-in Tools](#built-in-tools)
- [Skills](#skills)
- [Security](#security)
- [Troubleshoots](#troubleshoots)
- [Debug](#debug)
- [Configuration](#configuration)

## Why Mister Morph

What makes this project worth looking at:

- 🧩 **Reusable Go core**: Run the agent as a CLI, or embed it as a library/subprocess in other apps.
- 🔒 **Serious secure defaults**: Profile-based credential injection, Guard redaction, outbound policy controls, and async approvals with audit trails (see [docs/security.md](docs/security.md)).
- 🧰 **Practical Skills system**: Discover + inject `SKILL.md` from `file_state_dir/skills`, with simple on/off control (see [docs/skills.md](docs/skills.md)).
- 📚 **Beginner-friendly**: Built as a learning-first agent project, with detailed design docs in `docs/` and practical debugging tools like `--inspect-prompt` and `--inspect-request`.

## Quickstart

### Step 1: Install

Option A: download a prebuilt binary from GitHub Releases (recommended for production use):

```bash
curl -fsSL -o /tmp/install-mistermorph.sh https://raw.githubusercontent.com/quailyquaily/mistermorph/refs/heads/master/scripts/install-release.sh
sudo bash /tmp/install-mistermorph.sh
```

The installer supports:

- `bash install-release.sh <version-tag>`
- `INSTALL_DIR=$HOME/.local/bin bash install-release.sh <version-tag>`

Option B: install from source with Go:

```bash
go install github.com/quailyquaily/mistermorph/cmd/mistermorph@latest
```

### Step 2: Bootstrap the agent workspace

```bash
mistermorph install
# or
mistermorph install <dir>
```

The `install` command initializes `config.yaml` plus the core onboarding markdown files under `~/.morph/` (or a specified directory via `<dir>`).

When `config.yaml` does not already exist in the install target, `install` first tries to find a readable config in this order:

1. `--config` path
2. `<dir>/config.yaml`
3. `~/.morph/config.yaml`

If none is found, `install` runs an interactive setup wizard (TTY only) before writing `config.yaml`:

1. select LLM provider (`openai-compatible|gemini|anthropic|cloudflare`)
2. fill provider-specific required fields
3. set model
4. fill `IDENTITY.md`
5. choose a `SOUL.md` preset or customize it

Use `mistermorph install --yes` to skip interactive prompts.

### Step 3: Setup an API key

You can run without a `config.yaml` by using environment variables:

```bash
export MISTER_MORPH_LLM_API_KEY="YOUR_OPENAI_API_KEY_HERE"
# Optional explicit defaults:
export MISTER_MORPH_LLM_PROVIDER="openai"
export MISTER_MORPH_LLM_MODEL="gpt-5.4"
```

If you prefer file-based config, use `~/.morph/config.yaml`.

## Start Modes

### One-time run

```bash
mistermorph run --task "Hello!"
```

### Desktop App

The desktop wrapper starts the Console UI for you and handles first-run setup inside the app.

See [`docs/app.md`](docs/app.md) for build and run details.

### Other modes

Console server mode, Telegram, Slack, LINE, Lark, and runtime endpoint submission are documented in [`docs/modes.md`](docs/modes.md).

## Supported Models

> Model support may vary by specific model ID, provider endpoint capability, and tool-calling behavior.

| Model family | Model range | Status |
|---|---|---|
| GPT | `gpt-5*` | ✅ Full |
| GPT-OSS | `gpt-oss-120b` | ✅ Full |
| Grok | `grok-4+` | ✅ Full |
| Claude | `claude-3.5+` | ✅ Full |
| DeepSeek | `deepseek-3*` | ✅ Full |
| Gemini | `gemini-2.5+` | ✅ Full |
| Kimi | `kimi-2.5+` | ✅ Full |
| MiniMax | `minimax* / minimax-m2.5+` | ✅ Full |
| GLM | `glm-4.6+` | ✅ Full |
| Cloudflare Workers AI | `Workers AI model IDs` | ⚠️ Limited (no tool calling) |

## Embedding to other projects

See [`docs/integration.md`](docs/integration.md) for embedding patterns and examples.

## Built-in Tools

Core tools available to the agent:

- `read_file`: read local text files.
- `write_file`: write local text files under `file_cache_dir` or `file_state_dir`.
- `bash`: run a shell command (disabled by default).
- `url_fetch`: HTTP fetch with optional auth profiles.
- `web_search`: web search (DuckDuckGo HTML).
- `plan_create`: generate a structured plan.

Channel runtime tools:

- `telegram_send_file`: send a file in Telegram (Telegram only).
- `telegram_send_photo`: send a photo in Telegram (Telegram only).
- `telegram_send_voice`: send a voice message in Telegram (Telegram only).
- `message_react`: add an emoji reaction to a message (Telegram/Slack runtime; channel-specific params).

Please see [`docs/tools.md`](docs/tools.md) for detailed tool documentation.

## Skills

`mistermorph` discovers skills under `file_state_dir/skills` (recursively), and injects selected `SKILL.md` content into the system prompt.

By default, `run` uses `skills.enabled=true`; `skills.load=[]` loads all discovered skills, and unknown skill names are ignored.
You can also reference a skill directly in task text with `$skill-name` (or `$skill-id`) to trigger that skill for the run.

Docs: [`docs/skills.md`](docs/skills.md).

```bash
# list available skills
mistermorph skills list
# Use a specific skill in the run command
mistermorph run --task "..." --skills-enabled --skill skill-name
# install remote skills 
mistermorph skills install <remote-skill-url> 
```

### Security Mechanisms for Skills

1. Install audit: When installing remote skills, Mister Morph will preview the skill content and do a basic security audit (e.g., look for dangerous commands in scripts) before asking for user confirmation.
2. Auth profiles: Skills can declare required auth profiles in the `auth_profiles` field. This is only a declaration for prompt/context clarity. Actual permission still comes only from host config via `secrets.allow_profiles` and `auth_profiles` (see `assets/skills/moltbook` and the config sections).

## Security

Recommended systemd hardening and secret handling: [`docs/security.md`](docs/security.md).

## Troubleshoots

Known issues and workarounds: [`docs/troubleshoots.md`](docs/troubleshoots.md).

## Debug

### Logging

There is an argument `--log-level` set for logging level and format:

```bash
mistermorph run --log-level debug --task "..."
```

### Dump internal debug data

There are 2 arguments `--inspect-prompt`/`--inspect-request` for dumping internal state for debugging:

```bash
mistermorph run --inspect-prompt --inspect-request --task "..."
```

These arguments will dump the final system/user/tool prompts and the full LLM request/response JSON as plain text files to `./dump` directory. 

## Configuration

`mistermorph` uses Viper, so you can configure it via flags, env vars, or a config file.

- Config file: `--config /path/to/config.yaml` (supports `.yaml/.yml/.json/.toml/.ini`)
- Env var prefix: `MISTER_MORPH_`
- Nested keys: replace `.` and `-` with `_` (e.g. `tools.bash.enabled` → `MISTER_MORPH_TOOLS_BASH_ENABLED=true`)


### CLI flags

**Global (all commands)**
- `--config`
- `--log-level`
- `--log-format`
- `--log-add-source`
- `--log-include-thoughts`
- `--log-include-tool-params`
- `--log-include-skill-contents`
- `--log-max-thought-chars`
- `--log-max-json-bytes`
- `--log-max-string-value-chars`
- `--log-max-skill-content-chars`
- `--log-redact-key` (repeatable)

**run**
- `--task`
- `--provider`
- `--endpoint`
- `--model`
- `--api-key`
- `--llm-request-timeout`
- `--interactive`
- `--skills-dir` (repeatable)
- `--skill` (repeatable)
- `--skills-enabled`
- `--max-steps`
- `--parse-retries`
- `--max-token-budget`
- `--timeout`
- `--inspect-prompt`
- `--inspect-request`

**submit**
- `--task`
- `--server-url`
- `--auth-token`
- `--model`
- `--submit-timeout`
- `--wait`
- `--poll-interval`

**console serve**
- `--console-listen`
- `--console-base-path`
- `--console-static-dir` (optional; overrides the embedded Console SPA/static root)
- `--console-session-ttl`

**telegram**
- `--telegram-bot-token`
- `--telegram-allowed-chat-id` (repeatable)
- `--telegram-group-trigger-mode` (`strict|smart|talkative`)
- `--telegram-addressing-confidence-threshold`
- `--telegram-addressing-interject-threshold`
- `--telegram-poll-timeout`
- `--telegram-task-timeout`
- `--telegram-max-concurrency`

**slack**
- `--slack-bot-token`
- `--slack-app-token`
- `--slack-allowed-team-id` (repeatable)
- `--slack-allowed-channel-id` (repeatable)
- `--slack-group-trigger-mode` (`strict|smart|talkative`)
- `--slack-addressing-confidence-threshold`
- `--slack-addressing-interject-threshold`
- `--slack-task-timeout`
- `--slack-max-concurrency`

**skills**
- `skills list --skills-dir` (repeatable)
- `skills install --dest --dry-run --clean --skip-existing --timeout --max-bytes --yes`

**install**
- `install [dir]`
- `--yes`

### Environment variables

Common env vars (these map to config keys):

- `MISTER_MORPH_CONFIG`
- `MISTER_MORPH_LLM_PROVIDER`
- `MISTER_MORPH_LLM_ENDPOINT`
- `MISTER_MORPH_LLM_MODEL`
- `MISTER_MORPH_LLM_API_KEY`
- `MISTER_MORPH_LLM_REQUEST_TIMEOUT`
- `MISTER_MORPH_LOGGING_LEVEL`
- `MISTER_MORPH_LOGGING_FORMAT`
- `MISTER_MORPH_SERVER_AUTH_TOKEN`
- `MISTER_MORPH_CONSOLE_PASSWORD`
- `MISTER_MORPH_CONSOLE_PASSWORD_HASH`
- `MISTER_MORPH_TELEGRAM_BOT_TOKEN`
- `MISTER_MORPH_SLACK_BOT_TOKEN`
- `MISTER_MORPH_SLACK_APP_TOKEN`
- `MISTER_MORPH_FILE_CACHE_DIR`

Provider-specific settings use the same mapping, for example:
- `llm.azure.deployment` → `MISTER_MORPH_LLM_AZURE_DEPLOYMENT`
- `llm.bedrock.model_arn` → `MISTER_MORPH_LLM_BEDROCK_MODEL_ARN`

Tool toggles and limits also map to env vars, for example:

- `MISTER_MORPH_TOOLS_BASH_ENABLED`
- `MISTER_MORPH_TOOLS_URL_FETCH_ENABLED`
- `MISTER_MORPH_TOOLS_URL_FETCH_MAX_BYTES`

All string values in config support `${ENV_VAR}` syntax for environment variable expansion (e.g. `api_key: "${OPENAI_API_KEY}"`).

Key meanings (see `assets/config/config.example.yaml` for the canonical list):
- Core: `llm.provider` selects the backend. Most providers use `llm.endpoint`/`llm.api_key`/`llm.model`. Optional defaults `llm.temperature`, `llm.reasoning_effort`, and `llm.reasoning_budget_tokens` are forwarded to `uniai` only when set. Azure uses `llm.azure.deployment` for deployment name, while endpoint/key are still read from `llm.endpoint` and `llm.api_key`. Bedrock uses `llm.bedrock.*`. `llm.tools_emulation_mode` controls tool-call emulation for models without native tool calling (`off|fallback|force`). `llm.profiles` defines named profile overrides, and `llm.routes` routes semantic purposes such as `main_loop`, `addressing`, `heartbeat`, `plan_create`, and `memory_draft`.
- LLM secrets: use `${ENV_VAR}` syntax in any string field to reference environment variables. Example:
  ```yaml
  llm:
    api_key: "${OPENAI_API_KEY}"
    profiles:
      reasoning:
        provider: xai
        model: grok-4.1-fast-reasoning
        api_key: "${XAI_API_KEY}"
  ```
- LLM precedence: for config-sourced values, precedence is `CLI flag > MISTER_MORPH_* env > config.yaml > default`.
- Logging: `logging.level` (`info` shows progress; `debug` adds thoughts), `logging.format` (`text|json`), plus `logging.include_thoughts` and `logging.include_tool_params` (redacted).
- Loop: `max_steps` limits tool-call rounds; `parse_retries` retries invalid JSON; `max_token_budget` is a cumulative token cap (0 disables); `timeout` is the overall run timeout.
- Skills: `skills.enabled` controls whether skills are used; `file_state_dir` + `skills.dir_name` define the default skills root; `skills.load=[]` loads all discovered skills, otherwise it loads only listed skills (unknown names are ignored).
- Tools: all tool toggles live under `tools.*` (e.g. `tools.bash.enabled`, `tools.url_fetch.enabled`) with per-tool limits and timeouts.
- Auth profiles: `secrets.allow_profiles` is the runtime allowlist. `auth_profiles.<id>.credential.secret` holds the secret value (use `${ENV_VAR}` to reference env vars). If at least one allowlisted auth profile is configured, `bash` still works but `curl` is denied by default; authenticated HTTP should go through `url_fetch + auth_profile`.

## Star History

## Star History

[![Star History Chart](https://api.star-history.com/svg?repos=quailyquaily/mistermorph&type=date&legend=top-left)](https://www.star-history.com/#quailyquaily/mistermorph&type=date&legend=top-left)
