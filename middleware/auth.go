package middleware

import (
	"crypto/rand"
	"edge_server/models"
	"encoding/hex"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"golang.org/x/crypto/bcrypt"
)

type Session struct {
	Username  string
	CreatedAt time.Time
	ExpiresAt time.Time
}

var (
	sessions = make(map[string]*Session)
	mu       sync.RWMutex
)

func generateToken() string {
	b := make([]byte, 32)
	rand.Read(b)
	return hex.EncodeToString(b)
}

func Login(c *gin.Context) {
	var req struct {
		Username string `json:"username"`
		Password string `json:"password"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "请求参数错误"})
		return
	}

	var storedPassword string
	var enabled bool
	err := models.DB.QueryRow("SELECT password, enabled FROM users WHERE username=?", req.Username).Scan(&storedPassword, &enabled)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "用户名或密码错误"})
		return
	}

	if !enabled {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "用户已被禁用"})
		return
	}

	if err := bcrypt.CompareHashAndPassword([]byte(storedPassword), []byte(req.Password)); err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "用户名或密码错误"})
		return
	}

	token := generateToken()
	session := &Session{
		Username:  req.Username,
		CreatedAt: time.Now(),
		ExpiresAt: time.Now().Add(24 * time.Hour),
	}

	mu.Lock()
	sessions[token] = session
	mu.Unlock()

	c.JSON(http.StatusOK, gin.H{
		"token":    token,
		"username": req.Username,
	})
}

func AuthRequired() gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "未授权访问"})
			c.Abort()
			return
		}

		token := strings.TrimPrefix(authHeader, "Bearer ")
		if token == authHeader {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Token格式错误"})
			c.Abort()
			return
		}

		mu.RLock()
		session, exists := sessions[token]
		mu.RUnlock()

		if !exists {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "无效的Token"})
			c.Abort()
			return
		}

		if time.Now().After(session.ExpiresAt) {
			mu.Lock()
			delete(sessions, token)
			mu.Unlock()
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Token已过期"})
			c.Abort()
			return
		}

		c.Set("username", session.Username)
		c.Next()
	}
}

func CleanupExpiredSessions() {
	ticker := time.NewTicker(1 * time.Hour)
	go func() {
		for range ticker.C {
			mu.Lock()
			now := time.Now()
			for token, session := range sessions {
				if now.After(session.ExpiresAt) {
					delete(sessions, token)
				}
			}
			mu.Unlock()
		}
	}()
}
