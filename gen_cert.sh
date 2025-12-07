#!/bin/bash

set -e

echo "正在生成自签名SSL证书..."

CERT_FILE="server.crt"
KEY_FILE="server.key"

if [ -f "$CERT_FILE" ] && [ -f "$KEY_FILE" ]; then
    echo "证书文件已存在，跳过生成"
    exit 0
fi

openssl req -x509 -newkey rsa:4096 -keyout "$KEY_FILE" -out "$CERT_FILE" -days 365 -nodes \
    -subj "/C=CN/ST=Beijing/L=Beijing/O=EdgeServer/OU=IT/CN=vpn.example.com"

echo "SSL证书生成完成！"
echo "证书文件: $CERT_FILE"
echo "密钥文件: $KEY_FILE"
