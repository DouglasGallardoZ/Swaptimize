package system

import (
	"fmt"
	"log"

	"github.com/shirou/gopsutil/v3/disk"
)

// ValidationResult contiene información sobre la validación
type ValidationResult struct {
	CanCreate      bool
	Required       uint64
	Available      uint64
	BufferRequired uint64
	Message        string
}

// ValidateSwapCreation verifica si es posible crear un archivo swap
func ValidateSwapCreation(swapPath string, sizeMB int) *ValidationResult {
	result := &ValidationResult{
		Required:       uint64(sizeMB) * 1024 * 1024,
		BufferRequired: uint64(sizeMB) * 1024 * 1024 * 2, // 2x buffer
	}

	// Detectar filesystem para validaciones específicas
	fsInfo, _ := DetectFilesystem("/var/lib")

	// Para btrfs, requerir más buffer debido a metadatos de COW
	if fsInfo != nil && fsInfo.IsBtrfs() {
		result.BufferRequired = uint64(sizeMB) * 1024 * 1024 * 3 // 3x buffer for btrfs safety
	}

	// Obtener uso de disco de la partición where swapPath lives
	diskUsage, err := disk.Usage("/var/lib/swaptimize")
	if err != nil {
		// Fallback to root
		diskUsage, err = disk.Usage("/")
		if err != nil {
			result.Message = fmt.Sprintf("Error getting disk usage: %v", err)
			log.Printf("❌ %s\n", result.Message)
			return result
		}
	}

	result.Available = diskUsage.Free

	// Verificar espacio: necesitamos buffer (2x o 3x según filesystem)
	if result.Available < result.BufferRequired {
		result.CanCreate = false
		bufferMB := result.BufferRequired / (1024 * 1024)
		availMB := result.Available / (1024 * 1024)
		
		fsLabel := ""
		if fsInfo != nil && fsInfo.IsBtrfs() {
			fsLabel = " (btrfs requires 3x for COW metadata) "
		}
		
		result.Message = fmt.Sprintf(
			"Insufficient disk space%s: need %dMB, have %dMB",
			fsLabel,
			bufferMB,
			availMB,
		)
		log.Printf("⚠️  %s\n", result.Message)
		return result
	}

	result.CanCreate = true
	result.Message = fmt.Sprintf(
		"✓ Disk validation passed: %dMB free (need %dMB)",
		result.Available/(1024*1024),
		result.BufferRequired/(1024*1024),
	)

	return result
}

// ValidateSwapDeletion verifica que sea seguro borrar un archivo swap
func ValidateSwapDeletion(swapPath string) *ValidationResult {
	result := &ValidationResult{
		Message: "Deletion validation not implemented yet",
	}
	// Por ahora acepta cualquier borrado (será validado en /proc/swaps)
	result.CanCreate = true
	return result
}

// GetDiskFree retorna el espacio libre en disco (en MB)
func GetDiskFree() (uint64, error) {
	diskUsage, err := disk.Usage("/")
	if err != nil {
		return 0, err
	}
	return diskUsage.Free / (1024 * 1024), nil
}
