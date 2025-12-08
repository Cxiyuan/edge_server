package models

import (
	"database/sql"
	"time"
)

type UserGroup struct {
	ID          int       `json:"id"`
	Name        string    `json:"name"`
	Description string    `json:"description"`
	IPPool      string    `json:"ip_pool"`
	Routes      string    `json:"routes"`
	Policies    string    `json:"policies"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

type User struct {
	ID          int       `json:"id"`
	Username    string    `json:"username"`
	Password    string    `json:"-"`
	FullName    string    `json:"full_name"`
	Email       string    `json:"email"`
	GroupID     int       `json:"group_id"`
	GroupName   string    `json:"group_name"`
	CustomRoutes string   `json:"custom_routes"`
	CustomPolicies string `json:"custom_policies"`
	Enabled     bool      `json:"enabled"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

type OnlineUser struct {
	ID           int       `json:"id"`
	Username     string    `json:"username"`
	GroupName    string    `json:"group_name"`
	MAC          string    `json:"mac"`
	VirtualIP    string    `json:"virtual_ip"`
	RemoteIP     string    `json:"remote_ip"`
	Protocol     string    `json:"protocol"`
	VirtualDev   string    `json:"virtual_dev"`
	MTU          int       `json:"mtu"`
	UploadSpeed  int64     `json:"upload_speed"`
	DownloadSpeed int64    `json:"download_speed"`
	TotalUpload  int64     `json:"total_upload"`
	TotalDownload int64    `json:"total_download"`
	ConnectedAt  time.Time `json:"connected_at"`
}

type AuthLog struct {
	ID         int       `json:"id"`
	Username   string    `json:"username"`
	RemoteIP   string    `json:"remote_ip"`
	Action     string    `json:"action"`
	Success    bool      `json:"success"`
	Message    string    `json:"message"`
	CreatedAt  time.Time `json:"created_at"`
}

type AccessLog struct {
	ID          int       `json:"id"`
	Username    string    `json:"username"`
	SrcIP       string    `json:"src_ip"`
	DstIP       string    `json:"dst_ip"`
	DstPort     int       `json:"dst_port"`
	Protocol    string    `json:"protocol"`
	Action      string    `json:"action"`
	BytesSent   int64     `json:"bytes_sent"`
	BytesRecv   int64     `json:"bytes_recv"`
	CreatedAt   time.Time `json:"created_at"`
}

type SystemStats struct {
	CPUUsage      float64 `json:"cpu_usage"`
	MemoryUsage   float64 `json:"memory_usage"`
	DiskUsage     float64 `json:"disk_usage"`
	NetworkConnections int `json:"network_connections"`
	OnlineUsers   int     `json:"online_users"`
	Uptime        int64   `json:"uptime"`
}

var DB *sql.DB

func InitDB(dbPath string) error {
	var err error
	DB, err = sql.Open("sqlite3", dbPath)
	if err != nil {
		return err
	}

	if err := DB.Ping(); err != nil {
		return err
	}

	if err := createTables(); err != nil {
		return err
	}

	if err := initDefaultData(); err != nil {
		return err
	}

	return nil
}

func createTables() error {
	schema := `
	CREATE TABLE IF NOT EXISTS user_groups (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		name TEXT NOT NULL UNIQUE,
		description TEXT,
		ip_pool TEXT,
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

	CREATE TABLE IF NOT EXISTS ip_allocations (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		group_id INTEGER NOT NULL,
		ip_address TEXT NOT NULL,
		username TEXT,
		allocated_at DATETIME DEFAULT CURRENT_TIMESTAMP,
		FOREIGN KEY (group_id) REFERENCES user_groups(id),
		UNIQUE(group_id, ip_address)
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

	CREATE TABLE IF NOT EXISTS system_config (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		config_key TEXT NOT NULL UNIQUE,
		config_value TEXT,
		description TEXT,
		updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
	);

	CREATE INDEX IF NOT EXISTS idx_users_username ON users(username);
	CREATE INDEX IF NOT EXISTS idx_ip_allocations_username ON ip_allocations(username);
	CREATE INDEX IF NOT EXISTS idx_online_users_username ON online_users(username);
	CREATE INDEX IF NOT EXISTS idx_auth_logs_username ON auth_logs(username);
	CREATE INDEX IF NOT EXISTS idx_auth_logs_created ON auth_logs(created_at);
	CREATE INDEX IF NOT EXISTS idx_access_logs_username ON access_logs(username);
	CREATE INDEX IF NOT EXISTS idx_access_logs_created ON access_logs(created_at);
	CREATE INDEX IF NOT EXISTS idx_system_config_key ON system_config(config_key);
	`

	_, err := DB.Exec(schema)
	return err
}

func initDefaultData() error {
	var count int
	err := DB.QueryRow("SELECT COUNT(*) FROM user_groups").Scan(&count)
	if err != nil {
		return err
	}

	if count == 0 {
		_, err = DB.Exec(`
			INSERT INTO user_groups (name, description, ip_pool, routes, policies) 
			VALUES ('默认组', '系统默认用户组', '192.168.100.0/24', '192.168.10.0/24,10.0.0.0/8', '{"allow_internet":true,"allow_lan":true}')
		`)
		if err != nil {
			return err
		}
	}

	err = DB.QueryRow("SELECT COUNT(*) FROM users").Scan(&count)
	if err != nil {
		return err
	}

	if count == 0 {
		_, err = DB.Exec(`
			INSERT INTO users (username, password, full_name, email, group_id, enabled) 
			VALUES ('admin', '$2a$10$N9qo8uLOickgx2ZMRZoMye0J8YAR1WjxKRkzBCG.iHXE7BQOBZCVW', '管理员', 'admin@example.com', 1, 1)
		`)
		if err != nil {
			return err
		}
	}

	err = DB.QueryRow("SELECT COUNT(*) FROM system_config").Scan(&count)
	if err != nil {
		return err
	}

	if count == 0 {
		configs := []struct {
			Key   string
			Value string
			Desc  string
		}{
			{"default_ip_pool", "192.168.100.0/24", "默认VPN IP地址池"},
			{"default_dns1", "8.8.8.8", "默认DNS服务器1"},
			{"default_dns2", "8.8.4.4", "默认DNS服务器2"},
			{"default_mtu", "1400", "默认MTU值"},
			{"max_clients", "100", "最大客户端连接数"},
			{"idle_timeout", "3600", "空闲超时时间(秒)"},
			{"vpn_domain", "edge-vpn.local", "VPN域名"},
			{"vpn_device", "vpns", "VPN虚拟网卡名称"},
		}

		for _, cfg := range configs {
			_, err = DB.Exec(`
				INSERT INTO system_config (config_key, config_value, description) 
				VALUES (?, ?, ?)
			`, cfg.Key, cfg.Value, cfg.Desc)
			if err != nil {
				return err
			}
		}
	}

	return nil
}