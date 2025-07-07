package swap

import (
    "fmt"
    "log"
    "os"
    "os/exec"
    "path/filepath"
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

// Limpia archivos swap residuales al iniciar
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
