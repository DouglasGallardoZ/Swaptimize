package system

import (
	"log"
)

// ProfileManager gestiona la configuración de diferentes perfiles
type ProfileManager struct {
	currentProfile ConfigProfile
	environment    *SwapEnvironment
}

// NewProfileManager crea un nuevo gestor de perfiles
func NewProfileManager(env *SwapEnvironment) *ProfileManager {
	pm := &ProfileManager{
		environment: env,
	}

	// Seleccionar perfil inicial
	profileName, profile := RecommendProfile(env)
	pm.currentProfile = profile

	log.Printf("📋 Profile selected: %s\n", profileName)
	LogProfile(profile)

	return pm
}

// GetCurrentProfile retorna el perfil actual
func (pm *ProfileManager) GetCurrentProfile() ConfigProfile {
	return pm.currentProfile
}

// GetEnvironment retorna el ambiente detectado
func (pm *ProfileManager) GetEnvironment() *SwapEnvironment {
	return pm.environment
}

// AdaptiveThresholdHigh retorna el thresholdHigh adaptado
func (pm *ProfileManager) AdaptiveThresholdHigh() int {
	threshold := pm.currentProfile.ThresholdHigh

	// Adaptar según ambiente
	switch pm.environment.Type {
	case EnvironmentZSWAPOnly:
		// ZSWAP con compresión es muy eficiente
		threshold = 90 // Más alto, ZSWAP puede comprimir
	case EnvironmentZRAMOnly:
		// ZRAM es RAM comprimida, necesita backup swap conservador
		threshold = 85 // Solo crear backup cuando ZRAM casi saturado
	case EnvironmentTraditionalSwap:
		// Swap tradicional: muy conservador
		threshold = 95
	case EnvironmentPure:
		// Sin limitaciones, usar recomendado
		threshold = pm.currentProfile.ThresholdHigh
	}

	return threshold
}

// AdaptiveMaxSwapFiles retorna maxSwapFiles adaptado
func (pm *ProfileManager) AdaptiveMaxSwapFiles() int {
	maxFiles := pm.currentProfile.MaxSwapFiles

	// Reducir si hay poco espacio disco
	// En un caso real, habría que verificar GetDiskFree()
	// Por ahora solo aplicar basado en environment

	switch pm.environment.Type {
	case EnvironmentZSWAPWithBacking, EnvironmentHybridRedundant:
		maxFiles = 0 // No crear archivos
	case EnvironmentTraditionalSwap:
		maxFiles = 1 // Minimal expansion
	// Para ZRAM_ONLY y otros, usar valor del perfil (control inteligente en run.go)
	}

	return maxFiles
}

// ShouldAllowCreation verifica si se permite crear archivos swap
func (pm *ProfileManager) ShouldAllowCreation() bool {
	switch pm.environment.Type {
	case EnvironmentZSWAPWithBacking:
		return false
	case EnvironmentHybridRedundant:
		return false
	default:
		return pm.currentProfile.MaxSwapFiles > 0
	}
}

// GetModeDescription retorna descripción del modo actual
func (pm *ProfileManager) GetModeDescription() string {
	prefix := "Mode: "
	switch pm.environment.Type {
	case EnvironmentZSWAPOnly:
		return prefix + "ZSWAP backing store provider"
	case EnvironmentZRAMOnly:
		return prefix + "ZRAM complementer"
	case EnvironmentTraditionalSwap:
		return prefix + "Swap file expander"
	case EnvironmentZSWAPWithBacking:
		return prefix + "Monitor-only (ZSWAP + backing)"
	case EnvironmentHybridRedundant:
		return prefix + "Disabled (redundant config)"
	case EnvironmentPure:
		return prefix + "Full dynamic swap management"
	default:
		return prefix + "Unknown"
	}
}

// LogConfiguration registra la configuración actual
func (pm *ProfileManager) LogConfiguration() {
	profile := pm.currentProfile
	env := pm.environment
	
	log.Println("╔════════════════════════════════════════════════════════════╗")
	log.Println("║        PHASE 2: ENVIRONMENT DETECTION & PROFILE            ║")
	log.Println("╠════════════════════════════════════════════════════════════╣")
	log.Printf("║ Environment Detected: %-37s║\n", string(env.Type))
	log.Printf("║ Recommendation: %-44s║\n", env.Recommendation)
	log.Printf("║ ZSWAP Available: %v%-40s║\n", env.HasZSWAP, "")
	log.Printf("║ ZRAM Available: %v%-41s║\n", env.HasZRAM, "")
	log.Println("╠════════════════════════════════════════════════════════════╣")
	log.Printf("║ Profile Selected: %-42s║\n", profile.Name)
	log.Printf("║ Mode: %-55s║\n", pm.GetModeDescription())
	log.Println("╠════════════════════════════════════════════════════════════╣")
	log.Printf("║ ThresholdHigh: %d%%%-45s║\n", pm.AdaptiveThresholdHigh(), "")
	log.Printf("║ ThresholdLow: %d%%%-46s║\n", profile.ThresholdLow, "")
	log.Printf("║ MaxSwapFiles: %d%-46s║\n", pm.AdaptiveMaxSwapFiles(), "")
	log.Printf("║ SwapSize: %d MB%-42s║\n", profile.SwapSize, "")
	log.Printf("║ AllowCreation: %v%-45s║\n", pm.ShouldAllowCreation(), "")
	log.Println("╚════════════════════════════════════════════════════════════╝")
}
