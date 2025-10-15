// internal/security/rate_limiter.go
package security

import (
	"sync"
	"time"
)

type attempt struct {
	Count     int
	LockUntil time.Time
}

type LoginLimiter struct {
	mu       sync.Mutex
	store    map[string]*attempt
	MaxFail  int           // 5
	LockTime time.Duration // 10 * time.Minute
}

func NewLoginLimiter() *LoginLimiter {
	return &LoginLimiter{
		store:    make(map[string]*attempt),
		MaxFail:  5,
		LockTime: 10 * time.Minute,
	}
}

// key 建议 username（可加上 ip：username|ip）
func (l *LoginLimiter) Allow(key string) (ok bool, remaining int, unlockAt time.Time) {
	l.mu.Lock()
	defer l.mu.Unlock()

	at := l.store[key]
	if at == nil {
		return true, l.MaxFail, time.Time{}
	}
	if time.Now().Before(at.LockUntil) {
		return false, 0, at.LockUntil
	}
	// 锁已过期，重置
	if at.Count > 0 {
		at.Count = 0
	}
	return true, l.MaxFail, time.Time{}
}

func (l *LoginLimiter) OnFail(key string) (locked bool, unlockAt time.Time) {
	l.mu.Lock()
	defer l.mu.Unlock()

	at := l.store[key]
	if at == nil {
		at = &attempt{}
		l.store[key] = at
	}
	at.Count++
	if at.Count >= l.MaxFail {
		at.LockUntil = time.Now().Add(l.LockTime)
		at.Count = 0 // 进入锁定后计数清零
		return true, at.LockUntil
	}
	return false, time.Time{}
}

func (l *LoginLimiter) OnSuccess(key string) {
	l.mu.Lock()
	defer l.mu.Unlock()
	delete(l.store, key)
}
