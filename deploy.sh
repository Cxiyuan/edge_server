#!/bin/bash

echo "=== Edge Server 部署脚本 ==="
echo ""

INSTALL_DIR="/opt/edge_server"

echo "1. 检查依赖..."
if ! command -v sqlite3 &> /dev/null; then
    echo "安装 sqlite3..."
    yum install -y sqlite || apt-get install -y sqlite3
fi

if ! command -v openssl &> /dev/null; then
    echo "安装 openssl..."
    yum install -y openssl || apt-get install -y openssl
fi

echo ""
echo "2. 安装 ocserv..."
if ! command -v ocserv &> /dev/null; then
    if [ -f /etc/redhat-release ]; then
        yum install -y epel-release
        yum install -y ocserv gnutls-utils
    elif [ -f /etc/debian_version ]; then
        apt-get update
        apt-get install -y ocserv gnutls-bin
    fi
    
    if command -v ocserv &> /dev/null; then
        echo "ocserv 安装成功: $(ocserv -v | head -1)"
    else
        echo "错误: ocserv 安装失败"
        exit 1
    fi
else
    echo "ocserv 已安装: $(ocserv -v | head -1)"
fi

echo ""
echo "3. 创建安装目录..."
mkdir -p "$INSTALL_DIR"
cd "$INSTALL_DIR"

echo ""
echo "4. 生成 SSL 证书..."
if [ ! -f server.crt ]; then
    openssl req -new -newkey rsa:4096 -days 3650 -nodes -x509 \
        -subj "/C=CN/ST=Beijing/L=Beijing/O=EdgeVPN/CN=edge-vpn.local" \
        -keyout server.key -out server.crt
    chmod 600 server.key
    echo "证书生成成功"
else
    echo "证书已存在，跳过生成"
fi

echo ""
echo "5. 初始化数据库..."
if [ ! -f server.db ]; then
    ./init_db.sh 2>/dev/null || echo "需要编译后的程序来初始化数据库"
else
    echo "数据库已存在"
fi

echo ""
echo "6. 配置系统服务..."
cat > /etc/systemd/system/edge-server.service << EOF
[Unit]
Description=Edge VPN Server
After=network.target

[Service]
Type=simple
User=root
WorkingDirectory=$INSTALL_DIR
ExecStart=$INSTALL_DIR/edge-server
Restart=on-failure
RestartSec=5
StandardOutput=journal
StandardError=journal

[Install]
WantedBy=multi-user.target
EOF

systemctl daemon-reload
echo "系统服务已配置"

echo ""
echo "=== 部署完成 ==="
echo ""
echo "后续步骤:"
echo "1. 上传编译好的 edge-server 程序到 $INSTALL_DIR"
echo "2. 启动服务: systemctl start edge-server"
echo "3. 查看状态: systemctl status edge-server"
echo "4. 查看日志: journalctl -u edge-server -f"
echo "5. 开机自启: systemctl enable edge-server"
echo ""
echo "默认管理账号: admin / admin123"
echo "Web 管理界面: https://$(hostname -I | awk '{print $1}'):443"
