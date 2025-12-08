#!/bin/bash

echo "检查 ocserv 安装状态..."

if command -v ocserv &> /dev/null; then
    echo "ocserv 已安装"
    ocserv -v
    exit 0
fi

echo "开始安装 ocserv..."

if [ -f /etc/redhat-release ]; then
    echo "检测到 CentOS/RHEL 系统"
    yum install -y epel-release
    yum install -y ocserv
elif [ -f /etc/debian_version ]; then
    echo "检测到 Debian/Ubuntu 系统"
    apt-get update
    apt-get install -y ocserv
else
    echo "不支持的系统类型"
    exit 1
fi

echo "ocserv 安装完成"
ocserv -v
