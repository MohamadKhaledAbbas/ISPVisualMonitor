#!/usr/bin/env bash
# =============================================================================
#  lib/lab/cli.sh — Non-interactive CLI dispatcher and usage text
#  Sourced by lab.sh — do not execute directly.
# =============================================================================

cli_usage() {
    cat <<'EOF'
ISP Visual Monitor — Centralized Lab Controller

Usage:
  ./lab.sh                                      Interactive menu (default)
  ./lab.sh [flags] <section>                    Jump to interactive section
  ./lab.sh [flags] run <domain> <action> [args] Non-interactive command
  ./lab.sh [flags] preset <name>                Run a multi-step preset

Flags:
  --mode dev|full   Set lab mode (default: dev, or LAB_MODE env)
  --debug           Enable debug logging (LAB_DEBUG=1)
  -h, --help        Show this help

Interactive sections:
  bootstrap         Environment bootstrap menu
  lab               CHR lab control menu
  chr-scripts       CHR script wrappers menu
  inspect           Router inspection menu
  sim, simulate     Simulation & scenario menu
  demo              Demo scripts menu
  backend           Backend services menu
  checks            Connectivity checks menu
  repo-scripts      Repo utility scripts menu
  settings          Settings / lab mode menu

Domains and actions (./lab.sh run <domain> <action>):
  docker    start | stop | status            (Docker daemon control)
  lab       up | stop | down | reset | status | logs [service]
  chr       up | start | stop | down | reset | status | logs [service]
            health | interfaces | net | pppoe | dhcp | services
            ports | wait | dev-check | winbox | export
  demo      start | start-devcontainer | stop
            seed [host] [port] | reset [host] [port]
            scenario <name> [host] [port] | healthy [host] [port] | list
  backend   start | stop | status | infra-up | down
            logs [service] | db-shell | migrate | test-data
            build | test | fmt | tidy | health
  sim       pppoe-create <count> [prefix] [pass] [profile]
            pppoe-cleanup [prefix] | pppoe-show
            router-down <key> | router-up <key> | router-flap <key> [secs]
            api-disable <key> | api-enable <key> | disconnect-all
  repo      setup-dev | setup-prod | backup | backup-list
            backup-restore <file> | wait-lab | simulate-pppoe
  inspect   identity <key> | interfaces <key> | network <key>
            pppoe <key> | dhcp <key> | services <key> | custom <key> <cmd>
  checks    ports | wait-ready <key> | dev-check | port-map | ssh-probe <key>
            forwarded | recover <service|all>
  preset    quickstart | demo-showtime | chaos-suite | full-reset

Router keys:  dev  core  edge  acc  ppp

Examples:
  ./lab.sh                                      # interactive
  ./lab.sh run docker start                     # start Docker daemon first
  ./lab.sh --mode full run lab up               # start full 4-router lab
  ./lab.sh run backend infra-up                 # start postgres+redis+grafana
  ./lab.sh run demo scenario core-congestion    # trigger demo scenario
  ./lab.sh run sim router-flap edge 20          # flap edge router 20s
  ./lab.sh run inspect identity core            # show core router health
  ./lab.sh run checks ports                     # probe all lab ports
  ./lab.sh preset quickstart                    # bootstrap + lab up + seed + health
  ./lab.sh --mode full preset chaos-suite       # run full chaos drill
EOF
}

# ─── Domain Dispatchers ──────────────────────────────────────────────────────
cli_dispatch() {
    local domain="${1:-}"
    local action="${2:-}"

    if [[ -z "$domain" || -z "$action" ]]; then
        cli_usage
        return 1
    fi

    shift 2

    case "$domain" in
        docker)    _cli_docker "$action" "$@" ;;
        lab)       _cli_lab "$action" "$@" ;;
        chr)       _cli_chr "$action" "$@" ;;
        demo)      _cli_demo "$action" "$@" ;;
        backend)   _cli_backend "$action" "$@" ;;
        sim)       _cli_sim "$action" "$@" ;;
        repo)      _cli_repo "$action" "$@" ;;
        inspect)   _cli_inspect "$action" "$@" ;;
        checks)    _cli_checks "$action" "$@" ;;
        preset)    preset_run "$action" "$@" ;;
        *)
            log_error "Unknown domain: $domain"
            cli_usage
            return 1
            ;;
    esac
}

# ─── Per-Domain CLI Routers ──────────────────────────────────────────────────
_cli_docker() {
    local action="${1:-}"; shift || true
    case "$action" in
        start)   docker_start_daemon ;;
        stop)    docker_stop_daemon ;;
        status)  docker_daemon_status ;;
        *) log_error "Unknown docker action: $action. Use: start | stop | status"; return 1 ;;
    esac
}

_cli_lab() {
    local action="${1:-}"; shift || true
    case "$action" in
        up)           compose "$(lab_compose_file)" up -d ;;
        stop)         compose "$(lab_compose_file)" stop ;;
        down)         ensure_compose_down "$(lab_compose_file)" ;;
        reset)        compose "$(lab_compose_file)" down -v ;;
        status|ps)    compose_show_status "$(lab_compose_file)" ;;
        logs)         compose "$(lab_compose_file)" logs -f --tail=200 ${1:+"$1"} ;;
        *) log_error "Unknown lab action: $action"; return 1 ;;
    esac
}

_cli_chr() {
    local action="${1:-}"; shift || true
    export COMPOSE_FILE="$(lab_compose_file)"

    case "$action" in
        up|start|stop|down|reset|status)
            _run_script "$CHR_SCRIPTS/${action}.sh" ;;
        logs)
            _run_script "$CHR_SCRIPTS/logs.sh" ${1:+"$1"} ;;
        health)     _run_script "$CHR_SCRIPTS/show-health.sh" ;;
        interfaces) _run_script "$CHR_SCRIPTS/show-interfaces.sh" ;;
        net)        _run_script "$CHR_SCRIPTS/show-net.sh" ;;
        pppoe)      _run_script "$CHR_SCRIPTS/show-pppoe.sh" ;;
        dhcp)       _run_script "$CHR_SCRIPTS/show-dhcp.sh" ;;
        services)   _run_script "$CHR_SCRIPTS/show-services.sh" ;;
        ports)      _run_script "$CHR_SCRIPTS/ports-check.sh" ;;
        wait)       _run_script "$CHR_SCRIPTS/wait-ready.sh" ;;
        dev-check)  _run_script "$CHR_SCRIPTS/dev-check.sh" ;;
        winbox)     _run_script "$CHR_SCRIPTS/winbox-info.sh" ;;
        export)     _run_script "$CHR_SCRIPTS/export-config.sh" ;;
        *) log_error "Unknown chr action: $action"; return 1 ;;
    esac
}

_cli_demo() {
    local action="${1:-}"; shift || true
    case "$action" in
        start)              demo_start ;;
        start-devcontainer) demo_start_devcontainer ;;
        stop)               demo_stop ;;
        seed)               demo_seed "${1:-}" "${2:-}" ;;
        reset)              demo_reset "${1:-}" "${2:-}" ;;
        scenario)           demo_scenario "$@" ;;
        healthy)            demo_healthy "${1:-}" "${2:-}" ;;
        list)               demo_list_scenarios ;;
        *) log_error "Unknown demo action: $action"; return 1 ;;
    esac
}

_cli_backend() {
    local action="${1:-}"; shift || true
    case "$action" in
        start)     backend_start ;;
        stop)      backend_stop ;;
        status|ps) backend_status ;;
        infra-up)  backend_infra_up ;;
        down)      backend_down ;;
        logs)      backend_logs "${1:-}" ;;
        db-shell)  backend_db_shell ;;
        migrate)   backend_migrate ;;
        test-data) backend_test_data ;;
        build)     backend_build ;;
        test)      backend_test ;;
        fmt)       backend_fmt ;;
        tidy)      backend_tidy ;;
        health)    backend_health ;;
        *) log_error "Unknown backend action: $action"; return 1 ;;
    esac
}

_cli_sim() {
    local action="${1:-}"; shift || true
    case "$action" in
        pppoe-create)   sim_pppoe_create "$@" ;;
        pppoe-cleanup)  sim_pppoe_cleanup "${1:-cust}" ;;
        pppoe-show)     sim_pppoe_show ;;
        router-down)    sim_router_down "$@" ;;
        router-up)      sim_router_up "$@" ;;
        router-flap)    sim_router_flap "$@" ;;
        api-disable)    sim_api_disable "$@" ;;
        api-enable)     sim_api_enable "$@" ;;
        disconnect-all) sim_disconnect_all ;;
        peak-hour)      drill_peak_hour "${1:-50}" "${2:-60}" ;;
        churn-wave)     drill_churn_wave "${1:-3}" "${2:-30}" "${3:-20}" ;;
        full-chaos)     drill_full_chaos ;;
        *) log_error "Unknown sim action: $action"; return 1 ;;
    esac
}

_cli_repo() {
    local action="${1:-}"; shift || true
    case "$action" in
        setup-dev)       _run_script "$PROJECT_DIR/deploy/scripts/setup-dev.sh" ;;
        setup-prod)      _run_script "$PROJECT_DIR/deploy/scripts/setup-prod.sh" ;;
        backup)          _run_script "$PROJECT_DIR/deploy/scripts/backup.sh" backup ;;
        backup-list)     _run_script "$PROJECT_DIR/deploy/scripts/backup.sh" list ;;
        backup-restore)
            if [[ -z "${1:-}" ]]; then
                log_error "Missing filename. Usage: ./lab.sh run repo backup-restore <file>"
                return 1
            fi
            _run_script "$PROJECT_DIR/deploy/scripts/backup.sh" restore "$1"
            ;;
        wait-lab)        _run_script "$PROJECT_DIR/scripts/chr/wait-lab.sh" ;;
        simulate-pppoe)  _run_script "$PROJECT_DIR/scripts/chr/simulate-pppoe-activity.sh" ;;
        *) log_error "Unknown repo action: $action"; return 1 ;;
    esac
}

_cli_inspect() {
    local action="${1:-}"; shift || true
    case "$action" in
        identity)   inspect_identity "$@" ;;
        interfaces) inspect_interfaces "$@" ;;
        network)    inspect_network "$@" ;;
        pppoe)      inspect_pppoe "$@" ;;
        dhcp)       inspect_dhcp "$@" ;;
        services)   inspect_services "$@" ;;
        custom)     inspect_custom "$@" ;;
        *) log_error "Unknown inspect action: $action"; return 1 ;;
    esac
}

_cli_checks() {
    local action="${1:-}"; shift || true
    case "$action" in
        ports)       check_ports ;;
        wait-ready)  check_wait_ready "$@" ;;
        dev-check)   check_dev ;;
        port-map)    check_port_map ;;
        ssh-probe)   check_ssh_probe "$@" ;;
        forwarded)   check_forwarded_ports ;;
        recover)     check_recover_service "${1:-all}" ;;
        *) log_error "Unknown checks action: $action"; return 1 ;;
    esac
}
