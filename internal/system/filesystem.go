package system

import (
	"fmt"
	"log"
	"os"
	"os/exec"
	"strings"
)

// FilesystemType representa el tipo de sistema de archivos
type FilesystemType string

const (
	FilesystemBtrfs    FilesystemType = "btrfs"
	FilesystemExt4     FilesystemType = "ext4"
	FilesystemExt3     FilesystemType = "ext3"
	FilesystemExt2     FilesystemType = "ext2"
	FilesystemXFS      FilesystemType = "xfs"
	FilesystemF2FS     FilesystemType = "f2fs"
	FilesystemUnknown  FilesystemType = "unknown"
)

// FilesystemInfo contiene información del filesystem
type FilesystemInfo struct {
	Type FilesystemType
	Path string // Ruta montada del filesystem
}

// DetectFilesystem detecta el tipo de filesystem para una ruta
func DetectFilesystem(path string) (*FilesystemInfo, error) {
	// Obtener filesystem info usando statfs o df
	out, err := exec.Command("stat", "-f", "-c", "%T", path).Output()
	if err != nil {
		// Si stat no funciona, intentar con df
		out, err = exec.Command("df", "-T", path).Output()
		if err != nil {
			return nil, fmt.Errorf("cannot detect filesystem: %w", err)
		}

		// Parsear output de df
		lines := strings.Split(string(out), "\n")
		if len(lines) < 2 {
			return nil, fmt.Errorf("unexpected df output")
		}

		fields := strings.Fields(lines[1])
		if len(fields) < 2 {
			return nil, fmt.Errorf("cannot parse df output")
		}

		return &FilesystemInfo{
			Type: FilesystemType(fields[1]),
			Path: path,
		}, nil
	}

	fsType := strings.TrimSpace(string(out))
	return &FilesystemInfo{
		Type: FilesystemType(fsType),
		Path: path,
	}, nil
}

// IsBtrfs comprueba si el filesystem es btrfs
func (fi *FilesystemInfo) IsBtrfs() bool {
	return fi.Type == FilesystemBtrfs
}

// RequiresNoCOW comprueba si el filesystem requiere deshabilitar COW para swap
func (fi *FilesystemInfo) RequiresNoCOW() bool {
	return fi.IsBtrfs()
}

// DisableCOW desactiva Copy-on-Write en un archivo (para btrfs)
// CRÍTICO: Debe ser exitoso antes de mkswap o btrfs rechazará el archivo swap
func DisableCOW(filePath string) error {
	// Verificar que el archivo existe
	if _, err := os.Stat(filePath); err != nil {
		return fmt.Errorf("cannot access file: %w", err)
	}

	// Detectar filesystem
	fsInfo, _ := DetectFilesystem(filePath)

	if fsInfo != nil && fsInfo.IsBtrfs() {
		log.Printf("ℹ️  Btrfs detected - setting chattr +C on nocow file")
		
		// En btrfs moderno: El atributo +C debe existir en lsattr
		// Esto se hace con chattr, NO con btrfs property set
		cmd := exec.Command("chattr", "+C", filePath)
		if err := cmd.Run(); err != nil {
			return fmt.Errorf("chattr +C failed on btrfs file: %w", err)
		}
		log.Printf("✓ chattr +C ejecutado para %s", filePath)

		// IMPORTANTE: Verificar inmediatamente
		if !VerifyCOWDisabled(filePath) {
			return fmt.Errorf("chattr +C did not apply C attribute to btrfs file")
		}
		log.Printf("✓ Verificado: atributo C presente en %s", filePath)
	} else {
		// Para otros filesystems, chattr es suficiente
		cmd := exec.Command("chattr", "+C", filePath)
		if err := cmd.Run(); err != nil {
			return fmt.Errorf("chattr failed: %w", err)
		}
		log.Printf("✓ COW deshabilitado en %s (chattr +C)", filePath)
	}

	return nil
}

// VerifyCOWDisabled verifica que COW está deshabilitado (específico para btrfs)
func VerifyCOWDisabled(filePath string) bool {
	cmd := exec.Command("lsattr", filePath)
	output, err := cmd.Output()
	if err != nil {
		return false
	}
	
	outputStr := string(output)
	return strings.Contains(outputStr, "C") || strings.Contains(outputStr, "c")
}

// LogFilesystemInfo registra información del filesystem
func LogFilesystemInfo(fsInfo *FilesystemInfo) {
	log.Printf("📋 Filesystem detected: %s at %s\n", fsInfo.Type, fsInfo.Path)

	if fsInfo.IsBtrfs() {
		log.Println("⚠️  Special btrfs handling enabled:")
		log.Println("  • COW (Copy-on-Write) will be disabled for swap files")
		log.Println("  • This prevents data corruption and performance issues")
	}
}
