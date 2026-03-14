#!/usr/bin/env bash
# =============================================================================
#  lib/lab/core.sh — Foundation: logging, errors, retry, colors, lock
#  Sourced by lab.sh — do not execute directly.
# =============================================================================

# ─── Color System ─────────────────────────────────────────────────────────────
# Respects NO_COLOR (https://no-color.org/) and dumb terminals.
if [[ -z "${NO_COLOR:-}" && "${TERM:-dumb}" != "dumb" && -t 1 ]]; then
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
else
    C_RESET="" C_BOLD="" C_DIM="" C_RED="" C_GREEN=""
    C_YELLOW="" C_CYAN="" C_BLUE="" C_MAGENTA="" C_WHITE=""
fi

# ─── Logging ──────────────────────────────────────────────────────────────────
_ts() { date '+%H:%M:%S'; }

log_info()  { printf "${C_CYAN}[%s]${C_RESET} %s\n" "$(_ts)" "$*"; }
log_ok()    { printf "${C_GREEN}[%s] ✔${C_RESET} %s\n" "$(_ts)" "$*"; }
log_warn()  { printf "${C_YELLOW}[%s] ⚠${C_RESET} %s\n" "$(_ts)" "$*"; }
log_error() { printf "${C_RED}[%s] ✘${C_RESET} %s\n" "$(_ts)" "$*" >&2; }
log_debug() {
    [[ "${LAB_DEBUG:-0}" == "1" ]] && \
        printf "${C_DIM}[%s] DBG %s${C_RESET}\n" "$(_ts)" "$*"
    return 0
}

# Legacy aliases used throughout interactive menus
_ok()   { echo -e "  ${C_GREEN}✔${C_RESET}  $*"; }
_warn() { echo -e "  ${C_YELLOW}⚠${C_RESET}  $*"; }
_err()  { echo -e "  ${C_RED}✘${C_RESET}  $*"; }
_info() { echo -e "  ${C_CYAN}→${C_RESET}  $*"; }

# ─── Error Trap ───────────────────────────────────────────────────────────────
_trap_error() {
    local exit_code=$? line_no="${BASH_LINENO[0]}" cmd="${BASH_COMMAND}"
    if [[ $exit_code -ne 0 ]]; then
        log_error "Command failed (exit $exit_code) at line $line_no: $cmd"
    fi
}

_trap_exit() {
    _release_lock
    if [[ -n "${_LAB_TMPDIR:-}" && -d "$_LAB_TMPDIR" ]]; then
        rm -rf "$_LAB_TMPDIR"
    fi
}

# ─── Retry with Exponential Backoff ──────────────────────────────────────────
# Usage: retry_with_backoff <max_retries> <initial_delay_secs> <command...>
retry_with_backoff() {
    local max_retries="$1" delay="$2"
    shift 2
    local attempt=1

    while true; do
        if "$@"; then
            return 0
        fi
        if (( attempt >= max_retries )); then
            log_error "Failed after $max_retries attempts: $*"
            return 1
        fi
        log_warn "Attempt $attempt/$max_retries failed, retrying in ${delay}s..."
        sleep "$delay"
        delay=$(( delay * 2 ))
        (( attempt++ ))
    done
}

# ─── Generic Waiter ──────────────────────────────────────────────────────────
# Usage: wait_for <timeout_secs> <interval_secs> <description> <check_cmd...>
# Returns 0 when check_cmd succeeds, 1 on timeout.
wait_for() {
    local timeout="$1" interval="$2" desc="$3"
    shift 3
    local elapsed=0

    log_info "Waiting for ${desc} (timeout: ${timeout}s)..."
    while (( elapsed < timeout )); do
        if "$@" >/dev/null 2>&1; then
            log_ok "${desc} — ready (${elapsed}s)"
            return 0
        fi
        sleep "$interval"
        elapsed=$(( elapsed + interval ))
    done

    log_error "${desc} — timed out after ${timeout}s"
    return 1
}

# ─── Lock File ────────────────────────────────────────────────────────────────
_LOCK_FILE=""

_acquire_lock() {
    local name="${1:-lab}"
    _LOCK_FILE="/tmp/ispmonitor-${name}.lock"

    if [[ -f "$_LOCK_FILE" ]]; then
        local lock_pid
        lock_pid="$(cat "$_LOCK_FILE" 2>/dev/null || echo "")"
        if [[ -n "$lock_pid" ]] && kill -0 "$lock_pid" 2>/dev/null; then
            if [[ "${LAB_FORCE:-0}" != "1" ]]; then
                log_error "Another lab.sh running (PID $lock_pid). Set LAB_FORCE=1 to override."
                return 1
            fi
            log_warn "Forcing lock override (LAB_FORCE=1)"
        fi
    fi

    echo $$ > "$_LOCK_FILE"
    log_debug "Lock acquired: $_LOCK_FILE (PID $$)"
}

_release_lock() {
    if [[ -n "${_LOCK_FILE:-}" && -f "$_LOCK_FILE" ]]; then
        rm -f "$_LOCK_FILE"
        log_debug "Lock released"
    fi
}

# ─── Temp Directory ──────────────────────────────────────────────────────────
_LAB_TMPDIR=""

_lab_tmpdir() {
    if [[ -z "$_LAB_TMPDIR" ]]; then
        _LAB_TMPDIR="$(mktemp -d /tmp/ispmonitor-lab.XXXXXX)"
    fi
    echo "$_LAB_TMPDIR"
}

# ─── Validation ──────────────────────────────────────────────────────────────
require_cmd() {
    local cmd="$1" msg="${2:-}"
    if ! command -v "$cmd" >/dev/null 2>&1; then
        log_error "${msg:-Required command '$cmd' not found. Run: ./lab.sh bootstrap}"
        return 1
    fi
}

# ─── Script Runner ────────────────────────────────────────────────────────────
_run_script() {
    local script_path="$1"
    shift || true

    if [[ ! -f "$script_path" ]]; then
        log_error "Script not found: $script_path"
        return 1
    fi
    bash "$script_path" "$@"
}
