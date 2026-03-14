FROM mcr.microsoft.com/devcontainers/go:1-1.24-bookworm

# Remove stale/broken Yarn APT source entries so feature apt-get update cannot fail on NO_PUBKEY.
RUN set -eux; \
    rm -f /etc/apt/sources.list.d/yarn.list /etc/apt/sources.list.d/yarn*.list; \
    if [ -f /etc/apt/sources.list ]; then sed -i '/dl\.yarnpkg\.com/d' /etc/apt/sources.list; fi; \
    rm -f /usr/share/keyrings/yarnkey.gpg /etc/apt/keyrings/yarn.gpg
