FROM ubuntu:24.04

ENV DEBIAN_FRONTEND=noninteractive

RUN apt-get update && apt-get install -y --no-install-recommends \
    openssh-server \
    git \
    ca-certificates \
    docker.io \
    docker-compose-v2 \
    netcat-openbsd \
    curl \
    && rm -rf /var/lib/apt/lists/*

RUN useradd -m -s /bin/bash deploy \
    && usermod -aG docker deploy \
    && mkdir -p /var/run/sshd /home/deploy/.ssh \
    && chmod 700 /home/deploy/.ssh \
    && chown -R deploy:deploy /home/deploy/.ssh

COPY entrypoint.sh /entrypoint.sh
RUN sed -i 's/\r$//' /entrypoint.sh && chmod +x /entrypoint.sh

EXPOSE 22

ENTRYPOINT ["/entrypoint.sh"]
