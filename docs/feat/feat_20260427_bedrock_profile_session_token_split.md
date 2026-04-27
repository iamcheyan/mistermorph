---
title: Bedrock profile/session-token split plan
---

# Bedrock profile/session-token split plan

## Summary

This local branch validates the full Bedrock credential flow end-to-end:

- `llm.bedrock.aws_profile`
- `llm.bedrock.aws_session_token`
- AWS shared config / credentials resolution
- passing the resolved session token into the Bedrock client used by `uniai`

The branch is useful for local functional validation, but it is not the right final upstream PR shape because it currently uses:

```go
replace github.com/quailyquaily/uniai => ./third_party/uniai
```

That local replace vendors a full copy of `uniai` into `mistermorph` only so the end-to-end flow can be tested before upstreaming the dependency change.

## Why this should be split into two PRs

The feature boundary crosses two repositories:

1. `uniai`
   It owns the Bedrock provider implementation and is the place where `AwsSessionToken` must actually be accepted and forwarded into AWS credentials.

2. `mistermorph`
   It owns config parsing, profile inheritance, and resolution of AWS credentials from shared config / shared credentials files.

If `mistermorph` is changed without first updating `uniai`, the resolved session token stops at the `uniai.Config` boundary and never reaches the Bedrock request path.

If the whole local validation branch is proposed directly to `mistermorph`, the PR becomes much larger than necessary because it includes a vendored `third_party/uniai` tree instead of a normal upstream dependency update.

## What the local replace is doing

For local validation this branch rewires the module dependency:

```go
replace github.com/quailyquaily/uniai => ./third_party/uniai
```

That does two things:

- it prevents Go from using the published `github.com/quailyquaily/uniai v0.1.19`
- it makes this checkout use the local modified copy under `third_party/uniai`

Those local `uniai` changes are intentionally minimal:

- add `AwsSessionToken` to `uniai.Config`
- forward `AwsSessionToken` into the Bedrock provider config
- make the Bedrock provider pass the token into `credentials.NewStaticCredentials(...)`

This is appropriate for local testing, but not for the final `mistermorph` PR.

## Recommended upstream sequence

### PR 1: `uniai`

Suggested title:

- `feat(bedrock): support AWS session token credentials`

Suggested scope:

- add `AwsSessionToken` to `config.go`
- forward it through `client.go`
- pass it into `providers/bedrock/bedrock.go`
- add tests covering token propagation

Expected result:

- `uniai` publishes a version that natively supports temporary AWS credentials for Bedrock

Minimal content shape:

```text
config.go
client.go
providers/bedrock/bedrock.go
relevant bedrock/client tests
```

### PR 2: `mistermorph`

Suggested title:

- `feat(bedrock): resolve AWS profile credentials for uniai`

Suggested scope:

- add `llm.bedrock.aws_profile`
- add `llm.bedrock.aws_session_token`
- extend profile inheritance for those fields
- resolve AWS credentials via shared config / shared credentials using AWS SDK v2
- pass resolved values into `providers/uniai`
- remove the local `replace` and `third_party/uniai`
- update the `uniai` dependency version to the release containing `AwsSessionToken`

Expected result:

- users can configure Bedrock through explicit static credentials, temporary session credentials, AWS profiles, or the standard AWS credential chain

Minimal content shape:

```text
go.mod
go.sum
assets/config/config.example.yaml
internal/llmutil/llmutil.go
internal/llmutil/routes.go
internal/llmutil/llmutil_test.go
providers/uniai/client.go
providers/uniai/bedrock_credentials.go
providers/uniai/client_test.go
```

## What should not be in the final `mistermorph` PR

These local-validation-only pieces should be dropped before the final upstream PR:

- `replace github.com/quailyquaily/uniai => ./third_party/uniai`
- the entire `third_party/uniai/` tree

Instead, the final `mistermorph` PR should bump the `uniai` dependency to the released version that already contains PR 1.

## Validation used on this branch

Targeted checks passed:

- `go test ./providers/uniai ./internal/llmutil`

Functional validation performed locally:

- build console assets
- stage console assets
- build `./bin/mistermorph`
- confirm `./bin/mistermorph chat --help` starts cleanly

## Final handoff

Use this branch to validate behavior locally.

After validation:

1. upstream the minimal `uniai` support
2. replace the local `third_party/uniai` override with a normal module bump
3. submit the smaller `mistermorph` PR
