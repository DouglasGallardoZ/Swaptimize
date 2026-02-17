package swap

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"Swaptimize/internal/system"
)

// Ruta base para los archivos swap (puede hacerse configurable)
const swapDir = "/var/lib/swaptimize"

// Crea un archivo swap del tamaño especificado (en MB)
func CreateSwapFile(id int, sizeMB int) error {
	filePath := filepath.Join(swapDir, fmt.Sprintf("swap-%d", id))

	if err := os.MkdirAll(swapDir, 0755); err != nil {
		return fmt.Errorf("no se pudo crear directorio swap: %w", err)
	}

	log.Printf("🛠️ Creando archivo swap en %s (%dMB)", filePath, sizeMB)

	if _, err := os.Stat(filePath); err == nil {
		log.Printf("⚠️ El archivo swap ya existe: %s", filePath)
		return nil
	}

	// Crear archivo con tamaño exacto
	cmd := exec.Command("fallocate", "-l", fmt.Sprintf("%dM", sizeMB), filePath)
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("error al asignar espacio: %w", err)
	}

	// Preparar el archivo como swap
	if err := exec.Command("mkswap", filePath).Run(); err != nil {
		return fmt.Errorf("error al inicializar swap: %w", err)
	}

	// Activar el archivo de swap
	if err := exec.Command("swapon", filePath).Run(); err != nil {
		return fmt.Errorf("error al activar swap: %w", err)
	}

	log.Printf("✅ Archivo swap activado: %s", filePath)
	return nil
}

// CreateSwapFileWithRetry crea un archivo swap con reintentos
func CreateSwapFileWithRetry(ctx context.Context, idStr string, sizeMB int, retries int) error {
	filePath := filepath.Join(swapDir, fmt.Sprintf("swap-%s", idStr))

	// Validar espacio disco
	validation := system.ValidateSwapCreation(filePath, sizeMB)
	if !validation.CanCreate {
		log.Printf("❌ %s\n", validation.Message)
		return fmt.Errorf("disk validation failed: %s", validation.Message)
	}

	log.Printf("%s\n", validation.Message)

	// Crear directorio
	if err := os.MkdirAll(swapDir, 0755); err != nil {
		return fmt.Errorf("no se pudo crear directorio swap: %w", err)
	}

	// Verificar si ya existe
	if _, err := os.Stat(filePath); err == nil {
		log.Printf("⚠️ El archivo swap ya existe: %s", filePath)
		return nil
	}

	// Backoff strategy
	backoff := system.NewExponentialBackoff(retries, 500*time.Millisecond, 5*time.Second)

	// Paso 1: fallocate
	err := backoff.DoWithRetry(ctx, fmt.Sprintf("fallocate %s", filePath), func() error {
		cmd := exec.CommandContext(ctx, "fallocate", "-l", fmt.Sprintf("%dM", sizeMB), filePath)
		return cmd.Run()
	})
	if err != nil {
		return fmt.Errorf("fallocate failed: %w", err)
	}

	// Paso 2: mkswap
	err = backoff.DoWithRetry(ctx, fmt.Sprintf("mkswap %s", filePath), func() error {
		cmd := exec.CommandContext(ctx, "mkswap", filePath)
		return cmd.Run()
	})
	if err != nil {
		// Limpiar archivo si mkswap falla
		_ = os.Remove(filePath)
		return fmt.Errorf("mkswap failed: %w", err)
	}

	// Paso 3: swapon
	err = backoff.DoWithRetry(ctx, fmt.Sprintf("swapon %s", filePath), func() error {
		cmd := exec.CommandContext(ctx, "swapon", filePath)
		return cmd.Run()
	})
	if err != nil {
		// Limpiar archivo si swapon falla
		_ = os.Remove(filePath)
		return fmt.Errorf("swapon failed: %w", err)
	}

	log.Printf("✅ Archivo swap activado: %s\n", filePath)
	return nil
}

// RemoveSwapFileWithRetry desactiva y elimina el archivo de swap indicado con reintentos
func RemoveSwapFileWithRetry(ctx context.Context, idStr string, retries int) error {
	filePath := filepath.Join(swapDir, fmt.Sprintf("swap-%s", idStr))

	log.Printf("🧽 Desactivando swap en %s", filePath)

	backoff := system.NewExponentialBackoff(retries, 500*time.Millisecond, 5*time.Second)

	// Paso 1: swapoff
	_ = backoff.DoWithRetry(ctx, fmt.Sprintf("swapoff %s", filePath), func() error {
		cmd := exec.CommandContext(ctx, "swapoff", filePath)
		return cmd.Run()
	})
	// Ignorar errores de swapoff (podría ya estar inactivo)

	// Paso 2: remove
	err := backoff.DoWithRetry(ctx, fmt.Sprintf("remove %s", filePath), func() error {
		return os.Remove(filePath)
	})
	if err != nil {
		return fmt.Errorf("no se pudo eliminar archivo swap: %w", err)
	}

	log.Printf("🧹 Archivo swap eliminado: %s", filePath)
	return nil
}

// Desactiva y elimina el archivo de swap indicado
func RemoveSwapFile(id int) error {
	filePath := filepath.Join(swapDir, fmt.Sprintf("swap-%d", id))

	log.Printf("🧽 Desactivando swap en %s", filePath)

	if err := exec.Command("swapoff", filePath).Run(); err != nil {
		log.Printf("⚠️ Error al desactivar: %v", err)
	}

	if err := os.Remove(filePath); err != nil {
		return fmt.Errorf("no se pudo eliminar archivo swap: %w", err)
	}

	log.Printf("🧹 Archivo swap eliminado: %s", filePath)
	return nil
}

// Limpia archivos swap residuales al iniciar (deprecated, use CleanupOnStartup from recovery.go)
func CleanUpSwapFilesOnStartup() {
	files, err := filepath.Glob(filepath.Join(swapDir, "swap-*"))
	if err != nil {
		log.Printf("⚠️ Error al escanear archivos swap: %v\n", err)
		return
	}

	for _, f := range files {
		log.Printf("🧹 Eliminando swap residual: %s", f)

		_ = exec.Command("swapoff", f).Run() // Ignora errores si ya está inactivo
		if err := os.Remove(f); err != nil {
			log.Printf("⚠️ No se pudo eliminar archivo: %v\n", err)
		}
	}
}

// Devuelve el número de archivos swap creados por Swaptimize (deprecated, use SwapVerifier.CountActiveSwaps)
func CountActiveSwapFiles() (int, error) {
	files, err := filepath.Glob(filepath.Join(swapDir, "swap-*"))
	if err != nil {
		return 0, err
	}

	count := 0
	for _, file := range files {
		if strings.HasPrefix(file, swapDir+"/swap-") {
			count++
		}
	}

	return count, nil
}
