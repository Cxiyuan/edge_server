# 端点网络接入平台

基于 Go + Vue3 开发的 SSL VPN 服务器，兼容 Cisco AnyConnect 客户端。

## 功能特性

- ✅ 兼容 Cisco AnyConnect 客户端
- ✅ 基于 OpenConnect 协议的二次开发
- ✅ 现代化 Web 管理界面（Vue3 + Element Plus）
- ✅ 用户组管理和网络策略配置
- ✅ 实时在线用户监控
- ✅ 完整的日志审计功能
- ✅ SQLite 数据库，无需外部依赖
- ✅ 单一二进制文件部署

## 系统要求

- Linux (Ubuntu/CentOS)
- OpenSSL (用于生成证书)
- SQLite3 (用于初始化数据库，可选)

## 快速开始

### 下载预编译版本

从 GitHub Actions 下载最新的构建版本：

```bash
wget https://github.com/yourusername/edge_server/releases/latest/download/edge-server-linux-amd64.tar.gz
tar -xzf edge-server-linux-amd64.tar.gz
cd linux-amd64
```

### 初始化

1. 生成 SSL 证书：
```bash
./gen_cert.sh
```

2. 初始化数据库：
```bash
./init_db.sh
```

### 运行服务

```bash
./edge-server
```

服务启动后：
- Web 管理界面: http://localhost:8080
- VPN 服务端口: 443

默认管理员账号：
- 用户名: `admin`
- 密码: `admin123`

## 配置说明

编辑 `server.conf` 文件进行配置：

```ini
[server]
web_port = 8080        # Web 管理界面端口
vpn_port = 443         # VPN 服务端口
db_path = server.db    # 数据库文件路径

[ssl]
server_cert = server.crt
server_key = server.key

[network]
ip_pool = 192.168.100.0/24   # VPN IP 地址池
dns1 = 8.8.8.8
dns2 = 8.8.4.4
mtu = 1400

[system]
max_clients = 100
idle_timeout = 3600
```

## 功能说明

### 首页
- 系统资源监控（CPU、内存、磁盘）
- 在线用户统计
- 网络连接统计
- 系统运行时间

### 用户组配置
- 创建用户组
- 配置网络访问策略
- 设置路由规则

### 用户管理
- 创建/编辑/删除用户
- 分配用户组
- 为用户配置独立的路由和策略

### 在线用户
- 实时显示在线用户列表
- 查看连接信息（IP、MAC、协议等）
- 网络流量统计（上下行速率、总流量）

### 日志审计
- 用户认证日志
- 网络访问日志

## 开发构建

### 前端开发

```bash
cd frontend
npm install
npm run dev
```

### 后端开发

```bash
go mod download
go run main.go
```

### 构建

```bash
cd frontend
npm run build
cd ..
go build -o edge-server .
```

## 文件结构

```
edge_server/
├── main.go                 # 主程序入口
├── go.mod                  # Go 依赖
├── server.conf             # 配置文件
├── models/                 # 数据模型
│   └── database.go
├── handlers/               # API 处理器
│   └── api.go
├── vpn/                    # VPN 服务
│   └── ocserv.go
├── frontend/               # 前端项目
│   ├── package.json
│   ├── vite.config.js
│   ├── index.html
│   └── src/
│       ├── main.js
│       ├── App.vue
│       ├── router/
│       └── views/
├── static/                 # 前端构建输出
├── .github/
│   └── workflows/
│       └── build.yml       # GitHub Actions 配置
├── init_db.sh              # 数据库初始化脚本
└── gen_cert.sh             # 证书生成脚本
```

## 许可证

MIT License
