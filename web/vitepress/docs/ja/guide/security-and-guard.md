---
title: セキュリティと Guard
description: 本番運用向けの実践的な安全設定。
---

# セキュリティと Guard

## 推奨ベースライン

- API キーは環境変数で管理
- `guard.network.url_fetch.allowed_url_prefixes` で送信先制限
- 脱敏を有効化（`guard.redaction.enabled: true`）
- 常駐モードでは approval を有効化

## 最小 Guard 設定

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

本番運用の詳細は `docs/security.md` を参照。
