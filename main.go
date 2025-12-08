package main

import (
	"bufio"
	"embed"
	"edge_server/handlers"
	"edge_server/middleware"
	"edge_server/models"
	"edge_server/vpn"
	"io/fs"
	"log"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
	_ "github.com/mattn/go-sqlite3"
)

//go:embed static/*
var staticFiles embed.FS

type Config struct {
	WebPort      string
	VPNPort      string
	DBPath       string
	ServerCert   string
	ServerKey    string
	IPPool       string
	DNS          []string
	MTU          int
	MaxClients   int
	IdleTimeout  int
}

func loadConfig(configPath string) (*Config, error) {
	config := &Config{
		WebPort:     "8080",
		VPNPort:     "443",
		DBPath:      "server.db",
		ServerCert:  "server.crt",
		ServerKey:   "server.key",
		IPPool:      "192.168.100.0/24",
		DNS:         []string{"8.8.8.8", "8.8.4.4"},
		MTU:         1400,
		MaxClients:  100,
		IdleTimeout: 3600,
	}

	file, err := os.Open(configPath)
	if err != nil {
		log.Printf("无法打开配置文件 %s，使用默认配置: %v", configPath, err)
		return config, nil
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	currentSection := ""
	
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		
		if strings.HasPrefix(line, "[") && strings.HasSuffix(line, "]") {
			currentSection = strings.Trim(line, "[]")
			continue
		}
		
		parts := strings.SplitN(line, "=", 2)
		if len(parts) != 2 {
			continue
		}
		
		key := strings.TrimSpace(parts[0])
		value := strings.TrimSpace(parts[1])
		
		switch currentSection {
		case "server":
			switch key {
			case "web_port":
				config.WebPort = value
			case "vpn_port":
				config.VPNPort = value
			case "db_path":
				config.DBPath = value
			}
		case "ssl":
			switch key {
			case "server_cert":
				config.ServerCert = value
			case "server_key":
				config.ServerKey = value
			}
		case "network":
			switch key {
			case "ip_pool":
				config.IPPool = value
			case "dns1":
				if len(config.DNS) > 0 {
					config.DNS[0] = value
				}
			case "dns2":
				if len(config.DNS) > 1 {
					config.DNS[1] = value
				}
			case "mtu":
				if mtu, err := strconv.Atoi(value); err == nil {
					config.MTU = mtu
				}
			}
		case "system":
			switch key {
			case "max_clients":
				if max, err := strconv.Atoi(value); err == nil {
					config.MaxClients = max
				}
			case "idle_timeout":
				if timeout, err := strconv.Atoi(value); err == nil {
					config.IdleTimeout = timeout
				}
			}
		}
	}

	if err := scanner.Err(); err != nil {
		log.Printf("读取配置文件出错: %v", err)
	}

	return config, nil
}

func loadConfigFromDB(config *Config) {
	dbIPPool := models.GetConfig("default_ip_pool", config.IPPool)
	if dbIPPool != "" {
		config.IPPool = dbIPPool
	}

	dbDNS1 := models.GetConfig("default_dns1", config.DNS[0])
	dbDNS2 := models.GetConfig("default_dns2", config.DNS[1])
	if dbDNS1 != "" && dbDNS2 != "" {
		config.DNS = []string{dbDNS1, dbDNS2}
	}

	dbMTU := models.GetConfigInt("default_mtu", config.MTU)
	if dbMTU > 0 {
		config.MTU = dbMTU
	}

	dbMaxClients := models.GetConfigInt("max_clients", config.MaxClients)
	if dbMaxClients > 0 {
		config.MaxClients = dbMaxClients
	}

	dbIdleTimeout := models.GetConfigInt("idle_timeout", config.IdleTimeout)
	if dbIdleTimeout > 0 {
		config.IdleTimeout = dbIdleTimeout
	}

	log.Printf("已从数据库加载配置: IP池=%s, DNS=%v, MTU=%d, 最大客户端=%d, 超时=%d", 
		config.IPPool, config.DNS, config.MTU, config.MaxClients, config.IdleTimeout)
}

func main() {
	execDir, err := os.Executable()
	if err != nil {
		log.Fatal("获取执行目录失败:", err)
	}
	execDir = filepath.Dir(execDir)

	configPath := filepath.Join(execDir, "server.conf")
	config, err := loadConfig(configPath)
	if err != nil {
		log.Fatal("加载配置失败:", err)
	}

	dbPath := filepath.Join(execDir, config.DBPath)
	if err := models.InitDB(dbPath); err != nil {
		log.Fatal("初始化数据库失败:", err)
	}
	defer models.DB.Close()

	log.Println("数据库初始化成功")

	loadConfigFromDB(config)

	middleware.CleanupExpiredSessions()
	vpn.StartSessionCleanup(config.IdleTimeout)
	vpn.StartOCCtlMonitor()

	go func() {
		vpnConfig := &vpn.OCServConfig{
			ServerCert:  filepath.Join(execDir, config.ServerCert),
			ServerKey:   filepath.Join(execDir, config.ServerKey),
			ListenAddr:  ":" + config.VPNPort,
			IPPool:      config.IPPool,
			DNS:         config.DNS,
			MTU:         config.MTU,
			MaxClients:  config.MaxClients,
			IdleTimeout: config.IdleTimeout,
			ConfigDir:   filepath.Join(execDir, "ocserv_config"),
		}
		vpnServer := vpn.NewOCServServer(vpnConfig)
		if err := vpnServer.Start(); err != nil {
			log.Printf("VPN服务启动失败: %v", err)
		}
	}()

	router := gin.Default()

	staticFS, _ := fs.Sub(staticFiles, "static")
	router.GET("/", func(c *gin.Context) {
		data, err := fs.ReadFile(staticFS, "index.html")
		if err != nil {
			c.String(404, "Not Found")
			return
		}
		c.Data(200, "text/html; charset=utf-8", data)
	})

	router.GET("/assets/*filepath", func(c *gin.Context) {
		path := c.Param("filepath")
		data, err := fs.ReadFile(staticFS, "assets"+path)
		if err != nil {
			c.String(404, "Not Found")
			return
		}
		
		contentType := "application/octet-stream"
		if filepath.Ext(path) == ".js" {
			contentType = "application/javascript"
		} else if filepath.Ext(path) == ".css" {
			contentType = "text/css"
		}
		
		c.Data(200, contentType, data)
	})

	router.POST("/api/login", middleware.Login)

	api := router.Group("/api")
	api.Use(middleware.AuthRequired())
	{
		api.GET("/groups", handlers.GetUserGroups)
		api.POST("/groups", handlers.CreateUserGroup)
		api.PUT("/groups/:id", handlers.UpdateUserGroup)
		api.DELETE("/groups/:id", handlers.DeleteUserGroup)

		api.GET("/users", handlers.GetUsers)
		api.POST("/users", handlers.CreateUser)
		api.PUT("/users/:id", handlers.UpdateUser)
		api.DELETE("/users/:id", handlers.DeleteUser)

		api.GET("/online", handlers.GetOnlineUsers)
		api.POST("/online/:id/disconnect", handlers.DisconnectUser)

		api.GET("/logs/auth", handlers.GetAuthLogs)
		api.GET("/logs/access", handlers.GetAccessLogs)

		api.GET("/stats", handlers.GetSystemStats)

		api.GET("/config", handlers.GetSystemConfig)
		api.PUT("/config", handlers.UpdateSystemConfig)
		
		api.POST("/change-password", handlers.ChangePassword)
	}

	log.Printf("Web管理界面启动在端口 %s", config.WebPort)
	log.Printf("VPN服务端口 %s", config.VPNPort)
	
	certPath := filepath.Join(execDir, config.ServerCert)
	keyPath := filepath.Join(execDir, config.ServerKey)
	
	if _, err := os.Stat(certPath); os.IsNotExist(err) {
		log.Printf("警告: SSL证书文件 %s 不存在，使用HTTP模式", certPath)
		if err := router.Run(":" + config.WebPort); err != nil {
			log.Fatal("Web服务启动失败:", err)
		}
	} else {
		log.Printf("使用HTTPS模式，访问 https://localhost:%s", config.WebPort)
		if err := router.RunTLS(":" + config.WebPort, certPath, keyPath); err != nil {
			log.Fatal("Web服务启动失败:", err)
		}
	}
}