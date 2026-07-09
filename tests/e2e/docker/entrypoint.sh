#!/bin/bash
set -euo pipefail

DEPLOY_USER="deploy"
DEPLOY_HOME="/home/${DEPLOY_USER}"
AUTH_KEYS_FILE="${DEPLOY_HOME}/.ssh/authorized_keys"
BARE_REPO="/home/deploy/repos/sample-docker-app.git"
SOURCE_DIR="/opt/fixtures/sample-docker-app"

mkdir -p "${DEPLOY_HOME}/.ssh" /home/deploy/repos
chmod 700 "${DEPLOY_HOME}/.ssh"

if [ -n "${AUTHORIZED_KEYS:-}" ]; then
    echo "${AUTHORIZED_KEYS}" > "${AUTH_KEYS_FILE}"
elif [ -f /run/keys/id_ed25519.pub ]; then
    cp /run/keys/id_ed25519.pub "${AUTH_KEYS_FILE}"
else
    echo "No SSH authorized keys provided" >&2
    exit 1
fi

chmod 600 "${AUTH_KEYS_FILE}"
chown -R "${DEPLOY_USER}:${DEPLOY_USER}" "${DEPLOY_HOME}/.ssh" /home/deploy/repos

if [ ! -d "${BARE_REPO}" ] && [ -d "${SOURCE_DIR}" ]; then
    git init --bare "${BARE_REPO}"
    git --git-dir="${BARE_REPO}" --work-tree="${SOURCE_DIR}" add -A
    git --git-dir="${BARE_REPO}" --work-tree="${SOURCE_DIR}" \
        -c user.email="e2e@pablo.local" -c user.name="Pablo E2E" \
        commit -m "e2e fixture" || true
    git --git-dir="${BARE_REPO}" branch -M main 2>/dev/null || \
        git --git-dir="${BARE_REPO}" update-ref refs/heads/main HEAD
    chown -R "${DEPLOY_USER}:${DEPLOY_USER}" /home/deploy/repos
fi

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
