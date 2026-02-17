package system

import (
	"log"

	"Swaptimize/config"
)

// ConfigProfile define un perfil de configuración
type ConfigProfile struct {
	Name                   string
	SwapSize               int
	MaxSwapFiles           int
	ThresholdHigh          int
	ThresholdLow           int
	CreateRateLimit        int
	DeleteRateLimit        int
	Description            string
}

// PredefinedProfiles son perfiles recomendados por tipo de máquina
var PredefinedProfiles = map[string]ConfigProfile{
	"laptop": {
		Name:            "laptop",
		SwapSize:        512,
		MaxSwapFiles:    6,
		ThresholdHigh:   80,
		ThresholdLow:    35,
		CreateRateLimit: 300,
		DeleteRateLimit: 600,
		Description:     "Conservative profile for laptops (< 16GB RAM)",
	},
	"workstation": {
		Name:            "workstation",
		SwapSize:        1024,
		MaxSwapFiles:    8,
		ThresholdHigh:   85,
		ThresholdLow:    40,
		CreateRateLimit: 300,
		DeleteRateLimit: 600,
		Description:     "Balanced profile for workstations (16-32GB RAM)",
	},
	"server": {
		Name:            "server",
		SwapSize:        2048,
		MaxSwapFiles:    10,
		ThresholdHigh:   82,
		ThresholdLow:    38,
		CreateRateLimit: 300,
		DeleteRateLimit: 600,
		Description:     "Stable profile for non-critical servers (> 32GB RAM)",
	},
}

// RecommendProfile recomienda un perfil basado en configuración del ambiente
func RecommendProfile(env *SwapEnvironment) (string, ConfigProfile) {
	// Elegir perfil base basado en tipo de environment
	baseProfile := selectBaseProfile(env)

	// Crear perfil ajustado
	adjustedProfile := adjustProfileForEnvironment(baseProfile, env)

	return baseProfile, adjustedProfile
}

// selectBaseProfile selecciona uno de los perfiles predefinidos
func selectBaseProfile(env *SwapEnvironment) string {
	switch env.Type {
	case EnvironmentZSWAPWithBacking, EnvironmentHybridRedundant:
		// Minimal: no vamos a crear archivos
		return "laptop"
	case EnvironmentTraditionalSwap:
		// Conservative: solo expandir si es necesario
		return "laptop"
	case EnvironmentZSWAPOnly, EnvironmentZRAMOnly:
		// Balanced: normal activity
		return "workstation"
	case EnvironmentPure:
		// Aggressive: full control
		return "workstation"
	default:
		return "workstation"
	}
}

// adjustProfileForEnvironment ajusta el perfil según el environment específico
// IMPORTANTE: Solo ajusta THRESHOLDS, nunca SwapSize o MaxSwapFiles
func adjustProfileForEnvironment(baseName string, env *SwapEnvironment) ConfigProfile {
	profile := PredefinedProfiles[baseName]

	// SOLO ajustar thresholds según environment type, mantener SwapSize y MaxSwapFiles del perfil
	switch env.Type {
	case EnvironmentZSWAPOnly:
		// ZSWAP es muy eficiente con compresión
		profile.ThresholdHigh = 85
		profile.ThresholdLow = 50
		profile.Description = "ZSWAP backing store profile"

	case EnvironmentZRAMOnly:
		// ZRAM es RAM comprimida, ser algo más agresivo
		profile.ThresholdHigh = 75
		profile.ThresholdLow = 45
		profile.Description = "ZRAM complementar profile"

	case EnvironmentTraditionalSwap:
		// Swap tradicional es lento, muy conservador
		profile.ThresholdHigh = 95 // Solo crear si base > 95%
		profile.ThresholdLow = 85  // Sticky, no delete fácilmente
		profile.CreateRateLimit = 600 // Muy conservador
		profile.Description = "Traditional swap expansion profile"

	case EnvironmentZSWAPWithBacking:
		profile.MaxSwapFiles = 0 // NO crear archivos, solo monitorear
		profile.Description = "Monitor-only mode (ZSWAP + backing exists)"

	case EnvironmentHybridRedundant:
		profile.MaxSwapFiles = 0 // NO crear archivos
		profile.Description = "WARN: hybrid redundant, disabled"

	case EnvironmentPure:
		// Sin cambios, usar como está
		profile.Description = "Pure Swaptimize mode (full control)"
	}

	return profile
}

// ApplyProfile aplica un perfil de configuración a los settings
func ApplyProfile(profile ConfigProfile, settings *config.Settings) {
	settings.SwapSizeMB = profile.SwapSize
	settings.MaxSwapFiles = profile.MaxSwapFiles
	settings.ThresholdHigh = profile.ThresholdHigh
	settings.ThresholdLow = profile.ThresholdLow
	settings.SwapCreateRateLimitSec = profile.CreateRateLimit
	settings.SwapDeleteRateLimitSec = profile.DeleteRateLimit

	log.Printf("✓ Applied profile '%s': %s\n", profile.Name, profile.Description)
}

// LogProfile logs detailed profile information
func LogProfile(profile ConfigProfile) {
	log.Println("╔════════════════════════════════════════╗")
	log.Printf("║  CONFIGURATION PROFILE: %-16s║\n", profile.Name)
	log.Println("╠════════════════════════════════════════╣")
	log.Printf("║ SwapSize: %d MB%-23s║\n", profile.SwapSize, "")
	log.Printf("║ MaxFiles: %d%-27s║\n", profile.MaxSwapFiles, "")
	log.Printf("║ ThresholdHigh: %d%%%-20s║\n", profile.ThresholdHigh, "")
	log.Printf("║ ThresholdLow: %d%%%-21s║\n", profile.ThresholdLow, "")
	log.Printf("║ CreateRateLimit: %ds%-17s║\n", profile.CreateRateLimit, "")
	log.Printf("║ DeleteRateLimit: %ds%-17s║\n", profile.DeleteRateLimit, "")
	log.Println("╠════════════════════════════════════════╣")
	log.Printf("║ %s%-36s║\n", profile.Description, "")
	log.Println("╚════════════════════════════════════════╝")
}
