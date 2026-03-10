# Codespaces Notes

## Root Cause

The creation log shows this sequence:

1. Codespaces tried `devcontainer up --expect-existing-container`.
2. The expected devcontainer did not exist yet.
3. Codespaces opened a recovery container so the workspace remained accessible.
4. A later `devcontainer up` without `--expect-existing-container` succeeded.

This means recovery mode was triggered by the container attach/create flow, not by the Go app, PostgreSQL, or Redis crashing.

## Normal Codespaces Run Path

The intended setup is:

- Main service definitions live in [docker-compose.yml](docker-compose.yml).
- The workspace container is defined in [.devcontainer/docker-compose.devcontainer.yml](.devcontainer/docker-compose.devcontainer.yml).
- [.devcontainer/devcontainer.json](.devcontainer/devcontainer.json) starts only `postgres` and `redis` as sidecars.
- The app itself runs inside the devcontainer with `make demo-start`.

## What To Do If Recovery Mode Appears

1. Run `Codespaces: Rebuild Container`.
2. Wait for the normal devcontainer to build.
3. Verify you are not in recovery mode:

```bash
echo "Recovery mode: ${CODESPACES_RECOVERY_CONTAINER:-false}"
```

4. Start the app:

```bash
make demo-start
```

## Validation Checks

Inside the normal devcontainer, these should work:

```bash
psql -h postgres -U ispmonitor -d ispmonitor -c "SELECT 1;"
redis-cli -h redis ping
```

## Troubleshooting

If either sidecar is unavailable, use the fallback startup script only after rebuilding the container:

```bash
bash scripts/devcontainer-start.sh
```

That script now checks both PostgreSQL and Redis before starting the API and frontend so failures surface early instead of half-starting the stack.
