#!/bin/bash
set -e

echo "=== Edge VPN Server - Docker 启动 ==="

if [ ! -f /opt/edge_server/data/server.db ]; then
    echo "首次启动，初始化数据库..."
    mkdir -p /opt/edge_server/data
    touch /opt/edge_server/data/server.db
fi

if [ ! -f /opt/edge_server/data/server.crt ]; then
    echo "生成 SSL 证书..."
    cd /opt/edge_server/data
    openssl req -new -newkey rsa:4096 -days 3650 -nodes -x509 \
        -subj "/C=CN/ST=Beijing/L=Beijing/O=EdgeVPN/CN=${VPN_HOSTNAME:-edge-vpn.local}" \
        -keyout server.key -out server.crt
    chmod 600 server.key
    cd /opt/edge_server
fi

ln -sf /opt/edge_server/data/server.db /opt/edge_server/server.db
ln -sf /opt/edge_server/data/server.crt /opt/edge_server/server.crt
ln -sf /opt/edge_server/data/server.key /opt/edge_server/server.key

if [ -n "$DB_PATH" ]; then
    export DB_PATH="/opt/edge_server/data/server.db"
fi

echo "启用 IP 转发..."
echo 1 > /proc/sys/net/ipv4/ip_forward || true

echo "配置 iptables NAT..."
iptables -t nat -A POSTROUTING -s ${VPN_NETWORK:-192.168.100.0/24} -o eth0 -j MASQUERADE || true

echo "检查环境变量配置..."
if [ -n "$WEB_PORT" ]; then
    echo "Web端口: $WEB_PORT"
    sed -i "s/web_port = .*/web_port = $WEB_PORT/" /opt/edge_server/server.conf
fi

if [ -n "$VPN_PORT" ]; then
    echo "VPN端口: $VPN_PORT"
    sed -i "s/vpn_port = .*/vpn_port = $VPN_PORT/" /opt/edge_server/server.conf
fi

if [ -n "$VPN_NETWORK" ]; then
    echo "VPN网络: $VPN_NETWORK"
    sed -i "s|ip_pool = .*|ip_pool = $VPN_NETWORK|" /opt/edge_server/server.conf
fi

if [ -n "$DNS_SERVER1" ]; then
    sed -i "s/dns1 = .*/dns1 = $DNS_SERVER1/" /opt/edge_server/server.conf
fi

if [ -n "$DNS_SERVER2" ]; then
    sed -i "s/dns2 = .*/dns2 = $DNS_SERVER2/" /opt/edge_server/server.conf
fi

if [ -n "$MAX_CLIENTS" ]; then
    sed -i "s/max_clients = .*/max_clients = $MAX_CLIENTS/" /opt/edge_server/server.conf
fi

if [ -n "$IDLE_TIMEOUT" ]; then
    sed -i "s/idle_timeout = .*/idle_timeout = $IDLE_TIMEOUT/" /opt/edge_server/server.conf
fi

echo "启动 Edge VPN Server..."
echo "Web 管理界面: https://$(hostname -i):${WEB_PORT:-443}"
echo "默认账号: admin / admin123"
echo ""

exec "$@"
