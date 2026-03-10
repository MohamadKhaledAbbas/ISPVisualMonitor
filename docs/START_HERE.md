# Codespaces Start Here

## Recommended Path

1. Open the Command Palette.
2. Run `Codespaces: Rebuild Container`.
3. Wait for the normal devcontainer to finish building.
4. Start the app with `make demo-start`.

After the rebuild, Codespaces should start the workspace container plus the `postgres` and `redis` sidecars automatically.

## Why Recovery Mode Happened

The creation log shows Codespaces first tried to attach to an already-expected container using `--expect-existing-container` and that container did not exist yet. Codespaces then fell back to a recovery container so the workspace could still open.

That is a container startup-path problem, not an application crash.

## What Is Fixed In This Repo

- The devcontainer now reuses the main [docker-compose.yml](docker-compose.yml) definitions for `postgres` and `redis` instead of duplicating them.
- The devcontainer override only defines the workspace container in [.devcontainer/docker-compose.devcontainer.yml](.devcontainer/docker-compose.devcontainer.yml).
- The fallback startup script in [scripts/devcontainer-start.sh](scripts/devcontainer-start.sh) now verifies both PostgreSQL and Redis before starting the app.

## Supported Run Modes

For Codespaces or Dev Containers:

```bash
make demo-start
```

For local Docker-based development on your host:

```bash
docker compose up -d postgres redis
make demo-start
```

Do not install Docker manually inside a recovery container. Rebuild into the normal devcontainer instead.

## Quick Checks

```bash
echo "Recovery mode: ${CODESPACES_RECOVERY_CONTAINER:-false}"
echo "Codespaces: ${CODESPACES:-false}"
psql -h postgres -U ispmonitor -d ispmonitor -c "SELECT 1;"
```

If `CODESPACES_RECOVERY_CONTAINER=true`, rebuild the container before continuing.
