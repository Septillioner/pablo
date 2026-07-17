#!/bin/bash
set -euo pipefail

DEPLOY_USER="deploy"
DEPLOY_HOME="/home/${DEPLOY_USER}"
AUTH_KEYS_FILE="${DEPLOY_HOME}/.ssh/authorized_keys"

mkdir -p "${DEPLOY_HOME}/.ssh" /home/deploy/repos /var/www
chmod 700 "${DEPLOY_HOME}/.ssh"
# Own writable deploy roots (do not recurse into /opt/fixtures — it is a read-only mount).
chown "${DEPLOY_USER}:${DEPLOY_USER}" /var/www /home/deploy/repos
chmod 775 /var/www
# Allow deploy to create /opt/<app> targets without touching the fixtures mount.
if [ -d /opt ]; then
    chown root:deploy /opt
    chmod 775 /opt
fi

if [ -n "${AUTHORIZED_KEYS:-}" ]; then
    echo "${AUTHORIZED_KEYS}" > "${AUTH_KEYS_FILE}"
elif [ -f /run/keys/id_ed25519.pub ]; then
    cp /run/keys/id_ed25519.pub "${AUTH_KEYS_FILE}"
else
    echo "No SSH authorized keys provided" >&2
    exit 1
fi

chmod 600 "${AUTH_KEYS_FILE}"
chown -R "${DEPLOY_USER}:${DEPLOY_USER}" "${DEPLOY_HOME}/.ssh"

init_bare_repo() {
    local bare_repo="$1"
    local source_dir="$2"
    if [ -d "${bare_repo}" ] || [ ! -d "${source_dir}" ]; then
        return 0
    fi
    git init --bare "${bare_repo}"
    git --git-dir="${bare_repo}" --work-tree="${source_dir}" add -A
    git --git-dir="${bare_repo}" --work-tree="${source_dir}" \
        -c user.email="e2e@pablo.local" -c user.name="Pablo E2E" \
        commit -m "e2e fixture" || true
    git --git-dir="${bare_repo}" branch -M main 2>/dev/null || \
        git --git-dir="${bare_repo}" update-ref refs/heads/main HEAD
}

init_bare_repo "/home/deploy/repos/sample-docker-app.git" "/opt/fixtures/sample-docker-app"
init_bare_repo "/home/deploy/repos/sample-php-app.git" "/opt/fixtures/sample-php-app"
chown -R "${DEPLOY_USER}:${DEPLOY_USER}" /home/deploy/repos

if [ -S /var/run/docker.sock ]; then
    chmod 666 /var/run/docker.sock || true
fi

if [ -x /usr/bin/docker ] && [ ! -f /usr/local/bin/docker-wrapper-installed ]; then
    mv /usr/bin/docker /usr/bin/docker-cli
    cat > /usr/local/bin/docker << 'EOF'
#!/bin/sh
export DOCKER_API_VERSION=1.43
exec /usr/bin/docker-cli "$@"
EOF
    chmod +x /usr/local/bin/docker
    touch /usr/local/bin/docker-wrapper-installed
fi

exec /usr/sbin/sshd -D -e
