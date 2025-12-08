# Docker 部署指南

## 快速启动

### 1. 使用 docker-compose (推荐)

```bash
# 下载项目
git clone https://github.com/yourusername/edge_server.git
cd edge_server

# 从 GitHub Actions 下载编译好的程序
# 或者本地编译: go build -o edge-server .

# 启动服务
docker-compose up -d

# 查看日志
docker-compose logs -f

# 停止服务
docker-compose down
```

### 2. 使用 docker run

```bash
docker run -d \
  --name edge-vpn-server \
  --privileged \
  --cap-add=NET_ADMIN \
  --device=/dev/net/tun \
  -p 443:443/tcp \
  -p 443:443/udp \
  -p 8443:8443/tcp \
  -v $(pwd)/data:/opt/edge_server/data \
  -e VPN_NETWORK=192.168.100.0/24 \
  -e DNS_SERVER1=8.8.8.8 \
  -e DNS_SERVER2=8.8.4.4 \
  edge-server:latest
```

## 构建镜像

```bash
# 确保有编译好的 edge-server 可执行文件
chmod +x build-docker.sh
./build-docker.sh latest
```

## 环境变量配置

| 环境变量 | 说明 | 默认值 |
|---------|------|--------|
| `WEB_PORT` | Web管理界面端口 | 443 |
| `VPN_PORT` | VPN服务端口 | 8443 |
| `VPN_NETWORK` | VPN IP地址池 | 192.168.100.0/24 |
| `VPN_HOSTNAME` | VPN服务器域名 | edge-vpn.local |
| `DNS_SERVER1` | DNS服务器1 | 8.8.8.8 |
| `DNS_SERVER2` | DNS服务器2 | 8.8.4.4 |
| `MAX_CLIENTS` | 最大客户端数 | 100 |
| `IDLE_TIMEOUT` | 空闲超时(秒) | 3600 |
| `TZ` | 时区 | Asia/Shanghai |

## 端口说明

- **443/tcp, 443/udp**: VPN连接端口 (可通过环境变量修改)
- **8443/tcp**: Web管理界面 (可通过环境变量修改)

## 数据持久化

容器使用以下卷进行数据持久化:

- `/opt/edge_server/data`: 数据库、证书文件
- `/etc/ocserv`: ocserv配置文件

## 权限要求

容器需要以下权限才能正常运行 VPN:

- `--privileged`: 特权模式
- `--cap-add=NET_ADMIN`: 网络管理权限
- `--device=/dev/net/tun`: TUN/TAP设备访问

## 首次启动

1. 容器首次启动时会自动:
   - 生成 SSL 证书
   - 初始化数据库
   - 创建默认管理员账号: `admin / admin123`

2. 访问 Web 管理界面:
   ```
   https://<服务器IP>:443
   ```

3. 修改默认密码

## 客户端连接

使用 Cisco AnyConnect 客户端连接:

1. 服务器地址: `<服务器IP>:8443`
2. 用户名/密码: 在Web管理界面创建

## 常用命令

```bash
# 查看容器日志
docker-compose logs -f edge-vpn

# 进入容器
docker-compose exec edge-vpn bash

# 重启服务
docker-compose restart

# 查看在线用户
docker-compose exec edge-vpn occtl show users

# 备份数据
tar -czf edge-vpn-backup.tar.gz data/

# 恢复数据
tar -xzf edge-vpn-backup.tar.gz
docker-compose restart
```

## 故障排查

### 1. 容器无法启动

检查权限:
```bash
docker logs edge-vpn-server
```

### 2. VPN无法连接

检查防火墙:
```bash
# 宿主机开放端口
firewall-cmd --add-port=443/tcp --permanent
firewall-cmd --add-port=443/udp --permanent
firewall-cmd --add-port=8443/tcp --permanent
firewall-cmd --reload
```

检查IP转发:
```bash
sysctl net.ipv4.ip_forward
# 应该返回 1
```

### 3. 查看 ocserv 状态

```bash
docker-compose exec edge-vpn occtl show status
```

## 升级

```bash
# 1. 备份数据
tar -czf backup-$(date +%Y%m%d).tar.gz data/

# 2. 停止服务
docker-compose down

# 3. 拉取新镜像或重新构建
docker pull edge-server:latest
# 或
./build-docker.sh latest

# 4. 启动服务
docker-compose up -d
```

## 安全建议

1. **修改默认密码**: 首次登录后立即修改
2. **使用HTTPS**: 已默认启用
3. **限制访问**: 使用防火墙限制管理端口访问
4. **定期备份**: 定期备份 `data/` 目录
5. **监控日志**: 定期检查认证日志

## 性能优化

1. **调整最大客户端数**:
   ```yaml
   environment:
     - MAX_CLIENTS=500
   ```

2. **调整MTU**:
   ```yaml
   environment:
     - MTU=1400
   ```

3. **资源限制**:
   ```yaml
   deploy:
     resources:
       limits:
         cpus: '2'
         memory: 2G
       reservations:
         cpus: '1'
         memory: 1G
   ```
