#!/usr/bin/env bash
# =============================================================================
#  lib/lab/docker.sh — Docker & Compose operations with idempotency
#  Sourced by lab.sh — do not execute directly.
# =============================================================================

# ─── Environment Detection ────────────────────────────────────────────────────
# Returns 0 if we're inside a Codespace or dockerComposeFile-based devcontainer
# where the host Docker daemon manages sibling services (postgres, redis, etc.).
_in_devcontainer() {
    [[ "${CODESPACES:-}" == "true" ]] || [[ -n "${REMOTE_CONTAINERS:-}" ]] || \
    { [[ -f /.dockerenv ]] && [[ -n "${CODESPACE_NAME:-}" ]]; }
}

# Check if a compose service is already reachable on the network — meaning
# the host Docker (or another external process) is managing it.  This avoids
# trying to pull/start images inside a restricted inner Docker daemon.
_service_reachable() {
    local service="$1"
    case "$service" in
        postgres)
            command -v pg_isready >/dev/null 2>&1 && \
                pg_isready -h "${DB_HOST:-postgres}" -U "${DB_USER:-ispmonitor}" -q 2>/dev/null
            ;;
        redis)
            command -v redis-cli >/dev/null 2>&1 && \
                redis-cli -h "${REDIS_HOST:-redis}" ping >/dev/null 2>&1
            ;;
        prometheus)
            port_open "${PROMETHEUS_HOST:-localhost}" "${PROMETHEUS_PORT:-9090}"
            ;;
        grafana)
            port_open "${GRAFANA_HOST:-localhost}" "${GRAFANA_PORT:-3000}"
            ;;
        *)  return 1 ;;
    esac
}

# ─── Docker Daemon ────────────────────────────────────────────────────────────
# Detect whether dockerd is running; optionally auto-start it.
_DOCKER_STARTED_BY_LAB=0

docker_available() {
    if ! command -v docker >/dev/null 2>&1; then
        log_error "Docker is not installed. Run: ./lab.sh bootstrap"
        return 1
    fi
    if docker info >/dev/null 2>&1; then
        return 0
    fi

    # Attempt auto-start
    log_warn "Docker daemon is not running — attempting to start..."
    if docker_start_daemon; then
        return 0
    fi

    log_error "Docker daemon is not running and could not be started automatically."
    log_error "Try:  ./lab.sh run docker start   (or: sudo dockerd &)"
    return 1
}

# Start the Docker daemon. Handles Codespace / devcontainer environments where
# iptables and bridge creation may be blocked by the security context.
docker_start_daemon() {
    if docker info >/dev/null 2>&1; then
        log_ok "Docker daemon already running"
        return 0
    fi

    if ! command -v dockerd >/dev/null 2>&1; then
        log_error "dockerd not found on PATH — cannot start Docker daemon"
        return 1
    fi

    local dockerd_log="/tmp/dockerd-lab.log"
    local dockerd_pid_file="/tmp/dockerd-lab.pid"

    # Build flags — in restricted environments (Codespace, rootless containers)
    # we may need to disable iptables and bridge creation.
    local -a flags=()
    if ! sudo iptables -L -n >/dev/null 2>&1; then
        log_debug "iptables restricted — using --iptables=false --bridge=none"
        flags+=(--iptables=false --ip6tables=false --bridge=none)
    fi

    log_info "Starting dockerd (log: $dockerd_log)..."
    sudo dockerd "${flags[@]}" > "$dockerd_log" 2>&1 &
    local pid=$!
    echo "$pid" > "$dockerd_pid_file"

    # Wait for socket to appear (up to 30s)
    local waited=0
    while (( waited < 30 )); do
        # Fix socket permissions once it appears
        if [[ -S /var/run/docker.sock ]]; then
            sudo chmod 666 /var/run/docker.sock 2>/dev/null || true
        fi
        if docker info >/dev/null 2>&1; then
            _DOCKER_STARTED_BY_LAB=1
            log_ok "Docker daemon running (PID $pid)"
            return 0
        fi
        # If the process already exited, no point waiting
        if ! kill -0 "$pid" 2>/dev/null; then
            log_error "dockerd exited prematurely. Check $dockerd_log"
            return 1
        fi
        sleep 1
        (( waited++ ))
    done

    log_error "Docker daemon did not start within 30s. Check $dockerd_log"
    return 1
}

# Stop a daemon that was started by this lab session.
docker_stop_daemon() {
    local dockerd_pid_file="/tmp/dockerd-lab.pid"
    if [[ ! -f "$dockerd_pid_file" ]]; then
        log_warn "No lab-managed dockerd PID file found"
        return 0
    fi

    local pid
    pid="$(cat "$dockerd_pid_file")"
    if [[ -n "$pid" ]] && kill -0 "$pid" 2>/dev/null; then
        log_info "Stopping lab-managed dockerd (PID $pid)..."
        sudo kill "$pid"
        wait "$pid" 2>/dev/null || true
        rm -f "$dockerd_pid_file"
        log_ok "Docker daemon stopped"
    else
        log_info "dockerd (PID $pid) is not running"
        rm -f "$dockerd_pid_file"
    fi
}

docker_daemon_status() {
    if docker info >/dev/null 2>&1; then
        local server_version
        server_version="$(docker info --format '{{.ServerVersion}}' 2>/dev/null || echo 'unknown')"
        _ok "Docker daemon running (v${server_version})"
        local containers_running
        containers_running="$(docker info --format '{{.ContainersRunning}}' 2>/dev/null || echo '?')"
        _info "Containers running: $containers_running"
        local storage_driver
        storage_driver="$(docker info --format '{{.Driver}}' 2>/dev/null || echo '?')"
        _info "Storage driver: $storage_driver"
    else
        _err "Docker daemon is NOT running"
        _info "Start with: ./lab.sh run docker start"
    fi

    if _in_devcontainer; then
        echo
        _info "Environment: Codespace / devcontainer"
        _info "Host-managed services (reachable on network):"
        for svc in postgres redis prometheus grafana; do
            if _service_reachable "$svc"; then
                _ok "  $svc"
            else
                _err "  $svc — not reachable"
            fi
        done
    fi
}

# ─── Compose Detection ───────────────────────────────────────────────────────
# Auto-detects `docker compose` (plugin) or `docker-compose` (standalone).
declare -a COMPOSE_CMD=()

_detect_compose() {
    if [[ ${#COMPOSE_CMD[@]} -gt 0 ]]; then return 0; fi

    if command -v docker >/dev/null 2>&1 && docker compose version >/dev/null 2>&1; then
        COMPOSE_CMD=(docker compose)
        return 0
    fi
    if command -v docker-compose >/dev/null 2>&1; then
        COMPOSE_CMD=(docker-compose)
        return 0
    fi

    log_error "Docker Compose not available. Run: ./lab.sh bootstrap"
    return 1
}

# compose <file> <args...>  — thin wrapper around the detected compose command
compose() {
    local compose_file="$1"
    shift
    _detect_compose || return 1
    "${COMPOSE_CMD[@]}" -f "$compose_file" "$@"
}

# ─── Container State Queries ─────────────────────────────────────────────────
container_exists() {
    docker container inspect "$1" >/dev/null 2>&1
}

container_running() {
    local state
    state="$(docker container inspect -f '{{.State.Status}}' "$1" 2>/dev/null)" || return 1
    [[ "$state" == "running" ]]
}

container_healthy() {
    local health
    health="$(docker container inspect -f '{{.State.Health.Status}}' "$1" 2>/dev/null)" || return 1
    [[ "$health" == "healthy" ]]
}

# ─── Compose Service Queries ─────────────────────────────────────────────────
compose_service_running() {
    local compose_file="$1" service="$2"
    local status
    status="$(compose "$compose_file" ps --status running --format '{{.Name}}' "$service" 2>/dev/null)" || return 1
    [[ -n "$status" ]]
}

compose_services() {
    local compose_file="$1"
    compose "$compose_file" config --services 2>/dev/null
}

compose_running_services() {
    local compose_file="$1"
    compose "$compose_file" ps --status running --format '{{.Service}}' 2>/dev/null
}

# ─── Idempotent Compose Up ───────────────────────────────────────────────────
# Usage: ensure_compose_up <compose_file> [service1 service2 ...]
# - Detects host-managed services (Codespace) and skips them.
# - Skips services already running via compose.
# - Starts only what is needed.
# - Waits for health checks to pass before returning.
# - Retries on transient Docker failures.
ensure_compose_up() {
    local compose_file="$1"
    shift
    local -a services=("$@")
    local -a need_start=()
    local -a host_managed=()

    docker_available || return 1

    # Resolve full service list when none specified
    if [[ ${#services[@]} -eq 0 || ( ${#services[@]} -eq 1 && -z "${services[0]}" ) ]]; then
        mapfile -t services < <(compose_services "$compose_file")
    fi

    if [[ ${#services[@]} -eq 0 ]]; then
        log_error "No services found in $compose_file"
        return 1
    fi

    # Determine which services need starting
    for svc in "${services[@]}"; do
        if compose_service_running "$compose_file" "$svc"; then
            log_debug "Service '$svc' already running (compose) — skip"
        elif _in_devcontainer && _service_reachable "$svc"; then
            log_ok "Service '$svc' reachable (host-managed) — skip"
            host_managed+=("$svc")
        else
            need_start+=("$svc")
        fi
    done

    if [[ ${#need_start[@]} -eq 0 ]]; then
        log_ok "All services already running"
        [[ ${#host_managed[@]} -gt 0 ]] && \
            log_info "Host-managed: ${host_managed[*]}"
        return 0
    fi

    log_info "Starting: ${need_start[*]}"
    local compose_output
    compose_output="$(compose "$compose_file" up -d "${need_start[@]}" 2>&1)" || {
        # Detect the Codespace mount-namespace limitation
        if [[ "$compose_output" == *"mount namespace"* || "$compose_output" == *"operation not permitted"* ]]; then
            log_error "Cannot pull/start containers: mount namespace blocked"
            log_warn "Your Codespace/devcontainer restricts creating mount namespaces."
            log_warn "Fix: rebuild this Codespace — docker-in-docker feature has been added"
            log_warn "     to .devcontainer/devcontainer.json and will take effect on rebuild."
            log_warn ""
            log_warn "Meanwhile, services already reachable on the network: "
            for svc in "${need_start[@]}"; do
                if _service_reachable "$svc"; then
                    log_ok "  $svc  — reachable (host is running it)"
                else
                    log_error "  $svc  — NOT reachable"
                fi
            done
            return 1
        fi
        echo "$compose_output" >&2
        log_error "Failed to start: ${need_start[*]}"
        return 1
    }
    [[ -n "$compose_output" ]] && echo "$compose_output"
    log_ok "Started: ${need_start[*]}"

    # Wait for all services to become healthy (skip host-managed ones)
    local -a compose_managed=()
    for svc in "${services[@]}"; do
        local dominated=false
        for hm in "${host_managed[@]}"; do
            [[ "$svc" == "$hm" ]] && { dominated=true; break; }
        done
        $dominated || compose_managed+=("$svc")
    done

    if [[ ${#compose_managed[@]} -gt 0 ]]; then
        wait_compose_healthy "$compose_file" "${compose_managed[@]}"
    else
        log_ok "All services healthy (host-managed)"
    fi
}

# ─── Wait for Health ─────────────────────────────────────────────────────────
# Blocks until every listed service is healthy (or just running if no healthcheck).
wait_compose_healthy() {
    local compose_file="$1"
    shift
    local -a services=("$@")
    local timeout="${HEALTH_TIMEOUT:-120}"
    local interval=3
    local elapsed=0

    log_info "Waiting for health checks (timeout: ${timeout}s)..."

    while (( elapsed < timeout )); do
        local all_ready=true

        for svc in "${services[@]}"; do
            local cid
            cid="$(compose "$compose_file" ps -q "$svc" 2>/dev/null)" || { all_ready=false; continue; }
            [[ -z "$cid" ]] && { all_ready=false; continue; }

            # Does this container define a healthcheck?
            local has_hc
            has_hc="$(docker inspect -f '{{if .State.Health}}true{{else}}false{{end}}' "$cid" 2>/dev/null)" || { all_ready=false; continue; }

            if [[ "$has_hc" == "true" ]]; then
                container_healthy "$cid" || { all_ready=false; break; }
            else
                container_running "$cid" || { all_ready=false; break; }
            fi
        done

        if $all_ready; then
            log_ok "All services healthy (${elapsed}s)"
            return 0
        fi

        sleep "$interval"
        elapsed=$(( elapsed + interval ))
    done

    log_error "Health check timed out after ${timeout}s"
    compose "$compose_file" ps 2>/dev/null || true
    return 1
}

# ─── Idempotent Compose Down ─────────────────────────────────────────────────
# Usage: ensure_compose_down <compose_file> [extra_args...]
# Safe to call even when nothing is running.
ensure_compose_down() {
    local compose_file="$1"
    shift
    local -a extra_args=("$@")

    docker_available || return 1

    local running
    running="$(compose_running_services "$compose_file" 2>/dev/null || true)"

    if [[ -z "$running" ]]; then
        log_ok "No services running — nothing to stop"
        return 0
    fi

    log_info "Stopping services..."
    if ! compose "$compose_file" down "${extra_args[@]}"; then
        log_error "Failed to bring down services"
        return 1
    fi
    log_ok "All services stopped"
}

# ─── Formatted Status ────────────────────────────────────────────────────────
compose_show_status() {
    local compose_file="$1"
    docker_available || return 1
    compose "$compose_file" ps --format 'table {{.Name}}\t{{.Service}}\t{{.Status}}\t{{.Ports}}' 2>/dev/null || \
        compose "$compose_file" ps 2>/dev/null
}

# ─── Port Check ──────────────────────────────────────────────────────────────
port_open() {
    local host="${1:-127.0.0.1}" port="$2"
    if command -v nc &>/dev/null; then
        nc -z -w2 "$host" "$port" &>/dev/null
    elif command -v bash &>/dev/null; then
        (echo >/dev/tcp/"$host"/"$port") &>/dev/null
    else
        return 1
    fi
}
