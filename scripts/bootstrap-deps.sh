#!/usr/bin/env bash
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_DIR="$(dirname "$SCRIPT_DIR")"

CHECK_ONLY=false
SKIP_DOCKER=false
SKIP_REPO=false
REFRESH_NODE=false
NON_INTERACTIVE=false

while [[ $# -gt 0 ]]; do
  case "$1" in
    --check)
      CHECK_ONLY=true
      shift
      ;;
    --skip-docker)
      SKIP_DOCKER=true
      shift
      ;;
    --skip-repo)
      SKIP_REPO=true
      shift
      ;;
    --refresh-node)
      REFRESH_NODE=true
      shift
      ;;
    --non-interactive)
      NON_INTERACTIVE=true
      shift
      ;;
    *)
      echo "Unknown flag: $1" >&2
      echo "Usage: $0 [--check] [--skip-docker] [--skip-repo] [--refresh-node] [--non-interactive]" >&2
      exit 1
      ;;
  esac
done

log() {
  printf '%s\n' "$*"
}

have_cmd() {
  command -v "$1" >/dev/null 2>&1
}

have_docker_compose() {
  if have_cmd docker && docker compose version >/dev/null 2>&1; then
    return 0
  fi

  have_cmd docker-compose
}

repo_go_version() {
  if [[ ! -f "$PROJECT_DIR/go.mod" ]]; then
    return 1
  fi

  awk '/^go[[:space:]]+/ { print $2; exit }' "$PROJECT_DIR/go.mod"
}

current_go_version() {
  if ! have_cmd go; then
    return 1
  fi

  go env GOVERSION 2>/dev/null | sed 's/^go//'
}

version_gte() {
  local current="$1"
  local required="$2"

  [[ -n "$current" && -n "$required" ]] || return 1
  [[ "$(printf '%s\n%s\n' "$required" "$current" | sort -V | head -n1)" == "$required" ]]
}

need_sudo_prefix() {
  if [[ "$(id -u)" -eq 0 ]]; then
    return 0
  fi

  if have_cmd sudo; then
    return 0
  fi

  log "ERROR: root or sudo is required to install missing system packages." >&2
  exit 1
}

run_privileged() {
  if [[ "$(id -u)" -eq 0 ]]; then
    "$@"
  else
    sudo "$@"
  fi
}

disable_problematic_apt_sources() {
  local yarn_list="/etc/apt/sources.list.d/yarn.list"
  local disabled_yarn_list="/etc/apt/sources.list.d/yarn.list.disabled"

  if [[ -f "$yarn_list" ]]; then
    log "APT update failed; disabling problematic Yarn source and retrying."
    run_privileged mv "$yarn_list" "$disabled_yarn_list"
  fi
}

apt_update_with_retry() {
  if run_privileged apt-get update -qq; then
    return 0
  fi

  disable_problematic_apt_sources
  run_privileged apt-get update -qq
}

detect_pkg_manager() {
  if have_cmd apt-get; then
    echo apt
  elif have_cmd apk; then
    echo apk
  else
    echo ""
  fi
}

apt_pick_package() {
  local pkg
  for pkg in "$@"; do
    if apt-cache show "$pkg" >/dev/null 2>&1; then
      echo "$pkg"
      return 0
    fi
  done
  return 1
}

declare -a MISSING_FEATURES=()
declare -a INSTALL_PACKAGES=()

add_package() {
  local pkg
  for pkg in "$@"; do
    [[ -z "$pkg" ]] && continue
    if [[ " ${INSTALL_PACKAGES[*]} " != *" ${pkg} "* ]]; then
      INSTALL_PACKAGES+=("$pkg")
    fi
  done
}

note_missing() {
  local label="$1"
  if [[ " ${MISSING_FEATURES[*]} " != *" ${label} "* ]]; then
    MISSING_FEATURES+=("$label")
  fi
}

queue_if_missing() {
  local label="$1"
  local cmd_check="$2"
  local apt_pkg="$3"
  local apk_pkg="$4"
  local pkg_manager="$5"

  if have_cmd "$cmd_check"; then
    return 0
  fi

  note_missing "$label"
  case "$pkg_manager" in
    apt) add_package "$apt_pkg" ;;
    apk) add_package "$apk_pkg" ;;
  esac
}

queue_feature_packages() {
  local pkg_manager="$1"
  local required_go_version=""
  local installed_go_version=""

  queue_if_missing "bash" bash bash bash "$pkg_manager"
  queue_if_missing "curl" curl curl curl "$pkg_manager"
  queue_if_missing "git" git git git "$pkg_manager"
  queue_if_missing "ssh client" ssh openssh-client openssh-client "$pkg_manager"
  queue_if_missing "sshpass" sshpass sshpass sshpass "$pkg_manager"
  queue_if_missing "PostgreSQL client" psql postgresql-client postgresql-client "$pkg_manager"
  queue_if_missing "Redis CLI" redis-cli redis-tools redis "$pkg_manager"
  queue_if_missing "Node.js" node nodejs nodejs "$pkg_manager"
  queue_if_missing "npm" npm npm npm "$pkg_manager"
  queue_if_missing "lsof" lsof lsof lsof "$pkg_manager"
  queue_if_missing "netcat" nc netcat-openbsd netcat-openbsd "$pkg_manager"
  queue_if_missing "ping" ping iputils-ping iputils "$pkg_manager"
  queue_if_missing "pkill" pkill procps procps "$pkg_manager"

  required_go_version="$(repo_go_version || true)"
  if ! have_cmd go; then
    note_missing "Go toolchain"
    case "$pkg_manager" in
      apt) add_package golang-go ;;
      apk) add_package go ;;
    esac
  elif [[ -n "$required_go_version" ]]; then
    installed_go_version="$(current_go_version || true)"
    if [[ -n "$installed_go_version" ]] && ! version_gte "$installed_go_version" "$required_go_version"; then
      log "Go toolchain: local ${installed_go_version} is below repo requirement ${required_go_version}; bootstrap will use GOTOOLCHAIN=auto."
    fi
  fi

  if ! $SKIP_DOCKER; then
    if ! have_cmd docker; then
      note_missing "Docker CLI"
      case "$pkg_manager" in
        apt) add_package docker.io ;;
        apk) add_package docker ;;
      esac
    fi

    if ! have_docker_compose; then
      note_missing "Docker Compose"
      if [[ "$pkg_manager" == "apt" ]]; then
        local compose_pkg
        compose_pkg="$(apt_pick_package docker-compose-plugin docker-compose-v2 docker-compose || true)"
        add_package "$compose_pkg"
      else
        add_package docker-cli-compose
      fi
    fi
  fi
}

install_system_packages() {
  local pkg_manager="$1"

  if [[ ${#INSTALL_PACKAGES[@]} -eq 0 ]]; then
    log "System packages: all required commands already available."
    return 0
  fi

  need_sudo_prefix
  log "Installing system packages: ${INSTALL_PACKAGES[*]}"

  case "$pkg_manager" in
    apt)
      apt_update_with_retry
      run_privileged apt-get install -y -qq "${INSTALL_PACKAGES[@]}"
      ;;
    apk)
      run_privileged apk add --no-cache "${INSTALL_PACKAGES[@]}"
      ;;
    *)
      log "ERROR: Unsupported package manager." >&2
      exit 1
      ;;
  esac
}

ensure_repo_dependencies() {
  local required_go_version=""
  local installed_go_version=""

  if $SKIP_REPO; then
    return 0
  fi

  if [[ -f "$PROJECT_DIR/go.mod" ]] && have_cmd go; then
    required_go_version="$(repo_go_version || true)"
    installed_go_version="$(current_go_version || true)"
    if [[ -n "$required_go_version" && -n "$installed_go_version" ]] && ! version_gte "$installed_go_version" "$required_go_version"; then
      log "Ensuring Go modules with GOTOOLCHAIN=auto because local Go ${installed_go_version} is below ${required_go_version}..."
    else
      log "Ensuring Go modules..."
    fi
    (cd "$PROJECT_DIR" && GOTOOLCHAIN=auto go mod download)
  fi

  if [[ -f "$PROJECT_DIR/web/package.json" ]] && have_cmd npm; then
    if $REFRESH_NODE || [[ ! -d "$PROJECT_DIR/web/node_modules" ]]; then
      if [[ -f "$PROJECT_DIR/web/package-lock.json" ]]; then
        log "Ensuring frontend packages from web/package-lock.json..."
        (cd "$PROJECT_DIR/web" && npm ci)
      else
        log "Ensuring frontend packages from web/package.json..."
        (cd "$PROJECT_DIR/web" && npm install)
      fi
    else
      log "Frontend packages: web/node_modules already present; skip reinstall. Use --refresh-node to force reinstall."
    fi
  fi
}

main() {
  local pkg_manager
  pkg_manager="$(detect_pkg_manager)"

  if [[ -z "$pkg_manager" ]]; then
    log "ERROR: No supported package manager found. Supported: apt-get, apk." >&2
    exit 1
  fi

  queue_feature_packages "$pkg_manager"

  log "========================================"
  log "  ISPVisualMonitor Dependency Bootstrap"
  log "========================================"
  log "Package manager: $pkg_manager"
  if $SKIP_DOCKER; then
    log "Docker CLI: skipped by request"
  fi

  if [[ ${#MISSING_FEATURES[@]} -eq 0 ]]; then
    log "Missing commands: none"
  else
    log "Missing commands: ${MISSING_FEATURES[*]}"
  fi

  if $CHECK_ONLY; then
    if [[ ${#MISSING_FEATURES[@]} -eq 0 ]]; then
      exit 0
    fi
    exit 1
  fi

  install_system_packages "$pkg_manager"
  ensure_repo_dependencies

  log "Bootstrap complete."
}

main