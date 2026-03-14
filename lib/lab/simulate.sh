#!/usr/bin/env bash
# =============================================================================
#  lib/lab/simulate.sh — Simulation scenarios, PPPoE user management, drills
#  Sourced by lab.sh — do not execute directly.
# =============================================================================

# ─── PPPoE User Management (CLI) ─────────────────────────────────────────────
sim_pppoe_create() {
    local count="${1:?Usage: sim_pppoe_create <count> [prefix] [pass] [profile]}"
    local prefix="${2:-cust}" pass="${3:-pass123}" profile="${4:-pppoe-profile}"
    local key; key="$(pppoe_router_key)"

    if [[ ! "$count" =~ ^[0-9]+$ ]]; then
        log_error "Count must be a positive integer"
        return 1
    fi

    log_info "Creating $count PPPoE users (prefix=$prefix) on ${R_CONTAINER[$key]}"
    ROUTER_HOST=127.0.0.1 ROUTER_PORT="${R_SSH_PORT[$key]}" \
    USER_PREFIX="$prefix" PASS="$pass" PROFILE="$profile" \
        "$CHR_SCRIPTS/create-pppoe-users.sh" "$count"
    log_ok "PPPoE users created"
}

sim_pppoe_cleanup() {
    local prefix="${1:-cust}"
    local key; key="$(pppoe_router_key)"

    log_info "Cleaning up PPPoE users (prefix=$prefix) on ${R_CONTAINER[$key]}"
    ROUTER_HOST=127.0.0.1 ROUTER_PORT="${R_SSH_PORT[$key]}" \
    USER_PREFIX="$prefix" \
        "$CHR_SCRIPTS/cleanup-pppoe-users.sh"
    log_ok "PPPoE users cleaned up"
}

sim_pppoe_show() {
    local key; key="$(pppoe_router_key)"
    _run_ros_cmd "$key" "/ppp active print detail"
    echo
    _run_ros_cmd "$key" "/ppp active print count-only" 2>/dev/null || true
}

# ─── Router Scenarios (CLI) ──────────────────────────────────────────────────
sim_router_down() {
    local key="${1:?Usage: sim_router_down <router_key>}"
    _require_router_key "$key" || return 1
    log_info "Stopping ${R_CONTAINER[$key]} (simulates outage)"
    docker stop "${R_CONTAINER[$key]}" && log_ok "${R_CONTAINER[$key]} stopped" || log_error "Failed"
}

sim_router_up() {
    local key="${1:?Usage: sim_router_up <router_key>}"
    _require_router_key "$key" || return 1
    log_info "Starting ${R_CONTAINER[$key]}"
    docker start "${R_CONTAINER[$key]}" && log_ok "${R_CONTAINER[$key]} started" || log_error "Failed"
}

sim_router_flap() {
    local key="${1:?Usage: sim_router_flap <router_key> [duration_secs]}"
    local dur="${2:-15}"
    _require_router_key "$key" || return 1
    log_info "Flapping ${R_CONTAINER[$key]} for ${dur}s"
    if docker pause "${R_CONTAINER[$key]}"; then
        log_ok "Paused"
        sleep "$dur"
        docker unpause "${R_CONTAINER[$key]}" && log_ok "Unpaused after ${dur}s"
    else
        log_error "Failed to pause ${R_CONTAINER[$key]} — skipping flap"
        return 1
    fi
}

sim_api_disable() {
    local key="${1:?Usage: sim_api_disable <router_key>}"
    _require_router_key "$key" || return 1
    _run_ros_cmd "$key" "/ip service disable api"
    log_ok "API disabled on ${R_NAME[$key]} — SNMP fallback expected"
}

sim_api_enable() {
    local key="${1:?Usage: sim_api_enable <router_key>}"
    _require_router_key "$key" || return 1
    _run_ros_cmd "$key" "/ip service enable api"
    log_ok "API re-enabled on ${R_NAME[$key]}"
}

sim_disconnect_all() {
    local key; key="$(pppoe_router_key)"
    log_info "Disconnecting all PPPoE sessions on ${R_NAME[$key]}"
    _run_ros_cmd "$key" "/ppp active remove [find]"
    log_ok "All sessions disconnected"
}

# ─── Drills (CLI) ────────────────────────────────────────────────────────────
drill_peak_hour() {
    local peak="${1:-50}" hold="${2:-60}"
    local key; key="$(pppoe_router_key)"

    log_info "[Peak Hour] Ramping up $peak users, holding ${hold}s"
    sim_pppoe_create "$peak" "peak" "pass123" "${PPPOE_PROFILE:-pppoe-profile}"
    log_info "Holding for ${hold}s..."
    sleep "$hold"
    sim_pppoe_cleanup "peak"
    log_ok "Peak hour drill complete"
}

drill_churn_wave() {
    local waves="${1:-3}" batch="${2:-30}" pause_secs="${3:-20}"
    local key; key="$(pppoe_router_key)"

    log_info "[Churn Wave] $waves waves of $batch users, ${pause_secs}s between"
    local wave
    for (( wave=1; wave<=waves; wave++ )); do
        log_info "Wave ${wave}/${waves}: creating users..."
        sim_pppoe_create "$batch" "wave${wave}" "pass123" "${PPPOE_PROFILE:-pppoe-profile}"
        sleep "$pause_secs"
        log_info "Wave ${wave}/${waves}: removing users..."
        sim_pppoe_cleanup "wave${wave}"
        sleep 5
    done
    log_ok "Churn wave drill complete"
}

drill_full_chaos() {
    local key; key="$(pppoe_router_key)"

    log_info "[Full Chaos] Starting multi-router + session drill"
    log_info "[1/6] Creating 40 PPPoE users..."
    sim_pppoe_create 40 "chaos" "pass123" "${PPPOE_PROFILE:-pppoe-profile}"

    if [[ "$LAB_MODE" == "full" ]]; then
        log_info "[2/6] Flapping Access router for 8s..."
        if docker pause "${R_CONTAINER[acc]}"; then
            sleep 8
            docker unpause "${R_CONTAINER[acc]}"
        fi
        log_ok "Access router flap done"
        log_info "[3/6] Stopping Edge router..."
        docker stop "${R_CONTAINER[edge]}"
        log_ok "Edge router down"
    fi

    sleep 10
    log_info "[4/6] Disconnecting all PPPoE sessions..."
    _run_ros_cmd "$key" "/ppp active remove [find]" &>/dev/null || true
    log_ok "Sessions disconnected"

    sleep 10
    log_info "[5/6] Cleaning up chaos users..."
    sim_pppoe_cleanup "chaos"

    if [[ "$LAB_MODE" == "full" ]]; then
        log_info "[6/6] Restoring Edge router..."
        docker start "${R_CONTAINER[edge]}"
        log_ok "Edge router restored"
    fi

    log_ok "Full Chaos drill complete — check monitoring dashboards"
}

# ─── Validation ──────────────────────────────────────────────────────────────
_require_router_key() {
    local key="${1:-}"
    if [[ -z "$key" || -z "${R_CONTAINER[$key]:-}" ]]; then
        log_error "Invalid router key '$key'. Valid keys: dev core edge acc ppp"
        return 1
    fi
}

# ─── Interactive Menu ─────────────────────────────────────────────────────────
menu_simulate() {
    while true; do
        _header
        _section "Simulation & Scenario Testing"

        _menu_item 1  "PPPoE: Create N users"           "bulk add PPP secrets"
        _menu_item 2  "PPPoE: Cleanup users"             "remove secrets by prefix"
        _menu_item 3  "PPPoE: Show active sessions"      "live session list"
        _separator
        _menu_item 4  "Scenario: Router Down"            "docker stop a router"
        _menu_item 5  "Scenario: Router Up (restore)"    "docker start a router"
        _menu_item 6  "Scenario: Network Flap"           "pause → wait → unpause"
        _menu_item 7  "Scenario: API Disabled (SNMP)"    "disable RouterOS API service"
        _menu_item 8  "Scenario: API Re-enable"          "re-enable RouterOS API service"
        _menu_item 9  "Scenario: Disconnect all PPPoE"   "kick all active sessions"
        _separator
        _menu_item 10 "Drill: Peak Hour"                 "ramp users up then down"
        _menu_item 11 "Drill: Churn Wave"                "create batch → wait → delete loop"
        _menu_item 12 "Drill: Full Chaos"                "random flaps + session churn"
        _separator
        _menu_item 0  "Back"

        local c
        c="$(_pick_number "Choice" 12)"
        [[ -z "$c" ]] && return

        case "$c" in
            1)
                _header; _section "Create PPPoE Users"
                local key; key="$(pppoe_router_key)"
                _router_badge "$key"
                local count prefix pass profile
                count="$(_prompt_input "Number of users" "20")"
                prefix="$(_prompt_input "Username prefix" "cust")"
                pass="$(_prompt_input "Password" "pass123")"
                profile="$(_prompt_input "PPP Profile" "pppoe-profile")"
                sim_pppoe_create "$count" "$prefix" "$pass" "$profile"
                ;;
            2)
                _header; _section "Cleanup PPPoE Users"
                local key; key="$(pppoe_router_key)"
                _router_badge "$key"
                local prefix
                prefix="$(_prompt_input "Username prefix to remove" "cust")"
                _confirm "Remove all PPP secrets matching '${prefix}*'?" || { _pause; continue; }
                sim_pppoe_cleanup "$prefix"
                ;;
            3)
                _header; _section "Active PPPoE Sessions"
                sim_pppoe_show
                ;;
            4)
                _header; _section "Scenario: Router Down"
                local key; key="$(_pick_router)"
                [[ -z "$key" ]] && { _pause; continue; }
                _confirm "Stop ${R_CONTAINER[$key]}?" || { _pause; continue; }
                sim_router_down "$key"
                ;;
            5)
                _header; _section "Scenario: Router Up"
                local key; key="$(_pick_router)"
                [[ -z "$key" ]] && { _pause; continue; }
                sim_router_up "$key"
                ;;
            6)
                _header; _section "Scenario: Network Flap"
                local key; key="$(_pick_router)"
                [[ -z "$key" ]] && { _pause; continue; }
                local dur; dur="$(_prompt_input "Flap duration (seconds)" "15")"
                sim_router_flap "$key" "$dur"
                ;;
            7)
                _header; _section "Disable RouterOS API"
                local key; key="$(_pick_router)"
                [[ -z "$key" ]] && { _pause; continue; }
                _confirm "Disable API on ${R_NAME[$key]}?" || { _pause; continue; }
                sim_api_disable "$key"
                ;;
            8)
                _header; _section "Re-enable RouterOS API"
                local key; key="$(_pick_router)"
                [[ -z "$key" ]] && { _pause; continue; }
                sim_api_enable "$key"
                ;;
            9)
                _header; _section "Disconnect All PPPoE"
                _confirm "Kick ALL active PPPoE sessions?" || { _pause; continue; }
                sim_disconnect_all
                ;;
            10)
                _header; _section "Drill: Peak Hour"
                local peak hold
                peak="$(_prompt_input "Users to ramp up" "50")"
                hold="$(_prompt_input "Hold time (seconds)" "60")"
                drill_peak_hour "$peak" "$hold"
                ;;
            11)
                _header; _section "Drill: Churn Wave"
                drill_churn_wave
                ;;
            12)
                _header; _section "Drill: Full Chaos"
                _warn "This runs multi-router flaps + session churn"
                _confirm "Run Full Chaos drill?" || { _pause; continue; }
                drill_full_chaos
                ;;
        esac
        _pause
    done
}
