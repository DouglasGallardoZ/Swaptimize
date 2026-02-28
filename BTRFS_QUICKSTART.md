# 🚀 Guía Rápida: Swaptimize en Btrfs

## 1️⃣ Verificar que usando Btrfs

```bash
df -T /var/lib
```

Si ves `btrfs` en la columna TYPE, estás listo.

---

## 2️⃣ Instalar Swaptimize

```bash
git clone https://github.com/tu-usuario/Swaptimize
cd Swaptimize
sudo make install
```

---

## 3️⃣ Ver logs de detección btrfs

```bash
journalctl -u swaptimize -f
```

**Busca esto en los logs** (confirma que detectó btrfs):

```
📋 Filesystem detected: btrfs at /var/lib
⚠️  Special btrfs handling enabled:
  • COW (Copy-on-Write) will be disabled for swap files
...
║ Filesystem: btrfs                              ║
║ Btrfs Support: COW Disabled                    ║
...
✓ COW deshabilitado en /var/lib/swaptimize/swap-1 (chattr +C)
```

---

## 4️⃣ Verificar archivos swap

```bash
# Ver archivos swap creados
swapon --show

# Verificar que COW está deshabilitado (busca 'C')
lsattr /var/lib/swaptimize/swap-*
```

**Esperado**:
```
--------C------ /var/lib/swaptimize/swap-1
--------C------ /var/lib/swaptimize/swap-2
```

El `C` indica que COW está deshabilitado ✅

---

## 5️⃣ Test de funcionamiento

### Crear presión de RAM:

```bash
# Terminal 1: Ver logs en tiempo real
journalctl -u swaptimize -f

# Terminal 2: Crear carga de RAM (90%)
stress-ng --vm 1 --vm-bytes 90% --timeout 60s
```

**Esperado en logs**:
1. Detecta presión de alta RAM
2. Crea nuevo archivo swap en btrfs
3. Sin errores de COW
4. Sistema responde normalmente

---

## ⚠️ Troubleshooting

### "chattr not found"
```bash
# Instalar e2fsprogs
sudo apt install e2fsprogs    # Debian/Ubuntu
sudo dnf install e2fsprogs    # Fedora/RHEL

# O usar distro equivalente
```

**No es crítico**: Swaptimize continúa, pero es recomendado para btrfs.

### Verificar chattr funciona:
```bash
chattr +C test_file
lsattr test_file
# Debe mostrar 'C'
```

### Archivos swap no aparecen en `swapon --show`

- Verificar logs: `journalctl -u swaptimize -e`
- Espaciodisco: `df -h /var/lib`
- Permisos: ¿Ejecutas como root?

---

## 🎯 Ventajas en Btrfs

| Sin Swaptimize | Con Swaptimize |
|---|---|
| ❌ Sin swap dinámico | ✅ Swap bajo demanda |
| ❌ Riesgo de OOM kill | ✅ Presión manejable |
| ❌ Si creas swap manualmente + COW = corrupción | ✅ COW automáticamente deshabilitado |
| ❌ Deadlocks si fuerza swap + COW | ✅ Seguro, sin deadlocks |

---

## 📞 Más Información

- Lee `BTRFS_SUPPORT.md` para detalles técnicos
- Lee `README.md` para configuración avanzada
- Ver `prometheus.yml` para métricas si usas Prometheus

---

**¡Listo! Swaptimize en btrfs está completamente soportado y optimizado.** 🎉
