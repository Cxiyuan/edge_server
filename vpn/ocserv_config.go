package vpn

import (
	"fmt"
	"os"
	"path/filepath"
	"text/template"
)

const ocservConfigTemplate = `
# ocserv 配置文件 - 由 Edge Server 自动生成

auth = "plain[passwd=/run/ocserv/ocpasswd]"

tcp-port = {{.VPNPort}}
udp-port = {{.VPNPort}}

run-as-user = nobody
run-as-group = daemon

socket-file = /run/ocserv/ocserv-socket
chroot-dir = /var/lib/ocserv

max-clients = {{.MaxClients}}
max-same-clients = 2

server-cert = {{.ServerCert}}
server-key = {{.ServerKey}}

ca-cert = {{.ServerCert}}

isolate-workers = true

keepalive = 32400
dpd = 90
mobile-dpd = 1800

switch-to-tcp-timeout = 25

try-mtu-discovery = true
cert-user-oid = 0.9.2342.19200300.100.1.1

compression = true
no-compress-limit = 256

tls-priorities = "NORMAL:%SERVER_PRECEDENCE:%COMPAT:-RSA:-VERS-SSL3.0:-ARCFOUR-128"

auth-timeout = 240
idle-timeout = {{.IdleTimeout}}
mobile-idle-timeout = 2400

min-reauth-time = 300
max-ban-score = 80
ban-reset-time = 1200

cookie-timeout = 300
deny-roaming = false
rekey-time = 172800
rekey-method = ssl

use-occtl = true
pid-file = /run/ocserv/ocserv.pid

device = vpns
predictable-ips = true

default-domain = edge-vpn.local

ipv4-network = {{.IPPool}}

tunnel-all-dns = true
{{range .DNS}}
dns = {{.}}
{{end}}

ping-leases = false

cisco-client-compat = true
dtls-legacy = true

user-profile = /etc/ocserv/profile.xml
`

const userProfileTemplate = `<?xml version="1.0" encoding="UTF-8"?>
<AnyConnectProfile xmlns="http://schemas.xmlsoap.org/encoding/">
<ServerList>
    <HostEntry>
        <HostName>Edge VPN Server</HostName>
        <HostAddress>%{HOSTNAME}</HostAddress>
    </HostEntry>
</ServerList>
</AnyConnectProfile>
`

type OCServConfigParams struct {
	VPNPort      int
	MaxClients   int
	IdleTimeout  int
	ServerCert   string
	ServerKey    string
	IPPool       string
	DNS          []string
}

func GenerateOCServConfig(configPath string, params OCServConfigParams) error {
	tmpl, err := template.New("ocserv").Parse(ocservConfigTemplate)
	if err != nil {
		return fmt.Errorf("解析配置模板失败: %v", err)
	}

	file, err := os.Create(configPath)
	if err != nil {
		return fmt.Errorf("创建配置文件失败: %v", err)
	}
	defer file.Close()

	if err := tmpl.Execute(file, params); err != nil {
		return fmt.Errorf("生成配置文件失败: %v", err)
	}

	profilePath := filepath.Join(filepath.Dir(configPath), "profile.xml")
	if err := os.WriteFile(profilePath, []byte(userProfileTemplate), 0644); err != nil {
		return fmt.Errorf("创建用户配置文件失败: %v", err)
	}

	return nil
}
