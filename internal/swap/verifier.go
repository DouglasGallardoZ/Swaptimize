package swap

import (
	"bufio"
	"fmt"
	"log"
	"os"
	"strconv"
	"strings"
	"sync"
	"time"
)

// SwapInfo contiene información sobre un archivo de swap
type SwapInfo struct {
	Path     string
	Type     string
	SizeKB   int64
	UsedKB   int64
	Priority int
}

// SwapVerifier verifica el estado de archivos swap en /proc/swaps
type SwapVerifier struct {
	mu       sync.RWMutex
	cache    []SwapInfo
	cacheAge time.Time
	cacheTTL time.Duration
}

// NewSwapVerifier crea un nuevo verificador de swap
func NewSwapVerifier() *SwapVerifier {
	return &SwapVerifier{
		cacheTTL: 5 * time.Second,
	}
}

// ListActiveSwaps retorna todos los archivos swap activos
func (sv *SwapVerifier) ListActiveSwaps() ([]SwapInfo, error) {
	sv.mu.Lock()
	defer sv.mu.Unlock()

	// Usar cache si todavía es válido
	if !sv.cacheAge.IsZero() && time.Since(sv.cacheAge) < sv.cacheTTL {
		return sv.cache, nil
	}

	// Leer /proc/swaps
	file, err := os.Open("/proc/swaps")
	if err != nil {
		return nil, fmt.Errorf("cannot open /proc/swaps: %w", err)
	}
	defer file.Close()

	var swaps []SwapInfo
	scanner := bufio.NewScanner(file)

	// Skip header
	if !scanner.Scan() {
		return nil, fmt.Errorf("empty /proc/swaps file")
	}

	for scanner.Scan() {
		line := scanner.Text()
		fields := strings.Fields(line)

		if len(fields) < 5 {
			continue
		}

		size, _ := strconv.ParseInt(fields[2], 10, 64)
		used, _ := strconv.ParseInt(fields[3], 10, 64)
		priority, _ := strconv.Atoi(fields[4])

		swaps = append(swaps, SwapInfo{
			Path:     fields[0],
			Type:     fields[1],
			SizeKB:   size,
			UsedKB:   used,
			Priority: priority,
		})
	}

	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("error scanning /proc/swaps: %w", err)
	}

	// Actualizar cache
	sv.cache = swaps
	sv.cacheAge = time.Now()

	return swaps, nil
}

// IsSwapFileActive verifica si un archivo swap está activo en /proc/swaps
func (sv *SwapVerifier) IsSwapFileActive(filePath string) bool {
	swaps, err := sv.ListActiveSwaps()
	if err != nil {
		log.Printf("⚠️  Error verifying swap state: %v\n", err)
		return false
	}

	for _, swap := range swaps {
		if swap.Path == filePath {
			return true
		}
	}

	return false
}

// CountActiveSwaps retorna el número de archivos swap activos de Swaptimize
func (sv *SwapVerifier) CountActiveSwaps() (int, error) {
	swaps, err := sv.ListActiveSwaps()
	if err != nil {
		return 0, err
	}

	count := 0
	for _, swap := range swaps {
		// Contar solo swaps creados por Swaptimize
		if strings.HasPrefix(swap.Path, "/var/lib/swaptimize/swap-") {
			count++
		}
	}

	return count, nil
}

// GetTotalSwapUsage retorna el total de swap usado
func (sv *SwapVerifier) GetTotalSwapUsage() (int64, int64, error) {
	swaps, err := sv.ListActiveSwaps()
	if err != nil {
		return 0, 0, err
	}

	var totalSize, totalUsed int64

	for _, swap := range swaps {
		totalSize += swap.SizeKB
		totalUsed += swap.UsedKB
	}

	return totalSize, totalUsed, nil
}

// GetSwapPercent retorna el porcentaje de swap usado
func (sv *SwapVerifier) GetSwapPercent() (int, error) {
	totalSize, totalUsed, err := sv.GetTotalSwapUsage()
	if err != nil {
		return 0, err
	}

	if totalSize == 0 {
		return 0, nil
	}

	return int((totalUsed * 100) / totalSize), nil
}

// VerifyAndLog verifica consistencia y retorna un reporte
func (sv *SwapVerifier) VerifyAndLog() {
	swaps, err := sv.ListActiveSwaps()
	if err != nil {
		log.Printf("⚠️  Verification error: %v\n", err)
		return
	}

	log.Printf("📊 Swap status (%d active):\n", len(swaps))
	
	totalSize := int64(0)
	totalUsed := int64(0)
	
	for _, swap := range swaps {
		totalSize += swap.SizeKB
		totalUsed += swap.UsedKB
		
		percent := 0
		if swap.SizeKB > 0 {
			percent = int((swap.UsedKB * 100) / swap.SizeKB)
		}
		
		log.Printf("  ├─ %s (%s): %dMB / %dMB (%d%%), priority=%d\n",
			swap.Path,
			swap.Type,
			swap.UsedKB/1024,
			swap.SizeKB/1024,
			percent,
			swap.Priority,
		)
	}

	if totalSize > 0 {
		totalPercent := int((totalUsed * 100) / totalSize)
		log.Printf("  └─ TOTAL: %dMB / %dMB (%d%%)\n",
			totalUsed/1024,
			totalSize/1024,
			totalPercent,
		)
	}
}

// ClearCache limpia el cache
func (sv *SwapVerifier) ClearCache() {
	sv.mu.Lock()
	defer sv.mu.Unlock()
	sv.cache = nil
	sv.cacheAge = time.Time{}
}
