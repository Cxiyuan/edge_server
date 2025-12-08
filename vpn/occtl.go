package vpn

import (
	"bufio"
	"edge_server/models"
	"log"
	"os/exec"
	"strconv"
	"strings"
	"time"
)

func StartOCCtlMonitor() {
	ticker := time.NewTicker(10 * time.Second)
	go func() {
		for range ticker.C {
			updateOnlineUsersFromOCCtl()
		}
	}()
}

func updateOnlineUsersFromOCCtl() {
	cmd := exec.Command("occtl", "show", "users")
	output, err := cmd.Output()
	if err != nil {
		return
	}

	scanner := bufio.NewScanner(strings.NewReader(string(output)))
	activeUsers := make(map[string]bool)

	for scanner.Scan() {
		line := scanner.Text()
		fields := strings.Fields(line)
		
		if len(fields) < 5 {
			continue
		}

		username := fields[0]
		if username == "id" || username == "" {
			continue
		}

		activeUsers[username] = true

		var exists int
		models.DB.QueryRow("SELECT COUNT(*) FROM online_users WHERE username=?", username).Scan(&exists)
		
		if exists == 0 {
			virtualIP := ""
			remoteIP := ""
			if len(fields) >= 3 {
				virtualIP = fields[2]
			}
			if len(fields) >= 4 {
				remoteIP = fields[3]
			}

			var groupID int
			var groupName string
			models.DB.QueryRow(`
				SELECT u.group_id, COALESCE(g.name, '') FROM users u 
				LEFT JOIN user_groups g ON u.group_id = g.id 
				WHERE u.username=?
			`, username).Scan(&groupID, &groupName)

			models.DB.Exec(`
				INSERT INTO online_users 
				(username, group_name, virtual_ip, remote_ip, protocol, connected_at) 
				VALUES (?, ?, ?, ?, 'DTLS', CURRENT_TIMESTAMP)
			`, username, groupName, virtualIP, remoteIP)

			log.Printf("检测到新连接: %s (%s -> %s)", username, remoteIP, virtualIP)
		}
	}

	rows, _ := models.DB.Query("SELECT username, id FROM online_users")
	defer rows.Close()

	for rows.Next() {
		var username string
		var id int
		rows.Scan(&username, &id)

		if !activeUsers[username] {
			models.DB.Exec("DELETE FROM online_users WHERE id=?", id)
			
			var groupID int
			models.DB.QueryRow("SELECT group_id FROM users WHERE username=?", username).Scan(&groupID)
			models.ReleaseIP(username, groupID)
			
			log.Printf("用户已断开: %s", username)
		}
	}
}

func GetOCServStatus() map[string]interface{} {
	status := make(map[string]interface{})
	
	cmd := exec.Command("occtl", "show", "status")
	output, err := cmd.Output()
	if err != nil {
		status["running"] = false
		status["error"] = err.Error()
		return status
	}

	status["running"] = true
	lines := strings.Split(string(output), "\n")
	
	for _, line := range lines {
		if strings.Contains(line, "Active users:") {
			parts := strings.Split(line, ":")
			if len(parts) == 2 {
				count, _ := strconv.Atoi(strings.TrimSpace(parts[1]))
				status["active_users"] = count
			}
		}
	}

	return status
}

func DisconnectUserByOCCtl(username string) error {
	cmd := exec.Command("occtl", "disconnect", "user", username)
	return cmd.Run()
}
