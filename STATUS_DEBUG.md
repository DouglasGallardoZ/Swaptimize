# 🔴 Estado: Error exit 255 en swapon - Diagnóstico en Progreso

## 📋 Lo que Hemos Hecho

### Cambios Implementados en el Código

1. **`control.go`** - Mejoras significativas:
   - ✅ Agregar `chmod 0600` a archivos swap
   - ✅ Agregar `os.Chmod(0755)` al directorio `/var/lib/swaptimize`
   - ✅ Deshabilitar COW en DIRECTORIO + archivo (btrfs)
   - ✅ Captura de stderr en mkswap
   - ✅ Logs de diagnóstico: permisos, tamaño del archivo antes de swapon
   - ✅ Validación de que archivo existe y es accesible

2. **`filesystem.go`** - Extensión:
   - ✅ Agregar `btrfs property set compression none` (mejora btrfs)
   - ✅ Better logging para diagnóstico

3. **Documentación**:
   - ✅ Script diagnóstico: `diagnose-swap.sh`
   - ✅ Guía troubleshooting: `TROUBLESHOOT_EXIT255.md`

---

## 🚨 Problema Persistente

El error **sigue siendo `exit 255` en swapon** incluso con:
- ✓ Permisos correctos (0600)
- ✓ COW deshabilitado (chattr +C)
- ✓ mkswap ejecutándose
- ✓ Directorio con COW deshabilitado

**En Fedora con btrfs**, exit 255 es probablemente:

| Causa | Síntoma | Solución |
|-------|---------|----------|
| **SELinux bloqueando** | Logs limpios pero falla | `sudo setenforce 0` |
| **Límite de swap files** | Demasiados archivos swap existentes | Reducir cantidad |
| **Kernel strict mode** | Específico de la versión kernel | Verificar logs kernel |
| **btrfs sin soporte** | Filesystem reporta no supported | Usar ext4 para /var/lib |

---

## 🔍 Diagnóstico Next Step

### DEBE EJECUTAR:

```bash
# 1. Hacer diagnostico
chmod +x /home/dgallardo/Downloads/Swaptimize/diagnose-swap.sh
sudo /home/dgallardo/Downloads/Swaptimize/diagnose-swap.sh

# 2. Compartir el output COMPLETO

# 3. Probar deshabilitar SELinux temporalmente
sudo setenforce 0
sudo systemctl restart swaptimize
journalctl -u swaptimize -n 20 -f

# 4. Si funciona → SELinux es el culpable
# 5. Si sigue fallando → Hay otro problema
```

---

## 🎯 Según el Output del Diagnóstico:

### Escenario A: Test manual de swapon funciona pero systemd falla
→ **Es SELinux**
  - Solución: Configurar SELinux policy para swaptimize
  - O ejecutar: `sudo semanage permissive -a swaptimize_t`

### Escenario B: Test manual también falla
→ **Es problema del sistema, no swaptimize**
  - Solución: Ver mensaje de error en test manual
  - Problema típicos:
    - Btrfs no soporta swap en esa partición
    - Filesystem mounted con `noswap` flag
    - Kernel version incompatible

### Escenario C: mkswap falla en test manual
→ **Archivo corrupto o filesystem problema**
  - Solución: Limpiar `/var/lib/swaptimize`, reintentar

### Escenario D: fallocate falla
→ **Espacio disco insuficiente o permisos directorio**
  - Solución: `sudo rm -rf /var/lib/swaptimize && mkdir -p /var/lib/swaptimize`

---

## 📝 Archivos Nuevos/Modificados

| Archivo | Cambio | Propósito |
|---------|--------|----------|
| `control.go` | Actualizado | Mejor diagnóstico + COW en directorio |
| `filesystem.go` | Actualizado | `btrfs property set` + mejor logging |
| `diagnose-swap.sh` | ✨ NUEVO | Script diagnóstico automatizado |
| `TROUBLESHOOT_EXIT255.md` | ✨ NUEVO | Guía completa troubleshooting |

---

## 🚀 Próximas Acciones

**Necesitamos OUTPUT del diagnóstico para continuar**. El script te dirá exactamente qué está mal.

Comando para ejecutar:
```bash
sudo bash /home/dgallardo/Downloads/Swaptimize/diagnose-swap.sh 2>&1 | tee /tmp/diag.txt
cat /tmp/diag.txt
```

**Comparte el output de `/tmp/diag.txt` completo**

---

## 💡 Hipótesis Actual

**La causa más probable en Fedora:**

1. **SELinux** bloqueando `swapon` para usuario root bajo systemd
   - El servicio systemd tiene contexto diferente a shell interactivo
   - SELinux permite en shell pero niega en systemd

2. **Solución rápida test**: `sudo setenforce 0` + restart swaptimize

3. **Solución permanente**: Crear policy SELinux o usar `permissive` mode para swaptimize

---

## 📞 Contacto para Diagnosticar

Con el output del script diagnóstico podré identificar exactamente:
- Si es SELinux, kernel, btrfs, o capacidad del sistema
- Qué comando exacto está fallando
- Por qué fallen

**Por favor corre el script y comparte el output completo** 🙏
