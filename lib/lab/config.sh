#!/usr/bin/env bash
# =============================================================================
#  lib/lab/config.sh — Paths, router registry, environment
#  Sourced by lab.sh — do not execute directly.
# =============================================================================

# ─── Paths ────────────────────────────────────────────────────────────────────
# PROJECT_DIR is resolved relative to this library file (lib/lab/config.sh).
PROJECT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/../.." && pwd)"
LIB_DIR="$PROJECT_DIR/lib/lab"
CHR_SCRIPTS="$PROJECT_DIR/chr-lab/scripts"
CHR_CONFIGS="$PROJECT_DIR/chr-lab/configs"
CHR_DEV_COMPOSE="$PROJECT_DIR/chr-lab/docker-compose.chr.dev.yml"
CHR_FULL_COMPOSE="$PROJECT_DIR/chr-lab/docker-compose.chr.yml"
MAIN_COMPOSE="$PROJECT_DIR/docker-compose.yml"
PROD_COMPOSE="$PROJECT_DIR/docker-compose.prod.yml"
BOOTSTRAP_SCRIPT="$PROJECT_DIR/scripts/bootstrap-deps.sh"

export GOTOOLCHAIN="${GOTOOLCHAIN:-auto}"

# ─── Lab Mode ─────────────────────────────────────────────────────────────────
LAB_MODE="${LAB_MODE:-dev}"   # dev | full

lab_compose_file() {
    if [[ "$LAB_MODE" == "full" ]]; then echo "$CHR_FULL_COMPOSE"
    else echo "$CHR_DEV_COMPOSE"; fi
}

# ─── Router Registry ─────────────────────────────────────────────────────────
# Associative arrays storing router metadata by key.
declare -A R_NAME R_CONTAINER R_SSH_PORT R_API_PORT R_WEBFIG_PORT R_WINBOX_PORT R_CONFIG

_reg_router() {
    local key="$1"
    R_NAME[$key]="$2"
    R_CONTAINER[$key]="$3"
    R_SSH_PORT[$key]="$4"
    R_API_PORT[$key]="$5"
    R_WEBFIG_PORT[$key]="$6"
    R_WINBOX_PORT[$key]="$7"
    R_CONFIG[$key]="${8:-}"
}

# Single-router dev lab
_reg_router dev  "Dev (single)"          chr-01         2221 8728 18081 8291  ""

# Multi-router full lab
_reg_router core "Core Router"           chr-core-01    8722 8728 18081 8292  "$CHR_CONFIGS/chr-core-01.rsc"
_reg_router edge "Edge Router"           chr-edge-01    8723 8729 18082 8293  "$CHR_CONFIGS/chr-edge-01.rsc"
_reg_router acc  "Access Router (NAT)"   chr-access-01  8724 8730 18083 8294  "$CHR_CONFIGS/chr-access-01.rsc"
_reg_router ppp  "PPPoE Server"          chr-pppoe-01   8725 8731 18084 8295  "$CHR_CONFIGS/chr-pppoe-01.rsc"

FULL_ROUTERS=(core edge acc ppp)

# Get router keys for the current lab mode
active_router_keys() {
    if [[ "$LAB_MODE" == "full" ]]; then echo "${FULL_ROUTERS[@]}"
    else echo "dev"; fi
}

# PPPoE router key for the current lab mode
pppoe_router_key() {
    if [[ "$LAB_MODE" == "full" ]]; then echo "ppp"; else echo "dev"; fi
}

# ─── Database Defaults ───────────────────────────────────────────────────────
default_db_host() {
    if [[ -n "${DB_HOST:-}" ]]; then
        echo "$DB_HOST"
    elif command -v nc &>/dev/null && nc -z postgres 5432 &>/dev/null; then
        echo "postgres"
    elif command -v ping &>/dev/null && ping -c 1 -W 1 postgres &>/dev/null 2>&1; then
        echo "postgres"
    else
        echo "localhost"
    fi
}

DB_PORT="${DB_PORT:-5432}"
DB_USER="${DB_USER:-ispmonitor}"
DB_NAME="${DB_NAME:-ispmonitor}"
