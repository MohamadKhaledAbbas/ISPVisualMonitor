#!/usr/bin/env bash
# =============================================================================
#  lib/lab/menu.sh — Interactive menu helpers
#  Sourced by lab.sh — do not execute directly.
# =============================================================================

# ─── Display ──────────────────────────────────────────────────────────────────
_header() {
    clear
    echo -e "${C_BOLD}${C_CYAN}"
    echo "  ╔══════════════════════════════════════════════════════════╗"
    echo "  ║          ISP Visual Monitor — Lab Management            ║"
    printf "  ║  Mode: %-10s  Lab: %-33s║\n" \
        "$(echo "$LAB_MODE" | tr '[:lower:]' '[:upper:]')" \
        "$(basename "$(lab_compose_file)")"
    echo "  ╚══════════════════════════════════════════════════════════╝"
    echo -e "${C_RESET}"
}

_section() {
    echo -e "${C_BOLD}${C_BLUE}── ${1} ──${C_RESET}"
    echo
}

_separator() {
    echo -e "  ${C_DIM}──────────────────────────────────────${C_RESET}"
}

_menu_item() {
    local n="$1" label="$2" desc="${3:-}"
    printf "  ${C_WHITE}[%2s]${C_RESET} ${C_CYAN}%-28s${C_RESET} ${C_DIM}%s${C_RESET}\n" "$n" "$label" "$desc"
}

# ─── Input ────────────────────────────────────────────────────────────────────
_pause() {
    echo
    echo -e "${C_DIM}  Press Enter to continue...${C_RESET}"
    read -r
}

_confirm() {
    local msg="${1:-Are you sure?}"
    echo -e "${C_YELLOW}  ${msg} [y/N] ${C_RESET}"
    read -r ans
    [[ "$ans" =~ ^[Yy]$ ]]
}

_pick_number() {
    local prompt="$1" max="$2" n
    echo -ne "${C_WHITE}  ${prompt} (1-${max}, 0=back): ${C_RESET}" >&2
    read -r n </dev/tty
    if [[ "$n" == "0" || -z "$n" ]]; then echo ""; return; fi
    if [[ "$n" =~ ^[0-9]+$ && "$n" -ge 1 && "$n" -le "$max" ]]; then
        echo "$n"
    else
        echo -e "  ${C_YELLOW}⚠${C_RESET}  Invalid — enter 1-${max}, or 0 to go back." >&2
        echo ""
    fi
}

_prompt_input() {
    local prompt="$1" default="${2:-}" value
    if [[ -n "$default" ]]; then
        echo -ne "${C_WHITE}  ${prompt} (default: ${default}): ${C_RESET}"
    else
        echo -ne "${C_WHITE}  ${prompt}: ${C_RESET}"
    fi
    read -r value
    echo "${value:-$default}"
}

# ─── Router & Scenario Pickers ───────────────────────────────────────────────
_pick_router() {
    local opts
    if [[ "$LAB_MODE" == "full" ]]; then
        opts=("${FULL_ROUTERS[@]}")
    else
        opts=(dev)
    fi

    local i=1
    local -A idx_map
    echo >&2
    for k in "${opts[@]}"; do
        _menu_item "$i" "${R_NAME[$k]}" "container=${R_CONTAINER[$k]}  ssh=:${R_SSH_PORT[$k]}" >&2
        idx_map[$i]="$k"
        ((i++))
    done
    _separator >&2

    local choice
    choice="$(_pick_number "Pick router" "$((i-1))")"
    [[ -z "$choice" ]] && { echo ""; return; }
    echo "${idx_map[$choice]}"
}

_pick_demo_scenario() {
    local scenarios=(healthy router-offline core-congestion upstream-failure packet-loss high-sessions)
    local descriptions=(
        "reset to healthy baseline"
        "simulate one router outage"
        "simulate core uplink congestion"
        "simulate provider failure"
        "simulate uplink errors"
        "simulate PPPoE session surge"
    )

    echo >&2
    for i in "${!scenarios[@]}"; do
        _menu_item "$((i+1))" "${scenarios[$i]}" "${descriptions[$i]}" >&2
    done
    _separator >&2

    local choice
    choice="$(_pick_number "Pick scenario" "${#scenarios[@]}")"
    [[ -z "$choice" ]] && { echo ""; return; }
    echo "${scenarios[$((choice-1))]}"
}

_prompt_db_target() {
    local default_host default_port
    default_host="$(default_db_host)"
    default_port="${DB_PORT:-5432}"

    DEMO_DB_HOST="$(_prompt_input "Database host" "$default_host")"
    DEMO_DB_PORT="$(_prompt_input "Database port" "$default_port")"
}
