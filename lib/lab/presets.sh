#!/usr/bin/env bash
# =============================================================================
#  lib/lab/presets.sh — Multi-step preset workflows
#  Sourced by lab.sh — do not execute directly.
# =============================================================================

preset_run() {
    local name="${1:-}"
    shift || true

    case "$name" in
        quickstart)    preset_quickstart "$@" ;;
        demo-showtime) preset_demo_showtime "$@" ;;
        chaos-suite)   preset_chaos_suite "$@" ;;
        full-reset)    preset_full_reset "$@" ;;
        ""|list)       preset_list ;;
        *) log_error "Unknown preset: $name"; preset_list; return 1 ;;
    esac
}

preset_list() {
    echo
    echo -e "${C_BOLD}Available presets:${C_RESET}"
    echo
    printf "  ${C_CYAN}%-18s${C_RESET} %s\n" "quickstart"    "bootstrap → lab up → wait → seed → health check"
    printf "  ${C_CYAN}%-18s${C_RESET} %s\n" "demo-showtime" "infra up → seed → apply scenarios → open endpoints"
    printf "  ${C_CYAN}%-18s${C_RESET} %s\n" "chaos-suite"   "lab up → wait → full chaos drill → health report"
    printf "  ${C_CYAN}%-18s${C_RESET} %s\n" "full-reset"    "stop everything → wipe volumes → clean state"
    echo
}

# ─── Quickstart ──────────────────────────────────────────────────────────────
# Full from-zero setup: bootstrap, start lab + backend, seed, verify.
preset_quickstart() {
    log_info "=== Preset: Quickstart ==="
    local step=1

    log_info "[${step}/7] Checking dependencies..."
    bootstrap_check || {
        log_warn "Some dependencies missing — attempting install..."
        bootstrap_all || { log_error "Bootstrap failed"; return 1; }
    }
    (( step++ ))

    log_info "[${step}/7] Ensuring Docker daemon is running..."
    docker_available || { log_error "Docker is required. Run: ./lab.sh run docker start"; return 1; }
    (( step++ ))

    log_info "[${step}/7] Starting backend infrastructure..."
    backend_infra_up
    (( step++ ))

    log_info "[${step}/7] Running database migrations..."
    backend_migrate
    (( step++ ))

    log_info "[${step}/7] Starting CHR lab..."
    compose "$(lab_compose_file)" up -d
    (( step++ ))

    log_info "[${step}/7] Seeding demo data..."
    demo_seed
    (( step++ ))

    log_info "[${step}/7] Checking health..."
    backend_health
    check_ports

    log_ok "=== Quickstart complete ==="
    echo
    echo -e "  ${C_BOLD}Next steps:${C_RESET}"
    echo "    ./lab.sh run backend start     # start API server"
    echo "    ./lab.sh                        # interactive menu"
    echo "    ./lab.sh run demo scenario core-congestion"
}

# ─── Demo Showtime ───────────────────────────────────────────────────────────
# Prepare a full demo: infra + seed + some scenarios in quick succession.
preset_demo_showtime() {
    log_info "=== Preset: Demo Showtime ==="
    local step=1

    log_info "[${step}/6] Ensuring Docker daemon is running..."
    docker_available || { log_error "Docker is required"; return 1; }
    (( step++ ))

    log_info "[${step}/6] Starting backend infrastructure..."
    backend_infra_up
    (( step++ ))

    log_info "[${step}/6] Running database migrations..."
    backend_migrate
    (( step++ ))

    log_info "[${step}/6] Loading demo seed data..."
    demo_seed
    (( step++ ))

    log_info "[${step}/6] Resetting to healthy baseline..."
    demo_healthy
    (( step++ ))

    log_info "[${step}/6] Verifying endpoints..."
    backend_health
    check_ports

    log_ok "=== Demo Showtime ready ==="
    echo
    echo -e "  ${C_BOLD}Demo is live. Try some scenarios:${C_RESET}"
    echo "    ./lab.sh run demo scenario router-offline"
    echo "    ./lab.sh run demo scenario core-congestion"
    echo "    ./lab.sh run demo scenario upstream-failure"
    echo "    ./lab.sh run demo healthy   # reset"
}

# ─── Chaos Suite ─────────────────────────────────────────────────────────────
# Full chaos: ensure lab running, then run all drills in sequence.
preset_chaos_suite() {
    log_info "=== Preset: Chaos Suite ==="
    local step=1

    log_info "[${step}/6] Ensuring Docker daemon is running..."
    docker_available || { log_error "Docker is required"; return 1; }
    (( step++ ))

    log_info "[${step}/6] Ensuring CHR lab is running..."
    compose "$(lab_compose_file)" up -d
    (( step++ ))

    log_info "[${step}/6] Waiting for readiness..."
    sleep 5
    (( step++ ))

    log_info "[${step}/6] Running peak hour drill..."
    drill_peak_hour 30 30
    (( step++ ))

    log_info "[${step}/6] Running churn wave drill..."
    drill_churn_wave 2 20 15
    (( step++ ))

    log_info "[${step}/6] Running full chaos drill..."
    drill_full_chaos

    log_ok "=== Chaos Suite complete — check dashboards ==="
}

# ─── Full Reset ──────────────────────────────────────────────────────────────
# Nuclear option: stop everything, wipe volumes, clean state.
preset_full_reset() {
    log_warn "=== Preset: Full Reset — DESTRUCTIVE ==="
    echo
    log_warn "This will stop all services and remove ALL volumes."

    if [[ -t 0 ]]; then
        _confirm "Continue with full reset?" || {
            log_info "Cancelled."
            return 0
        }
    fi

    local step=1

    log_info "[${step}/4] Stopping backend services..."
    ensure_compose_down "$MAIN_COMPOSE" || true
    (( step++ ))

    log_info "[${step}/4] Stopping CHR lab and wiping volumes..."
    compose "$(lab_compose_file)" down -v 2>/dev/null || true
    (( step++ ))

    log_info "[${step}/4] Stopping full CHR lab (if any)..."
    compose "$CHR_FULL_COMPOSE" down -v 2>/dev/null || true
    (( step++ ))

    log_info "[${step}/4] Stopping dev CHR lab (if any)..."
    compose "$CHR_DEV_COMPOSE" down -v 2>/dev/null || true

    log_ok "=== Full Reset complete — clean slate ==="
}
