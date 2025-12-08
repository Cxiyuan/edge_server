package vpn

import (
	"edge_server/models"
	"log"
	"sync"
	"time"
)

type VPNSession struct {
	Username   string
	GroupID    int
	VirtualIP  string
	RemoteIP   string
	ConnectedAt time.Time
	LastActivity time.Time
	TotalUpload int64
	TotalDownload int64
	mu sync.Mutex
}

var (
	sessions = make(map[string]*VPNSession)
	sessionsMu sync.RWMutex
)

func AddSession(username, virtualIP, remoteIP string, groupID int) {
	sessionsMu.Lock()
	defer sessionsMu.Unlock()
	
	sessions[username] = &VPNSession{
		Username:   username,
		GroupID:    groupID,
		VirtualIP:  virtualIP,
		RemoteIP:   remoteIP,
		ConnectedAt: time.Now(),
		LastActivity: time.Now(),
	}
}

func RemoveSession(username string) {
	sessionsMu.Lock()
	defer sessionsMu.Unlock()
	
	if session, exists := sessions[username]; exists {
		models.ReleaseIP(username, session.GroupID)
		delete(sessions, username)
	}
}

func GetSession(username string) (*VPNSession, bool) {
	sessionsMu.RLock()
	defer sessionsMu.RUnlock()
	
	session, exists := sessions[username]
	return session, exists
}

func UpdateSessionActivity(username string, uploadBytes, downloadBytes int64) {
	sessionsMu.RLock()
	session, exists := sessions[username]
	sessionsMu.RUnlock()
	
	if !exists {
		return
	}
	
	session.mu.Lock()
	session.LastActivity = time.Now()
	session.TotalUpload += uploadBytes
	session.TotalDownload += downloadBytes
	session.mu.Unlock()
	
	models.DB.Exec(`
		UPDATE online_users 
		SET total_upload = ?, total_download = ?, upload_speed = ?, download_speed = ?
		WHERE username = ?
	`, session.TotalUpload, session.TotalDownload, uploadBytes, downloadBytes, username)
}

func StartSessionCleanup(idleTimeout int) {
	ticker := time.NewTicker(30 * time.Second)
	go func() {
		for range ticker.C {
			sessionsMu.Lock()
			now := time.Now()
			for username, session := range sessions {
				session.mu.Lock()
				idle := now.Sub(session.LastActivity).Seconds()
				session.mu.Unlock()
				
				if idle > float64(idleTimeout) {
					log.Printf("会话超时，断开用户: %s (空闲 %.0f 秒)", username, idle)
					
					models.DB.Exec("DELETE FROM online_users WHERE username=?", username)
					models.ReleaseIP(username, session.GroupID)
					
					delete(sessions, username)
				}
			}
			sessionsMu.Unlock()
		}
	}()
}
