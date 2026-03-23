#!/usr/bin/env bash
set -euo pipefail

ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
SOURCE_DIR="${ROOT_DIR}/web/console/dist"
TARGET_DIR="${ROOT_DIR}/cmd/mistermorph/consolecmd/static"

if [[ ! -f "${SOURCE_DIR}/index.html" ]]; then
  echo "missing ${SOURCE_DIR}/index.html; run 'pnpm --dir web/console build' first" >&2
  exit 1
fi

mkdir -p "${TARGET_DIR}"
find "${TARGET_DIR}" -mindepth 1 -maxdepth 1 ! -name '.gitkeep' -exec rm -rf {} +
cp -R "${SOURCE_DIR}/." "${TARGET_DIR}/"

if [[ ! -f "${TARGET_DIR}/index.html" ]]; then
  echo "failed to stage console assets into ${TARGET_DIR}" >&2
  exit 1
fi

echo "Staged console assets:"
echo "  source: ${SOURCE_DIR}"
echo "  target: ${TARGET_DIR}"
