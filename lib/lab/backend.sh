#!/usr/bin/env bash
# =============================================================================
#  lib/lab/backend.sh — Backend services: Docker infra, DB, Go build/test
#  Sourced by lab.sh — do not execute directly.
# =============================================================================

# ─── Compose Helpers ──────────────────────────────────────────────────────────
_main_compose() {
    compose "$MAIN_COMPOSE" "$@"
}

# ─── CLI Functions ────────────────────────────────────────────────────────────
backend_start() {
    _run_script "$PROJECT_DIR/scripts/dev-start.sh"
}

backend_stop() {
    _run_script "$PROJECT_DIR/scripts/dev-stop.sh"
}

backend_status() {
    compose_show_status "$MAIN_COMPOSE"
}

backend_infra_up() {
    ensure_compose_up "$MAIN_COMPOSE" postgres redis prometheus grafana
}

backend_down() {
    ensure_compose_down "$MAIN_COMPOSE"
}

backend_logs() {
    local service="${1:-}"
    _main_compose logs -f --tail=300 ${service:+"$service"}
}

backend_db_shell() {
    _main_compose exec postgres psql -U "$DB_USER" -d "$DB_NAME"
}

backend_migrate() {
    local f
    for f in "$PROJECT_DIR/db/migrations/"*.sql; do
        log_info "Applying: $(basename "$f")"
        _main_compose exec -T postgres \
            psql -U "$DB_USER" -d "$DB_NAME" < "$f" && log_ok "OK" || log_error "Failed: $f"
    done
}

backend_test_data() {
    local tf="$PROJECT_DIR/db/examples/test_data.sql"
    if [[ ! -f "$tf" ]]; then
        log_error "Test data file not found: $tf"
        return 1
    fi
    _main_compose exec -T postgres \
        psql -U "$DB_USER" -d "$DB_NAME" < "$tf" && log_ok "Test data loaded" || log_error "Failed"
}

backend_build() {
    log_info "Building Go binary..."
    mkdir -p "$PROJECT_DIR/bin"
    (cd "$PROJECT_DIR" && CGO_ENABLED=0 go build -o bin/ispmonitor ./cmd/ispmonitor)
    log_ok "Binary: bin/ispmonitor"
}

backend_test() {
    log_info "Running tests..."
    (cd "$PROJECT_DIR" && go test -v -race ./...)
}

backend_fmt() {
    (cd "$PROJECT_DIR" && go fmt ./...)
    log_ok "Code formatted"
}

backend_tidy() {
    (cd "$PROJECT_DIR" && go mod tidy)
    log_ok "Modules tidied"
}

backend_health() {
    local api_url="${API_URL:-http://localhost:8080}"
    local tmpdir hc_file ep code body
    tmpdir="$(_lab_tmpdir)"
    hc_file="${tmpdir}/hc_response.json"
    echo
    for ep in /health /ready /live; do
        if code=$(curl -s -o "$hc_file" -w "%{http_code}" "${api_url}${ep}" 2>/dev/null); then
            :
        else
            code="000"
        fi
        body="$(cat "$hc_file" 2>/dev/null | tr -d '\n' | cut -c1-80)"
        if [[ "$code" == "200" ]]; then
            printf "  ${C_GREEN}%-8s HTTP %s${C_RESET}  %s\n" "$ep" "$code" "$body"
        else
            printf "  ${C_RED}%-8s HTTP %s${C_RESET}  (unreachable or error)\n" "$ep" "$code"
        fi
    done
}

# ─── Interactive Menu ─────────────────────────────────────────────────────────
menu_backend() {
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
        _menu_item 13 "API: Build Binary"        "go build ./cmd/ispmonitor"
        _menu_item 14 "API: Run Tests"           "go test ./..."
        _menu_item 15 "API: Health Check"        "curl /health + /ready"
        _menu_item 16 "Go: Format"               "go fmt ./..."
        _menu_item 17 "Go: Tidy Modules"         "go mod tidy"
        _separator
        _menu_item 0  "Back"

        local c
        c="$(_pick_number "Choice" 17)"
        [[ -z "$c" ]] && return

        _header
        case "$c" in
            1)  _section "Starting All Dev Services"; backend_start ;;
            2)  _section "Stopping All Dev Services"; backend_stop ;;
            3)  _section "Docker Status"; backend_status ;;
            4)  _section "Starting Infra Only"; backend_infra_up ;;
            5)  _section "Docker Down"
                _confirm "Stop and remove all main stack containers?" && backend_down
                ;;
            6)  backend_logs api ;;
            7)  backend_logs postgres ;;
            8)  backend_logs redis ;;
            9)  backend_logs ;;
            10) _section "Database Shell"; backend_db_shell ;;
            11) _section "Running Migrations"; backend_migrate ;;
            12) _section "Loading Test Data"; backend_test_data ;;
            13) _section "Building Go Binary"; backend_build ;;
            14) _section "Running Tests"; backend_test ;;
            15) _section "API Health Check"; backend_health ;;
            16) _section "Formatting Go Code"; backend_fmt ;;
            17) _section "Tidying Go Modules"; backend_tidy ;;
        esac
        _pause
    done
}
