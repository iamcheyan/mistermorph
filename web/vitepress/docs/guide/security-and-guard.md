---
title: Security and Guard
description: Practical baseline for safe deployment.
---

# Security and Guard

## Recommended Baseline

- Keep API keys in env vars, not committed config
- Restrict outbound domains with `guard.network.url_fetch.allowed_url_prefixes`
- Keep redaction enabled (`guard.redaction.enabled: true`)
- Enable approvals for risky actions in long-running modes

## Minimal Guard Snippet

```yaml
guard:
  enabled: true
  network:
    url_fetch:
      allowed_url_prefixes: ["https://"]
      deny_private_ips: true
      follow_redirects: false
  redaction:
    enabled: true
```

For production details, read `docs/security.md`.
