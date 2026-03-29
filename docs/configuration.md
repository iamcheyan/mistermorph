# Configuration

This page holds the config details moved out of the top-level README.

The canonical config template is [../assets/config/config.example.yaml](../assets/config/config.example.yaml).

## Sources and Precedence

`mistermorph` uses Viper. You can configure it with:

- CLI flags
- environment variables
- a config file

Precedence:

`CLI flag > MISTER_MORPH_* env > config.yaml > default`

Supported config file formats:

- `.yaml`
- `.yml`
- `.json`
- `.toml`
- `.ini`

Env var rules:

- prefix: `MISTER_MORPH_`
- nested keys: replace `.` and `-` with `_`
- example: `tools.bash.enabled` -> `MISTER_MORPH_TOOLS_BASH_ENABLED=true`

All string values in config support `${ENV_VAR}` expansion.

## Common CLI Flags

Global flags:

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

`run`:

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

`submit`:

- `--task`
- `--server-url`
- `--auth-token`
- `--model`
- `--submit-timeout`
- `--wait`
- `--poll-interval`

`console serve`:

- `--console-listen`
- `--console-base-path`
- `--console-static-dir`
- `--console-session-ttl`
- `--allow-empty-password`

`telegram`:

- `--telegram-bot-token`
- `--telegram-allowed-chat-id` (repeatable)
- `--telegram-group-trigger-mode`
- `--telegram-addressing-confidence-threshold`
- `--telegram-addressing-interject-threshold`
- `--telegram-poll-timeout`
- `--telegram-task-timeout`
- `--telegram-max-concurrency`

`slack`:

- `--slack-bot-token`
- `--slack-app-token`
- `--slack-allowed-team-id` (repeatable)
- `--slack-allowed-channel-id` (repeatable)
- `--slack-group-trigger-mode`
- `--slack-addressing-confidence-threshold`
- `--slack-addressing-interject-threshold`
- `--slack-task-timeout`
- `--slack-max-concurrency`

`skills`:

- `skills list --skills-dir` (repeatable)
- `skills install --dest --dry-run --clean --skip-existing --timeout --max-bytes --yes`

`install`:

- `install [dir]`
- `--yes`

## Common Environment Variables

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

Provider-specific values use the same mapping. Examples:

- `llm.azure.deployment` -> `MISTER_MORPH_LLM_AZURE_DEPLOYMENT`
- `llm.bedrock.model_arn` -> `MISTER_MORPH_LLM_BEDROCK_MODEL_ARN`

## Key Config Areas

Core LLM:

- `llm.provider` selects the backend.
- Most providers use `llm.endpoint`, `llm.api_key`, and `llm.model`.
- Azure uses `llm.azure.deployment`.
- Bedrock uses `llm.bedrock.*`.
- `llm.tools_emulation_mode` controls tool-call emulation for models without native tool calling.
- `llm.profiles` defines named profile overrides.
- `llm.fallback_profiles` defines an ordered fallback chain for the implicit default profile when it hits transient errors (`timeout`, `429`, `529`).
- `llm.routes` routes semantic purposes such as `main_loop`, `addressing`, `heartbeat`, `plan_create`, and `memory_draft`.

Logging and runtime limits:

- `logging.level`
- `logging.format`
- `logging.include_thoughts`
- `logging.include_tool_params`
- `max_steps`
- `parse_retries`
- `max_token_budget`
- `timeout`

Skills:

- `skills.enabled`
- `skills.load`
- `file_state_dir`
- `skills.dir_name`

Tools:

- all tool toggles live under `tools.*`
- examples: `tools.bash.enabled`, `tools.url_fetch.enabled`

Console:

- `console.listen`
- `console.base_path`
- `console.static_dir`
- `console.session_ttl`

Auth profiles and secrets:

- `secrets.allow_profiles` is the runtime allowlist.
- `auth_profiles.<id>.credential.secret` holds the secret value.
- Use `${ENV_VAR}` for secret references.

If you configure at least one allowlisted auth profile, `bash` still works but `curl` is denied by default. Use `url_fetch` for authenticated HTTP.

## Example

```yaml
llm:
  provider: openai
  model: gpt-5.4
  api_key: "${OPENAI_API_KEY}"
  profiles:
    reasoning:
      provider: xai
      model: grok-4.1-fast-reasoning
      api_key: "${XAI_API_KEY}"
```
