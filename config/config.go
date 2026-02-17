package config

import (
	"os"
	"strconv"

	"github.com/joho/godotenv"
)

type Settings struct {
	SleepInterval            int
	ThresholdHigh            int
	ThresholdLow             int
	SwapSizeMB               int
	MaxSwapFiles             int
	SwapEmergencyInterval    int
	SwapCreateRateLimitSec   int // Phase 1: rate limiting
	SwapDeleteRateLimitSec   int // Phase 1: rate limiting
	WorkerPoolSize           int // Phase 1: async operations
	RetryMaxAttempts         int // Phase 1: retry logic
	RetryBaseDelayMs         int // Phase 1: exponential backoff
	RetryMaxDelayMs          int // Phase 1: exponential backoff
	PSIHighThreshold         float64 // Phase 2: PSI pressure threshold
	PSILowThreshold          float64 // Phase 2: PSI recovery threshold
	// v2.1 Hybrid Adaptive - Thresholds (configurable)
	SwapAlertThreshold       int     // Trigger swap creation at SwapPercent >= this
	SwapDeleteThreshold      int     // Only delete swap if SwapPercent <= this
	SwapUseThreshold         int     // Require SwapPercent >= this before creating more
	MemDeltaThreshold        float64 // Peak detection: RAM delta > this %
	MemDeltaWindowSec        int     // Peak detection: window in seconds
	SwapDeltaThreshold       int     // Fast filling: swap delta > this %
	SwapDeltaWindowSec       int     // Fast filling: window in seconds
	PeakResponseSecSec       int     // Peak response rate-limit (seconds)
	GradualResponseSecSec    int     // Gradual response rate-limit (seconds)
	CreateDeleteHysteresisSec int   // Wait between create and delete actions
	DeleteWaitSec            int     // Wait after delete before allowing new creates
}

func LoadSettings(path string) (*Settings, error) {
	_ = godotenv.Load(path)

	return &Settings{
		SleepInterval:          getEnvInt("SWAP_SLEEP_INTERVAL", 30),
		ThresholdHigh:          getEnvInt("SWAP_THRESHOLD_HIGH", 85),
		ThresholdLow:           getEnvInt("SWAP_THRESHOLD_LOW", 40),
		SwapSizeMB:             getEnvInt("SWAP_SIZE", 4096),
		MaxSwapFiles:           getEnvInt("MAX_SWAP_FILES", 4),
		SwapEmergencyInterval:  getEnvInt("SWAP_EMERGENCY_INTERVAL", 10),
		SwapCreateRateLimitSec: getEnvInt("SWAP_CREATE_RATE_LIMIT", 300),
		SwapDeleteRateLimitSec: getEnvInt("SWAP_DELETE_RATE_LIMIT", 600),
		WorkerPoolSize:         getEnvInt("WORKER_POOL_SIZE", 1),
		RetryMaxAttempts:       getEnvInt("RETRY_MAX_ATTEMPTS", 3),
		RetryBaseDelayMs:       getEnvInt("RETRY_BASE_DELAY_MS", 500),
		RetryMaxDelayMs:        getEnvInt("RETRY_MAX_DELAY_MS", 5000),
		PSIHighThreshold:       getEnvFloat("PSI_HIGH_THRESHOLD", 80.0),
		PSILowThreshold:        getEnvFloat("PSI_LOW_THRESHOLD", 40.0),
		// v2.1 Hybrid Adaptive - Configurable thresholds
		SwapAlertThreshold:       getEnvInt("SWAP_ALERT_THRESHOLD", 85),       // Create swap at 85%
		SwapDeleteThreshold:      getEnvInt("SWAP_DELETE_THRESHOLD", 30),      // Delete swap at 30%
		SwapUseThreshold:         getEnvInt("SWAP_USE_THRESHOLD", 60),         // Need 60% usage before creating more
		MemDeltaThreshold:        getEnvFloat("MEM_DELTA_THRESHOLD", 10.0),    // Peak: 10% jump
		MemDeltaWindowSec:        getEnvInt("MEM_DELTA_WINDOW_SEC", 60),       // Peak: in 60 seconds
		SwapDeltaThreshold:       getEnvInt("SWAP_DELTA_THRESHOLD", 20),       // Fast: grow 20%
		SwapDeltaWindowSec:       getEnvInt("SWAP_DELTA_WINDOW_SEC", 45),      // Fast: in 45 seconds
		PeakResponseSecSec:       getEnvInt("PEAK_RESPONSE_SEC", 10),          // Peak: respond in 10s
		GradualResponseSecSec:    getEnvInt("GRADUAL_RESPONSE_SEC", 300),      // Gradual: wait 300s
		CreateDeleteHysteresisSec: getEnvInt("CREATE_DELETE_HYSTERESIS_SEC", 30), // 30s between create/delete
		DeleteWaitSec:            getEnvInt("DELETE_WAIT_SEC", 120),           // Wait 120s after delete
	}, nil
}

func getEnvInt(key string, defaultVal int) int {
	valStr := os.Getenv(key)
	val, err := strconv.Atoi(valStr)
	if err != nil {
		return defaultVal
	}
	return val
}

func getEnvFloat(key string, defaultVal float64) float64 {
	valStr := os.Getenv(key)
	val, err := strconv.ParseFloat(valStr, 64)
	if err != nil {
		return defaultVal
	}
	return val
}
