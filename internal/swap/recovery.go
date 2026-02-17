package swap

import (
	"fmt"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"
)

// SwapRecovery maneja la limpieza de swaps huérfanos
type SwapRecovery struct {
	Verifier *SwapVerifier
}

// NewSwapRecovery crea un nuevo gestor de recovery
func NewSwapRecovery() *SwapRecovery {
	return &SwapRecovery{
		Verifier: NewSwapVerifier(),
	}
}

// FindOrphanSwaps detecta archivos swap que no están activos
func (sr *SwapRecovery) FindOrphanSwaps() ([]string, error) {
	// Listar archivos en /var/lib/swaptimize
	entries, err := os.ReadDir(swapDir)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil // Directorio no existe, no hay huérfanos
		}
		return nil, fmt.Errorf("cannot read swap directory: %w", err)
	}

	// Obtener swaps activos
	activeSwaps, err := sr.Verifier.ListActiveSwaps()
	if err != nil {
		return nil, fmt.Errorf("cannot list active swaps: %w", err)
	}

	// Crear mapa de paths activos
	activePaths := make(map[string]bool)
	for _, swap := range activeSwaps {
		activePaths[swap.Path] = true
	}

	var orphans []string

	// Buscar huérfanos
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}

		filePath := filepath.Join(swapDir, entry.Name())

		// Solo marcar como huérfano si es un archivo swap de Swaptimize
		if strings.HasPrefix(entry.Name(), "swap-") && !activePaths[filePath] {
			orphans = append(orphans, filePath)
		}
	}

	return orphans, nil
}

// CleanupOrphanSwaps elimina los archivos swap huérfanos
func (sr *SwapRecovery) CleanupOrphanSwaps() (int, error) {
	orphans, err := sr.FindOrphanSwaps()
	if err != nil {
		return 0, err
	}

	if len(orphans) == 0 {
		return 0, nil
	}

	log.Printf("🧹 Found %d orphan swap files\n", len(orphans))

	cleaned := 0

	for _, filePath := range orphans {
		log.Printf("🧽 Cleaning orphan: %s\n", filePath)

		// Intentar desactivar por si acaso
		_ = exec.Command("swapoff", filePath).Run()

		// Eliminar archivo
		if err := os.Remove(filePath); err != nil {
			log.Printf("⚠️  Failed to remove orphan %s: %v\n", filePath, err)
			continue
		}

		cleaned++
		log.Printf("✓ Removed orphan: %s\n", filePath)
	}

	return cleaned, nil
}

// VerifySwapConsistency verifica consistencia entre fs y /proc/swaps
func (sr *SwapRecovery) VerifySwapConsistency() map[string]interface{} {
	result := map[string]interface{}{
		"timestamp":          time.Now(),
		"files_on_disk":      0,
		"files_in_proc":      0,
		"orphan_files":       0,
		"consistency_ok":     true,
		"issues":             []string{},
	}

	// Contar archivos en disco
	entries, err := os.ReadDir(swapDir)
	if err != nil && !os.IsNotExist(err) {
		result["consistency_ok"] = false
		result["issues"] = append(result["issues"].([]string), fmt.Sprintf("Cannot read swap dir: %v", err))
		return result
	}

	filesOnDisk := 0
	for _, entry := range entries {
		if !entry.IsDir() && strings.HasPrefix(entry.Name(), "swap-") {
			filesOnDisk++
		}
	}
	result["files_on_disk"] = filesOnDisk

	// Contar archivos en /proc/swaps
	count, err := sr.Verifier.CountActiveSwaps()
	if err != nil {
		result["consistency_ok"] = false
		result["issues"] = append(result["issues"].([]string), fmt.Sprintf("Cannot count active swaps: %v", err))
		return result
	}
	result["files_in_proc"] = count

	// Detectar huérfanos
	orphans, err := sr.FindOrphanSwaps()
	if err != nil {
		result["consistency_ok"] = false
		result["issues"] = append(result["issues"].([]string), fmt.Sprintf("Cannot find orphans: %v", err))
		return result
	}
	result["orphan_files"] = len(orphans)

	// Verificar consistencia
	if filesOnDisk != count {
		result["consistency_ok"] = false
		result["issues"] = append(result["issues"].([]string),
			fmt.Sprintf("Mismatch: %d files on disk, %d in /proc/swaps", filesOnDisk, count))
	}

	if len(orphans) > 0 {
		result["consistency_ok"] = false
		for _, orphan := range orphans {
			result["issues"] = append(result["issues"].([]string),
				fmt.Sprintf("Orphan file found: %s", orphan))
		}
	}

	return result
}

// CleanupOnStartup realiza limpieza de archivos residuales al arrancar
func CleanupOnStartup(verifier *SwapVerifier) {
	log.Println("🧹 Running startup cleanup...")

	recovery := NewSwapRecovery()
	recovery.Verifier = verifier

	// Intentar limpiar huérfanos
	cleaned, err := recovery.CleanupOrphanSwaps()
	if err != nil {
		log.Printf("⚠️  Error during cleanup: %v\n", err)
	} else if cleaned > 0 {
		log.Printf("✓ Cleaned %d orphan swaps\n", cleaned)
	} else {
		log.Println("✓ No orphans found, everything clean")
	}

	// Verificar consistencia
	consistency := recovery.VerifySwapConsistency()
	if consistency["consistency_ok"].(bool) {
		log.Printf("✓ Swap consistency check passed\n")
	} else {
		issues := consistency["issues"].([]string)
		for _, issue := range issues {
			log.Printf("⚠️  %s\n", issue)
		}
	}
}
