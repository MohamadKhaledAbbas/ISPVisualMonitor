#!/usr/bin/env bash
# =============================================================================
#  lib/lab/demo.sh — Demo data management: seed, reset, scenarios
#  Sourced by lab.sh — do not execute directly.
# =============================================================================

# ─── CLI Functions ────────────────────────────────────────────────────────────
demo_start() {
    _run_script "$PROJECT_DIR/scripts/dev-start.sh"
}

demo_start_devcontainer() {
    _run_script "$PROJECT_DIR/scripts/devcontainer-start.sh"
}

demo_stop() {
    _run_script "$PROJECT_DIR/scripts/dev-stop.sh"
}

demo_seed() {
    local host="${1:-$(default_db_host)}" port="${2:-${DB_PORT:-5432}}"
    log_info "Seeding demo data → ${host}:${port}"
    bash "$PROJECT_DIR/scripts/demo-seed.sh" --host "$host" --port "$port"
}

demo_reset() {
    local host="${1:-$(default_db_host)}" port="${2:-${DB_PORT:-5432}}"
    log_info "Resetting demo data on ${host}:${port}"
    bash "$PROJECT_DIR/scripts/demo-reset.sh" --host "$host" --port "$port"
}

demo_scenario() {
    local scenario="${1:?Usage: demo_scenario <name> [host] [port]}"
    local host="${2:-$(default_db_host)}" port="${3:-${DB_PORT:-5432}}"
    log_info "Applying scenario '$scenario' → ${host}:${port}"
    bash "$PROJECT_DIR/scripts/demo-scenarios.sh" "$scenario" --host "$host" --port "$port"
}

demo_healthy() {
    local host="${1:-$(default_db_host)}" port="${2:-${DB_PORT:-5432}}"
    demo_scenario "healthy" "$host" "$port"
}

demo_list_scenarios() {
    bash "$PROJECT_DIR/scripts/demo-scenarios.sh"
}

# ─── Interactive Menu ─────────────────────────────────────────────────────────
menu_demo() {
    while true; do
        _header
        _section "Demo Scripts"

        _menu_item 1 "Demo: Start"                "run scripts/dev-start.sh"
        _menu_item 2 "Demo: Start (devcontainer)" "run scripts/devcontainer-start.sh"
        _menu_item 3 "Demo: Stop"                 "run scripts/dev-stop.sh"
        _separator
        _menu_item 4 "Demo: Seed Data"            "load demo data into database"
        _menu_item 5 "Demo: Reset Data"           "wipe + re-seed demo data"
        _menu_item 6 "Demo: Scenario Picker"      "apply a failure scenario"
        _menu_item 7 "Demo: Healthy Baseline"     "quick reset to healthy state"
        _menu_item 8 "Demo: Show Scenario Help"   "list available scenarios"
        _separator
        _menu_item 0 "Back"

        local c
        c="$(_pick_number "Choice" 8)"
        [[ -z "$c" ]] && return

        _header
        case "$c" in
            1) _section "Demo: Start"; demo_start ;;
            2) _section "Demo: Start (devcontainer)"; demo_start_devcontainer ;;
            3) _section "Demo: Stop"; demo_stop ;;
            4)
                _section "Demo: Seed Data"
                _prompt_db_target
                demo_seed "$DEMO_DB_HOST" "$DEMO_DB_PORT"
                ;;
            5)
                _section "Demo: Reset Data"
                _prompt_db_target
                _warn "This runs the full demo reset against ${DEMO_DB_HOST}:${DEMO_DB_PORT}."
                demo_reset "$DEMO_DB_HOST" "$DEMO_DB_PORT"
                ;;
            6)
                _section "Demo: Scenario Picker"
                local scenario
                scenario="$(_pick_demo_scenario)"
                [[ -z "$scenario" ]] && { _pause; continue; }
                _prompt_db_target
                demo_scenario "$scenario" "$DEMO_DB_HOST" "$DEMO_DB_PORT"
                ;;
            7)
                _section "Demo: Healthy Baseline"
                _prompt_db_target
                demo_healthy "$DEMO_DB_HOST" "$DEMO_DB_PORT"
                ;;
            8)
                _section "Demo: Scenario Help"
                demo_list_scenarios
                ;;
        esac
        _pause
    done
}
