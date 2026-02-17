package health

import (
	"log"
	"sync"
	"time"
)

// CheckStatus definne el estado de un chequeo
type CheckStatus int

const (
	CheckUnknown CheckStatus = iota
	CheckPassing
	CheckWarning
	CheckFailing
)

// Check representa un chequeo de salud
type Check struct {
	Name    string
	Status  CheckStatus
	Message string
	LastRun time.Time
}

// HealthMonitor gestiona los chequeos de salud del sistema
type HealthMonitor struct {
	mu     sync.RWMutex
	checks map[string]*Check

	// Callbacks
	onStatusChange func(name string, oldStatus, newStatus CheckStatus)
}

// NewHealthMonitor crea un nuevo monitor de salud
func NewHealthMonitor() *HealthMonitor {
	return &HealthMonitor{
		checks: make(map[string]*Check),
	}
}

// RegisterCheck registra un nuevo chequeo
func (hm *HealthMonitor) RegisterCheck(name string) {
	hm.mu.Lock()
	defer hm.mu.Unlock()

	hm.checks[name] = &Check{
		Name:    name,
		Status:  CheckUnknown,
		Message: "Not yet evaluated",
		LastRun: time.Now(),
	}
}

// UpdateCheck actualiza el estado de un chequeo
func (hm *HealthMonitor) UpdateCheck(name string, status CheckStatus, message string) {
	hm.mu.Lock()
	defer hm.mu.Unlock()

	check, exists := hm.checks[name]
	if !exists {
		check = &Check{Name: name}
		hm.checks[name] = check
	}

	oldStatus := check.Status
	check.Status = status
	check.Message = message
	check.LastRun = time.Now()

	hm.mu.Unlock()

	// Llamar callback fuera del lock
	if oldStatus != status && hm.onStatusChange != nil {
		hm.onStatusChange(name, oldStatus, status)
	}

	hm.mu.Lock()
}

// SetStatusChangeCallback define el callback de cambio de estado
func (hm *HealthMonitor) SetStatusChangeCallback(f func(name string, oldStatus, newStatus CheckStatus)) {
	hm.mu.Lock()
	defer hm.mu.Unlock()
	hm.onStatusChange = f
}

// GetCheckStatus retorna el estado de un chequeo
func (hm *HealthMonitor) GetCheckStatus(name string) (CheckStatus, string) {
	hm.mu.RLock()
	defer hm.mu.RUnlock()

	check, exists := hm.checks[name]
	if !exists {
		return CheckUnknown, "Check not found"
	}

	return check.Status, check.Message
}

// GetAllChecks retorna todos los chequeos
func (hm *HealthMonitor) GetAllChecks() []*Check {
	hm.mu.RLock()
	defer hm.mu.RUnlock()

	checks := make([]*Check, 0, len(hm.checks))
	for _, check := range hm.checks {
		checks = append(checks, &Check{
			Name:    check.Name,
			Status:  check.Status,
			Message: check.Message,
			LastRun: check.LastRun,
		})
	}

	return checks
}

// GetOverallStatus retorna el estado general
func (hm *HealthMonitor) GetOverallStatus() CheckStatus {
	hm.mu.RLock()
	defer hm.mu.RUnlock()

	if len(hm.checks) == 0 {
		return CheckUnknown
	}

	// Prioridad: Failing > Warning > Passing > Unknown
	overallStatus := CheckPassing

	for _, check := range hm.checks {
		if check.Status == CheckFailing {
			return CheckFailing
		}
		if check.Status == CheckWarning {
			overallStatus = CheckWarning
		}
		if check.Status == CheckUnknown && overallStatus == CheckPassing {
			overallStatus = CheckUnknown
		}
	}

	return overallStatus
}

// StatusString retorna la representación string del estado
func (status CheckStatus) String() string {
	switch status {
	case CheckPassing:
		return "passing"
	case CheckWarning:
		return "warning"
	case CheckFailing:
		return "failing"
	case CheckUnknown:
		return "unknown"
	default:
		return "unknown"
	}
}

// LogAllChecks registra todos los chequeos
func (hm *HealthMonitor) LogAllChecks() {
	checks := hm.GetAllChecks()

	log.Println("╔════════════════════════════════════════╗")
	log.Println("║        SYSTEM HEALTH STATUS            ║")
	log.Println("╠════════════════════════════════════════╣")

	for _, check := range checks {
		emoji := "❓"
		switch check.Status {
		case CheckPassing:
			emoji = "✅"
		case CheckWarning:
			emoji = "⚠️ "
		case CheckFailing:
			emoji = "❌"
		}

		log.Printf("║ %s %-35s║\n", emoji, check.Name)
		log.Printf("║   Message: %-30s║\n", check.Message)
	}

	overall := hm.GetOverallStatus().String()
	emoji := "❓"
	switch hm.GetOverallStatus() {
	case CheckPassing:
		emoji = "✅"
	case CheckWarning:
		emoji = "⚠️ "
	case CheckFailing:
		emoji = "❌"
	}

	log.Printf("║ %s Overall: %-29s║\n", emoji, overall)
	log.Println("╚════════════════════════════════════════╝")
}

// StandardChecks contiene los chequeos estándar de Swaptimize
const (
	CheckPSIMonitoringAvailable   = "psi_monitoring"
	CheckZSWAPConfigured          = "zswap_configured"
	CheckZRAMConfigured           = "zram_configured"
	CheckSwapAvailable            = "swap_available"
	CheckSwapFilesCreatable       = "swap_files_creatable"
	CheckDiskSpaceAdequate        = "disk_space_adequate"
	CheckMemoryPressureNormal     = "memory_pressure_normal"
	CheckWorkerPoolHealthy        = "worker_pool_healthy"
	CheckMetricsServerRunning     = "metrics_server"
	CheckNoDaemonConflict         = "no_daemon_conflict"
)

// InitializeStandardChecks inicializa los chequeos estándar
func (hm *HealthMonitor) InitializeStandardChecks() {
	hm.RegisterCheck(CheckPSIMonitoringAvailable)
	hm.RegisterCheck(CheckZSWAPConfigured)
	hm.RegisterCheck(CheckZRAMConfigured)
	hm.RegisterCheck(CheckSwapAvailable)
	hm.RegisterCheck(CheckSwapFilesCreatable)
	hm.RegisterCheck(CheckDiskSpaceAdequate)
	hm.RegisterCheck(CheckMemoryPressureNormal)
	hm.RegisterCheck(CheckWorkerPoolHealthy)
	hm.RegisterCheck(CheckMetricsServerRunning)
	hm.RegisterCheck(CheckNoDaemonConflict)

	log.Println("📋 Standard health checks initialized")
}
