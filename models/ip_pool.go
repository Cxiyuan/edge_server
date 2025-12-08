package models

import (
	"database/sql"
	"fmt"
	"net"
	"strings"
)

func AllocateIP(username string, groupID int) (string, error) {
	var ipPool string
	err := DB.QueryRow("SELECT ip_pool FROM user_groups WHERE id=?", groupID).Scan(&ipPool)
	if err != nil {
		return "", fmt.Errorf("获取用户组IP池失败: %v", err)
	}

	if ipPool == "" {
		return "", fmt.Errorf("用户组未配置IP地址池")
	}

	_, ipNet, err := net.ParseCIDR(ipPool)
	if err != nil {
		return "", fmt.Errorf("IP池格式错误: %v", err)
	}

	var existingIP string
	err = DB.QueryRow(`
		SELECT ip_address FROM ip_allocations 
		WHERE username=? AND group_id=?
	`, username, groupID).Scan(&existingIP)
	if err == nil {
		return existingIP, nil
	}

	ip := ipNet.IP
	for ip := ip.Mask(ipNet.Mask); ipNet.Contains(ip); incrementIP(ip) {
		ipStr := ip.String()
		
		if isReservedIP(ip, ipNet) {
			continue
		}

		var count int
		err = DB.QueryRow(`
			SELECT COUNT(*) FROM ip_allocations 
			WHERE group_id=? AND ip_address=?
		`, groupID, ipStr).Scan(&count)
		if err != nil {
			continue
		}

		if count == 0 {
			_, err = DB.Exec(`
				INSERT INTO ip_allocations (group_id, ip_address, username) 
				VALUES (?, ?, ?)
			`, groupID, ipStr, username)
			if err != nil {
				return "", fmt.Errorf("分配IP失败: %v", err)
			}
			return ipStr, nil
		}
	}

	return "", fmt.Errorf("IP地址池已耗尽")
}

func ReleaseIP(username string, groupID int) error {
	_, err := DB.Exec(`
		DELETE FROM ip_allocations 
		WHERE username=? AND group_id=?
	`, username, groupID)
	return err
}

func GetAllocatedIPs(groupID int) ([]string, error) {
	rows, err := DB.Query(`
		SELECT ip_address, username, allocated_at 
		FROM ip_allocations 
		WHERE group_id=?
		ORDER BY ip_address
	`, groupID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var ips []string
	for rows.Next() {
		var ip, username string
		var allocatedAt string
		if err := rows.Scan(&ip, &username, &allocatedAt); err != nil {
			continue
		}
		ips = append(ips, fmt.Sprintf("%s (%s)", ip, username))
	}

	return ips, nil
}

func incrementIP(ip net.IP) {
	for j := len(ip) - 1; j >= 0; j-- {
		ip[j]++
		if ip[j] > 0 {
			break
		}
	}
}

func isReservedIP(ip net.IP, ipNet *net.IPNet) bool {
	ipv4 := ip.To4()
	if ipv4 == nil {
		return false
	}

	networkIP := ipNet.IP.To4()
	broadcastIP := make(net.IP, len(networkIP))
	copy(broadcastIP, networkIP)
	for i := range broadcastIP {
		broadcastIP[i] |= ^ipNet.Mask[i]
	}

	if ipv4.Equal(networkIP) || ipv4.Equal(broadcastIP) {
		return true
	}

	lastOctet := ipv4[3]
	if lastOctet == 0 || lastOctet == 255 {
		return true
	}

	return false
}
