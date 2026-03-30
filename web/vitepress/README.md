# VitePress Docs (web/docs)

## Run locally

```bash
pnpm install
pnpm dev
```

## Build

```bash
pnpm build
pnpm serve
```

## Deploy To Cloudflare Workers

This repo keeps the docs deploy setup fork-safe:

- The committed `wrangler.base.jsonc` only contains reusable asset settings.
- The final deploy config is generated into `wrangler.generated.jsonc`.
- Production-only bindings such as worker name, custom domain, account ID, and API token stay outside the repo.

### Local preview deploy

```bash
CF_DOCS_WORKER_NAME=my-docs-preview pnpm deploy:preview
```

This deploys the built docs to a `workers.dev` URL by default.

### Local production-style deploy

```bash
CF_DOCS_WORKER_NAME=mistermorph-docs \
CF_DOCS_CUSTOM_DOMAIN=docs.mistermorph.com \
CF_DOCS_WORKERS_DEV=false \
pnpm deploy:production
```

### GitHub Actions production deploy

The workflow at `.github/workflows/deploy_docs.yml` only runs when the repository variable `CF_DOCS_DEPLOY_ENABLED` is set to `true`.

Set these repository variables on the upstream repo:

- `CF_DOCS_DEPLOY_ENABLED=true`
- `CF_DOCS_WORKER_NAME=mistermorph-docs`
- `CF_DOCS_CUSTOM_DOMAIN=docs.mistermorph.com`

Set these org or repository secrets:

- `CF_ACCOUNT_ID=<your Cloudflare account id>`
- `CF_WORKER_API_TOKEN=<token with Workers deploy permissions>`

The workflow builds the VitePress site, renders `wrangler.generated.jsonc`, and runs `wrangler deploy`.

## Agent-friendly outputs

After `pnpm build`, generated files include:

- `dist/llms.txt`
- `dist/llms-full.txt`
- per-page `*.md`
