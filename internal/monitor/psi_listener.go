package monitor

import (
	"context"
	"fmt"
	"log"
	"sync"
	"time"

	"github.com/fsnotify/fsnotify"
)

// MemoryPressureEvent es un evento de presión de memoria
type MemoryPressureEvent struct {
	Type       string    // "high", "low", "critical"
	SomeAvg10  float64
	Timestamp  time.Time
	Description string
}

// PSIListener escucha cambios en /proc/pressure/memory usando inotify
type PSIListener struct {
	reader         *PSIReader
	watcher        *fsnotify.Watcher
	eventChan      chan MemoryPressureEvent
	ctx            context.Context
	cancel         context.CancelFunc
	mu             sync.RWMutex
	isRunning      bool
	lastEvent      *MemoryPressureEvent
	highThreshold  float64
	lowThreshold   float64
	lastPressure   float64
	stabilizeCount int
}

// NewPSIListener crea un nuevo listener de PSI
func NewPSIListener(ctx context.Context, highThreshold, lowThreshold float64) (*PSIListener, error) {
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		return nil, err
	}

	pctx, cancel := context.WithCancel(ctx)

	listener := &PSIListener{
		reader:        NewPSIReader(),
		watcher:       watcher,
		eventChan:     make(chan MemoryPressureEvent, 10),
		ctx:           pctx,
		cancel:        cancel,
		highThreshold: highThreshold,
		lowThreshold:  lowThreshold,
		lastPressure:  0,
	}

	// Agregar watch a /proc/pressure/memory
	err = watcher.Add("/proc/pressure/memory")
	if err != nil {
		watcher.Close()
		return nil, err
	}

	log.Println("✓ PSI listener initialized (inotify)")
	return listener, nil
}

// Start comienza a escuchar eventos de PSI
func (pl *PSIListener) Start() {
	pl.mu.Lock()
	if pl.isRunning {
		pl.mu.Unlock()
		return
	}
	pl.isRunning = true
	pl.mu.Unlock()

	go pl.listenLoop()
	go pl.heartbeat() // Verificar periódicamente aunque no haya eventos
}

// listenLoop es el loop principal del listener
func (pl *PSIListener) listenLoop() {
	defer func() {
		pl.mu.Lock()
		pl.isRunning = false
		pl.mu.Unlock()
	}()

	for {
		select {
		case <-pl.ctx.Done():
			log.Println("⏹️  PSI listener stopped")
			return

		case event, ok := <-pl.watcher.Events:
			if !ok {
				return
			}

			// /proc/pressure/memory fue modificado
			if event.Op&fsnotify.Write == fsnotify.Write {
				metrics, err := pl.reader.ReadMemoryPSI()
				if err != nil {
					log.Printf("⚠️  Error reading PSI: %v\n", err)
					continue
				}

				pl.processPressure(metrics)
			}

		case err, ok := <-pl.watcher.Errors:
			if !ok {
				return
			}
			log.Printf("⚠️  Watcher error: %v\n", err)
		}
	}
}

// heartbeat verifica periódicamente PSI aunque no haya eventos inotify
func (pl *PSIListener) heartbeat() {
	ticker := time.NewTicker(5 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-pl.ctx.Done():
			return
		case <-ticker.C:
			metrics, err := pl.reader.ReadMemoryPSI()
			if err != nil {
				continue
			}
			pl.processPressure(metrics)
		}
	}
}

// processPressure procesa cambios de presión y genera eventos
func (pl *PSIListener) processPressure(metrics *PSIMetrics) {
	pl.mu.Lock()
	defer pl.mu.Unlock()

	currentPressure := metrics.SomeAvg10
	pressureChange := currentPressure - pl.lastPressure

	// Detectar transiciones de presión
	var eventType string
	var shouldGenerate bool

	switch {
	case currentPressure >= pl.highThreshold && pl.lastPressure < pl.highThreshold:
		eventType = "high"
		shouldGenerate = true
		log.Printf("🔴 HIGH PRESSURE: %.2f%% (some_avg10)\n", currentPressure)

	case currentPressure < pl.lowThreshold && pl.lastPressure >= pl.lowThreshold:
		eventType = "low"
		shouldGenerate = true
		log.Printf("🟢 PRESSURE RELIEVED: %.2f%%\n", currentPressure)

	case currentPressure >= 50 && pressureChange > 5:
		eventType = "critical"
		shouldGenerate = true
		log.Printf("🟠 CRITICAL SPIKE: %.2f%% (+%.2f%%)\n", currentPressure, pressureChange)
	}

	pl.lastPressure = currentPressure

	if shouldGenerate {
		event := MemoryPressureEvent{
			Type:      eventType,
			SomeAvg10: currentPressure,
			Timestamp: metrics.Timestamp,
			Description: fmt.Sprintf("Memory pressure: %.2f%% (type=%s)", currentPressure, eventType),
		}

		// Enviar evento sin bloquear
		select {
		case pl.eventChan <- event:
			pl.lastEvent = &event
		default:
			log.Println("⚠️  Event channel full, dropping event")
		}
	}
}

// Events retorna el channel de eventos
func (pl *PSIListener) Events() <-chan MemoryPressureEvent {
	return pl.eventChan
}

// Stop detiene el listener
func (pl *PSIListener) Stop() error {
	log.Println("⏹️  Stopping PSI listener...")
	pl.cancel()
	time.Sleep(100 * time.Millisecond) // Dar tiempo a goroutines
	return pl.watcher.Close()
}

// IsRunning verifica si el listener está activo
func (pl *PSIListener) IsRunning() bool {
	pl.mu.RLock()
	defer pl.mu.RUnlock()
	return pl.isRunning
}

// GetLastEvent retorna el último evento generado
func (pl *PSIListener) GetLastEvent() *MemoryPressureEvent {
	pl.mu.RLock()
	defer pl.mu.RUnlock()
	return pl.lastEvent
}

// Stats retorna estadísticas del listener
func (pl *PSIListener) Stats() map[string]interface{} {
	pl.mu.RLock()
	defer pl.mu.RUnlock()

	return map[string]interface{}{
		"running":         pl.isRunning,
		"last_pressure":   pl.lastPressure,
		"high_threshold":  pl.highThreshold,
		"low_threshold":   pl.lowThreshold,
		"last_event":      pl.lastEvent,
	}
}
