package handlers

import (
	"edge_server/models"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"golang.org/x/crypto/bcrypt"
)

func GetSystemConfig(c *gin.Context) {
	config, err := models.GetAllConfig()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": config})
}

func UpdateSystemConfig(c *gin.Context) {
	var req map[string]string
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	validKeys := map[string]bool{
		"default_ip_pool": true,
		"default_dns1":    true,
		"default_dns2":    true,
		"default_mtu":     true,
		"max_clients":     true,
		"idle_timeout":    true,
		"vpn_domain":      true,
		"vpn_device":      true,
	}

	for key, value := range req {
		if !validKeys[key] {
			continue
		}

		if key == "default_mtu" || key == "max_clients" || key == "idle_timeout" {
			if _, err := strconv.Atoi(value); err != nil {
				c.JSON(http.StatusBadRequest, gin.H{"error": key + " 必须是数字"})
				return
			}
		}

		if err := models.SetConfig(key, value); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
	}

	c.JSON(http.StatusOK, gin.H{"message": "配置更新成功，重启服务后生效"})
}

func ChangePassword(c *gin.Context) {
	username, exists := c.Get("username")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "未认证"})
		return
	}

	var req struct {
		OldPassword string `json:"old_password" binding:"required"`
		NewPassword string `json:"new_password" binding:"required,min=6"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "请求参数错误"})
		return
	}

	var storedPassword string
	err := models.DB.QueryRow("SELECT password FROM users WHERE username=?", username).Scan(&storedPassword)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "查询用户失败"})
		return
	}

	if err := bcrypt.CompareHashAndPassword([]byte(storedPassword), []byte(req.OldPassword)); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "原密码错误"})
		return
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.NewPassword), bcrypt.DefaultCost)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "密码加密失败"})
		return
	}

	_, err = models.DB.Exec("UPDATE users SET password=?, updated_at=CURRENT_TIMESTAMP WHERE username=?", 
		string(hashedPassword), username)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "更新密码失败"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "密码修改成功"})
}