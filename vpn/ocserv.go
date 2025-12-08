package vpn

import (
	"bufio"
	"edge_server/models"
	"fmt"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
	"time"
)

type OCServConfig struct {
	ServerCert  string
	ServerKey   string
	ListenAddr  string
	IPPool      string
	DNS         []string
	MTU         int
	MaxClients  int
	IdleTimeout int
	ConfigDir   string
}

type OCServServer struct {
	config  *OCServConfig
	cmd     *exec.Cmd
	mu      sync.Mutex
	running bool
}

func NewOCServServer(config *OCServConfig) *OCServServer {
	return &OCServServer{
		config: config,
	}
}

func (s *OCServServer) Start() error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.running {
		return fmt.Errorf("ocserv 已在运行")
	}

	if err := s.prepareConfig(); err != nil {
		return fmt.Errorf("准备配置失败: %v", err)
	}

	port, _ := strconv.Atoi(strings.TrimPrefix(s.config.ListenAddr, ":"))
	configPath := filepath.Join(s.config.ConfigDir, "ocserv.conf")

	s.cmd = exec.Command("ocserv", 
		"-f",
		"-c", configPath,
	)

	stdout, _ := s.cmd.StdoutPipe()
	stderr, _ := s.cmd.StderrPipe()

	if err := s.cmd.Start(); err != nil {
		return fmt.Errorf("启动 ocserv 失败: %v", err)
	}

	s.running = true
	log.Printf("ocserv VPN 服务已启动，端口: %d", port)

	go s.monitorLogs(stdout, "STDOUT")
	go s.monitorLogs(stderr, "STDERR")

	go func() {
		err := s.cmd.Wait()
		s.mu.Lock()
		s.running = false
		s.mu.Unlock()
		if err != nil {
			log.Printf("ocserv 进程退出: %v", err)
		}
	}()

	return nil
}

func (s *OCServServer) Stop() error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if !s.running || s.cmd == nil || s.cmd.Process == nil {
		return fmt.Errorf("ocserv 未在运行")
	}

	if err := s.cmd.Process.Signal(os.Interrupt); err != nil {
		return fmt.Errorf("停止 ocserv 失败: %v", err)
	}

	s.running = false
	return nil
}

func (s *OCServServer) prepareConfig() error {
	if s.config.ConfigDir == "" {
		s.config.ConfigDir = "/etc/ocserv"
	}

	os.MkdirAll(s.config.ConfigDir, 0755)
	os.MkdirAll("/run/ocserv", 0755)
	os.MkdirAll("/var/lib/ocserv", 0755)

	port, _ := strconv.Atoi(strings.TrimPrefix(s.config.ListenAddr, ":"))
	
	params := OCServConfigParams{
		VPNPort:     port,
		MaxClients:  s.config.MaxClients,
		IdleTimeout: s.config.IdleTimeout,
		ServerCert:  s.config.ServerCert,
		ServerKey:   s.config.ServerKey,
		IPPool:      s.config.IPPool,
		DNS:         s.config.DNS,
	}

	configPath := filepath.Join(s.config.ConfigDir, "ocserv.conf")
	if err := GenerateOCServConfig(configPath, params); err != nil {
		return err
	}

	passwdPath := "/run/ocserv/ocpasswd"
	if err := s.generatePasswordFile(passwdPath); err != nil {
		return fmt.Errorf("生成密码文件失败: %v", err)
	}

	return nil
}

func (s *OCServServer) generatePasswordFile(path string) error {
	rows, err := models.DB.Query("SELECT username, password FROM users WHERE enabled=1")
	if err != nil {
		return err
	}
	defer rows.Close()

	file, err := os.Create(path)
	if err != nil {
		return err
	}
	defer file.Close()

	for rows.Next() {
		var username, password string
		if err := rows.Scan(&username, &password); err != nil {
			continue
		}
		fmt.Fprintf(file, "%s:%s\n", username, password)
	}

	return nil
}

func (s *OCServServer) monitorLogs(pipe *os.File, source string) {
	if pipe == nil {
		return
	}
	defer pipe.Close()

	scanner := bufio.NewScanner(pipe)
	for scanner.Scan() {
		line := scanner.Text()
		log.Printf("[ocserv-%s] %s", source, line)
		
		s.parseLogLine(line)
	}
}

func (s *OCServServer) parseLogLine(line string) {
	if strings.Contains(line, "user") && strings.Contains(line, "connected") {
		
	} else if strings.Contains(line, "disconnected") {
		
	}
}

func generateMAC() string {
	return fmt.Sprintf("00:50:56:%02x:%02x:%02x", 
		time.Now().Unix()%256, 
		time.Now().UnixNano()%256, 
		time.Now().Nanosecond()%256)
}