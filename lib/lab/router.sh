#!/usr/bin/env bash
# =============================================================================
#  lib/lab/router.sh — RouterOS SSH interaction and config management
#  Sourced by lab.sh — do not execute directly.
# =============================================================================

# ─── SSH Command Execution ────────────────────────────────────────────────────
_run_ros_cmd() {
    local key="$1" cmd="$2"
    local host="127.0.0.1"
    local port="${R_SSH_PORT[$key]}"
    local user="${ROUTER_USER:-admin}"
    local pass="${ROUTER_PASS:-admin}"

    require_cmd sshpass "sshpass required for RouterOS SSH. Run: ./lab.sh bootstrap" || return 1

    log_debug "SSH → ${R_CONTAINER[$key]} (port $port): $cmd"
    echo -e "  ${C_DIM}→ SSH to ${R_CONTAINER[$key]} (port $port): $cmd${C_RESET}"
    echo

    ROUTER_HOST="$host" ROUTER_PORT="$port" ROUTER_USER="$user" ROUTER_PASS="$pass" \
        "$CHR_SCRIPTS/routeros-run.sh" "$cmd" 2>&1 || true
}

# ─── Display ──────────────────────────────────────────────────────────────────
_router_badge() {
    local key="$1"
    echo -e "  ${C_BOLD}Router:${C_RESET} ${C_MAGENTA}${R_NAME[$key]}${C_RESET}  container=${C_CYAN}${R_CONTAINER[$key]}${C_RESET}  ssh=:${R_SSH_PORT[$key]}  api=:${R_API_PORT[$key]}  webfig=:${R_WEBFIG_PORT[$key]}"
    echo
}

# ─── Config Management ───────────────────────────────────────────────────────
_apply_router_config() {
    local key="$1" cfg="${2:-}"

    if [[ -z "$cfg" ]]; then
        cfg="${R_CONFIG[$key]}"
    fi

    if [[ -z "$cfg" || ! -f "$cfg" ]]; then
        log_error "Config file not found: ${cfg:-<none>}"
        return 1
    fi

    log_info "Applying $cfg → ${R_CONTAINER[$key]}"
    if ! docker cp "$cfg" "${R_CONTAINER[$key]}:/apply.rsc"; then
        log_error "Failed to copy config to container"
        return 1
    fi

    _run_ros_cmd "$key" "/import apply.rsc"
    log_ok "Config applied to ${R_NAME[$key]}"
}

_export_router_config() {
    local key="$1"
    local out="${R_CONTAINER[$key]}-export-$(date +%Y%m%d-%H%M%S).rsc"
    local dest="$PROJECT_DIR/$out"

    log_info "Exporting config → $out"
    _run_ros_cmd "$key" "/export show-sensitive" >"$dest" 2>&1
    log_ok "Config saved: $out"
}
