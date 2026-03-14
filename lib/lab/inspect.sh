#!/usr/bin/env bash
# =============================================================================
#  lib/lab/inspect.sh — Router inspection: health, interfaces, PPPoE, DHCP
#  Sourced by lab.sh — do not execute directly.
# =============================================================================

# ─── CLI Functions ────────────────────────────────────────────────────────────
inspect_identity() {
    local key="${1:?Usage: inspect_identity <router_key>}"
    _run_ros_cmd "$key" "/system identity print"
    echo; _run_ros_cmd "$key" "/system resource print"
    echo; _run_ros_cmd "$key" "/system clock print"
}

inspect_interfaces() {
    local key="${1:?Usage: inspect_interfaces <router_key>}"
    _run_ros_cmd "$key" "/interface print"
    echo; _run_ros_cmd "$key" "/interface monitor-traffic [find] once"
}

inspect_network() {
    local key="${1:?Usage: inspect_network <router_key>}"
    _run_ros_cmd "$key" "/interface print"
    echo; _run_ros_cmd "$key" "/ip address print"
    echo; _run_ros_cmd "$key" "/ip route print detail"
    echo; _run_ros_cmd "$key" "/ip firewall nat print"
    echo; _run_ros_cmd "$key" "/ip firewall filter print"
}

inspect_pppoe() {
    local key="${1:?Usage: inspect_pppoe <router_key>}"
    _run_ros_cmd "$key" "/ppp active print detail"
    echo; _run_ros_cmd "$key" "/ppp secret print detail"
}

inspect_dhcp() {
    local key="${1:?Usage: inspect_dhcp <router_key>}"
    _run_ros_cmd "$key" "/ip dhcp-server print detail"
    echo; _run_ros_cmd "$key" "/ip dhcp-server lease print detail"
}

inspect_services() {
    local key="${1:?Usage: inspect_services <router_key>}"
    _run_ros_cmd "$key" "/ip service print detail"
    echo; _run_ros_cmd "$key" "/snmp print"
}

inspect_custom() {
    local key="${1:?Usage: inspect_custom <router_key> <command>}"
    local cmd="${2:?Usage: inspect_custom <router_key> <command>}"
    _run_ros_cmd "$key" "$cmd"
}

# ─── Interactive Menu ─────────────────────────────────────────────────────────
menu_inspect() {
    local key
    while true; do
        _header
        _section "Router Inspection — Pick a Router"
        key="$(_pick_router)"
        [[ -z "$key" ]] && return

        while true; do
            _header
            _router_badge "$key"
            _section "What to Inspect?"

            _menu_item 1 "Identity & Health"    "system identity, resource, clock"
            _menu_item 2 "Interfaces"           "interface list + live traffic"
            _menu_item 3 "Network / Routes"     "IPs, routes, NAT, firewall"
            _menu_item 4 "PPPoE Sessions"       "active sessions + secrets"
            _menu_item 5 "DHCP Leases"          "DHCP server + leases"
            _menu_item 6 "Services & SNMP"      "ip services, snmp config"
            _menu_item 7 "Custom Command"       "run any RouterOS command"
            _separator
            _menu_item 0 "Back (pick another router)"

            local c
            c="$(_pick_number "Choice" 7)"
            [[ -z "$c" ]] && break

            _header
            _router_badge "$key"
            case "$c" in
                1) _section "Identity & Health"; inspect_identity "$key" ;;
                2) _section "Interfaces"; inspect_interfaces "$key" ;;
                3) _section "Network / Routes / NAT"; inspect_network "$key" ;;
                4) _section "PPPoE Sessions"; inspect_pppoe "$key" ;;
                5) _section "DHCP Leases"; inspect_dhcp "$key" ;;
                6) _section "Services & SNMP"; inspect_services "$key" ;;
                7)
                    _section "Custom Command"
                    local custom_cmd
                    custom_cmd="$(_prompt_input "RouterOS command" "")"
                    [[ -z "$custom_cmd" ]] && continue
                    inspect_custom "$key" "$custom_cmd"
                    ;;
            esac
            _pause
        done
    done
}
