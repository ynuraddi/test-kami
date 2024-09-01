package application

import (
	"sync"
	"time"
)

type MutexManager struct {
	mutexes         map[string]*mutexWithTime
	mutexesLock     sync.Mutex
	cleanupInterval time.Duration
	expiration      time.Duration
}

type mutexWithTime struct {
	mu       *sync.Mutex
	lastUsed time.Time
}

func NewMutexManager(cleanupInterval, expiration time.Duration) *MutexManager {
	mm := &MutexManager{
		mutexes:         make(map[string]*mutexWithTime, 100),
		cleanupInterval: cleanupInterval,
		expiration:      expiration,
	}
	go mm.startCleanupRoutine()
	return mm
}

func (mm *MutexManager) GetMutex(roomID string) *sync.Mutex {
	mm.mutexesLock.Lock()
	defer mm.mutexesLock.Unlock()

	if roomMutex, ok := mm.mutexes[roomID]; ok {
		roomMutex.lastUsed = time.Now()
		return roomMutex.mu
	}

	m := &sync.Mutex{}
	mm.mutexes[roomID] = &mutexWithTime{
		mu:       m,
		lastUsed: time.Now(),
	}
	return m
}

func (mm *MutexManager) startCleanupRoutine() {
	ticker := time.NewTicker(mm.cleanupInterval)

	for {
		<-ticker.C
		mm.cleanupMutexes()
	}
}

func (mm *MutexManager) cleanupMutexes() {
	mm.mutexesLock.Lock()
	defer mm.mutexesLock.Unlock()

	now := time.Now()
	for k, v := range mm.mutexes {
		if now.Sub(v.lastUsed) > mm.expiration {
			delete(mm.mutexes, k)
		}
	}
}
