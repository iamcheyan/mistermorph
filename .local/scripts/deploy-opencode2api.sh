#!/usr/bin/env bash
#
# deploy-opencode2api.sh — One-click deploy opencode2api gateway for Mister Morph
#
# Usage:
#   ./scripts/deploy-opencode2api.sh [OPTIONS]
#
# Options:
#   -d, --dir DIR       Install directory (default: ~/opencode2api)
#   -p, --port PORT     Gateway port (default: 10000)
#   -k, --key KEY       API key for gateway (default: morph-local-key)
#   --node PATH         Node.js binary path (auto-detected if omitted)
#   --opencode PATH     OpenCode CLI path (auto-detected if omitted)
#   -n, --dry-run       Print what would be done without executing
#   -h, --help          Show this help
#
# Prerequisites:
#   - Node.js 18+ and npm
#   - OpenCode CLI installed and logged in (opencode login)
#   - git, curl, systemd --user
#
# Example:
#   ./scripts/deploy-opencode2api.sh
#   ./scripts/deploy-opencode2api.sh -p 10001 -k my-secret-key
#

set -euo pipefail

# ─── Defaults ────────────────────────────────────────────────────────────────
INSTALL_DIR="${HOME}/opencode2api"
GATEWAY_PORT="10000"
API_KEY="morph-local-key"
BIND_HOST="127.0.0.1"
NODE_BIN=""
OPENCODE_BIN=""
DRY_RUN=false

# opencode2api git repo
OPENCODE2API_REPO="https://github.com/TiaraBasori/opencode2api.git"

# ─── Colors ──────────────────────────────────────────────────────────────────
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

info()  { printf "${BLUE}[INFO]${NC}  %s\n" "$*"; }
ok()    { printf "${GREEN}[OK]${NC}    %s\n" "$*"; }
warn()  { printf "${YELLOW}[WARN]${NC}  %s\n" "$*"; }
err()   { printf "${RED}[ERR]${NC}   %s\n" "$*" >&2; }

die() { err "$*"; exit 1; }

# ─── Helpers ─────────────────────────────────────────────────────────────────
run() {
    if [[ "$DRY_RUN" == true ]]; then
        printf "${YELLOW}[DRY]${NC}   %s\n" "$*"
        return 0
    fi
    "$@"
}

print_help() {
    sed -n '/^# Usage:/,/^# Example:/p' "$0" | sed 's/^# //'
}

# ─── Argument Parsing ────────────────────────────────────────────────────────
while [[ $# -gt 0 ]]; do
    case "$1" in
        -d|--dir)       INSTALL_DIR="$2"; shift 2 ;;
        -p|--port)      GATEWAY_PORT="$2"; shift 2 ;;
        -k|--key)       API_KEY="$2"; shift 2 ;;
        --node)         NODE_BIN="$2"; shift 2 ;;
        --opencode)     OPENCODE_BIN="$2"; shift 2 ;;
        -n|--dry-run)   DRY_RUN=true; shift ;;
        -h|--help)      print_help; exit 0 ;;
        *)              die "Unknown option: $1" ;;
    esac
done

# ─── Dry-run banner ──────────────────────────────────────────────────────────
if [[ "$DRY_RUN" == true ]]; then
    warn "DRY RUN MODE — no changes will be made"
    echo ""
fi

# ─── Step 0: Check prerequisites ─────────────────────────────────────────────
info "Checking prerequisites..."

# git
if ! command -v git &>/dev/null; then
    die "git is required but not installed."
fi

# Node.js
if [[ -z "$NODE_BIN" ]]; then
    NODE_BIN="$(command -v node || true)"
    if [[ -z "$NODE_BIN" ]] && [[ -d "${HOME}/.nvm" ]]; then
        # Try nvm
        export NVM_DIR="${HOME}/.nvm"
        # shellcheck source=/dev/null
        [[ -s "$NVM_DIR/nvm.sh" ]] && . "$NVM_DIR/nvm.sh"
        NODE_BIN="$(command -v node || true)"
    fi
fi

if [[ -z "$NODE_BIN" ]] || [[ ! -x "$NODE_BIN" ]]; then
    die "Node.js is required but not found. Install Node.js 18+ or specify with --node."
fi

NODE_VERSION="$($NODE_BIN --version | sed 's/^v//')"
NODE_MAJOR="${NODE_VERSION%%.*}"
if [[ "$NODE_MAJOR" -lt 18 ]]; then
    die "Node.js 18+ required, found v${NODE_VERSION}"
fi
ok "Node.js v${NODE_VERSION} found at ${NODE_BIN}"

# npm
if ! command -v npm &>/dev/null; then
    die "npm is required but not found."
fi
ok "npm found"

# OpenCode CLI
if [[ -z "$OPENCODE_BIN" ]]; then
    OPENCODE_BIN="$(command -v opencode || true)"
    if [[ -z "$OPENCODE_BIN" ]] && [[ -x "${HOME}/.opencode/bin/opencode" ]]; then
        OPENCODE_BIN="${HOME}/.opencode/bin/opencode"
    fi
fi

if [[ -z "$OPENCODE_BIN" ]] || [[ ! -x "$OPENCODE_BIN" ]]; then
    die "OpenCode CLI not found. Install it first:\n" \
        "  curl -fsSL https://opencode.ai/install | bash\n" \
        "  or: npm install -g opencode-ai"
fi

OPENCODE_VERSION="$($OPENCODE_BIN --version 2>/dev/null || echo "unknown")"
ok "OpenCode CLI v${OPENCODE_VERSION} found at ${OPENCODE_BIN}"

# Check OpenCode login status
OPENCODE_DB="${HOME}/.local/share/opencode/opencode.db"
if [[ ! -f "$OPENCODE_DB" ]]; then
    warn "OpenCode login state not found at ${OPENCODE_DB}"
    warn "You may need to run: opencode login"
fi

# systemd --user
if ! systemctl --user &>/dev/null; then
    die "systemd user session is not available."
fi
ok "systemd user session available"

# ─── Step 1: Clone / update opencode2api ─────────────────────────────────────
info "Installing opencode2api to ${INSTALL_DIR}..."

if [[ -d "${INSTALL_DIR}/.git" ]]; then
    info "Existing repo found, pulling latest..."
    run git -C "$INSTALL_DIR" pull --ff-only
elif [[ -d "$INSTALL_DIR" ]]; then
    die "Directory ${INSTALL_DIR} exists but is not a git repo. Remove it or choose another directory with -d."
else
    run git clone "$OPENCODE2API_REPO" "$INSTALL_DIR"
fi
ok "opencode2api ready at ${INSTALL_DIR}"

# ─── Step 2: npm install ─────────────────────────────────────────────────────
info "Installing Node dependencies..."
run "${NODE_BIN}" -e "process.chdir('${INSTALL_DIR}'); require('child_process').execSync('npm install', {stdio: 'inherit'})"
ok "Dependencies installed"

# ─── Step 3: Write config.json ───────────────────────────────────────────────
info "Writing config.json..."

CONFIG_JSON="${INSTALL_DIR}/config.json"

# Resolve opencode path for config (must be absolute)
OPENCODE_ABS="$(cd "$(dirname "$OPENCODE_BIN")" && pwd)/$(basename "$OPENCODE_BIN")"

cat > "${CONFIG_JSON}" <<EOF
{
    "PORT": ${GATEWAY_PORT},
    "API_KEY": "${API_KEY}",
    "BIND_HOST": "${BIND_HOST}",
    "DISABLE_TOOLS": false,
    "EXTERNAL_TOOLS_MODE": "proxy-bridge",
    "EXTERNAL_TOOLS_CONFLICT_POLICY": "namespace",
    "USE_ISOLATED_HOME": false,
    "PROMPT_MODE": "standard",
    "OMIT_SYSTEM_PROMPT": false,
    "AUTO_CLEANUP_CONVERSATIONS": true,
    "CLEANUP_INTERVAL_MS": 43200000,
    "CLEANUP_MAX_AGE_MS": 86400000,
    "DEBUG": false,
    "OPENCODE_SERVER_URL": "http://127.0.0.1:$((GATEWAY_PORT + 1))",
    "OPENCODE_PATH": "${OPENCODE_ABS}",
    "REQUEST_TIMEOUT_MS": 180000,
    "MANAGE_BACKEND": true
}
EOF

ok "config.json written"

# ─── Step 4: Install systemd user service ────────────────────────────────────
info "Installing systemd user service..."

SERVICE_NAME="opencode2api"
SERVICE_FILE="${HOME}/.config/systemd/user/${SERVICE_NAME}.service"

# Build PATH for service: include common Node/npm paths
SERVICE_PATH="/usr/bin"
[[ -d "${HOME}/.nvm/current/bin" ]] && SERVICE_PATH="${SERVICE_PATH}:${HOME}/.nvm/current/bin"
[[ -d "${HOME}/.nvm/versions/node" ]] && SERVICE_PATH="${SERVICE_PATH}:$(find "${HOME}/.nvm/versions/node" -maxdepth 1 -name 'v*' -type d | sort -V | tail -n1)/bin"
[[ -d "${HOME}/.local/bin" ]]       && SERVICE_PATH="${SERVICE_PATH}:${HOME}/.local/bin"
[[ -d "${HOME}/.npm-global/bin" ]]  && SERVICE_PATH="${SERVICE_PATH}:${HOME}/.npm-global/bin"
[[ -d "${HOME}/bin" ]]              && SERVICE_PATH="${SERVICE_PATH}:${HOME}/bin"
SERVICE_PATH="${SERVICE_PATH}:/usr/local/bin:/bin"

run mkdir -p "${HOME}/.config/systemd/user"

cat > "${SERVICE_FILE}" <<EOF
[Unit]
Description=OpenCode2API - OpenAI-compatible gateway for OpenCode
After=network-online.target
Wants=network-online.target
StartLimitBurst=5
StartLimitIntervalSec=60

[Service]
Type=simple
WorkingDirectory=${INSTALL_DIR}
ExecStart=${NODE_BIN} ${INSTALL_DIR}/index.js
Restart=always
RestartSec=5
RestartPreventExitStatus=78
TimeoutStopSec=30
TimeoutStartSec=60
SuccessExitStatus=0 143
KillMode=control-group
Environment=HOME=${HOME}
Environment=TMPDIR=/tmp
Environment=PATH=${SERVICE_PATH}

[Install]
WantedBy=default.target
EOF

ok "Service file written: ${SERVICE_FILE}"

# ─── Step 5: Start service ───────────────────────────────────────────────────
info "Starting opencode2api service..."

run systemctl --user daemon-reload

# Stop existing if running
if systemctl --user is-active "$SERVICE_NAME" &>/dev/null; then
    warn "Service already running, stopping first..."
    run systemctl --user stop "$SERVICE_NAME"
fi

run systemctl --user enable --now "$SERVICE_NAME"
ok "Service enabled and started"

# ─── Step 6: Wait for health check ───────────────────────────────────────────
info "Waiting for service to be ready..."

HEALTH_URL="http://${BIND_HOST}:${GATEWAY_PORT}/health"
MAX_WAIT=30
WAITED=0

while [[ $WAITED -lt $MAX_WAIT ]]; do
    if curl -fsS "$HEALTH_URL" &>/dev/null; then
        ok "Health check passed: ${HEALTH_URL}"
        break
    fi
    sleep 1
    ((WAITED++))
done

if [[ $WAITED -ge $MAX_WAIT ]]; then
    warn "Health check timed out after ${MAX_WAIT}s"
    warn "Check logs: journalctl --user -u ${SERVICE_NAME} -n 20"
fi

# ─── Step 7: Verify models endpoint ──────────────────────────────────────────
info "Verifying /v1/models endpoint..."

MODELS_URL="http://${BIND_HOST}:${GATEWAY_PORT}/v1/models"
if curl -fsS -H "Authorization: Bearer ${API_KEY}" "$MODELS_URL" &>/dev/null; then
    ok "Models endpoint responding"
else
    warn "Models endpoint not responding yet (this is normal on first start)"
fi

# ─── Summary ─────────────────────────────────────────────────────────────────
echo ""
echo "═══════════════════════════════════════════════════════════════"
echo "  OpenCode2API Deployment Complete"
echo "═══════════════════════════════════════════════════════════════"
echo ""
echo "  Install directory:  ${INSTALL_DIR}"
echo "  Config file:        ${CONFIG_JSON}"
echo "  Service file:       ${SERVICE_FILE}"
echo "  Gateway URL:        http://${BIND_HOST}:${GATEWAY_PORT}"
echo "  API Key:            ${API_KEY}"
echo "  OpenCode backend:   http://${BIND_HOST}:$((GATEWAY_PORT + 1))"
echo ""
echo "  Service status:     systemctl --user status ${SERVICE_NAME}"
echo "  View logs:          journalctl --user -u ${SERVICE_NAME} -f"
echo "  Restart:            systemctl --user restart ${SERVICE_NAME}"
echo "  Stop:               systemctl --user stop ${SERVICE_NAME}"
echo ""
echo "───────────────────────────────────────────────────────────────"
echo "  Next steps for Mister Morph integration:"
echo "───────────────────────────────────────────────────────────────"
echo ""
echo "  1. Add the following to ~/.morph/config.yaml under llm.profiles:"
echo ""
cat <<EOF
     opencode:
       provider: openai_custom
       model: opencode/big-pickle
       api_key: "${API_KEY}"
       endpoint: "http://${BIND_HOST}:${GATEWAY_PORT}/v1"
EOF
echo ""
echo "  2. Add 'opencode' to your route fallback_profiles list."
echo ""
echo "  3. Test with: mistermorph run --profile opencode --task 'hello'"
echo ""
echo "═══════════════════════════════════════════════════════════════"
