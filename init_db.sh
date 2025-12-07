#!/bin/bash

set -e

echo "正在创建数据库..."

DB_FILE="server.db"

if [ -f "$DB_FILE" ]; then
    echo "数据库文件已存在，跳过创建"
    exit 0
fi

sqlite3 "$DB_FILE" <<EOF
CREATE TABLE IF NOT EXISTS user_groups (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    name TEXT NOT NULL UNIQUE,
    description TEXT,
    routes TEXT,
    policies TEXT,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE IF NOT EXISTS users (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    username TEXT NOT NULL UNIQUE,
    password TEXT NOT NULL,
    full_name TEXT,
    email TEXT,
    group_id INTEGER,
    custom_routes TEXT,
    custom_policies TEXT,
    enabled INTEGER DEFAULT 1,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (group_id) REFERENCES user_groups(id)
);

CREATE TABLE IF NOT EXISTS online_users (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    username TEXT NOT NULL,
    group_name TEXT,
    mac TEXT,
    virtual_ip TEXT,
    remote_ip TEXT,
    protocol TEXT,
    virtual_dev TEXT,
    mtu INTEGER,
    upload_speed INTEGER DEFAULT 0,
    download_speed INTEGER DEFAULT 0,
    total_upload INTEGER DEFAULT 0,
    total_download INTEGER DEFAULT 0,
    connected_at DATETIME DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE IF NOT EXISTS auth_logs (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    username TEXT NOT NULL,
    remote_ip TEXT,
    action TEXT,
    success INTEGER,
    message TEXT,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE IF NOT EXISTS access_logs (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    username TEXT NOT NULL,
    src_ip TEXT,
    dst_ip TEXT,
    dst_port INTEGER,
    protocol TEXT,
    action TEXT,
    bytes_sent INTEGER DEFAULT 0,
    bytes_recv INTEGER DEFAULT 0,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX IF NOT EXISTS idx_users_username ON users(username);
CREATE INDEX IF NOT EXISTS idx_online_users_username ON online_users(username);
CREATE INDEX IF NOT EXISTS idx_auth_logs_username ON auth_logs(username);
CREATE INDEX IF NOT EXISTS idx_auth_logs_created ON auth_logs(created_at);
CREATE INDEX IF NOT EXISTS idx_access_logs_username ON access_logs(username);
CREATE INDEX IF NOT EXISTS idx_access_logs_created ON access_logs(created_at);

INSERT INTO user_groups (name, description, routes, policies) 
VALUES ('默认组', '系统默认用户组', '192.168.10.0/24,10.0.0.0/8', '{"allow_internet":true,"allow_lan":true}');

INSERT INTO users (username, password, full_name, email, group_id, enabled) 
VALUES ('admin', '\$2a\$10\$N9qo8uLOickgx2ZMRZoMye0J8YAR1WjxKRkzBCG.iHXE7BQOBZCVW', '管理员', 'admin@example.com', 1, 1);

EOF

echo "数据库创建完成！"
echo "默认管理员账号: admin"
echo "默认管理员密码: admin123"
