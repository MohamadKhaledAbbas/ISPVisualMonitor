#!/usr/bin/env bash
# =============================================================================
#  ISP Visual Monitor — Centralized Lab Controller
#
#  Usage:
#    ./lab.sh                                  Interactive menu
#    ./lab.sh [flags] <section>                Jump to interactive section
#    ./lab.sh [flags] run <domain> <action>    Non-interactive command
#    ./lab.sh [flags] preset <name>            Multi-step workflow
#    ./lab.sh --help                           Show full reference
#
#  Modules loaded from: lib/lab/
# =============================================================================
set -uo pipefail

# ─── Module Loader ────────────────────────────────────────────────────────────
_LIB_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/lib/lab" && pwd)"

# Load order matters: config first (paths/registry), then core (logging),
# then docker (compose), then everything else.
_MODULES=(
    config
    core
    docker
    menu
    router
    bootstrap
    backend
    checks
    inspect
    simulate
    demo
    presets
    cli
    menus
)

for _mod in "${_MODULES[@]}"; do
    _mod_path="$_LIB_DIR/${_mod}.sh"
    if [[ -f "$_mod_path" ]]; then
        # shellcheck source=/dev/null
        source "$_mod_path"
    else
        echo "FATAL: Missing module: $_mod_path" >&2
        exit 1
    fi
done
unset _mod _mod_path _MODULES

# ─── Traps ────────────────────────────────────────────────────────────────────
trap '_trap_error' ERR
trap '_trap_exit'  EXIT

# ─── Argument Parsing ─────────────────────────────────────────────────────────
while [[ $# -gt 0 ]]; do
    case "$1" in
        --mode)
            LAB_MODE="${2:-}"
            if [[ "$LAB_MODE" != "dev" && "$LAB_MODE" != "full" ]]; then
                log_error "--mode must be 'dev' or 'full'"
                exit 1
            fi
            shift 2
            ;;
        --debug)
            export LAB_DEBUG=1
            shift
            ;;
        --help|-h|help)
            cli_usage
            exit 0
            ;;
        --)
            shift
            break
            ;;
        *)
            break
            ;;
    esac
done

# ─── Route ────────────────────────────────────────────────────────────────────
case "${1:-}" in
    # Interactive menus
    ""|menu|interactive) menu_main ;;
    bootstrap)           menu_bootstrap ;;
    lab)                 menu_lab_control ;;
    chr-scripts)         menu_chr_scripts ;;
    inspect)             menu_inspect ;;
    sim|simulate)        menu_simulate ;;
    demo)                menu_demo ;;
    backend)             menu_backend ;;
    check|checks)        menu_checks ;;
    repo-scripts)        menu_repo_scripts ;;
    settings)            menu_settings ;;

    # Non-interactive dispatch
    run|cmd)             shift; cli_dispatch "$@" ;;

    # Presets
    preset)              shift; preset_run "$@" ;;

    *)
        log_error "Unknown command: $1"
        echo
        cli_usage
        exit 1
        ;;
esac
