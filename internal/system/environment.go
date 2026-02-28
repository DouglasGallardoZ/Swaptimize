package system

import (
	"fmt"
	"log"
	"os"
	"strconv"
	"strings"
)

// EnvironmentType representa el tipo de sistema detectado
type EnvironmentType string

const (
	EnvironmentZSWAPOnly        EnvironmentType = "ZSWAP_ONLY"
	EnvironmentZRAMOnly         EnvironmentType = "ZRAM_ONLY"
	EnvironmentTraditionalSwap  EnvironmentType = "TRADITIONAL_SWAP"
	EnvironmentZSWAPWithBacking EnvironmentType = "ZSWAP_WITH_BACKING"
	EnvironmentHybridRedundant  EnvironmentType = "HYBRID_REDUNDANT"
	EnvironmentPure             EnvironmentType = "PURE_SWAPTIMIZE"
)

// SwapEnvironment contiene información sobre la configuración de swap del sistema
type SwapEnvironment struct {
	Type            EnvironmentType
	HasZSWAP        bool
	HasZRAM         bool
	ExistingSwaps   []ExistingSwapInfo
	ZSWAPCompress   string // Compresor usado (zstd, lz4, etc)
	ZSWAPPoolPct    int    // Porcentaje de memoria pool usado
	Recommendation  string // Recomendación de comportamiento
	Filesystem      *FilesystemInfo // Información del filesystem donde se crearán swaps
}

// ExistingSwapInfo contiene info de un swap existente
type ExistingSwapInfo struct {
	Path     string
	Type     string // "file", "partition", "zram"
	SizeMB   int
	UsedMB   int
	Priority int
}

// DetectEnvironment detecta la configuración de swap del sistema
func DetectEnvironment() (*SwapEnvironment, error) {
	env := &SwapEnvironment{
		ExistingSwaps: make([]ExistingSwapInfo, 0),
	}

	// 1. Detectar filesystem (donde se creará /var/lib/swaptimize)
	fsInfo, err := DetectFilesystem("/var/lib")
	if err != nil {
		log.Printf("⚠️  Error detecting filesystem: %v\n", err)
	} else {
		env.Filesystem = fsInfo
		LogFilesystemInfo(fsInfo)
	}

	// 2. Detectar ZSWAP
	zswapEnabled := CheckZSWAPEnabled()
	env.HasZSWAP = zswapEnabled
	if zswapEnabled {
		env.ZSWAPCompress = GetZSWAPCompressor()
		env.ZSWAPPoolPct = GetZSWAPPoolPercent()
		log.Printf("✓ ZSWAP detected: enabled, compressor=%s, pool=%d%%\n",
			env.ZSWAPCompress, env.ZSWAPPoolPct)
	}

	// 3. Detectar ZRAM
	zramEnabled := CheckZRAMEnabled()
	env.HasZRAM = zramEnabled
	if zramEnabled {
		log.Println("✓ ZRAM detected: enabled")
	}

	// 4. Detectar swaps existentes (no creados por Swaptimize)
	existingSwaps, err := GetExistingSwaps()
	if err != nil {
		log.Printf("⚠️ Error detecting existing swaps: %v\n", err)
	} else {
		for _, swap := range existingSwaps {
			// Filtrar swaps creados por Swaptimize
			if !strings.Contains(swap.Path, "swaptimize") {
				env.ExistingSwaps = append(env.ExistingSwaps, swap)
				log.Printf("✓ Existing swap: %s (%s) %dMB\n", swap.Path, swap.Type, swap.SizeMB)
			}
		}
	}

	// 5. Calcular recomendación
	env.Type = DetermineEnvironmentType(env)
	env.Recommendation = GenerateRecommendation(env)

	return env, nil
}

// CheckZSWAPEnabled verifica si ZSWAP está habilitado
func CheckZSWAPEnabled() bool {
	data, err := os.ReadFile("/sys/module/zswap/parameters/enabled")
	if err != nil {
		return false
	}
	value := strings.TrimSpace(string(data))
	return value == "Y" || value == "1"
}

// GetZSWAPCompressor obtiene el compresor de ZSWAP
func GetZSWAPCompressor() string {
	data, err := os.ReadFile("/sys/module/zswap/parameters/compressor")
	if err != nil {
		return "unknown"
	}
	return strings.TrimSpace(string(data))
}

// GetZSWAPPoolPercent obtiene el porcentaje del pool usado
func GetZSWAPPoolPercent() int {
	// Intentar leer desde debugfs
	debugPath := "/sys/kernel/debug/zswap/pool_total_size"

	if _, err := os.Stat(debugPath); os.IsNotExist(err) {
		// Si debugfs no está disponible, retornar estimación
		return 50
	}

	// Leer valores si están disponibles
	dataTotal, err := os.ReadFile(debugPath)
	if err != nil {
		return 50
	}

	totalStr := strings.TrimSpace(string(dataTotal))
	total, err := strconv.Atoi(totalStr)
	if err != nil || total == 0 {
		return 50
	}

	// Calcular porcentaje (esto es aproximado)
	return (total * 100) / (1024 * 1024) // Asumir 1GB límite por defecto
}

// CheckZRAMEnabled verifica si hay ZRAM activo
func CheckZRAMEnabled() bool {
	// Buscar /dev/zram* en /proc/swaps
	data, err := os.ReadFile("/proc/swaps")
	if err != nil {
		return false
	}
	return strings.Contains(string(data), "/dev/zram")
}

// GetExistingSwaps obtiene lista de swaps existentes del sistema
func GetExistingSwaps() ([]ExistingSwapInfo, error) {
	file, err := os.Open("/proc/swaps")
	if err != nil {
		return nil, fmt.Errorf("cannot open /proc/swaps: %w", err)
	}
	defer file.Close()

	var swaps []ExistingSwapInfo
	data, err := os.ReadFile("/proc/swaps")
	if err != nil {
		return nil, err
	}

	lines := strings.Split(string(data), "\n")
	for i, line := range lines {
		if i == 0 || strings.TrimSpace(line) == "" {
			continue // Skip header and empty lines
		}

		fields := strings.Fields(line)
		if len(fields) < 5 {
			continue
		}

		size, _ := strconv.Atoi(fields[2])
		used, _ := strconv.Atoi(fields[3])
		priority, _ := strconv.Atoi(fields[4])

		// Determinar tipo
		swapType := "unknown"
		if strings.HasPrefix(fields[0], "/dev/") {
			if strings.Contains(fields[0], "zram") {
				swapType = "zram"
			} else {
				swapType = "partition"
			}
		} else {
			swapType = "file"
		}

		swaps = append(swaps, ExistingSwapInfo{
			Path:     fields[0],
			Type:     swapType,
			SizeMB:   size / 1024,
			UsedMB:   used / 1024,
			Priority: priority,
		})
	}

	return swaps, nil
}

// DetermineEnvironmentType determina el tipo de ambiente basado en lo detectado
func DetermineEnvironmentType(env *SwapEnvironment) EnvironmentType {
	hasExisting := len(env.ExistingSwaps) > 0

	switch {
	// REDUNDANCIA REAL: 3 capas (ZSWAP + ZRAM + existing)
	case env.HasZSWAP && env.HasZRAM && hasExisting:
		return EnvironmentHybridRedundant // Demasiadas capas

	// ZSWAP con backing: ZSWAP + existing swap
	case env.HasZSWAP && hasExisting && !env.HasZRAM:
		return EnvironmentZSWAPWithBacking // ZSWAP ya tiene backing

	// ZSWAP sin backing
	case env.HasZSWAP && !env.HasZRAM && !hasExisting:
		return EnvironmentZSWAPOnly // ZSWAP sin backing

	// ZRAM (con o sin swap existente) - ZRAM es válido por sí solo
	case env.HasZRAM && !env.HasZSWAP:
		return EnvironmentZRAMOnly // ZRAM como compressión, swap heredado es opcional

	// Swap tradicional sin ZSWAP ni ZRAM
	case hasExisting && !env.HasZSWAP && !env.HasZRAM:
		return EnvironmentTraditionalSwap // Swap tradicional

	// Sistema sin limitaciones
	default:
		return EnvironmentPure // Sistema limpio
	}
}

// GenerateRecommendation genera recomendación basada en el tipo
func GenerateRecommendation(env *SwapEnvironment) string {
	switch env.Type {
	case EnvironmentZSWAPOnly:
		return "ZSWAP detected without backing store. Swaptimize will create swap files as backing store."
	case EnvironmentZRAMOnly:
		return "ZRAM detected. Swaptimize will expand when ZRAM is saturated."
	case EnvironmentTraditionalSwap:
		return "Traditional swap file/partition detected. Swaptimize will expand only if usage exceeds 95%."
	case EnvironmentZSWAPWithBacking:
		return "ZSWAP with backing store already configured. Swaptimize will monitor only, no file creation."
	case EnvironmentHybridRedundant:
		return "WARNING: Redundant swap configuration detected. This may cause thrashing. Disabling Swaptimize file creation."
	case EnvironmentPure:
		return "No existing swap detected. Swaptimize will manage swap dynamically from scratch."
	default:
		return "Unknown environment"
	}
}

// LogEnvironment logs detailed information about detected environment
func LogEnvironment(env *SwapEnvironment) {
	log.Println("╔════════════════════════════════════════╗")
	log.Println("║   SWAP ENVIRONMENT DETECTION REPORT    ║")
	log.Println("╚════════════════════════════════════════╝")
	log.Printf("Type: %s\n", env.Type)
	log.Printf("ZSWAP: %v", env.HasZSWAP)
	if env.HasZSWAP {
		log.Printf(" (compressor=%s, pool=%d%%)", env.ZSWAPCompress, env.ZSWAPPoolPct)
	}
	log.Println()
	log.Printf("ZRAM: %v\n", env.HasZRAM)

	if len(env.ExistingSwaps) > 0 {
		log.Printf("Existing swaps: %d\n", len(env.ExistingSwaps))
		for _, swap := range env.ExistingSwaps {
			log.Printf("  ├─ %s (%s): %dMB / %dMB, priority=%d\n",
				swap.Path, swap.Type, swap.UsedMB, swap.SizeMB, swap.Priority)
		}
	} else {
		log.Println("Existing swaps: 0")
	}

	log.Printf("Recommendation: %s\n", env.Recommendation)
	log.Println("╚════════════════════════════════════════╝")
}
