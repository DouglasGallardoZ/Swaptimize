package metrics

import (
	"fmt"
	"log"
	"net/http"
	"sync"
	"time"
)

// PrometheusMetrics gestiona las métricas de Prometheus
type PrometheusMetrics struct {
	mu sync.RWMutex

	// Gauge metrics
	MemoryUsageBytes        float64
	MemoryAvailableBytes    float64
	MemoryUsagePercent      float64
	SwapUsageBytes          float64
	SwapAvailableBytes      float64
	SwapUsagePercent        float64
	PSISomeAvg10            float64
	PSISomeAvg60            float64
	PSISomeAvg300           float64
	PSIFullAvg10            float64
	PSIFullAvg60            float64
	PSIFullAvg300           float64

	// Counter metrics
	SwapFilesCreated        int64
	SwapFilesDeleted        int64
	OperationsTotal         int64
	ErrorsTotal             int64
	RetryAttemptsTotal      int64

	// Summary metrics (latest values)
	LastSwapCreationMS      float64
	LastSwapDeletionMS      float64
	LastMonitorCycleMS      float64

	// Metadata
	StartTime      time.Time
	LastUpdateTime time.Time
	ServerRunning  bool
	Port           int
	server         *http.Server

	// Configuration (Profile & Environment)
	EnvironmentType   string
	EnvironmentDesc   string
	ProfileName       string
	ProfileMode       string
	ThresholdHigh     int
	ThresholdLow      int
	MaxSwapFiles      int
	SwapSize          int
	AllowCreation     bool
}

// NewPrometheusMetrics crea un nuevo gestor de métricas
func NewPrometheusMetrics(port int) *PrometheusMetrics {
	pm := &PrometheusMetrics{
		Port:      port,
		StartTime: time.Now(),
	}
	return pm
}

// UpdateGauges actualiza las métricas gauge
func (pm *PrometheusMetrics) UpdateGauges(
	memUsage, memAvail, memPercent,
	swapUsage, swapAvail, swapPercent float64,
) {
	pm.mu.Lock()
	defer pm.mu.Unlock()

	pm.MemoryUsageBytes = memUsage
	pm.MemoryAvailableBytes = memAvail
	pm.MemoryUsagePercent = memPercent
	pm.SwapUsageBytes = swapUsage
	pm.SwapAvailableBytes = swapAvail
	pm.SwapUsagePercent = swapPercent
	pm.LastUpdateTime = time.Now()
}

// UpdatePSI actualiza las métricas de presión
func (pm *PrometheusMetrics) UpdatePSI(
	someAvg10, someAvg60, someAvg300,
	fullAvg10, fullAvg60, fullAvg300 float64,
) {
	pm.mu.Lock()
	defer pm.mu.Unlock()

	pm.PSISomeAvg10 = someAvg10
	pm.PSISomeAvg60 = someAvg60
	pm.PSISomeAvg300 = someAvg300
	pm.PSIFullAvg10 = fullAvg10
	pm.PSIFullAvg60 = fullAvg60
	pm.PSIFullAvg300 = fullAvg300
	pm.LastUpdateTime = time.Now()
}

// RecordSwapCreation registra la creación de un archivo swap
func (pm *PrometheusMetrics) RecordSwapCreation(durationMS float64) {
	pm.mu.Lock()
	defer pm.mu.Unlock()

	pm.SwapFilesCreated++
	pm.OperationsTotal++
	pm.LastSwapCreationMS = durationMS
}

// RecordSwapDeletion registra la eliminación de un archivo swap
func (pm *PrometheusMetrics) RecordSwapDeletion(durationMS float64) {
	pm.mu.Lock()
	defer pm.mu.Unlock()

	pm.SwapFilesDeleted++
	pm.OperationsTotal++
	pm.LastSwapDeletionMS = durationMS
}

// RecordError registra un error
func (pm *PrometheusMetrics) RecordError() {
	pm.mu.Lock()
	defer pm.mu.Unlock()

	pm.ErrorsTotal++
}

// RecordRetry registra un reintento
func (pm *PrometheusMetrics) RecordRetry() {
	pm.mu.Lock()
	defer pm.mu.Unlock()

	pm.RetryAttemptsTotal++
}

// RecordMonitorCycle registra la duración de un ciclo de monitoreo
func (pm *PrometheusMetrics) RecordMonitorCycle(durationMS float64) {
	pm.mu.Lock()
	defer pm.mu.Unlock()

	pm.LastMonitorCycleMS = durationMS
}

// UpdateConfiguration actualiza la configuraci\u00f3n del perfil detectado
func (pm *PrometheusMetrics) UpdateConfiguration(
	environmentType, environmentDesc, profileName, profileMode string,
	thresholdHigh, thresholdLow, maxSwapFiles, swapSize int,
	allowCreation bool,
) {
	pm.mu.Lock()
	defer pm.mu.Unlock()

	pm.EnvironmentType = environmentType
	pm.EnvironmentDesc = environmentDesc
	pm.ProfileName = profileName
	pm.ProfileMode = profileMode
	pm.ThresholdHigh = thresholdHigh
	pm.ThresholdLow = thresholdLow
	pm.MaxSwapFiles = maxSwapFiles
	pm.SwapSize = swapSize
	pm.AllowCreation = allowCreation
}

// Start inicia el servidor de metricas
func (pm *PrometheusMetrics) Start() error {
	mux := http.NewServeMux()
	mux.HandleFunc("/metrics", pm.handleMetrics)
	mux.HandleFunc("/health", pm.handleHealth)

	addr := fmt.Sprintf(":%d", pm.Port)
	pm.server = &http.Server{
		Addr:    addr,
		Handler: mux,
	}

	pm.ServerRunning = true
	go func() {
		if err := pm.server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Printf("❌ Metrics server error: %v\n", err)
			pm.ServerRunning = false
		}
	}()

	log.Printf("📊 Metrics server started on http://localhost%s/metrics\n", addr)
	return nil
}

// Stop detiene el servidor de métricas
func (pm *PrometheusMetrics) Stop() error {
	if pm.server != nil {
		pm.ServerRunning = false
		return pm.server.Close()
	}
	return nil
}

// handleMetrics genera la respuesta de Prometheus
func (pm *PrometheusMetrics) handleMetrics(w http.ResponseWriter, r *http.Request) {
	pm.mu.RLock()
	defer pm.mu.RUnlock()

	w.Header().Set("Content-Type", "text/plain; charset=utf-8")

	uptime := time.Since(pm.StartTime).Seconds()

	// HELP y TYPE
	fmt.Fprintf(w, "# HELP swaptimize_memory_usage_bytes Memory usage in bytes\n")
	fmt.Fprintf(w, "# TYPE swaptimize_memory_usage_bytes gauge\n")
	fmt.Fprintf(w, "swaptimize_memory_usage_bytes %.0f\n", pm.MemoryUsageBytes)

	fmt.Fprintf(w, "# HELP swaptimize_memory_available_bytes Available memory in bytes\n")
	fmt.Fprintf(w, "# TYPE swaptimize_memory_available_bytes gauge\n")
	fmt.Fprintf(w, "swaptimize_memory_available_bytes %.0f\n", pm.MemoryAvailableBytes)

	fmt.Fprintf(w, "# HELP swaptimize_memory_usage_percent Memory usage percentage\n")
	fmt.Fprintf(w, "# TYPE swaptimize_memory_usage_percent gauge\n")
	fmt.Fprintf(w, "swaptimize_memory_usage_percent %.2f\n", pm.MemoryUsagePercent)

	fmt.Fprintf(w, "# HELP swaptimize_swap_usage_bytes Swap usage in bytes\n")
	fmt.Fprintf(w, "# TYPE swaptimize_swap_usage_bytes gauge\n")
	fmt.Fprintf(w, "swaptimize_swap_usage_bytes %.0f\n", pm.SwapUsageBytes)

	fmt.Fprintf(w, "# HELP swaptimize_swap_available_bytes Available swap in bytes\n")
	fmt.Fprintf(w, "# TYPE swaptimize_swap_available_bytes gauge\n")
	fmt.Fprintf(w, "swaptimize_swap_available_bytes %.0f\n", pm.SwapAvailableBytes)

	fmt.Fprintf(w, "# HELP swaptimize_swap_usage_percent Swap usage percentage\n")
	fmt.Fprintf(w, "# TYPE swaptimize_swap_usage_percent gauge\n")
	fmt.Fprintf(w, "swaptimize_swap_usage_percent %.2f\n", pm.SwapUsagePercent)

	// PSI metrics
	fmt.Fprintf(w, "# HELP swaptimize_psi_some_avg10 PSI 'some' 10-second average\n")
	fmt.Fprintf(w, "# TYPE swaptimize_psi_some_avg10 gauge\n")
	fmt.Fprintf(w, "swaptimize_psi_some_avg10 %.2f\n", pm.PSISomeAvg10)

	fmt.Fprintf(w, "# HELP swaptimize_psi_some_avg60 PSI 'some' 60-second average\n")
	fmt.Fprintf(w, "# TYPE swaptimize_psi_some_avg60 gauge\n")
	fmt.Fprintf(w, "swaptimize_psi_some_avg60 %.2f\n", pm.PSISomeAvg60)

	fmt.Fprintf(w, "# HELP swaptimize_psi_some_avg300 PSI 'some' 300-second average\n")
	fmt.Fprintf(w, "# TYPE swaptimize_psi_some_avg300 gauge\n")
	fmt.Fprintf(w, "swaptimize_psi_some_avg300 %.2f\n", pm.PSISomeAvg300)

	fmt.Fprintf(w, "# HELP swaptimize_psi_full_avg10 PSI 'full' 10-second average\n")
	fmt.Fprintf(w, "# TYPE swaptimize_psi_full_avg10 gauge\n")
	fmt.Fprintf(w, "swaptimize_psi_full_avg10 %.2f\n", pm.PSIFullAvg10)

	fmt.Fprintf(w, "# HELP swaptimize_psi_full_avg60 PSI 'full' 60-second average\n")
	fmt.Fprintf(w, "# TYPE swaptimize_psi_full_avg60 gauge\n")
	fmt.Fprintf(w, "swaptimize_psi_full_avg60 %.2f\n", pm.PSIFullAvg60)

	fmt.Fprintf(w, "# HELP swaptimize_psi_full_avg300 PSI 'full' 300-second average\n")
	fmt.Fprintf(w, "# TYPE swaptimize_psi_full_avg300 gauge\n")
	fmt.Fprintf(w, "swaptimize_psi_full_avg300 %.2f\n", pm.PSIFullAvg300)

	// Counter metrics
	fmt.Fprintf(w, "# HELP swaptimize_swap_files_created_total Total swap files created\n")
	fmt.Fprintf(w, "# TYPE swaptimize_swap_files_created_total counter\n")
	fmt.Fprintf(w, "swaptimize_swap_files_created_total %d\n", pm.SwapFilesCreated)

	fmt.Fprintf(w, "# HELP swaptimize_swap_files_deleted_total Total swap files deleted\n")
	fmt.Fprintf(w, "# TYPE swaptimize_swap_files_deleted_total counter\n")
	fmt.Fprintf(w, "swaptimize_swap_files_deleted_total %d\n", pm.SwapFilesDeleted)

	fmt.Fprintf(w, "# HELP swaptimize_operations_total Total operations performed\n")
	fmt.Fprintf(w, "# TYPE swaptimize_operations_total counter\n")
	fmt.Fprintf(w, "swaptimize_operations_total %d\n", pm.OperationsTotal)

	fmt.Fprintf(w, "# HELP swaptimize_errors_total Total errors encountered\n")
	fmt.Fprintf(w, "# TYPE swaptimize_errors_total counter\n")
	fmt.Fprintf(w, "swaptimize_errors_total %d\n", pm.ErrorsTotal)

	fmt.Fprintf(w, "# HELP swaptimize_retry_attempts_total Total retry attempts\n")
	fmt.Fprintf(w, "# TYPE swaptimize_retry_attempts_total counter\n")
	fmt.Fprintf(w, "swaptimize_retry_attempts_total %d\n", pm.RetryAttemptsTotal)

	// Summary metrics
	fmt.Fprintf(w, "# HELP swaptimize_last_swap_creation_ms Last swap creation duration\n")
	fmt.Fprintf(w, "# TYPE swaptimize_last_swap_creation_ms gauge\n")
	fmt.Fprintf(w, "swaptimize_last_swap_creation_ms %.2f\n", pm.LastSwapCreationMS)

	fmt.Fprintf(w, "# HELP swaptimize_last_swap_deletion_ms Last swap deletion duration\n")
	fmt.Fprintf(w, "# TYPE swaptimize_last_swap_deletion_ms gauge\n")
	fmt.Fprintf(w, "swaptimize_last_swap_deletion_ms %.2f\n", pm.LastSwapDeletionMS)

	fmt.Fprintf(w, "# HELP swaptimize_last_monitor_cycle_ms Last monitor cycle duration\n")
	fmt.Fprintf(w, "# TYPE swaptimize_last_monitor_cycle_ms gauge\n")
	fmt.Fprintf(w, "swaptimize_last_monitor_cycle_ms %.2f\n", pm.LastMonitorCycleMS)

	// System info
	fmt.Fprintf(w, "# HELP swaptimize_uptime_seconds Daemon uptime in seconds\n")
	fmt.Fprintf(w, "# TYPE swaptimize_uptime_seconds gauge\n")
	fmt.Fprintf(w, "swaptimize_uptime_seconds %.0f\n", uptime)
}

// handleHealth genera la respuesta de health check
func (pm *PrometheusMetrics) handleHealth(w http.ResponseWriter, r *http.Request) {
	pm.mu.RLock()
	defer pm.mu.RUnlock()

	w.Header().Set("Content-Type", "application/json; charset=utf-8")

	status := "healthy"
	statusCode := http.StatusOK

	// Simple checks
	if pm.LastUpdateTime.IsZero() || time.Since(pm.LastUpdateTime) > 2*time.Minute {
		status = "stale"
		statusCode = http.StatusServiceUnavailable
	}

	if pm.ErrorsTotal > 0 {
		status = "degraded"
	}

	w.WriteHeader(statusCode)

	fmt.Fprintf(w, `{
  "status": "%s",
  "timestamp": "%s",
  "uptime_seconds": %.0f,
  "last_update": "%s",
  "configuration": {
    "environment": {
      "type": "%s",
      "description": "%s"
    },
    "profile": {
      "name": "%s",
      "mode": "%s",
      "threshold_high_percent": %d,
      "threshold_low_percent": %d,
      "max_swap_files": %d,
      "swap_size_mb": %d,
      "allow_creation": %v
    }
  },
  "metrics": {
    "memory_usage_percent": %.2f,
    "swap_usage_percent": %.2f,
    "swap_files_created": %d,
    "swap_files_deleted": %d,
    "operations_total": %d,
    "errors_total": %d,
    "retry_attempts_total": %d
  }
}
`, status, time.Now().Format(time.RFC3339), time.Since(pm.StartTime).Seconds(), 
		pm.LastUpdateTime.Format(time.RFC3339),
		pm.EnvironmentType, pm.EnvironmentDesc,
		pm.ProfileName, pm.ProfileMode,
		pm.ThresholdHigh, pm.ThresholdLow, pm.MaxSwapFiles, pm.SwapSize, pm.AllowCreation,
		pm.MemoryUsagePercent, pm.SwapUsagePercent,
		pm.SwapFilesCreated, pm.SwapFilesDeleted, pm.OperationsTotal, pm.ErrorsTotal, pm.RetryAttemptsTotal)

}

// GetCurrentMetrics retorna un snapshot de las métricas actuales
func (pm *PrometheusMetrics) GetCurrentMetrics() map[string]interface{} {
	pm.mu.RLock()
	defer pm.mu.RUnlock()

	return map[string]interface{}{
		"memory_usage_bytes":    pm.MemoryUsageBytes,
		"memory_available_bytes": pm.MemoryAvailableBytes,
		"memory_usage_percent":   pm.MemoryUsagePercent,
		"swap_usage_bytes":       pm.SwapUsageBytes,
		"swap_available_bytes":   pm.SwapAvailableBytes,
		"swap_usage_percent":     pm.SwapUsagePercent,
		"psi_some_avg10":         pm.PSISomeAvg10,
		"psi_some_avg60":         pm.PSISomeAvg60,
		"psi_some_avg300":        pm.PSISomeAvg300,
		"psi_full_avg10":         pm.PSIFullAvg10,
		"psi_full_avg60":         pm.PSIFullAvg60,
		"psi_full_avg300":        pm.PSIFullAvg300,
		"swap_files_created":     pm.SwapFilesCreated,
		"swap_files_deleted":     pm.SwapFilesDeleted,
		"operations_total":       pm.OperationsTotal,
		"errors_total":           pm.ErrorsTotal,
		"uptime_seconds":         time.Since(pm.StartTime).Seconds(),
	}
}
