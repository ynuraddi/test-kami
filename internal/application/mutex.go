package application

import (
	"sync"
	"time"
)

type MutexManager struct {
	mutexes         sync.Map
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
		mutexes:         sync.Map{},
		cleanupInterval: cleanupInterval,
		expiration:      expiration,
	}
	go mm.startCleanupRoutine()
	return mm
}

func (mm *MutexManager) GetMutex(roomID string) *sync.Mutex {
	mm.mutexesLock.Lock()
	defer mm.mutexesLock.Unlock()

	if val, ok := mm.mutexes.Load(roomID); ok {
		roomMutex := val.(*mutexWithTime)
		roomMutex.lastUsed = time.Now()
		return roomMutex.mu
	}

	m := &sync.Mutex{}
	mm.mutexes.Store(roomID, &mutexWithTime{
		mu:       m,
		lastUsed: time.Now(),
	})
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
	mm.mutexes.Range(func(key, value any) bool {
		roomMutex := value.(*mutexWithTime)
		if now.Sub(roomMutex.lastUsed) > mm.expiration {
			mm.mutexes.Delete(key)
		}
		return true
	})
}
