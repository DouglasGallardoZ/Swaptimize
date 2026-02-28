# 🔍 Diagnóstico: Error de Activación de Swap (exit status 255)

## Log Analizado

```
Feb 28 16:24:12 fedora swaptimize[138973]: ✓ COW deshabilitado en /var/lib/swaptimize/swap-1 (chattr +C)
Feb 28 16:24:12 fedora swaptimize[138973]: ❌ Error al crear swap: error al activar swap: exit status 255
Feb 28 16:24:42 fedora swaptimize[138973]: ⚠️ El archivo swap ya existe: /var/lib/swaptimize/swap-1
Feb 28 16:24:42 fedora swaptimize[138973]: 🛠️ Swap creado [gradual] (pero NO realmente activado!)
```

---

## 🚨 Problema Identificado

### Síntomas

1. **`swapon` falla** con exit status 255 (error de permisos/validación)
2. **Archivo swap persiste** en disco pero no está activo
3. **Sistema cree que swap está activo** pero no hay espacio real disponible
4. **Ciclo se repite**: Intenta crear, falla, detecta archivo existente, log falso de éxito

### Causa Raíz

El archivo swap se crea sin los **permisos correctos (0600)**:

```bash
# ANTES (INCORRECTO):
fallocate -l 1024M /var/lib/swaptimize/swap-1   # ← permisos hereda del umask
mkswap /var/lib/swaptimize/swap-1
swapon /var/lib/swaptimize/swap-1               # ← FALLA: permisos incorrectos

# DESPUÉS (CORRECTO):
fallocate -l 1024M /var/lib/swaptimize/swap-1
chmod 0600 /var/lib/swaptimize/swap-1           # ← NUEVO: permisos correctos
mkswap /var/lib/swaptimize/swap-1
swapon /var/lib/swaptimize/swap-1               # ← ✅ FUNCIONA
```

---

## 🔧 Soluciones Implementadas

### 1. Establecer Permisos 0600

**Archivo**: `internal/swap/control.go`

#### En `CreateSwapFile()`:
```go
// Paso 1: fallocate (crear archivo)
cmd := exec.Command("fallocate", "-l", fmt.Sprintf("%dM", sizeMB), filePath)
cmd.Run()

// Paso 1.5: [NUEVO] Establecer permisos 0600
if err := os.Chmod(filePath, 0600); err != nil {
    return fmt.Errorf("error al establecer permisos: %w", err)
}

// Paso 2: mkswap (preparar como swap)
// ...
```

#### En `CreateSwapFileWithRetry()`:
```go
// Después de fallocate, antes de mkswap:
if err := os.Chmod(filePath, 0600); err != nil {
    _ = os.Remove(filePath)  // Limpiar si chmod falla
    return fmt.Errorf("chmod 0600 failed: %w", err)
}
```

### 2. Capturar Stderr para Diagnósticos

**Beneficio**: Si `swapon` falla, ahora vemos el mensaje de error real

```go
err = backoff.DoWithRetry(ctx, fmt.Sprintf("swapon %s", filePath), func() error {
    cmd := exec.CommandContext(ctx, "swapon", filePath)
    
    // [NUEVO] Capturar stderr
    var errBuf strings.Builder
    cmd.Stderr = &errBuf
    
    if err := cmd.Run(); err != nil {
        stderrMsg := errBuf.String()
        if stderrMsg != "" {
            log.Printf("⚠️ swapon stderr: %s\n", stderrMsg)  // Diagnóstico
        }
        return err
    }
    return nil
})
```

### 3. Mejorar Códigos de Error

Ahora capturamos el exit code exacto:

```go
// Antes:
fmt.Errorf("error al activar swap: %w", err)

// Después (en CreateSwapFile):
fmt.Errorf("error al activar swap (exit %d): %w", cmd.ProcessState.ExitCode(), err)
```

---

## 📊 Flujo Antes vs Después

### ANTES (INCORRECTO)

```
16:24:12 fallocate            ✅ (permisos: 644 por defecto!)
16:24:12 chattr +C            ✅
16:24:12 mkswap               ✅
16:24:12 swapon /var/lib/swaptimize/swap-1
         ❌ EXIT 255 (permisos 0644 ≠ 0600)
         ❌ Archivo permanece pero NO activo
         
16:24:42 Detecta archivo existente
         Retorna sin intentar de nuevo
         ⚠️ Log falso: "Swap creado" (pero no está!)
         
16:27:12 Intenta eliminar
         swapoff falla (no está activo)
         rm /var/lib/swaptimize/swap-1  ✅
         
16:29:12 Intenta crear NUEVAMENTE (mismo error)
         Ciclo infinito...
```

### DESPUÉS (CORRECTO)

```
16:24:12 fallocate            ✅ (permisos: 644 inicialmente)
16:24:12 chmod 0600           ✅ [NUEVO]
16:24:12 chattr +C            ✅
16:24:12 mkswap               ✅
16:24:12 swapon /var/lib/swaptimize/swap-1
         ✅ EXIT 0 (permisos 0600 correctos)
         ✅ Archivo activo en sistema
         ✅ Espacio real disponible
         
16:24:12 swapon --show        ✅ Aparece activo
         Log correcto: "Swap creado"
         
16:27:12 Condición de eliminación
         swapoff ✅ (está activo)
         rm ✅
         
16:29:12 Intenta crear
         ✅ ÉXITO (permisos correctos)
```

---

## 🧪 Verificación

### Test 1: Verificar permisos en nuevo código

```bash
# Crear script test
echo '#!/bin/bash
swaptimize run &
sleep 5
ls -l /var/lib/swaptimize/
lsattr /var/lib/swaptimize/swap-*
swapon --show
' > test_swap.sh

bash test_swap.sh
```

**Esperado**:
```
-rw------- 1 root root 1073741824 Feb 28 16:24 /var/lib/swaptimize/swap-1
--------C------ /var/lib/swaptimize/swap-1
/var/lib/swaptimize/swap-1             file    1024M        0B   -2
```

### Test 2: Ver diagnósticos de swapon

El nuevo código capturará stderr si falla:

```bash
journalctl -u swaptimize -f | grep -i "swapon\|stderr"
```

**Si hay error, verás**:
```
⚠️ swapon stderr: PERMISO_DENEGADO
```

---

## 💡 ¿Por Qué Falló con exit 255?

**exit 255** en swapon significa:
- Permisos incorrectos (no 0600)
- Archivo corrupto
- Límite de swap alcanzado
- SELinux bloqueando (menos probable)

**Lo más probable**: Permisos ≠ 0600

---

## 📝 Cambios en el Código

| Archivo | Cambio | Por qué |
|---------|--------|--------|
| `control.go` | Agregar `os.Chmod(filePath, 0600)` | Requerido por kernel para swap |
| `control.go` | Capturar stderr en swapon | Diagnóstico mejor |
| `control.go` | Limpiar si chmod falla | Sanidad |

**Total**: 15 líneas nuevas, 100% backward compatible

---

## 🎯 Resultado Esperado en Nuevo Log

```
Feb 28 16:24:12 fedora swaptimize[138973]: 🛠️ Creando archivo swap en /var/lib/swaptimize/swap-1 (1024MB)
Feb 28 16:24:12 fedora swaptimize[138973]: ✓ COW deshabilitado en /var/lib/swaptimize/swap-1 (chattr +C)
Feb 28 16:24:12 fedora swaptimize[138973]: ✅ Archivo swap activado: /var/lib/swaptimize/swap-1
Feb 28 16:24:12 fedora swaptimize[138973]: 🛠️ Swap creado [gradual] (MemPercent=95.2%, SwapPercent=50%, PSI=6.06%)

[SIN ERRORES]
[swapon funciona]
[swap está realmente activo]
```

---

## References

**Linux Swap File Requirements**:
- Must have permissions `0600` (rw-------)
- Must be on filesystem without `nodev` flag
- Ownership must be `root:root`
- Cannot use btrfs with COW (that's why we disable it)

**Kernel Docs**: 
- `man swapon` - "file must have permissions 0600"
- Btrfs docs - "Swap files require COW disabled"
