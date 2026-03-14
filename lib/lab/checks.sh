#!/usr/bin/env bash
# =============================================================================
#  lib/lab/checks.sh — Connectivity checks: ports, readiness, health
#  Sourced by lab.sh — do not execute directly.
# =============================================================================

# ─── CLI Functions ────────────────────────────────────────────────────────────
check_ports() {
    local keys
    read -ra keys <<< "$(active_router_keys)"

    for k in "${keys[@]}"; do
        echo -e "  ${C_BOLD}${R_NAME[$k]}${C_RESET}"
        local port
        for port in "${R_SSH_PORT[$k]}" "${R_API_PORT[$k]}" "${R_WEBFIG_PORT[$k]}" "${R_WINBOX_PORT[$k]}"; do
            if port_open 127.0.0.1 "$port"; then
                _ok "127.0.0.1:$port  OPEN"
            else
                _err "127.0.0.1:$port  CLOSED"
            fi
        done
        echo
    done
    _info "Backend API:"
    port_open 127.0.0.1 8080 && _ok "API  127.0.0.1:8080  OPEN" || _warn "API  127.0.0.1:8080  CLOSED"
}

check_wait_ready() {
    local key="${1:?Usage: check_wait_ready <router_key>}"
    ROUTER_HOST=127.0.0.1 ROUTER_PORT="${R_SSH_PORT[$key]}" \
    API_PORT="${R_API_PORT[$key]}" WEBFIG_PORT="${R_WEBFIG_PORT[$key]}" \
        "$CHR_SCRIPTS/wait-ready.sh"
}

check_dev() {
    ROUTER_HOST=127.0.0.1 ROUTER_PORT="${R_SSH_PORT[dev]}" \
    API_PORT="${R_API_PORT[dev]}" WEBFIG_PORT="${R_WEBFIG_PORT[dev]}" \
        "$CHR_SCRIPTS/dev-check.sh"
}

check_port_map() {
    echo
    printf "  ${C_BOLD}%-22s %-14s %-8s %-8s %-8s %-8s${C_RESET}\n" \
        "Router" "Container" "SSH" "API" "WebFig" "Winbox"
    _separator
    local all_keys=(dev "${FULL_ROUTERS[@]}")
    local k
    for k in "${all_keys[@]}"; do
        printf "  ${C_CYAN}%-22s${C_RESET} %-14s :%-7s :%-7s :%-7s :%-7s\n" \
            "${R_NAME[$k]}" "${R_CONTAINER[$k]}" \
            "${R_SSH_PORT[$k]}" "${R_API_PORT[$k]}" \
            "${R_WEBFIG_PORT[$k]}" "${R_WINBOX_PORT[$k]}"
    done
    echo
    echo -e "  ${C_BOLD}Backend stack${C_RESET}"
    echo "  API :8080   Postgres :5432   Redis :6379   Prometheus :9090   Grafana :3000"
}

check_ssh_probe() {
    local key="${1:?Usage: check_ssh_probe <router_key>}"
    local p="${R_SSH_PORT[$key]}"
    port_open 127.0.0.1 "$p" && _ok "127.0.0.1:$p OPEN" || _err "127.0.0.1:$p CLOSED"
}

# ─── Interactive Menu ─────────────────────────────────────────────────────────
menu_checks() {
    while true; do
        _header
        _section "Connectivity Checks"

        _menu_item 1 "Docker Daemon Status"         "check if dockerd is running"
        _menu_item 2 "Start Docker Daemon"           "auto-start with env detection"
        _separator
        _menu_item 3 "Port Check — current lab"     "nc probe all known ports"
        _menu_item 4 "Wait for Router Ready"        "block until API+WebFig respond"
        _menu_item 5 "Dev Check (ports + health)"   "wait-ready + port-check"
        _menu_item 6 "Show All Port Mappings"       "all routers + backend ports"
        _menu_item 7 "Ping: Router SSH port test"   "quick reachability probe"
        _separator
        _menu_item 0 "Back"

        local c
        c="$(_pick_number "Choice" 7)"
        [[ -z "$c" ]] && return

        _header
        case "$c" in
            1) _section "Docker Daemon Status"; docker_daemon_status ;;
            2) _section "Starting Docker Daemon"; docker_start_daemon ;;
            3) _section "Port Check — All Lab Routers"; check_ports ;;
            4)
                _section "Waiting for Router Ready"
                local key
                key="$(_pick_router)"
                [[ -z "$key" ]] && { _pause; continue; }
                _router_badge "$key"
                check_wait_ready "$key"
                ;;
            5) _section "Dev Check"; check_dev ;;
            6) _section "All Port Mappings"; check_port_map ;;
            7)
                _section "SSH Port Probe"
                local key
                key="$(_pick_router)"
                [[ -z "$key" ]] && { _pause; continue; }
                check_ssh_probe "$key"
                ;;
        esac
        _pause
    done
}
