import { readFile, writeFile } from 'node:fs/promises'
import path from 'node:path'
import { fileURLToPath } from 'node:url'

const __filename = fileURLToPath(import.meta.url)
const __dirname = path.dirname(__filename)
const projectRoot = path.resolve(__dirname, '..')

function getArgValue(flag) {
  const index = process.argv.indexOf(flag)
  if (index === -1) {
    return null
  }
  return process.argv[index + 1] ?? null
}

function envOrDefault(name, fallback) {
  const value = process.env[name]
  return value && value.trim() ? value.trim() : fallback
}

function parseBoolean(value, fallback) {
  if (value == null || value === '') {
    return fallback
  }

  switch (value.trim().toLowerCase()) {
    case '1':
    case 'true':
    case 'yes':
    case 'on':
      return true
    case '0':
    case 'false':
    case 'no':
    case 'off':
      return false
    default:
      throw new Error(`Invalid boolean value: ${value}`)
  }
}

function sanitizeName(value) {
  return value
    .toLowerCase()
    .replace(/[^a-z0-9-]+/g, '-')
    .replace(/-{2,}/g, '-')
    .replace(/^-+|-+$/g, '')
}

function deriveDefaultWorkerName(mode) {
  const repoSlug = envOrDefault(
    'GITHUB_REPOSITORY',
    path.basename(path.resolve(projectRoot, '../..'))
  ).split('/').pop()

  const base = sanitizeName(repoSlug || 'docs-site') || 'docs-site'
  return mode === 'production' ? `${base}-docs` : `${base}-docs-${mode}`
}

const mode = envOrDefault('CF_DOCS_DEPLOY_ENV', getArgValue('--env') || 'preview')
const outPath = path.resolve(
  projectRoot,
  envOrDefault('CF_DOCS_WRANGLER_OUT', getArgValue('--out') || 'wrangler.generated.jsonc')
)
const baseConfigPath = path.resolve(projectRoot, 'wrangler.base.jsonc')
const baseConfig = JSON.parse(await readFile(baseConfigPath, 'utf8'))

const workerName = envOrDefault('CF_DOCS_WORKER_NAME', deriveDefaultWorkerName(mode))
const customDomain = process.env.CF_DOCS_CUSTOM_DOMAIN?.trim()
const workersDev = parseBoolean(
  process.env.CF_DOCS_WORKERS_DEV,
  customDomain ? false : true
)

const config = {
  ...baseConfig,
  compatibility_date: envOrDefault(
    'CF_DOCS_COMPATIBILITY_DATE',
    baseConfig.compatibility_date
  ),
  name: workerName,
  workers_dev: workersDev
}

if (customDomain) {
  config.routes = [
    {
      pattern: customDomain,
      custom_domain: true
    }
  ]
}

await writeFile(outPath, `${JSON.stringify(config, null, 2)}\n`, 'utf8')

console.log(
  JSON.stringify(
    {
      mode,
      workerName,
      workersDev,
      customDomain: customDomain || null,
      configPath: path.relative(projectRoot, outPath)
    },
    null,
    2
  )
)
