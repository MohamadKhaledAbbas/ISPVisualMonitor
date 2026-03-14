#!/usr/bin/env bash
# =============================================================================
#  lib/lab/menus.sh — Interactive menus: lab control, CHR wrappers, repo, settings
#  Sourced by lab.sh — do not execute directly.
# =============================================================================

# =============================================================================
#  CHR Lab Control
# =============================================================================
menu_lab_control() {
    while true; do
        _header
        _section "CHR Lab Control"

        _menu_item 1  "Lab: Start (up -d)"        "bring lab up detached"
        _menu_item 2  "Lab: Stop (stop)"           "pause, keep containers"
        _menu_item 3  "Lab: Down (down)"           "remove containers"
        _menu_item 4  "Lab: Full Reset (down -v)"  "WARNING: destroys volumes"
        _menu_item 5  "Lab: Status"                "show running containers"
        _menu_item 6  "Lab: Logs (all/service)"    "follow compose logs"
        _separator
        _menu_item 7  "Per-Router: Control"        "start/stop/pause/shell per router"
        _menu_item 8  "Per-Router: Apply Config"   "import .rsc config file"
        _menu_item 9  "Per-Router: Export Config"  "export running config to stdout"
        _menu_item 10 "Winbox / WebFig Info"       "connection details"
        _separator
        _menu_item 0  "Back"

        echo
        local choice
        choice="$(_pick_number "Choice" 10)"
        [[ -z "$choice" ]] && return

        case "$choice" in
            1)
                _header; _section "Starting CHR Lab"
                log_info "Compose: $(lab_compose_file)"
                compose "$(lab_compose_file)" up -d
                log_ok "Lab started."
                ;;
            2)
                _header; _section "Stopping CHR Lab"
                compose "$(lab_compose_file)" stop
                log_ok "Lab stopped (containers preserved)."
                ;;
            3)
                _header; _section "Bringing Down CHR Lab"
                _confirm "Remove all lab containers?" && {
                    ensure_compose_down "$(lab_compose_file)"
                }
                ;;
            4)
                _header; _section "Full Reset — Destroys All Volumes"
                _warn "This will wipe ALL router data and configuration!"
                _confirm "Continue with full reset?" || { _warn "Cancelled."; _pause; continue; }
                compose "$(lab_compose_file)" down -v
                log_ok "Lab fully reset. All volumes removed."
                ;;
            5)
                _header; _section "Lab Status"
                compose_show_status "$(lab_compose_file)"
                ;;
            6)
                _header; _section "Lab Logs"
                echo -e "  ${C_DIM}Available services:${C_RESET}"
                if [[ "$LAB_MODE" == "full" ]]; then
                    echo "  chr-core-01  chr-edge-01  chr-access-01  chr-pppoe-01"
                else
                    echo "  chr-01"
                fi
                echo
                local svc
                svc="$(_prompt_input "Service name (Enter=all)" "")"
                compose "$(lab_compose_file)" logs -f --tail=200 ${svc:+"$svc"}
                ;;
            7)  _menu_per_router_control ;;
            8)  _menu_apply_config ;;
            9)  _menu_export_config ;;
            10) _menu_winbox_info ;;
        esac
        _pause
    done
}

_menu_per_router_control() {
    local key
    key="$(_pick_router)"
    [[ -z "$key" ]] && return

    while true; do
        _header
        _router_badge "$key"
        _section "Per-Router Control"

        _menu_item 1 "Start"            "docker start ${R_CONTAINER[$key]}"
        _menu_item 2 "Stop"             "docker stop ${R_CONTAINER[$key]}"
        _menu_item 3 "Restart"          "docker restart ${R_CONTAINER[$key]}"
        _menu_item 4 "Pause"            "freeze CPU/network (flap simulation)"
        _menu_item 5 "Unpause"          "resume after pause"
        _menu_item 6 "Network Flap"     "pause 10s then unpause automatically"
        _menu_item 7 "Follow Logs"      "docker logs -f"
        _menu_item 8 "Open Shell"       "exec /bin/sh inside container"
        _separator
        _menu_item 0 "Back"

        local c container="${R_CONTAINER[$key]}"
        c="$(_pick_number "Choice" 8)"
        [[ -z "$c" ]] && return

        case "$c" in
            1)  _header; _section "Starting $container"
                docker start "$container" && _ok "Started." || _err "Failed." ;;
            2)  _header; _section "Stopping $container"
                docker stop "$container" && _ok "Stopped." || _err "Failed." ;;
            3)  _header; _section "Restarting $container"
                docker restart "$container" && _ok "Restarted." || _err "Failed." ;;
            4)  _header; _section "Pausing $container"
                docker pause "$container" && _ok "Paused (CPU frozen)." || _err "Failed." ;;
            5)  _header; _section "Unpausing $container"
                docker unpause "$container" && _ok "Unpaused." || _err "Failed." ;;
            6)  _header; _section "Network Flap: $container"
                _info "Pausing for 10 seconds..."
                if docker pause "$container"; then
                    _ok "Paused."
                    sleep 10
                    docker unpause "$container" && _ok "Unpaused after 10s — flap complete."
                else
                    _err "Failed to pause $container"
                fi ;;
            7)  _header; _section "Logs: $container"
                docker logs -f --tail=200 "$container" ;;
            8)  _header; _section "Shell: $container"
                _info "Entering container (Ctrl+D or exit to leave)"
                docker exec -it "$container" /bin/sh || true ;;
        esac
        _pause
    done
}

_menu_apply_config() {
    local key
    key="$(_pick_router)"
    [[ -z "$key" ]] && return

    _header
    _router_badge "$key"
    _section "Apply RSC Config"

    local cfg="${R_CONFIG[$key]}"
    if [[ -z "$cfg" || ! -f "$cfg" ]]; then
        _warn "No config file found for ${R_NAME[$key]}"
        cfg="$(_prompt_input "Path to .rsc file" "")"
    fi

    if [[ ! -f "$cfg" ]]; then
        _err "File not found: $cfg"
        return
    fi

    _apply_router_config "$key" "$cfg"
}

_menu_export_config() {
    local key
    key="$(_pick_router)"
    [[ -z "$key" ]] && return

    _header
    _router_badge "$key"
    _section "Export Running Config"
    _export_router_config "$key"
}

_menu_winbox_info() {
    _header; _section "Winbox / WebFig Connection Info"

    local keys
    if [[ "$LAB_MODE" == "full" ]]; then
        keys=("${FULL_ROUTERS[@]}")
    else
        keys=(dev)
    fi

    printf "  ${C_BOLD}%-18s %-22s %-20s %-20s${C_RESET}\n" "Router" "Container" "WebFig" "Winbox/API"
    _separator
    local k
    for k in "${keys[@]}"; do
        printf "  ${C_CYAN}%-18s${C_RESET} %-22s ${C_GREEN}http://127.0.0.1:%-5s${C_RESET} ${C_YELLOW}127.0.0.1:%-5s${C_RESET} (API :%-5s)\n" \
            "${R_NAME[$k]}" "${R_CONTAINER[$k]}" "${R_WEBFIG_PORT[$k]}" "${R_WINBOX_PORT[$k]}" "${R_API_PORT[$k]}"
    done
    echo
    _info "Winbox must connect manually (no neighbours via Docker NAT)."
    _info "Default user: admin / admin  |  Monitor user: monitor / monitor123"
}

# =============================================================================
#  CHR Script Wrappers
# =============================================================================
menu_chr_scripts() {
    while true; do
        _header
        _section "CHR Script Wrappers"

        _menu_item 1  "Compose Up"           "run chr-lab/scripts/up.sh"
        _menu_item 2  "Compose Start"        "run chr-lab/scripts/start.sh"
        _menu_item 3  "Compose Stop"         "run chr-lab/scripts/stop.sh"
        _menu_item 4  "Compose Down"         "run chr-lab/scripts/down.sh"
        _menu_item 5  "Compose Reset"        "run chr-lab/scripts/reset.sh"
        _menu_item 6  "Compose Status"       "run chr-lab/scripts/status.sh"
        _menu_item 7  "Compose Logs"         "run chr-lab/scripts/logs.sh"
        _separator
        _menu_item 8  "Show Health"          "run chr-lab/scripts/show-health.sh"
        _menu_item 9  "Show Interfaces"      "run chr-lab/scripts/show-interfaces.sh"
        _menu_item 10 "Show Network"         "run chr-lab/scripts/show-net.sh"
        _menu_item 11 "Show PPPoE"           "run chr-lab/scripts/show-pppoe.sh"
        _menu_item 12 "Show DHCP"            "run chr-lab/scripts/show-dhcp.sh"
        _menu_item 13 "Show Services"        "run chr-lab/scripts/show-services.sh"
        _menu_item 14 "Ports Check"          "run chr-lab/scripts/ports-check.sh"
        _menu_item 15 "Wait Ready"           "run chr-lab/scripts/wait-ready.sh"
        _menu_item 16 "Dev Check"            "run chr-lab/scripts/dev-check.sh"
        _menu_item 17 "Winbox Info"          "run chr-lab/scripts/winbox-info.sh"
        _menu_item 18 "Export Config"        "run chr-lab/scripts/export-config.sh"
        _separator
        _menu_item 0  "Back"

        local c
        c="$(_pick_number "Choice" 18)"
        [[ -z "$c" ]] && return

        _header
        _section "CHR Script Wrappers"
        export COMPOSE_FILE="$(lab_compose_file)"

        case "$c" in
            1)  _run_script "$CHR_SCRIPTS/up.sh" ;;
            2)  _run_script "$CHR_SCRIPTS/start.sh" ;;
            3)  _run_script "$CHR_SCRIPTS/stop.sh" ;;
            4)  _run_script "$CHR_SCRIPTS/down.sh" ;;
            5)  _run_script "$CHR_SCRIPTS/reset.sh" ;;
            6)  _run_script "$CHR_SCRIPTS/status.sh" ;;
            7)
                local svc
                svc="$(_prompt_input "Service name (Enter=all)" "")"
                _run_script "$CHR_SCRIPTS/logs.sh" ${svc:+"$svc"}
                ;;
            8)  _run_script "$CHR_SCRIPTS/show-health.sh" ;;
            9)  _run_script "$CHR_SCRIPTS/show-interfaces.sh" ;;
            10) _run_script "$CHR_SCRIPTS/show-net.sh" ;;
            11) _run_script "$CHR_SCRIPTS/show-pppoe.sh" ;;
            12) _run_script "$CHR_SCRIPTS/show-dhcp.sh" ;;
            13) _run_script "$CHR_SCRIPTS/show-services.sh" ;;
            14) _run_script "$CHR_SCRIPTS/ports-check.sh" ;;
            15) _run_script "$CHR_SCRIPTS/wait-ready.sh" ;;
            16) _run_script "$CHR_SCRIPTS/dev-check.sh" ;;
            17) _run_script "$CHR_SCRIPTS/winbox-info.sh" ;;
            18) _run_script "$CHR_SCRIPTS/export-config.sh" ;;
        esac
        _pause
    done
}

# =============================================================================
#  Repo Utility Scripts
# =============================================================================
menu_repo_scripts() {
    while true; do
        _header
        _section "Repo Utility Scripts"

        _menu_item 1 "Dev Setup"               "run deploy/scripts/setup-dev.sh"
        _menu_item 2 "Prod Setup"              "run deploy/scripts/setup-prod.sh"
        _menu_item 3 "Backup: Create"          "run deploy/scripts/backup.sh backup"
        _menu_item 4 "Backup: List"            "run deploy/scripts/backup.sh list"
        _menu_item 5 "Backup: Restore"         "run deploy/scripts/backup.sh restore"
        _separator
        _menu_item 6 "CHR Wait Lab"            "run scripts/chr/wait-lab.sh"
        _menu_item 7 "CHR PPPoE Activity Sim"  "run scripts/chr/simulate-pppoe-activity.sh"
        _separator
        _menu_item 0 "Back"

        local c
        c="$(_pick_number "Choice" 7)"
        [[ -z "$c" ]] && return

        _header
        _section "Repo Utility Scripts"
        case "$c" in
            1) _run_script "$PROJECT_DIR/deploy/scripts/setup-dev.sh" ;;
            2) _run_script "$PROJECT_DIR/deploy/scripts/setup-prod.sh" ;;
            3) _run_script "$PROJECT_DIR/deploy/scripts/backup.sh" backup ;;
            4) _run_script "$PROJECT_DIR/deploy/scripts/backup.sh" list ;;
            5)
                local restore_file
                restore_file="$(_prompt_input "Backup file to restore" "")"
                [[ -n "$restore_file" ]] && _run_script "$PROJECT_DIR/deploy/scripts/backup.sh" restore "$restore_file"
                ;;
            6) _run_script "$PROJECT_DIR/scripts/chr/wait-lab.sh" ;;
            7) _run_script "$PROJECT_DIR/scripts/chr/simulate-pppoe-activity.sh" ;;
        esac
        _pause
    done
}

# =============================================================================
#  Settings / Mode Switch
# =============================================================================
menu_settings() {
    while true; do
        _header
        _section "Settings"

        echo -e "  Current lab mode:  ${C_BOLD}${C_MAGENTA}${LAB_MODE}${C_RESET}"
        echo
        _menu_item 1 "Switch to DEV mode"    "single chr-01 router, dev compose"
        _menu_item 2 "Switch to FULL mode"   "4 routers: core/edge/access/pppoe"
        _separator
        _menu_item 3 "Set Router Credentials" "change default SSH user/pass"
        _menu_item 4 "Show Version / Info"    "lib path, compose file, router count"
        _separator
        _menu_item 0 "Back"

        local c
        c="$(_pick_number "Choice" 4)"
        [[ -z "$c" ]] && return

        case "$c" in
            1) LAB_MODE=dev;  _ok "Switched to DEV mode (single router)" ;;
            2) LAB_MODE=full; _ok "Switched to FULL mode (4 routers)" ;;
            3)
                local u p
                u="$(_prompt_input "RouterOS SSH user" "${ROUTER_USER:-admin}")"
                echo -ne "${C_WHITE}  RouterOS SSH pass (current: ${ROUTER_PASS:-admin}): ${C_RESET}"
                read -rs p; echo
                [[ -n "$u" ]] && export ROUTER_USER="$u"
                [[ -n "$p" ]] && export ROUTER_PASS="$p"
                _ok "Credentials updated for this session."
                ;;
            4)
                echo
                echo -e "  ${C_BOLD}Project:${C_RESET}  $PROJECT_DIR"
                echo -e "  ${C_BOLD}Lib:${C_RESET}      $LIB_DIR"
                echo -e "  ${C_BOLD}Mode:${C_RESET}     $LAB_MODE"
                echo -e "  ${C_BOLD}Compose:${C_RESET}  $(lab_compose_file)"
                echo -e "  ${C_BOLD}Routers:${C_RESET}  $(active_router_keys)"
                echo -e "  ${C_BOLD}Modules:${C_RESET}  $(ls "$LIB_DIR"/*.sh | wc -l) files in lib/lab/"
                ;;
        esac
        _pause
    done
}

# =============================================================================
#  Docker Daemon & Compose Control
# =============================================================================
menu_docker() {
    while true; do
        _header
        _section "Docker & Compose Control"

        # Live status indicator
        if docker info >/dev/null 2>&1; then
            echo -e "  ${C_GREEN}● Docker daemon: RUNNING${C_RESET}"
        else
            echo -e "  ${C_RED}● Docker daemon: STOPPED${C_RESET}"
        fi
        if command -v docker >/dev/null 2>&1 && docker compose version >/dev/null 2>&1; then
            echo -e "  ${C_GREEN}● Compose plugin: $(docker compose version --short 2>/dev/null || echo 'available')${C_RESET}"
        elif command -v docker-compose >/dev/null 2>&1; then
            echo -e "  ${C_YELLOW}● Compose standalone: $(docker-compose version --short 2>/dev/null || echo 'available')${C_RESET}"
        else
            echo -e "  ${C_RED}● Docker Compose: NOT FOUND${C_RESET}"
        fi
        echo

        _menu_item 1  "Start Docker Daemon"       "auto-detects Codespace/devcontainer mode"
        _menu_item 2  "Stop Docker Daemon"         "stop lab-managed dockerd"
        _menu_item 3  "Docker Daemon Status"       "detailed info"
        _separator
        _menu_item 4  "Compose: Start Backend"     "postgres, redis, prometheus, grafana"
        _menu_item 5  "Compose: Start CHR Lab"     "router containers (current mode)"
        _menu_item 6  "Compose: Stop Backend"      "stop main stack"
        _menu_item 7  "Compose: Stop CHR Lab"      "stop router containers"
        _menu_item 8  "Compose: Status — Backend"  "docker compose ps (main)"
        _menu_item 9  "Compose: Status — CHR Lab"  "docker compose ps (lab)"
        _menu_item 10 "Compose: Logs — Backend"    "follow main stack logs"
        _menu_item 11 "Compose: Logs — CHR Lab"    "follow CHR lab logs"
        _separator
        _menu_item 12 "Docker: List Containers"    "docker ps -a"
        _menu_item 13 "Docker: List Images"        "docker images"
        _menu_item 14 "Docker: Prune (cleanup)"    "remove stopped containers + dangling images"
        _separator
        _menu_item 0  "Back"

        local c
        c="$(_pick_number "Choice" 14)"
        [[ -z "$c" ]] && return

        _header
        case "$c" in
            1)  _section "Starting Docker Daemon"
                docker_start_daemon
                ;;
            2)  _section "Stopping Docker Daemon"
                docker_stop_daemon
                ;;
            3)  _section "Docker Daemon Status"
                docker_daemon_status
                echo
                if docker info >/dev/null 2>&1; then
                    echo -e "  ${C_BOLD}Docker Info:${C_RESET}"
                    docker info 2>/dev/null | sed 's/^/  /' | head -20
                fi
                ;;
            4)  _section "Starting Backend Infrastructure"
                docker_available || { _pause; continue; }
                ensure_compose_up "$MAIN_COMPOSE" postgres redis prometheus grafana
                ;;
            5)  _section "Starting CHR Lab"
                docker_available || { _pause; continue; }
                compose "$(lab_compose_file)" up -d
                log_ok "CHR lab started"
                ;;
            6)  _section "Stopping Backend"
                docker_available || { _pause; continue; }
                _confirm "Stop all backend services?" && ensure_compose_down "$MAIN_COMPOSE"
                ;;
            7)  _section "Stopping CHR Lab"
                docker_available || { _pause; continue; }
                _confirm "Stop all CHR containers?" && ensure_compose_down "$(lab_compose_file)"
                ;;
            8)  _section "Backend Status"
                docker_available || { _pause; continue; }
                compose_show_status "$MAIN_COMPOSE"
                ;;
            9)  _section "CHR Lab Status"
                docker_available || { _pause; continue; }
                compose_show_status "$(lab_compose_file)"
                ;;
            10) docker_available || { _pause; continue; }
                local svc
                svc="$(_prompt_input "Service name (Enter=all)" "")"
                compose "$MAIN_COMPOSE" logs -f --tail=200 ${svc:+"$svc"}
                ;;
            11) docker_available || { _pause; continue; }
                local svc
                svc="$(_prompt_input "Service name (Enter=all)" "")"
                compose "$(lab_compose_file)" logs -f --tail=200 ${svc:+"$svc"}
                ;;
            12) _section "All Containers"
                docker_available || { _pause; continue; }
                docker ps -a --format 'table {{.Names}}\t{{.Image}}\t{{.Status}}\t{{.Ports}}' 2>/dev/null
                ;;
            13) _section "Docker Images"
                docker_available || { _pause; continue; }
                docker images --format 'table {{.Repository}}\t{{.Tag}}\t{{.Size}}\t{{.CreatedSince}}' 2>/dev/null
                ;;
            14) _section "Docker Prune"
                _warn "This removes stopped containers and dangling images."
                if _confirm "Continue?"; then
                    docker_available || { _pause; continue; }
                    docker container prune -f 2>/dev/null
                    docker image prune -f 2>/dev/null
                    log_ok "Prune complete"
                fi
                ;;
        esac
        _pause
    done
}

# =============================================================================
#  Main Menu
# =============================================================================
menu_main() {
    while true; do
        _header
        printf "  ${C_DIM}Lab mode: %s   Compose: %s${C_RESET}\n\n" \
            "$LAB_MODE" "$(basename "$(lab_compose_file)")"

        _menu_item 1  "Docker & Compose"      "daemon start/stop, compose up/down/logs"
        _menu_item 2  "Environment Bootstrap" "install and verify lab/app dependencies"
        _menu_item 3  "CHR Lab Control"       "start/stop/reset lab, per-router ops"
        _menu_item 4  "CHR Script Wrappers"   "direct access to chr-lab script entry points"
        _menu_item 5  "Router Inspect"        "health, interfaces, PPPoE, DHCP"
        _menu_item 6  "Simulate & Scenarios"  "failure drills, PPPoE churn, chaos"
        _menu_item 7  "Demo Scripts"          "seed/reset/start/stop/demo scenarios"
        _menu_item 8  "Backend Services"      "API, Docker, DB, logs, build, test"
        _menu_item 9  "Connectivity Checks"   "port probes, wait-ready, port map"
        _separator
        _menu_item 10 "Repo Utility Scripts"  "deploy/setup/backup/chr extra scripts"
        _menu_item 11 "Settings / Lab Mode"   "switch dev↔full, set credentials, info"
        _menu_item 12 "Run Preset Workflow"   "quickstart, demo-showtime, chaos-suite"
        _separator
        _menu_item 0  "Quit"

        echo
        local c
        c="$(_pick_number "Choose" 12)"
        case "$c" in
            1)  menu_docker ;;
            2)  menu_bootstrap ;;
            3)  menu_lab_control ;;
            4)  menu_chr_scripts ;;
            5)  menu_inspect ;;
            6)  menu_simulate ;;
            7)  menu_demo ;;
            8)  menu_backend ;;
            9)  menu_checks ;;
            10) menu_repo_scripts ;;
            11) menu_settings ;;
            12) _menu_preset_picker ;;
            "") return 0 ;;
        esac
    done
}

_menu_preset_picker() {
    _header
    _section "Preset Workflows"

    _menu_item 1 "Quickstart"    "bootstrap → lab up → seed → health"
    _menu_item 2 "Demo Showtime" "infra → seed → healthy → verify"
    _menu_item 3 "Chaos Suite"   "lab up → peak → churn → full chaos"
    _menu_item 4 "Full Reset"    "stop everything, wipe all volumes"
    _separator
    _menu_item 0 "Back"

    local c
    c="$(_pick_number "Choice" 4)"
    [[ -z "$c" ]] && return

    case "$c" in
        1) preset_quickstart ;;
        2) preset_demo_showtime ;;
        3) preset_chaos_suite ;;
        4) preset_full_reset ;;
    esac
    _pause
}
