FROM ubuntu:24.04

LABEL maintainer="Edge VPN Server"
LABEL description="SSL VPN Server with Web Management Interface"

ENV DEBIAN_FRONTEND=noninteractive
ENV TZ=Asia/Shanghai

RUN apt-get update && apt-get install -y \
    ocserv \
    gnutls-bin \
    sqlite3 \
    openssl \
    iptables \
    iproute2 \
    net-tools \
    ca-certificates \
    tzdata \
    && apt-get clean \
    && rm -rf /var/lib/apt/lists/*

RUN mkdir -p /opt/edge_server \
    /etc/ocserv \
    /run/ocserv \
    /var/lib/ocserv \
    /var/log/edge_server

WORKDIR /opt/edge_server

COPY edge-server /opt/edge_server/
COPY server.conf /opt/edge_server/
COPY scripts/ /opt/edge_server/scripts/
COPY docker-entrypoint.sh /opt/edge_server/

RUN chmod +x /opt/edge_server/edge-server \
    && chmod +x /opt/edge_server/scripts/*.sh \
    && chmod +x /opt/edge_server/docker-entrypoint.sh

RUN echo "net.ipv4.ip_forward=1" >> /etc/sysctl.conf

EXPOSE 443/tcp 443/udp 8080/tcp

VOLUME ["/opt/edge_server/data", "/etc/ocserv"]

ENTRYPOINT ["/opt/edge_server/docker-entrypoint.sh"]
CMD ["edge-server"]
