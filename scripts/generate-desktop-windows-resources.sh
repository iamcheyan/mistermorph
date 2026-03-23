#!/usr/bin/env bash
set -euo pipefail

ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
ICON_PNG="${ICON_PNG:-${ROOT_DIR}/desktop/wails/packaging/appicon.png}"
WINDOWS_PACKAGING_DIR="${WINDOWS_PACKAGING_DIR:-${ROOT_DIR}/desktop/wails/packaging/windows}"
ICON_ICO="${ICON_ICO:-${WINDOWS_PACKAGING_DIR}/appicon.ico}"
MANIFEST_PATH="${MANIFEST_PATH:-${WINDOWS_PACKAGING_DIR}/wails.exe.manifest}"
ARCH="${ARCH:-amd64}"
SYSO_OUT="${SYSO_OUT:-${ROOT_DIR}/desktop/wails/rsrc_windows_${ARCH}.syso}"

if [[ ! -f "${ICON_PNG}" ]]; then
  echo "missing icon PNG: ${ICON_PNG}" >&2
  exit 1
fi

if [[ ! -f "${MANIFEST_PATH}" ]]; then
  echo "missing Windows manifest: ${MANIFEST_PATH}" >&2
  exit 1
fi

mkdir -p "${WINDOWS_PACKAGING_DIR}"
rm -f "${ICON_ICO}" "${SYSO_OUT}"

echo "==> Generating Windows .ico from ${ICON_PNG}"
go run github.com/wailsapp/wails/v3/cmd/wails3 generate icons \
  -input "${ICON_PNG}" \
  -windowsFilename "${ICON_ICO}"

echo "==> Generating Windows .syso for ${ARCH}"
go run github.com/wailsapp/wails/v3/cmd/wails3 generate syso \
  -arch "${ARCH}" \
  -icon "${ICON_ICO}" \
  -manifest "${MANIFEST_PATH}" \
  -out "${SYSO_OUT}"

echo
echo "Generated:"
echo "  icon: ${ICON_ICO}"
echo "  syso: ${SYSO_OUT}"
