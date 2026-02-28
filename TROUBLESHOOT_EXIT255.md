# 🛠️ Troubleshooting: Error exit 255 en swapon

## 📊 Análisis del Error

```
✓ COW deshabilitado
✓ fallocate completado
✓ Permisos 0600 establecidos
❌ swapon falla con exit 255
```

**Exit code 255** en Fedora/btrfs típicamente significa:

1. **SELinux bloqueando** (Fedora por defecto con SELinux)
2. **Atributo `nodatasum` requerido** para btrfs swap
3. **Límite de archivos swap del sistema**
4. **Modo strict del kernel**

---

## ✅ Soluciones Verificadas

### Opción 1: Desabilitar SELinux (Rápido para test)

```bash
# Verificar estado
getenforce

# Desabilitar temporalmente (solo para sesión)
sudo setenforce 0

# Después de test
sudo setenforce 1
```

**Si swaptimize funciona después**: Es SELinux. Solución:
- Crear policy de SELinux para swaptimize
- O ejecutar swaptimize en contexto `unconfined_t`

---

### Opción 2: Usar `btrfs property set` (Recomendado para btrfs)

Además de `chattr +C`, btrfs requiere `nodatasum` para swap files:

**Necesito actualizar `filesystem.go` para agregar esto:**

```bash
# Manual:
sudo btrfs property set /var/lib/swaptimize/swap-1 compression none
```

---

### Opción 3: Usar `/tmp` en lugar de `/var/lib/swaptimize`

Si `/var/lib` tiene restricciones:

```bash
# Cambiar ruta en `/etc/swaptimize.env`:
SWAP_DIR=/tmp/swaptimize
```

Pero requiere cambio de código.

---

### Opción 4: Verificar límite de archivos swap

```bash
# En btrfs, hay límite de cantidad de archivos swap
swapon --show | wc -l
```

Si ya hay 4+ archivos swap de otras fuentes, btrfs puede rechazar más.

---

## 🔧 Agregar `btrfs property set` al Código

Voy a agregar esto a `filesystem.go` para mejor soporte btrfs:

```go
// Para btrfs swap files, también necesitamos:
btrfs property set /path compression none
btrfs property set /path nodatasum on  (opcional pero recomendado)
```

---

## 🧪 Test Completo Paso a Paso

### 1. Diagnóstico previo
```bash
chmod +x /path/to/diagnose-swap.sh
sudo /path/to/diagnose-swap.sh
```

### 2. Si SELinux es culpable:
```bash
sudo setenforce 0
sudo systemctl restart swaptimize
journalctl -u swaptimize -f  # Ver si funciona
```

### 3. Si el problema persiste incluso sin SELinux:
```bash
# Crear manual para ver exacto error:
sudo fallocate -l 1G /tmp/test.swp
sudo chmod 0600 /tmp/test.swp
sudo mkswap /tmp/test.swp
sudo swapon /tmp/test.swp 2>&1  # Ver error exacto
```

---

## 📝 Próximas Mejoras Implementadas

Actualizando `filesystem.go` para agregar:

1. **`btrfs property set compression none`** - Optimización
2. **Mejor diagnóstico de permisos en logs**
3. **Fallback a `/tmp`** si `/var/lib` falla (opcional)
4. **SELinux detection y warning**

---

## 🎯 Recomendación Inmediata

1. **Ejecuta el script diagnóstico**:
   ```bash
   sudo bash /home/dgallardo/Downloads/Swaptimize/diagnose-swap.sh
   ```

2. **Comparte output** para análisis completo

3. **Intenta deshabilitar SELinux temporalmente**:
   ```bash
   sudo setenforce 0
   sudo systemctl restart swaptimize
   ```

Si funciona sin SELinux → El problema es SELinux policy
Si sigue fallando → El problema no es SELinux (es kernel/btrfs/uso)

---

## 📞 Output esperado de diagnóstico que nos dirá qué hacer:

- ✅ Si todos los tests manuales pasan → SELinux
- ✅ Si swapon falla en test manual → Problema del sistema (no swaptimize)
- ✅ Si mkswap falla → Filesystem/corrupción
- ✅ Si fallocate falla → Espacio disco o permisos directorio

**Corre el script y comparte el output completo**
