package main

import (
	"embed"
	"edge_server/handlers"
	"edge_server/models"
	"edge_server/vpn"
	"io/fs"
	"log"
	"os"
	"path/filepath"

	"github.com/gin-gonic/gin"
	_ "github.com/mattn/go-sqlite3"
)

//go:embed static/*
var staticFiles embed.FS

type Config struct {
	WebPort    string
	VPNPort    string
	DBPath     string
	ServerCert string
	ServerKey  string
	IPPool     string
	DNS        []string
	MTU        int
}

func loadConfig(configPath string) (*Config, error) {
	config := &Config{
		WebPort:    "8080",
		VPNPort:    "443",
		DBPath:     "server.db",
		ServerCert: "server.crt",
		ServerKey:  "server.key",
		IPPool:     "192.168.100.0/24",
		DNS:        []string{"8.8.8.8", "8.8.4.4"},
		MTU:        1400,
	}

	return config, nil
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

	go func() {
		vpnConfig := &vpn.OCServConfig{
			ServerCert: filepath.Join(execDir, config.ServerCert),
			ServerKey:  filepath.Join(execDir, config.ServerKey),
			ListenAddr: ":" + config.VPNPort,
			IPPool:     config.IPPool,
			DNS:        config.DNS,
			MTU:        config.MTU,
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

	api := router.Group("/api")
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

		api.GET("/logs/auth", handlers.GetAuthLogs)
		api.GET("/logs/access", handlers.GetAccessLogs)

		api.GET("/stats", handlers.GetSystemStats)
	}

	log.Printf("Web管理界面启动在端口 %s", config.WebPort)
	log.Printf("访问 http://localhost:%s 进入管理界面", config.WebPort)
	
	if err := router.Run(":" + config.WebPort); err != nil {
		log.Fatal("Web服务启动失败:", err)
	}
}