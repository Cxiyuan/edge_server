package vpn

import (
	"crypto/tls"
	"edge_server/models"
	"fmt"
	"log"
	"net/http"
	"strings"
	"time"

	"golang.org/x/crypto/bcrypt"
)

type OCServConfig struct {
	ServerCert string
	ServerKey  string
	ListenAddr string
	IPPool     string
	DNS        []string
	MTU        int
}

type OCServServer struct {
	config *OCServConfig
	server *http.Server
}

func NewOCServServer(config *OCServConfig) *OCServServer {
	return &OCServServer{
		config: config,
	}
}

func (s *OCServServer) Start() error {
	mux := http.NewServeMux()
	mux.HandleFunc("/", s.handleAuth)
	mux.HandleFunc("/auth", s.handleAuth)
	mux.HandleFunc("/connect", s.handleConnect)

	cert, err := tls.LoadX509KeyPair(s.config.ServerCert, s.config.ServerKey)
	if err != nil {
		return fmt.Errorf("加载证书失败: %v", err)
	}

	s.server = &http.Server{
		Addr:    s.config.ListenAddr,
		Handler: mux,
		TLSConfig: &tls.Config{
			Certificates: []tls.Certificate{cert},
			MinVersion:   tls.VersionTLS12,
		},
	}

	log.Printf("OpenConnect VPN 服务启动在 %s", s.config.ListenAddr)
	return s.server.ListenAndServeTLS("", "")
}

func (s *OCServServer) handleAuth(w http.ResponseWriter, r *http.Request) {
	username := r.FormValue("username")
	password := r.FormValue("password")

	remoteIP := r.RemoteAddr
	if idx := strings.LastIndex(remoteIP, ":"); idx != -1 {
		remoteIP = remoteIP[:idx]
	}

	if username == "" || password == "" {
		s.logAuth(username, remoteIP, "login", false, "用户名或密码为空")
		http.Error(w, "认证失败", http.StatusUnauthorized)
		return
	}

	var storedPassword string
	var enabled bool
	err := models.DB.QueryRow("SELECT password, enabled FROM users WHERE username=?", username).Scan(&storedPassword, &enabled)
	if err != nil {
		s.logAuth(username, remoteIP, "login", false, "用户不存在")
		http.Error(w, "认证失败", http.StatusUnauthorized)
		return
	}

	if !enabled {
		s.logAuth(username, remoteIP, "login", false, "用户已禁用")
		http.Error(w, "用户已禁用", http.StatusUnauthorized)
		return
	}

	if err := bcrypt.CompareHashAndPassword([]byte(storedPassword), []byte(password)); err != nil {
		s.logAuth(username, remoteIP, "login", false, "密码错误")
		http.Error(w, "认证失败", http.StatusUnauthorized)
		return
	}

	s.logAuth(username, remoteIP, "login", true, "认证成功")

	xmlResp := `<?xml version="1.0" encoding="UTF-8"?>
<auth id="success">
<message>认证成功</message>
</auth>`
	w.Header().Set("Content-Type", "text/xml")
	w.Write([]byte(xmlResp))
}

func (s *OCServServer) handleConnect(w http.ResponseWriter, r *http.Request) {
	username := r.Header.Get("X-CSTP-Username")
	if username == "" {
		http.Error(w, "未认证", http.StatusUnauthorized)
		return
	}

	remoteIP := r.RemoteAddr
	if idx := strings.LastIndex(remoteIP, ":"); idx != -1 {
		remoteIP = remoteIP[:idx]
	}

	virtualIP := s.allocateIP()
	mac := generateMAC()

	var groupName string
	models.DB.QueryRow(`
		SELECT g.name FROM users u 
		LEFT JOIN user_groups g ON u.group_id = g.id 
		WHERE u.username=?
	`, username).Scan(&groupName)

	_, err := models.DB.Exec(`
		INSERT INTO online_users (username, group_name, mac, virtual_ip, remote_ip, protocol, virtual_dev, mtu, connected_at) 
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)
	`, username, groupName, mac, virtualIP, remoteIP, "DTLS", "vpns0", s.config.MTU, time.Now())

	if err != nil {
		log.Printf("记录在线用户失败: %v", err)
	}

	configXML := fmt.Sprintf(`<?xml version="1.0" encoding="UTF-8"?>
<config>
<vpn-tunnel-type>full</vpn-tunnel-type>
<session-token>%s</session-token>
<ipv4>%s</ipv4>
<netmask>255.255.255.0</netmask>
<dns>%s</dns>
<mtu>%d</mtu>
</config>`, username, virtualIP, strings.Join(s.config.DNS, ","), s.config.MTU)

	w.Header().Set("Content-Type", "text/xml")
	w.Write([]byte(configXML))

	s.logAuth(username, remoteIP, "connect", true, fmt.Sprintf("分配IP: %s", virtualIP))
}

func (s *OCServServer) allocateIP() string {
	return "192.168.100.10"
}

func generateMAC() string {
	return fmt.Sprintf("00:50:56:%02x:%02x:%02x", 
		time.Now().Unix()%256, 
		time.Now().UnixNano()%256, 
		time.Now().Nanosecond()%256)
}

func (s *OCServServer) logAuth(username, remoteIP, action string, success bool, message string) {
	_, err := models.DB.Exec(`
		INSERT INTO auth_logs (username, remote_ip, action, success, message) 
		VALUES (?, ?, ?, ?, ?)
	`, username, remoteIP, action, success, message)
	if err != nil {
		log.Printf("记录认证日志失败: %v", err)
	}
}