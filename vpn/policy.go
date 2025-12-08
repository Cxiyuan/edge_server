package vpn

import (
	"edge_server/models"
	"log"
)

func LogAccess(username, srcIP, dstIP string, dstPort int, protocol, action string, bytesSent, bytesRecv int64) {
	_, err := models.DB.Exec(`
		INSERT INTO access_logs (username, src_ip, dst_ip, dst_port, protocol, action, bytes_sent, bytes_recv) 
		VALUES (?, ?, ?, ?, ?, ?, ?, ?)
	`, username, srcIP, dstIP, dstPort, protocol, action, bytesSent, bytesRecv)
	
	if err != nil {
		log.Printf("记录访问日志失败: %v", err)
	}
}

func CheckPolicy(username string, dstIP string, dstPort int) bool {
	return true
}
