#!/bin/bash

echo "=== 构建 Edge VPN Server Docker 镜像 ==="

VERSION=${1:-latest}

echo "构建版本: $VERSION"

if [ ! -f edge-server ]; then
    echo "错误: edge-server 可执行文件不存在"
    echo "请先编译或从 GitHub Actions 下载"
    exit 1
fi

echo "开始构建 Docker 镜像..."
docker build -t edge-server:${VERSION} .

if [ $? -eq 0 ]; then
    echo ""
    echo "✅ 镜像构建成功!"
    echo ""
    echo "镜像标签: edge-server:${VERSION}"
    echo ""
    echo "使用方法:"
    echo "  docker-compose up -d"
    echo ""
    echo "或者:"
    echo "  docker run -d --name edge-vpn \\"
    echo "    --privileged \\"
    echo "    --cap-add=NET_ADMIN \\"
    echo "    --device=/dev/net/tun \\"
    echo "    -p 443:443/tcp \\"
    echo "    -p 443:443/udp \\"
    echo "    -p 8443:8443 \\"
    echo "    -v \$(pwd)/data:/opt/edge_server/data \\"
    echo "    edge-server:${VERSION}"
else
    echo "❌ 镜像构建失败"
    exit 1
fi
