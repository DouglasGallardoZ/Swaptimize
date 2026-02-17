package system

import (
	"log"
	"sync"
	"time"
)

// RateLimiter controla la frecuencia de operaciones
type RateLimiter struct {
	mu                      sync.RWMutex
	lastCreate              time.Time
	lastDelete              time.Time
	minIntervalCreate       time.Duration
	minIntervalDelete       time.Duration
	createAttemptsSinceOk   int
	deleteAttemptsSinceOk   int
}

// NewRateLimiter crea un nuevo rate limiter
func NewRateLimiter(createIntervalSec, deleteIntervalSec int) *RateLimiter {
	return &RateLimiter{
		minIntervalCreate: time.Duration(createIntervalSec) * time.Second,
		minIntervalDelete: time.Duration(deleteIntervalSec) * time.Second,
		lastCreate:        time.Now().Add(-24 * time.Hour), // Permitir primera creación
		lastDelete:        time.Now().Add(-24 * time.Hour),
	}
}

// CanCreate verifica si se puede crear otro archivo swap
func (rl *RateLimiter) CanCreate() bool {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	elapsed := time.Since(rl.lastCreate)
	canCreate := elapsed >= rl.minIntervalCreate

	if canCreate {
		rl.createAttemptsSinceOk = 0
	} else {
		rl.createAttemptsSinceOk++
		waitTime := rl.minIntervalCreate - elapsed
		log.Printf("⏳ Rate limiter: blocked create attempt #%d, wait %d more seconds\n",
			rl.createAttemptsSinceOk, int(waitTime.Seconds()))
	}

	return canCreate
}

// CanDelete verifica si se puede borrar un archivo swap
func (rl *RateLimiter) CanDelete() bool {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	elapsed := time.Since(rl.lastDelete)
	canDelete := elapsed >= rl.minIntervalDelete

	if canDelete {
		rl.deleteAttemptsSinceOk = 0
	} else {
		rl.deleteAttemptsSinceOk++
		waitTime := rl.minIntervalDelete - elapsed
		log.Printf("⏳ Rate limiter: blocked delete attempt #%d, wait %d more seconds\n",
			rl.deleteAttemptsSinceOk, int(waitTime.Seconds()))
	}

	return canDelete
}

// RecordCreate marca que se creó un archivo swap
func (rl *RateLimiter) RecordCreate() {
	rl.mu.Lock()
	defer rl.mu.Unlock()
	rl.lastCreate = time.Now()
}

// RecordDelete marca que se borró un archivo swap
func (rl *RateLimiter) RecordDelete() {
	rl.mu.Lock()
	defer rl.mu.Unlock()
	rl.lastDelete = time.Now()
}

// Stats retorna estadísticas del rate limiter
func (rl *RateLimiter) Stats() map[string]interface{} {
	rl.mu.RLock()
	defer rl.mu.RUnlock()

	return map[string]interface{}{
		"last_create":                rl.lastCreate,
		"last_delete":                rl.lastDelete,
		"min_interval_create_sec":    int(rl.minIntervalCreate.Seconds()),
		"min_interval_delete_sec":    int(rl.minIntervalDelete.Seconds()),
		"create_attempts_since_ok":   rl.createAttemptsSinceOk,
		"delete_attempts_since_ok":   rl.deleteAttemptsSinceOk,
	}
}

// Reset resetea el rate limiter
func (rl *RateLimiter) Reset() {
	rl.mu.Lock()
	defer rl.mu.Unlock()
	rl.lastCreate = time.Now().Add(-24 * time.Hour)
	rl.lastDelete = time.Now().Add(-24 * time.Hour)
	rl.createAttemptsSinceOk = 0
	rl.deleteAttemptsSinceOk = 0
}
