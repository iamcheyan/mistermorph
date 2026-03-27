---
title: 安全与 Guard
description: 生产场景下的安全基线建议。
---

# 安全与 Guard

## 推荐基线

- API Key 放环境变量，不提交到仓库
- 用 `guard.network.url_fetch.allowed_url_prefixes` 限制外连
- 保持脱敏开启（`guard.redaction.enabled: true`）
- 长期运行模式下启用 approval

## 最小 Guard 配置

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

上线细节请看 `docs/security.md`。
