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

check_forwarded_ports() {
    echo
    printf "  ${C_BOLD}%-6s %-12s %-19s %-19s %-34s${C_RESET}\n" "Port" "Service" "Localhost" "Service-Net" "Recover"
    _separator

    local ports=(5173 8080 3000 9090 5432 6379)
    local p svc cmd local_state net_state local_ok net_ok
    for p in "${ports[@]}"; do
        case "$p" in
            5173)
                svc="frontend"
                cmd="./lab.sh run checks recover frontend"
                if port_open 127.0.0.1 5173; then local_ok=0; else local_ok=1; fi
                net_ok=$local_ok
                ;;
            8080)
                svc="api"
                cmd="./lab.sh run checks recover api"
                if port_open 127.0.0.1 8080; then local_ok=0; else local_ok=1; fi
                net_ok=$local_ok
                ;;
            3000)
                svc="grafana"
                cmd="./lab.sh run checks recover grafana"
                if port_open 127.0.0.1 3000; then local_ok=0; else local_ok=1; fi
                if _service_reachable grafana; then net_ok=0; else net_ok=1; fi
                ;;
            9090)
                svc="prometheus"
                cmd="./lab.sh run checks recover prometheus"
                if port_open 127.0.0.1 9090; then local_ok=0; else local_ok=1; fi
                if _service_reachable prometheus; then net_ok=0; else net_ok=1; fi
                ;;
            5432)
                svc="postgres"
                cmd="./lab.sh run checks recover postgres"
                if port_open 127.0.0.1 5432; then local_ok=0; else local_ok=1; fi
                if _service_reachable postgres; then net_ok=0; else net_ok=1; fi
                ;;
            6379)
                svc="redis"
                cmd="./lab.sh run checks recover redis"
                if port_open 127.0.0.1 6379; then local_ok=0; else local_ok=1; fi
                if _service_reachable redis; then net_ok=0; else net_ok=1; fi
                ;;
        esac

        if [[ "$local_ok" -eq 0 ]]; then
            local_state="${C_GREEN}OPEN${C_RESET}"
        else
            local_state="${C_RED}CLOSED${C_RESET}"
        fi

        if [[ "$net_ok" -eq 0 ]]; then
            net_state="${C_GREEN}OPEN${C_RESET}"
        else
            net_state="${C_RED}CLOSED${C_RESET}"
        fi

        printf "  %-6s %-12s %-19b %-19b %s\n" ":$p" "$svc" "$local_state" "$net_state" "$cmd"
    done

    echo
    _info "Localhost column matches VS Code Ports tab in this container"
    _info "Service-Net column is reachability to sidecar services (e.g., postgres/redis in Codespaces)"
    _info "Recover all inactive services: ./lab.sh run checks recover all"
}

check_recover_service() {
    local target="${1:-all}"
    case "$target" in
        all)
            backend_infra_up
            compose "$(lab_compose_file)" up -d
            backend_start
            ;;
        frontend|api)
            backend_start
            ;;
        postgres|redis|prometheus|grafana)
            ensure_compose_up "$MAIN_COMPOSE" "$target"
            ;;
        chr|lab)
            compose "$(lab_compose_file)" up -d
            ;;
        *)
            log_error "Unknown service '$target'. Use: all | frontend | api | postgres | redis | prometheus | grafana | chr"
            return 1
            ;;
    esac

    check_forwarded_ports
}

menu_recover_service() {
    _header
    _section "Recover Service"

    echo "  1) all"
    echo "  2) frontend"
    echo "  3) api"
    echo "  4) postgres"
    echo "  5) redis"
    echo "  6) prometheus"
    echo "  7) grafana"
    echo "  8) chr"
    echo

    local c target
    c="$(_pick_number "Choice" 8)"
    [[ -z "$c" ]] && return

    case "$c" in
        1) target="all" ;;
        2) target="frontend" ;;
        3) target="api" ;;
        4) target="postgres" ;;
        5) target="redis" ;;
        6) target="prometheus" ;;
        7) target="grafana" ;;
        8) target="chr" ;;
    esac

    _header
    _section "Recover: $target"
    check_recover_service "$target"
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
        _menu_item 8 "Forwarded Ports Dashboard"    "status of forwarded app ports"
        _menu_item 9 "Recover Inactive Service"     "start/restart stopped services"
        _separator
        _menu_item 0 "Back"

        local c
        c="$(_pick_number "Choice" 9)"
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
            8) _section "Forwarded Ports Dashboard"; check_forwarded_ports ;;
            9) menu_recover_service ;;
        esac
        _pause
    done
}
