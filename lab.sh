#!/usr/bin/env bash
# =============================================================================
#  ISP Visual Monitor — Interactive Lab Management
#  Usage: ./lab.sh
# =============================================================================
set -uo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
CHR_SCRIPTS="$SCRIPT_DIR/chr-lab/scripts"
CHR_CONFIGS="$SCRIPT_DIR/chr-lab/configs"
CHR_DEV_COMPOSE="$SCRIPT_DIR/chr-lab/docker-compose.chr.dev.yml"
CHR_FULL_COMPOSE="$SCRIPT_DIR/chr-lab/docker-compose.chr.yml"
MAIN_COMPOSE="$SCRIPT_DIR/docker-compose.yml"

# ─── Colours ──────────────────────────────────────────────────────────────────
C_RESET="\033[0m"
C_BOLD="\033[1m"
C_DIM="\033[2m"
C_RED="\033[0;31m"
C_GREEN="\033[0;32m"
C_YELLOW="\033[1;33m"
C_CYAN="\033[0;36m"
C_BLUE="\033[0;34m"
C_MAGENTA="\033[0;35m"
C_WHITE="\033[1;37m"

# ─── Router definitions ───────────────────────────────────────────────────────
# Format: role | container | ssh_port | api_host_port | webfig_port | winbox_port | config_file
declare -A R_NAME R_CONTAINER R_SSH_PORT R_API_PORT R_WEBFIG_PORT R_WINBOX_PORT R_CONFIG

_reg_router() {
  local key="$1"
  R_NAME[$key]="$2"
  R_CONTAINER[$key]="$3"
  R_SSH_PORT[$key]="$4"
  R_API_PORT[$key]="$5"
  R_WEBFIG_PORT[$key]="$6"
  R_WINBOX_PORT[$key]="$7"
  R_CONFIG[$key]="${8:-}"
}

# Single-router dev lab
_reg_router dev  "Dev (single)"         chr-01         2221 8728 18081 8291  ""

# Multi-router full lab
_reg_router core "Core Router"          chr-core-01    8722 8728 18081 8292  "$CHR_CONFIGS/chr-core-01.rsc"
_reg_router edge "Edge Router"          chr-edge-01    8723 8729 18082 8293  "$CHR_CONFIGS/chr-edge-01.rsc"
_reg_router acc  "Access Router (NAT)"  chr-access-01  8724 8730 18083 8294  "$CHR_CONFIGS/chr-access-01.rsc"
_reg_router ppp  "PPPoE Server"         chr-pppoe-01   8725 8731 18084 8295  "$CHR_CONFIGS/chr-pppoe-01.rsc"

FULL_ROUTERS=(core edge acc ppp)

# ─── Lab mode ─────────────────────────────────────────────────────────────────
LAB_MODE="${LAB_MODE:-dev}"   # dev | full

_compose_file() {
  if [[ "$LAB_MODE" == "full" ]]; then echo "$CHR_FULL_COMPOSE"
  else                                  echo "$CHR_DEV_COMPOSE"
  fi
}

_compose() { COMPOSE_FILE="$(_compose_file)" docker compose -f "$(_compose_file)" "$@"; }

# ─── Helpers ──────────────────────────────────────────────────────────────────
_header() {
  clear
  echo -e "${C_BOLD}${C_CYAN}"
  echo "  ╔══════════════════════════════════════════════════════════╗"
  echo "  ║          ISP Visual Monitor — Lab Management            ║"
  printf "  ║  Mode: %-10s  Lab: %-33s║\n" \
    "$(echo "$LAB_MODE" | tr '[:lower:]' '[:upper:]')" \
    "$(basename "$(_compose_file)")"
  echo "  ╚══════════════════════════════════════════════════════════╝"
  echo -e "${C_RESET}"
}

_section() {
  echo -e "${C_BOLD}${C_BLUE}── ${1} ──${C_RESET}"
  echo
}

_ok()   { echo -e "  ${C_GREEN}✔${C_RESET}  $*"; }
_warn() { echo -e "  ${C_YELLOW}⚠${C_RESET}  $*"; }
_err()  { echo -e "  ${C_RED}✘${C_RESET}  $*"; }
_info() { echo -e "  ${C_CYAN}→${C_RESET}  $*"; }

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
  # _pick_number <prompt> <max>  → echoes picked number, or empty on back/quit
  # All display goes to stderr so $() only captures the return value.
  local prompt="$1" max="$2" n
  echo -ne "${C_WHITE}  ${prompt} (1-${max}, 0=back): ${C_RESET}" >&2
  read -r n </dev/tty
  if [[ "$n" == "0" || -z "$n" ]]; then echo ""; return; fi
  if [[ "$n" =~ ^[0-9]+$ && "$n" -ge 1 && "$n" -le "$max" ]]; then
    echo "$n"
  else
    echo -e "  ${C_YELLOW}⚠${C_RESET}  Invalid choice — enter a number between 1 and ${max}, or 0 to go back." >&2
    echo ""
  fi
}

_run_ros_cmd() {
  local key="$1" cmd="$2"
  local host="127.0.0.1"
  local port="${R_SSH_PORT[$key]}"
  local user="${ROUTER_USER:-admin}"
  local pass="${ROUTER_PASS:-admin}"

  if ! command -v sshpass &>/dev/null; then
    _warn "sshpass not found. Install: sudo apt install -y sshpass"
    return 1
  fi
  echo -e "  ${C_DIM}→ SSH to ${R_CONTAINER[$key]} (port $port): $cmd${C_RESET}"
  echo
  ROUTER_HOST="$host" ROUTER_PORT="$port" ROUTER_USER="$user" ROUTER_PASS="$pass" \
    "$CHR_SCRIPTS/routeros-run.sh" "$cmd" 2>&1 || true
}

_menu_item() {
  local n="$1" label="$2" desc="${3:-}"
  printf "  ${C_WHITE}[%2s]${C_RESET} ${C_CYAN}%-28s${C_RESET} ${C_DIM}%s${C_RESET}\n" "$n" "$label" "$desc"
}

_separator() {
  echo -e "  ${C_DIM}──────────────────────────────────────${C_RESET}"
}

_router_badge() {
  local key="$1"
  echo -e "  ${C_BOLD}Router:${C_RESET} ${C_MAGENTA}${R_NAME[$key]}${C_RESET}  container=${C_CYAN}${R_CONTAINER[$key]}${C_RESET}  ssh=:${R_SSH_PORT[$key]}  api=:${R_API_PORT[$key]}  webfig=:${R_WEBFIG_PORT[$key]}"
  echo
}

_pick_router() {
  # Presents a router picker; echoes the key (dev|core|edge|acc|ppp) or ""
  # All display goes to stderr so $() only captures the key string.
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
  [[ -z "$choice" ]] && echo "" && return
  echo "${idx_map[$choice]}"
}

# =============================================================================
#  SECTION 1 — CHR Lab Control
# =============================================================================
_lab_control() {
  while true; do
    _header
    _section "CHR Lab Control"

    _menu_item 1 "Lab: Start (up -d)"         "bring lab up detached"
    _menu_item 2 "Lab: Stop (stop)"            "pause, keep containers"
    _menu_item 3 "Lab: Down (down)"            "remove containers"
    _menu_item 4 "Lab: Full Reset (down -v)"   "WARNING: destroys volumes"
    _menu_item 5 "Lab: Status"                 "show running containers"
    _menu_item 6 "Lab: Logs (all/service)"     "follow compose logs"
    _separator
    _menu_item 7 "Per-Router: Control"         "start/stop/pause/shell per router"
    _menu_item 8 "Per-Router: Apply Config"    "import .rsc config file"
    _menu_item 9 "Per-Router: Export Config"   "export running config to stdout"
    _menu_item 10 "Winbox / WebFig Info"       "connection details"
    _separator
    _menu_item 0 "Back"

    echo
    local choice
    choice="$(_pick_number "Choice" 10)"
    [[ -z "$choice" ]] && return

    case "$choice" in
      1)
        _header; _section "Starting CHR Lab"
        _info "Compose: $(_compose_file)"
        _compose up -d
        echo
        _ok "Lab started."
        ;;
      2)
        _header; _section "Stopping CHR Lab (stop)"
        _compose stop
        _ok "Lab stopped (containers preserved)."
        ;;
      3)
        _header; _section "Bringing Down CHR Lab"
        _confirm "Remove all lab containers?" && { _compose down; _ok "Done."; }
        ;;
      4)
        _header; _section "Full Reset — Destroys All Volumes"
        _warn "This will wipe ALL router data and configuration!"
        _confirm "Continue with full reset?" || { _warn "Cancelled."; _pause; continue; }
        _compose down -v
        _ok "Lab fully reset. All volumes removed."
        ;;
      5)
        _header; _section "Lab Status"
        _compose ps
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
        echo -ne "${C_WHITE}  Service name (Enter=all): ${C_RESET}"
        read -r svc
        _compose logs -f --tail=200 ${svc:-}
        ;;
      7)  _per_router_control ;;
      8)  _apply_config ;;
      9)  _export_config ;;
      10) _winbox_info ;;
    esac
    _pause
  done
}

_per_router_control() {
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
      1)
        _header; _section "Starting $container"
        docker start "$container" && _ok "Started." || _err "Failed."
        ;;
      2)
        _header; _section "Stopping $container"
        docker stop "$container" && _ok "Stopped." || _err "Failed."
        ;;
      3)
        _header; _section "Restarting $container"
        docker restart "$container" && _ok "Restarted." || _err "Failed."
        ;;
      4)
        _header; _section "Pausing $container"
        docker pause "$container" && _ok "Paused (CPU frozen, TCP sessions will time out)." || _err "Failed."
        ;;
      5)
        _header; _section "Unpausing $container"
        docker unpause "$container" && _ok "Unpaused." || _err "Failed."
        ;;
      6)
        _header; _section "Network Flap: $container"
        _info "Pausing for 10 seconds..."
        docker pause "$container" && _ok "Paused."
        sleep 10
        docker unpause "$container" && _ok "Unpaused after 10s — flap complete."
        ;;
      7)
        _header; _section "Logs: $container"
        docker logs -f --tail=200 "$container"
        ;;
      8)
        _header; _section "Shell: $container"
        _info "Entering container (Ctrl+D or exit to leave)"
        docker exec -it "$container" /bin/sh || true
        ;;
    esac
    _pause
  done
}

_apply_config() {
  local key
  key="$(_pick_router)"
  [[ -z "$key" ]] && return

  _header
  _router_badge "$key"
  _section "Apply RSC Config"

  local cfg="${R_CONFIG[$key]}"
  if [[ -z "$cfg" || ! -f "$cfg" ]]; then
    _warn "No config file found for ${R_NAME[$key]}"
    echo -ne "${C_WHITE}  Path to .rsc file: ${C_RESET}"
    read -r cfg
  fi

  if [[ ! -f "$cfg" ]]; then
    _err "File not found: $cfg"
    return
  fi

  _info "Copying $cfg → ${R_CONTAINER[$key]}:/apply.rsc"
  docker cp "$cfg" "${R_CONTAINER[$key]}:/apply.rsc"
  _info "Running /import apply.rsc via SSH..."
  _run_ros_cmd "$key" "/import apply.rsc"
  _ok "Config applied."
}

_export_config() {
  local key
  key="$(_pick_router)"
  [[ -z "$key" ]] && return

  _header
  _router_badge "$key"
  _section "Export Running Config"

  local out="${R_CONTAINER[$key]}-export-$(date +%Y%m%d-%H%M%S).rsc"
  _info "Exporting to $SCRIPT_DIR/$out"
  _run_ros_cmd "$key" "/export show-sensitive" >"$SCRIPT_DIR/$out" 2>&1
  _ok "Saved to: $out"
}

_winbox_info() {
  _header; _section "Winbox / WebFig Connection Info"

  if [[ "$LAB_MODE" == "full" ]]; then
    local keys=("${FULL_ROUTERS[@]}")
  else
    local keys=(dev)
  fi

  printf "  ${C_BOLD}%-18s %-22s %-20s %-20s${C_RESET}\n" "Router" "Container" "WebFig" "Winbox/API"
  _separator
  for k in "${keys[@]}"; do
    printf "  ${C_CYAN}%-18s${C_RESET} %-22s ${C_GREEN}http://127.0.0.1:%-5s${C_RESET} ${C_YELLOW}127.0.0.1:%-5s${C_RESET} (API :%-5s)\n" \
      "${R_NAME[$k]}" "${R_CONTAINER[$k]}" "${R_WEBFIG_PORT[$k]}" "${R_WINBOX_PORT[$k]}" "${R_API_PORT[$k]}"
  done
  echo
  _info "Winbox must connect manually (no neighbours via Docker NAT)."
  _info "Default user: admin / admin  |  Monitor user: monitor / monitor123"
}

# =============================================================================
#  SECTION 2 — Router Inspection
# =============================================================================
_router_inspect() {
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
        1) _section "Identity & Health"
           _run_ros_cmd "$key" "/system identity print"
           echo; _run_ros_cmd "$key" "/system resource print"
           echo; _run_ros_cmd "$key" "/system clock print"
           ;;
        2) _section "Interfaces"
           _run_ros_cmd "$key" "/interface print"
           echo; _run_ros_cmd "$key" "/interface monitor-traffic [find] once"
           ;;
        3) _section "Network / Routes / NAT"
           _run_ros_cmd "$key" "/interface print"
           echo; _run_ros_cmd "$key" "/ip address print"
           echo; _run_ros_cmd "$key" "/ip route print detail"
           echo; _run_ros_cmd "$key" "/ip firewall nat print"
           echo; _run_ros_cmd "$key" "/ip firewall filter print"
           ;;
        4) _section "PPPoE Sessions"
           _run_ros_cmd "$key" "/ppp active print detail"
           echo; _run_ros_cmd "$key" "/ppp secret print detail"
           ;;
        5) _section "DHCP Leases"
           _run_ros_cmd "$key" "/ip dhcp-server print detail"
           echo; _run_ros_cmd "$key" "/ip dhcp-server lease print detail"
           ;;
        6) _section "Services & SNMP"
           _run_ros_cmd "$key" "/ip service print detail"
           echo; _run_ros_cmd "$key" "/snmp print"
           ;;
        7) _section "Custom Command"
           echo -ne "${C_WHITE}  RouterOS command: ${C_RESET}"
           read -r custom_cmd
           [[ -z "$custom_cmd" ]] && continue
           _run_ros_cmd "$key" "$custom_cmd"
           ;;
      esac
      _pause
    done
  done
}

# =============================================================================
#  SECTION 3 — Simulation & Scenarios
# =============================================================================
_simulate() {
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
    _menu_item 7  "Scenario: API Disabled (SNMP only)" "disable RouterOS API service"
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
      1)  _sim_create_users ;;
      2)  _sim_cleanup_users ;;
      3)  _sim_show_sessions ;;
      4)  _sim_router_down ;;
      5)  _sim_router_up ;;
      6)  _sim_network_flap ;;
      7)  _sim_api_disable ;;
      8)  _sim_api_enable ;;
      9)  _sim_disconnect_all ;;
      10) _drill_peak_hour ;;
      11) _drill_churn_wave ;;
      12) _drill_full_chaos ;;
    esac
    _pause
  done
}

_sim_pppoe_key() {
  if [[ "$LAB_MODE" == "full" ]]; then echo "ppp"; else echo "dev"; fi
}

_sim_create_users() {
  _header; _section "Create PPPoE Users"
  local key; key="$(_sim_pppoe_key)"
  _router_badge "$key"
  echo -ne "${C_WHITE}  Number of users to create (default 20): ${C_RESET}"; read -r count
  echo -ne "${C_WHITE}  Username prefix (default 'cust'):        ${C_RESET}"; read -r prefix
  echo -ne "${C_WHITE}  Password (default 'pass123'):            ${C_RESET}"; read -r pass
  echo -ne "${C_WHITE}  PPP Profile (default 'pppoe-profile'):   ${C_RESET}"; read -r profile

  count="${count:-20}"
  prefix="${prefix:-cust}"
  pass="${pass:-pass123}"
  profile="${profile:-pppoe-profile}"

  _info "Creating ${count} users: ${prefix}001.. on ${R_CONTAINER[$key]}"
  ROUTER_HOST=127.0.0.1 ROUTER_PORT="${R_SSH_PORT[$key]}" \
  USER_PREFIX="$prefix" PASS="$pass" PROFILE="$profile" \
    "$CHR_SCRIPTS/create-pppoe-users.sh" "$count"
  _ok "Users created."
}

_sim_cleanup_users() {
  _header; _section "Cleanup PPPoE Users"
  local key; key="$(_sim_pppoe_key)"
  _router_badge "$key"
  echo -ne "${C_WHITE}  Username prefix to remove (default 'cust'): ${C_RESET}"; read -r prefix
  prefix="${prefix:-cust}"
  _confirm "Remove all PPP secrets + active sessions matching '${prefix}*'?" || return
  ROUTER_HOST=127.0.0.1 ROUTER_PORT="${R_SSH_PORT[$key]}" \
  USER_PREFIX="$prefix" \
    "$CHR_SCRIPTS/cleanup-pppoe-users.sh"
  _ok "Cleaned up."
}

_sim_show_sessions() {
  _header; _section "Active PPPoE Sessions"
  local key; key="$(_sim_pppoe_key)"
  _router_badge "$key"
  _run_ros_cmd "$key" "/ppp active print detail"
  echo
  echo -e "  ${C_DIM}Total:${C_RESET}"
  _run_ros_cmd "$key" "/ppp active print count-only" 2>/dev/null || true
}

_sim_router_down() {
  _header; _section "Scenario: Router Down"
  local key
  key="$(_pick_router)"
  [[ -z "$key" ]] && return
  _confirm "Stop ${R_CONTAINER[$key]} (simulates router down)?" || return
  docker stop "${R_CONTAINER[$key]}" && _ok "${R_CONTAINER[$key]} stopped — monitoring should detect it DOWN." || _err "Failed."
}

_sim_router_up() {
  _header; _section "Scenario: Router Up (Restore)"
  local key
  key="$(_pick_router)"
  [[ -z "$key" ]] && return
  docker start "${R_CONTAINER[$key]}" && _ok "${R_CONTAINER[$key]} started." || _err "Failed."
}

_sim_network_flap() {
  _header; _section "Scenario: Network Flap"
  local key
  key="$(_pick_router)"
  [[ -z "$key" ]] && return
  echo -ne "${C_WHITE}  Flap duration in seconds (default 15): ${C_RESET}"; read -r dur
  dur="${dur:-15}"
  _info "Pausing ${R_CONTAINER[$key]} for ${dur}s..."
  docker pause "${R_CONTAINER[$key]}" && _ok "PAUSED — sessions will start failing."
  sleep "$dur"
  docker unpause "${R_CONTAINER[$key]}" && _ok "UNPAUSED after ${dur}s — checking recovery..."
}

_sim_api_disable() {
  _header; _section "Scenario: Disable RouterOS API (forces SNMP fallback)"
  local key
  key="$(_pick_router)"
  [[ -z "$key" ]] && return
  _router_badge "$key"
  _confirm "Disable API service on ${R_NAME[$key]}?" || return
  _run_ros_cmd "$key" "/ip service disable api"
  _ok "API disabled. Monitor should fall back to SNMP within next poll cycle."
}

_sim_api_enable() {
  _header; _section "Scenario: Re-enable RouterOS API"
  local key
  key="$(_pick_router)"
  [[ -z "$key" ]] && return
  _router_badge "$key"
  _run_ros_cmd "$key" "/ip service enable api"
  _ok "API re-enabled."
}

_sim_disconnect_all() {
  _header; _section "Scenario: Disconnect All PPPoE Sessions"
  local key; key="$(_sim_pppoe_key)"
  _router_badge "$key"
  _confirm "Kick ALL active PPPoE sessions on ${R_NAME[$key]}?" || return
  _run_ros_cmd "$key" "/ppp active remove [find]"
  _ok "All PPPoE sessions disconnected."
}

_drill_peak_hour() {
  _header; _section "Drill: Peak Hour Simulation"
  local key; key="$(_sim_pppoe_key)"
  _router_badge "$key"

  echo -ne "${C_WHITE}  Users to ramp up (default 50):  ${C_RESET}"; read -r peak
  peak="${peak:-50}"
  echo -ne "${C_WHITE}  Hold time in seconds (default 60): ${C_RESET}"; read -r hold
  hold="${hold:-60}"

  _info "Ramp-up: creating ${peak} users..."
  ROUTER_HOST=127.0.0.1 ROUTER_PORT="${R_SSH_PORT[$key]}" \
  USER_PREFIX="peak" PASS="pass123" PROFILE="${PPPOE_PROFILE:-pppoe-profile}" \
    "$CHR_SCRIPTS/create-pppoe-users.sh" "$peak"
  _ok "Peak users created."

  _info "Holding for ${hold}s (simulate sustained load)..."
  sleep "$hold"

  _info "Ramp-down: removing peak users..."
  ROUTER_HOST=127.0.0.1 ROUTER_PORT="${R_SSH_PORT[$key]}" \
  USER_PREFIX="peak" \
    "$CHR_SCRIPTS/cleanup-pppoe-users.sh"
  _ok "Peak hour drill complete."
}

_drill_churn_wave() {
  _header; _section "Drill: Churn Wave (3 waves)"
  local key; key="$(_sim_pppoe_key)"
  _router_badge "$key"
  _info "Running 3 waves: create 30 users → wait 20s → remove → repeat"

  for wave in 1 2 3; do
    _info "Wave ${wave}/3: creating users..."
    ROUTER_HOST=127.0.0.1 ROUTER_PORT="${R_SSH_PORT[$key]}" \
    USER_PREFIX="wave${wave}" PASS="pass123" PROFILE="${PPPOE_PROFILE:-pppoe-profile}" \
      "$CHR_SCRIPTS/create-pppoe-users.sh" 30
    _ok "Wave ${wave} users created."
    sleep 20
    _info "Wave ${wave}/3: removing users..."
    ROUTER_HOST=127.0.0.1 ROUTER_PORT="${R_SSH_PORT[$key]}" \
    USER_PREFIX="wave${wave}" \
      "$CHR_SCRIPTS/cleanup-pppoe-users.sh"
    _ok "Wave ${wave} cleared."
    sleep 5
  done
  _ok "Churn wave drill complete."
}

_drill_full_chaos() {
  _header; _section "Drill: Full Chaos (multi-router + sessions)"
  _warn "This will:  flap Access router  +  create 40 PPPoE users  +  disconnect them  +  flap Edge router"
  _confirm "Run Full Chaos drill?" || return

  local ppp_key; ppp_key="$(_sim_pppoe_key)"

  _info "[1/6] Creating 40 PPPoE users..."
  ROUTER_HOST=127.0.0.1 ROUTER_PORT="${R_SSH_PORT[$ppp_key]}" \
  USER_PREFIX="chaos" PASS="pass123" PROFILE="${PPPOE_PROFILE:-pppoe-profile}" \
    "$CHR_SCRIPTS/create-pppoe-users.sh" 40 &
  wait
  _ok "PPPoE users created."

  if [[ "$LAB_MODE" == "full" ]]; then
    _info "[2/6] Flapping Access router for 8s..."
    docker pause chr-access-01 && sleep 8 && docker unpause chr-access-01
    _ok "Access router flap done."

    _info "[3/6] Stopping Edge router..."
    docker stop chr-edge-01
    _ok "Edge router down."
  fi

  sleep 10

  _info "[4/6] Disconnecting all PPPoE sessions..."
  _run_ros_cmd "$ppp_key" "/ppp active remove [find]" &>/dev/null || true
  _ok "Sessions disconnected."

  sleep 10

  _info "[5/6] Cleaning up chaos users..."
  ROUTER_HOST=127.0.0.1 ROUTER_PORT="${R_SSH_PORT[$ppp_key]}" \
  USER_PREFIX="chaos" \
    "$CHR_SCRIPTS/cleanup-pppoe-users.sh"
  _ok "Chaos users cleaned."

  if [[ "$LAB_MODE" == "full" ]]; then
    _info "[6/6] Restoring Edge router..."
    docker start chr-edge-01
    _ok "Edge router restored."
  fi

  _ok "Full Chaos drill complete. Check your monitoring dashboards!"
}

# =============================================================================
#  SECTION 4 — Backend Management
# =============================================================================
_backend() {
  while true; do
    _header
    _section "Backend Services"

    _menu_item 1  "Dev: Start All"           "postgres+redis+prometheus+grafana+API"
    _menu_item 2  "Dev: Stop All"            "kill API + docker compose down"
    _separator
    _menu_item 3  "Docker: Status"           "docker compose ps"
    _menu_item 4  "Docker: Start Infra Only" "postgres, redis, prometheus, grafana"
    _menu_item 5  "Docker: Down"             "stop + remove all containers"
    _separator
    _menu_item 6  "Logs: API"
    _menu_item 7  "Logs: Postgres"
    _menu_item 8  "Logs: Redis"
    _menu_item 9  "Logs: All Services"
    _separator
    _menu_item 10 "DB: Shell (psql)"
    _menu_item 11 "DB: Run Migrations"
    _menu_item 12 "DB: Load Test Data"       "db/examples/test_data.sql"
    _separator
    _menu_item 13 "API: Build Binary"        "make build"
    _menu_item 14 "API: Run Tests"           "make test"
    _menu_item 15 "API: Health Check"        "curl /health + /ready"
    _separator
    _menu_item 0  "Back"

    local c
    c="$(_pick_number "Choice" 15)"
    [[ -z "$c" ]] && return

    _header
    case "$c" in
      1)  _section "Starting All Dev Services"
          bash "$SCRIPT_DIR/scripts/dev-start.sh"
          ;;
      2)  _section "Stopping All Dev Services"
          bash "$SCRIPT_DIR/scripts/dev-stop.sh"
          ;;
      3)  _section "Docker Status"
          docker compose -f "$MAIN_COMPOSE" ps
          ;;
      4)  _section "Starting Infra Only"
          docker compose -f "$MAIN_COMPOSE" up -d postgres redis prometheus grafana
          ;;
      5)  _section "Docker Down"
          _confirm "Stop and remove all main stack containers?" && \
            docker compose -f "$MAIN_COMPOSE" down
          ;;
      6)  docker compose -f "$MAIN_COMPOSE" logs -f --tail=300 api ;;
      7)  docker compose -f "$MAIN_COMPOSE" logs -f --tail=300 postgres ;;
      8)  docker compose -f "$MAIN_COMPOSE" logs -f --tail=300 redis ;;
      9)  docker compose -f "$MAIN_COMPOSE" logs -f --tail=300 ;;
      10) _section "Database Shell"
          docker compose -f "$MAIN_COMPOSE" exec postgres psql -U ispmonitor -d ispmonitor
          ;;
      11) _section "Running Migrations"
          for f in "$SCRIPT_DIR/db/migrations/"*.sql; do
            _info "Applying: $(basename "$f")"
            docker compose -f "$MAIN_COMPOSE" exec -T postgres \
              psql -U ispmonitor -d ispmonitor < "$f" && _ok "OK" || _err "Failed: $f"
          done
          ;;
      12) _section "Loading Test Data"
          if [[ -f "$SCRIPT_DIR/db/examples/test_data.sql" ]]; then
            docker compose -f "$MAIN_COMPOSE" exec -T postgres \
              psql -U ispmonitor -d ispmonitor < "$SCRIPT_DIR/db/examples/test_data.sql" \
              && _ok "Test data loaded." || _err "Failed."
          else
            _warn "No test data file found at db/examples/test_data.sql"
          fi
          ;;
      13) _section "Building Go Binary"
          cd "$SCRIPT_DIR" && make build
          ;;
      14) _section "Running Tests"
          cd "$SCRIPT_DIR" && make test
          ;;
      15) _section "API Health Check"
          local api_url="http://localhost:8080"
          echo
          for ep in /health /ready /live; do
            printf "  ${C_CYAN}%-8s${C_RESET} " "$ep"
            code=$(curl -s -o /tmp/isp_hc.json -w "%{http_code}" "${api_url}${ep}" 2>/dev/null || echo "000")
            if [[ "$code" == "200" ]]; then
              echo -e "${C_GREEN}HTTP $code${C_RESET}  $(cat /tmp/isp_hc.json 2>/dev/null | tr -d '\n' | cut -c1-80)"
            else
              echo -e "${C_RED}HTTP $code${C_RESET}  (unreachable or error)"
            fi
          done
          ;;
    esac
    _pause
  done
}

# =============================================================================
#  SECTION 5 — Connectivity Checks
# =============================================================================
_checks() {
  while true; do
    _header
    _section "Connectivity Checks"

    _menu_item 1 "Port Check — current lab"     "nc probe all known ports"
    _menu_item 2 "Wait for Router Ready"        "block until API+WebFig respond"
    _menu_item 3 "Dev Check (ports + health)"   "wait-ready + port-check"
    _menu_item 4 "Show All Port Mappings"       "all routers + backend ports"
    _menu_item 5 "Ping: Router SSH port test"   "quick reachability probe"
    _separator
    _menu_item 0 "Back"

    local c
    c="$(_pick_number "Choice" 5)"
    [[ -z "$c" ]] && return

    _header
    case "$c" in
      1)  _section "Port Check — All Lab Routers"
          local keys
          if [[ "$LAB_MODE" == "full" ]]; then keys=("${FULL_ROUTERS[@]}"); else keys=(dev); fi
          for k in "${keys[@]}"; do
            echo -e "  ${C_BOLD}${R_NAME[$k]}${C_RESET}"
            for port in "${R_SSH_PORT[$k]}" "${R_API_PORT[$k]}" "${R_WEBFIG_PORT[$k]}" "${R_WINBOX_PORT[$k]}"; do
              if nc -z 127.0.0.1 "$port" &>/dev/null; then
                _ok "127.0.0.1:$port  OPEN"
              else
                _err "127.0.0.1:$port  CLOSED"
              fi
            done
            echo
          done
          _info "Backend API:"
          nc -z 127.0.0.1 8080 &>/dev/null && _ok "API  127.0.0.1:8080  OPEN" || _warn "API  127.0.0.1:8080  CLOSED (start dev server first)"
          ;;
      2)  _section "Waiting for Router Ready"
          local key
          key="$(_pick_router)"
          [[ -z "$key" ]] && continue
          _router_badge "$key"
          ROUTER_HOST=127.0.0.1 ROUTER_PORT="${R_SSH_PORT[$key]}" \
          API_PORT="${R_API_PORT[$key]}" WEBFIG_PORT="${R_WEBFIG_PORT[$key]}" \
            "$CHR_SCRIPTS/wait-ready.sh"
          ;;
      3)  _section "Dev Check"
          ROUTER_HOST=127.0.0.1 ROUTER_PORT="${R_SSH_PORT[dev]}" \
          API_PORT="${R_API_PORT[dev]}" WEBFIG_PORT="${R_WEBFIG_PORT[dev]}" \
            "$CHR_SCRIPTS/dev-check.sh"
          ;;
      4)  _section "All Port Mappings"
          echo
          printf "  ${C_BOLD}%-22s %-14s %-8s %-8s %-8s %-8s${C_RESET}\n" \
            "Router" "Container" "SSH" "API" "WebFig" "Winbox"
          _separator
          local all_keys=(dev "${FULL_ROUTERS[@]}")
          for k in "${all_keys[@]}"; do
            printf "  ${C_CYAN}%-22s${C_RESET} %-14s :%-7s :%-7s :%-7s :%-7s\n" \
              "${R_NAME[$k]}" "${R_CONTAINER[$k]}" \
              "${R_SSH_PORT[$k]}" "${R_API_PORT[$k]}" \
              "${R_WEBFIG_PORT[$k]}" "${R_WINBOX_PORT[$k]}"
          done
          echo
          echo -e "  ${C_BOLD}Backend stack${C_RESET}"
          echo "  API :8080   Postgres :5432   Redis :6379   Prometheus :9090   Grafana :3000"
          ;;
      5)  _section "SSH Port Probe"
          local key
          key="$(_pick_router)"
          [[ -z "$key" ]] && continue
          local p="${R_SSH_PORT[$key]}"
          nc -z -w3 127.0.0.1 "$p" && _ok "127.0.0.1:$p OPEN" || _err "127.0.0.1:$p CLOSED"
          ;;
    esac
    _pause
  done
}

# =============================================================================
#  SECTION 6 — Settings / Mode Switch
# =============================================================================
_settings() {
  while true; do
    _header
    _section "Settings"

    echo -e "  Current lab mode:  ${C_BOLD}${C_MAGENTA}${LAB_MODE}${C_RESET}"
    echo
    _menu_item 1 "Switch to DEV mode"   "single chr-01 router, dev compose"
    _menu_item 2 "Switch to FULL mode"  "4 routers: core/edge/access/pppoe"
    _separator
    _menu_item 3 "Set Router Credentials" "change default SSH user/pass"
    _separator
    _menu_item 0 "Back"

    local c
    c="$(_pick_number "Choice" 3)"
    [[ -z "$c" ]] && return

    case "$c" in
      1) LAB_MODE=dev;  _ok "Switched to DEV mode (single router)" ;;
      2) LAB_MODE=full; _ok "Switched to FULL mode (4 routers)" ;;
      3)
        echo -ne "${C_WHITE}  RouterOS SSH user (current: ${ROUTER_USER:-admin}): ${C_RESET}"
        read -r u
        echo -ne "${C_WHITE}  RouterOS SSH pass (current: ${ROUTER_PASS:-admin}): ${C_RESET}"
        read -rs p; echo
        [[ -n "$u" ]] && export ROUTER_USER="$u"
        [[ -n "$p" ]] && export ROUTER_PASS="$p"
        _ok "Credentials updated for this session."
        ;;
    esac
    _pause
  done
}

# =============================================================================
#  MAIN MENU
# =============================================================================
_main() {
  while true; do
    _header
    printf "  ${C_DIM}Lab mode: %s   Compose: %s${C_RESET}\n\n" \
      "$LAB_MODE" "$(basename "$(_compose_file)")"

    _menu_item 1 "CHR Lab Control"      "start/stop/reset lab, per-router ops"
    _menu_item 2 "Router Inspect"       "health, interfaces, PPPoE, DHCP"
    _menu_item 3 "Simulate & Scenarios" "failure drills, PPPoE churn, chaos"
    _menu_item 4 "Backend Services"     "API, Docker, DB, logs, build, test"
    _menu_item 5 "Connectivity Checks"  "port probes, wait-ready, port map"
    _separator
    _menu_item 6 "Settings / Lab Mode"  "switch dev↔full, set credentials"
    _separator
    _menu_item 0 "Quit"

    echo
    local c
    c="$(_pick_number "Choose" 6)"
    case "$c" in
      1) _lab_control ;;
      2) _router_inspect ;;
      3) _simulate ;;
      4) _backend ;;
      5) _checks ;;
      6) _settings ;;
      "") return 0 ;;
    esac
  done
}

# ─── Entry Point ──────────────────────────────────────────────────────────────
# Allow running a specific section directly: ./lab.sh simulate | backend | check
case "${1:-}" in
  lab)      _lab_control ;;
  inspect)  _router_inspect ;;
  sim|simulate) _simulate ;;
  backend)  _backend ;;
  check|checks) _checks ;;
  *) _main ;;
esac
