package handlers

import (
	"bufio"
	"edge_server/models"
	"edge_server/vpn"
	"net/http"
	"os"
	"strconv"
	"strings"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"
	"golang.org/x/crypto/bcrypt"
)

func GetUserGroups(c *gin.Context) {
	rows, err := models.DB.Query(`
		SELECT id, name, description, ip_pool, routes, policies, created_at, updated_at 
		FROM user_groups
		ORDER BY created_at DESC
	`)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	defer rows.Close()

	var groups []models.UserGroup
	for rows.Next() {
		var g models.UserGroup
		if err := rows.Scan(&g.ID, &g.Name, &g.Description, &g.IPPool, &g.Routes, &g.Policies, &g.CreatedAt, &g.UpdatedAt); err != nil {
			continue
		}
		groups = append(groups, g)
	}

	c.JSON(http.StatusOK, gin.H{"data": groups})
}

func CreateUserGroup(c *gin.Context) {
	var group models.UserGroup
	if err := c.ShouldBindJSON(&group); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	result, err := models.DB.Exec(`
		INSERT INTO user_groups (name, description, ip_pool, routes, policies) 
		VALUES (?, ?, ?, ?, ?)
	`, group.Name, group.Description, group.IPPool, group.Routes, group.Policies)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	id, _ := result.LastInsertId()
	group.ID = int(id)
	c.JSON(http.StatusOK, gin.H{"data": group})
}

func UpdateUserGroup(c *gin.Context) {
	id := c.Param("id")
	var group models.UserGroup
	if err := c.ShouldBindJSON(&group); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	_, err := models.DB.Exec(`
		UPDATE user_groups 
		SET name=?, description=?, ip_pool=?, routes=?, policies=?, updated_at=CURRENT_TIMESTAMP 
		WHERE id=?
	`, group.Name, group.Description, group.IPPool, group.Routes, group.Policies, id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "更新成功"})
}

func DeleteUserGroup(c *gin.Context) {
	id := c.Param("id")
	_, err := models.DB.Exec("DELETE FROM user_groups WHERE id=?", id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "删除成功"})
}

func GetUsers(c *gin.Context) {
	rows, err := models.DB.Query(`
		SELECT u.id, u.username, u.full_name, u.email, u.group_id, g.name as group_name, 
		       u.custom_routes, u.custom_policies, u.enabled, u.created_at, u.updated_at 
		FROM users u
		LEFT JOIN user_groups g ON u.group_id = g.id
		ORDER BY u.created_at DESC
	`)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	defer rows.Close()

	var users []models.User
	for rows.Next() {
		var u models.User
		if err := rows.Scan(&u.ID, &u.Username, &u.FullName, &u.Email, &u.GroupID, &u.GroupName, 
			&u.CustomRoutes, &u.CustomPolicies, &u.Enabled, &u.CreatedAt, &u.UpdatedAt); err != nil {
			continue
		}
		users = append(users, u)
	}

	c.JSON(http.StatusOK, gin.H{"data": users})
}

func CreateUser(c *gin.Context) {
	var user models.User
	if err := c.ShouldBindJSON(&user); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(user.Password), bcrypt.DefaultCost)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "密码加密失败"})
		return
	}

	result, err := models.DB.Exec(`
		INSERT INTO users (username, password, full_name, email, group_id, custom_routes, custom_policies, enabled) 
		VALUES (?, ?, ?, ?, ?, ?, ?, ?)
	`, user.Username, string(hashedPassword), user.FullName, user.Email, user.GroupID, user.CustomRoutes, user.CustomPolicies, user.Enabled)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	id, _ := result.LastInsertId()
	user.ID = int(id)
	c.JSON(http.StatusOK, gin.H{"data": user})
}

func UpdateUser(c *gin.Context) {
	id := c.Param("id")
	var user models.User
	if err := c.ShouldBindJSON(&user); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if user.Password != "" {
		hashedPassword, err := bcrypt.GenerateFromPassword([]byte(user.Password), bcrypt.DefaultCost)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "密码加密失败"})
			return
		}
		_, err = models.DB.Exec(`
			UPDATE users 
			SET username=?, password=?, full_name=?, email=?, group_id=?, custom_routes=?, custom_policies=?, enabled=?, updated_at=CURRENT_TIMESTAMP 
			WHERE id=?
		`, user.Username, string(hashedPassword), user.FullName, user.Email, user.GroupID, user.CustomRoutes, user.CustomPolicies, user.Enabled, id)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
	} else {
		_, err := models.DB.Exec(`
			UPDATE users 
			SET username=?, full_name=?, email=?, group_id=?, custom_routes=?, custom_policies=?, enabled=?, updated_at=CURRENT_TIMESTAMP 
			WHERE id=?
		`, user.Username, user.FullName, user.Email, user.GroupID, user.CustomRoutes, user.CustomPolicies, user.Enabled, id)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
	}

	c.JSON(http.StatusOK, gin.H{"message": "更新成功"})
}

func DeleteUser(c *gin.Context) {
	id := c.Param("id")
	_, err := models.DB.Exec("DELETE FROM users WHERE id=?", id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "删除成功"})
}

func GetOnlineUsers(c *gin.Context) {
	rows, err := models.DB.Query(`
		SELECT id, username, group_name, mac, virtual_ip, remote_ip, protocol, virtual_dev, mtu, 
		       upload_speed, download_speed, total_upload, total_download, connected_at 
		FROM online_users
		ORDER BY connected_at DESC
	`)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	defer rows.Close()

	var users []models.OnlineUser
	for rows.Next() {
		var u models.OnlineUser
		if err := rows.Scan(&u.ID, &u.Username, &u.GroupName, &u.MAC, &u.VirtualIP, &u.RemoteIP, 
			&u.Protocol, &u.VirtualDev, &u.MTU, &u.UploadSpeed, &u.DownloadSpeed, 
			&u.TotalUpload, &u.TotalDownload, &u.ConnectedAt); err != nil {
			continue
		}
		users = append(users, u)
	}

	c.JSON(http.StatusOK, gin.H{"data": users})
}

func DisconnectUser(c *gin.Context) {
	id := c.Param("id")
	
	var username string
	var groupID int
	var virtualIP string
	err := models.DB.QueryRow(`
		SELECT u.username, u.group_id, ou.virtual_ip 
		FROM online_users ou
		JOIN users u ON ou.username = u.username
		WHERE ou.id = ?
	`, id).Scan(&username, &groupID, &virtualIP)
	
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "未找到在线用户"})
		return
	}

	vpn.DisconnectUserByOCCtl(username)

	_, err = models.DB.Exec("DELETE FROM online_users WHERE id=?", id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	models.ReleaseIP(username, groupID)

	c.JSON(http.StatusOK, gin.H{"message": "断开成功"})
}

func GetAuthLogs(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("pageSize", "20"))
	offset := (page - 1) * pageSize

	var total int
	models.DB.QueryRow("SELECT COUNT(*) FROM auth_logs").Scan(&total)

	rows, err := models.DB.Query(`
		SELECT id, username, remote_ip, action, success, message, created_at 
		FROM auth_logs
		ORDER BY created_at DESC
		LIMIT ? OFFSET ?
	`, pageSize, offset)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	defer rows.Close()

	var logs []models.AuthLog
	for rows.Next() {
		var l models.AuthLog
		if err := rows.Scan(&l.ID, &l.Username, &l.RemoteIP, &l.Action, &l.Success, &l.Message, &l.CreatedAt); err != nil {
			continue
		}
		logs = append(logs, l)
	}

	c.JSON(http.StatusOK, gin.H{
		"data":     logs,
		"total":    total,
		"page":     page,
		"pageSize": pageSize,
	})
}

func GetAccessLogs(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("pageSize", "20"))
	offset := (page - 1) * pageSize

	var total int
	models.DB.QueryRow("SELECT COUNT(*) FROM access_logs").Scan(&total)

	rows, err := models.DB.Query(`
		SELECT id, username, src_ip, dst_ip, dst_port, protocol, action, bytes_sent, bytes_recv, created_at 
		FROM access_logs
		ORDER BY created_at DESC
		LIMIT ? OFFSET ?
	`, pageSize, offset)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	defer rows.Close()

	var logs []models.AccessLog
	for rows.Next() {
		var l models.AccessLog
		if err := rows.Scan(&l.ID, &l.Username, &l.SrcIP, &l.DstIP, &l.DstPort, &l.Protocol, &l.Action, &l.BytesSent, &l.BytesRecv, &l.CreatedAt); err != nil {
			continue
		}
		logs = append(logs, l)
	}

	c.JSON(http.StatusOK, gin.H{
		"data":     logs,
		"total":    total,
		"page":     page,
		"pageSize": pageSize,
	})
}

func GetSystemStats(c *gin.Context) {
	var onlineCount int
	models.DB.QueryRow("SELECT COUNT(*) FROM online_users").Scan(&onlineCount)

	stats := models.SystemStats{
		CPUUsage:           getCPUUsage(),
		MemoryUsage:        getMemoryUsage(),
		DiskUsage:          getDiskUsage(),
		NetworkConnections: getNetworkConnections(),
		OnlineUsers:        onlineCount,
		Uptime:             getUptime(),
	}

	c.JSON(http.StatusOK, gin.H{"data": stats})
}

var (
	lastCPUTotal uint64
	lastCPUIdle  uint64
	startTime    = time.Now()
)

func getCPUUsage() float64 {
	file, err := os.Open("/proc/stat")
	if err != nil {
		return 0.0
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	if !scanner.Scan() {
		return 0.0
	}

	line := scanner.Text()
	if !strings.HasPrefix(line, "cpu ") {
		return 0.0
	}

	fields := strings.Fields(line)
	if len(fields) < 5 {
		return 0.0
	}

	user, _ := strconv.ParseUint(fields[1], 10, 64)
	nice, _ := strconv.ParseUint(fields[2], 10, 64)
	system, _ := strconv.ParseUint(fields[3], 10, 64)
	idle, _ := strconv.ParseUint(fields[4], 10, 64)
	iowait, _ := strconv.ParseUint(fields[5], 10, 64)
	irq, _ := strconv.ParseUint(fields[6], 10, 64)
	softirq, _ := strconv.ParseUint(fields[7], 10, 64)

	total := user + nice + system + idle + iowait + irq + softirq

	if lastCPUTotal == 0 {
		lastCPUTotal = total
		lastCPUIdle = idle
		return 0.0
	}

	totalDelta := total - lastCPUTotal
	idleDelta := idle - lastCPUIdle

	lastCPUTotal = total
	lastCPUIdle = idle

	if totalDelta == 0 {
		return 0.0
	}

	return float64(totalDelta-idleDelta) / float64(totalDelta) * 100.0
}

func getMemoryUsage() float64 {
	file, err := os.Open("/proc/meminfo")
	if err != nil {
		return 0.0
	}
	defer file.Close()

	var memTotal, memAvailable uint64
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := scanner.Text()
		fields := strings.Fields(line)
		if len(fields) < 2 {
			continue
		}

		if strings.HasPrefix(line, "MemTotal:") {
			memTotal, _ = strconv.ParseUint(fields[1], 10, 64)
		} else if strings.HasPrefix(line, "MemAvailable:") {
			memAvailable, _ = strconv.ParseUint(fields[1], 10, 64)
		}
	}

	if memTotal == 0 {
		return 0.0
	}

	memUsed := memTotal - memAvailable
	return float64(memUsed) / float64(memTotal) * 100.0
}

func getDiskUsage() float64 {
	var stat syscall.Statfs_t
	err := syscall.Statfs("/", &stat)
	if err != nil {
		return 0.0
	}

	total := stat.Blocks * uint64(stat.Bsize)
	free := stat.Bfree * uint64(stat.Bsize)
	used := total - free

	if total == 0 {
		return 0.0
	}

	return float64(used) / float64(total) * 100.0
}

func getNetworkConnections() int {
	file, err := os.Open("/proc/net/tcp")
	if err != nil {
		return 0
	}
	defer file.Close()

	count := -1
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		count++
	}

	file6, err := os.Open("/proc/net/tcp6")
	if err == nil {
		defer file6.Close()
		scanner6 := bufio.NewScanner(file6)
		for scanner6.Scan() {
			count++
		}
	}

	return count
}

func getUptime() int64 {
	return int64(time.Since(startTime).Seconds())
}