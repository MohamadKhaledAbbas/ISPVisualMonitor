#!/usr/bin/env bash
# =============================================================================
#  lib/lab/bootstrap.sh — Environment bootstrap and dependency checking
#  Sourced by lab.sh — do not execute directly.
# =============================================================================

# ─── CLI Functions ────────────────────────────────────────────────────────────
bootstrap_check() {
    _run_bootstrap "Check Dependencies" --check
}

bootstrap_all() {
    _run_bootstrap "Install All Dependencies"
}

bootstrap_app() {
    _run_bootstrap "Install App Dependencies" --skip-docker
}

bootstrap_frontend() {
    _run_bootstrap "Refresh Frontend Packages" --skip-docker --refresh-node
}

_run_bootstrap() {
    local title="$1"
    shift

    if [[ ! -x "$BOOTSTRAP_SCRIPT" ]]; then
        log_error "Bootstrap script missing or not executable: $BOOTSTRAP_SCRIPT"
        return 1
    fi

    log_info "$title"
    bash "$BOOTSTRAP_SCRIPT" "$@"
}

# ─── Interactive Menu ─────────────────────────────────────────────────────────
menu_bootstrap() {
    while true; do
        _header
        _section "Environment Bootstrap"

        _menu_item 1 "Check Dependencies"        "audit required commands and packages"
        _menu_item 2 "Install All Dependencies"  "system packages + Go modules + Node packages"
        _menu_item 3 "Install App Dependencies"  "skip Docker CLI packages"
        _menu_item 4 "Refresh Frontend Packages" "force reinstall from web/package.json"
        _separator
        _menu_item 0 "Back"

        local c
        c="$(_pick_number "Choice" 4)"
        [[ -z "$c" ]] && return

        case "$c" in
            1)
                if bootstrap_check; then
                    _ok "All tracked dependencies are available."
                else
                    _warn "Some dependencies are missing. Run 'Install All Dependencies'."
                fi
                ;;
            2) bootstrap_all ;;
            3) bootstrap_app ;;
            4) bootstrap_frontend ;;
        esac
        _pause
    done
}
