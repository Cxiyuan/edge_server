package handlers

import (
	"edge_server/models"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
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