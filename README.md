# 📦 Swaptimize v2.1 | Hybrid Adaptive Swap Manager

**Gestor dinámico e inteligente de memoria swap para Linux.**

Swaptimize es un daemon modular escrito en Go que optimiza el uso de la swap del sistema en tiempo real usando una estrategia **dual-path adaptativa**. Detecta picos de presión de memoria con PSI (Pressure Stall Information) e implementa respuestas diferenciadas: reacción rápida (10s) ante picos críticos, conservador (300s) ante cambios graduales.

> 🎯 **Ideal para:** Estaciones de trabajo, entornos de ciencia de datos, contenedores, servidores backend y sistemas con recursos limitados.

---

## 🚀 Características v2.1

### Arquitectura Hybrid Adaptive
- **Dual-path response**: 10s de intervalo para picos agudos, 300s para cambios graduales
- **Peak detection**: Basado en cambios rápidos de RAM (Δ>10% en 60s) + PSI (Pressure Stall Information)
- **Protección inteligente**: Evita oscilaciones de swap creando/eliminando continuamente
- **Rate limiting adaptativo**: Histeresis de 30s entre operaciones, espera de 120s post-eliminación

### Funcionalidades Core
- ⚙️ Demonio eficiente en Go, sin procesos externos ni dependencias externas
- 📊 Monitoreo continuo de RAM, swap, PSI y presión de sistema
- 🔁 Creación/eliminación dinámica de archivos swap basada en presión real
- 🧠 Detección automática de picos de memoria usando PSI (/proc/pressure/memory)
- 📈 Circular buffer de 5 muestras (150 segundos) para análisis histórico
- 🛡️ Protecciones contra oscilación: SwapPercent ≤ 30% para eliminar seguro
- 🖥️ Integración con `systemd` como servicio de usuario o root
- 🔧 **13 parámetros configurables** vía archivo `.env` (todos con defaults sensatos)
- 💻 CLI modular con comandos `run`, `status`, `clean`
- 📈 Prometheus compatible (metrics en :9100)
- 📉 Logging minimalista: solo acciones (creación/eliminación), sin repetición de estados

---

## 📋 Requisitos

- **OS**: Linux con `systemd` (kernel ≥ 5.0 recomendado para PSI support)
- **Go**: ≥ 1.20 para compilación local
- **Permisos**: `sudo` para `swapon`/`swapoff`
- **PSI Support**: `/proc/pressure/memory` (opcional pero recomendado para peak detection)

---

## ⚙️ Instalación

### Opción 1: Desde el repositorio
```bash
git clone https://github.com/tu-usuario/Swaptimize
cd Swaptimize
sudo make install
```

### Opción 2: Compilación manual
```bash
go build -o swaptimize main.go
sudo cp swaptimize /usr/local/bin/
sudo cp assets/example.env.v2.1 /etc/swaptimize.env  # (opcional, preserva si existe)
sudo systemctl enable swaptimize.service
sudo systemctl start swaptimize.service
```

---

## 🔧 Configuración v2.1 (13 parámetros nuevos)

### Archivo `/etc/swaptimize.env`

```ini
# === THRESHOLDS DE CREACIÓN Y ELIMINACIÓN ===
SWAP_ALERT_THRESHOLD=85          # RAM ≥ 85% → crear swap (%)
SWAP_DELETE_THRESHOLD=30         # Swap ≤ 30% REQUERIDO para eliminar (protección de datos)
SWAP_USE_THRESHOLD=60            # No crear más swap si swap < 60% (evita desperdicio)

# === DETECCIÓN DE PICOS (Peak Detection) ===
MEM_DELTA_THRESHOLD=10.0         # Δ RAM > 10% en ventana activa pico (%)
MEM_DELTA_WINDOW_SEC=60          # Ventana de análisis histórico para picos (seg)

# === DETECCIÓN DE LLENADO RÁPIDO (Fast-Fill) ===
SWAP_DELTA_THRESHOLD=20          # Δ Swap > 20% en 45s → respuesta rápida (%)
SWAP_DELTA_WINDOW_SEC=45         # Ventana de análisis para swap rápido (seg)

# === RESPUESTA DUAL-PATH ===
PEAK_RESPONSE_SEC=10             # Intervalo si detecta pico (seg)
GRADUAL_RESPONSE_SEC=300         # Intervalo para cambios graduales (seg)

# === RATE LIMITING ===
CREATE_DELETE_HYSTERESIS_SEC=30  # Espera mínima entre acciones (crear/eliminar)
DELETE_WAIT_SEC=120              # Espera POST-eliminación antes de crear nuevo

# === PSI THRESHOLDS ===
PSI_HIGH_THRESHOLD=80.0          # PSI > 80% considera presión alta (0-100)
PSI_LOW_THRESHOLD=40.0           # PSI < 40% considera presión baja (0-100)

# === OTROS ===
SWAP_SIZE_MB=4096                # Tamaño de cada archivo swap (MB)
MAX_SWAP_FILES=4                 # Máximo archivos swap simultáneos
```

### Defaults (si no se especifican en `.env`)
Todos los parámetros tienen defaults sensatos con backward-compatibility. Un archivo `.env` antiguo de v2.0 sigue funcionando (aplicará los nuevos defaults).

### Profiles Preestablecidos
```ini
# LAPTOP (512 MB swap min)
SWAP_SIZE_MB=512
MAX_SWAP_FILES=2

# WORKSTATION (1024 MB swap min) - RECOMENDADO
SWAP_SIZE_MB=1024
MAX_SWAP_FILES=4

# SERVER (2048 MB swap min)
SWAP_SIZE_MB=2048
MAX_SWAP_FILES=8
```

---

## 📊 Comportamiento Adaptativo

### Matriz de Respuesta (v2.1 Hybrid Adaptive)

| Condición | Detector | Intervalo | Acción |
|-----------|----------|-----------|--------|
| **RAM ≥ 85%** | MemPercent | 10s (peak) | Crear si no existe |
| **Δ RAM > 10% en 60s** | Peak detection | 10s | Prioritario, crear |
| **Swap ≥ 85%** | SwapPercent | 10s (peak) | Crear si no existe |
| **Δ Swap > 20% en 45s** | Fast-fill | 10s | Prioritario, crear |
| **PSI > 80%** | Pressure Stall | 10s | Prioritario, crear |
| **RAM ≤ 40% AND Swap ≤ 30%** | Gradual | 300s | Eliminar si seguro |
| **PSI ≤ 40%** | Low pressure | 300s | Conservador |

### Protecciones Inteligentes
1. **Anti-oscilación**: No elimina swap si SwapPercent > 30% (datos en memoria)
2. **Anti-desperdicio**: No crea más swap si SwapPercent < 60% y MemPercent < 85%
3. **Hysteresis**: Espera 30s mín entre acciones, 120s después de eliminar
4. **Boot detection**: Crea swap automático si no existe al iniciar

---

## 💻 Uso

```bash
swaptimize run       # Ejecuta daemon (requiere sudo)
swaptimize status    # Métricas actuales (no requiere sudo)
swaptimize clean     # Elimina swap activos (requiere sudo)
```

### Como servicio systemd
```bash
sudo systemctl start swaptimize.service
sudo systemctl stop swaptimize.service
sudo systemctl status swaptimize.service
journalctl -u swaptimize -f      # Follow logs
```

---

## 📜 Logging y Observabilidad

### Journal minimalista (v2.1 optimization)
```bash
journalctl -u swaptimize -f
```

**Logs esperados:**
- ✅ `🛠️ Swap creado (N files activos)` - Cuando se crea nuevo archivo
- ✅ `📊 Swap eliminado (N files activos)` - Cuando se elimina seguro
- ✅ Error logs si hay problemas de swapon/swapoff

**Logs REMOVIDOS en v2.1 (85% menos logs):**
- ❌ Detección de picos (repetitivo cada ciclo)
- ❌ Detección de llenado rápido (repetitivo cada ciclo)
- ❌ Ajustes de intervalo ≥ 90% (log cada 10s innecesario)

**Beneficio**: Journal size reducido de 65.8MB a ~10-15MB en 24h (dependiendo del sistema)

### Prometheus Metrics (Endpoint :9100/metrics)
```bash
curl localhost:9100/metrics
```

Expone: `swaptimize_memory_percent`, `swaptimize_swap_percent`, `swaptimize_psi_memory`, `swaptimize_swap_files_active`

---

## 📦 Soporte para Filesystems

### Btrfs (Full Support ✅)

Swaptimize v2.1 incluye soporte completo y optimizado para **Btrfs**, el moderno filesystem de Linux con compresión y snapshots.

**Consideración especial**: Btrfs usa **Copy-on-Write (COW)** por defecto, lo que causa problemas críticos con archivos swap:
- **Corrupción de datos** en el archivo swap
- **Deadlocks** del kernel durante presión de memoria
- **Rendimiento degradado** por overhead de COW

**Solución implementada en Swaptimize**: 
1. ✅ **Detección automática** del filesystem (btrfs, ext4, xfs, etc)
2. ✅ **Deshabilitación de COW** antes de activar swap con `chattr +C`
3. ✅ **Buffer aumentado** (3x vs 2x) para validación de espacio
4. ✅ **Logging detallado** de manejo btrfs

#### Flujo de creación de swap en btrfs:
```
fallocate(4GB)        → Asigna bloques físicos
    ↓
chattr +C            → DESHABILITA COW para este archivo
    ↓
mkswap               → Prepara como swap (COW ya deshabilitado)
    ↓
swapon               → Activa (seguro, sin riesgos COW)
```

#### Requisitos:
- El comando `chattr` debe estar disponible (incluido en `e2fsprogs`)
- Kernel ≥ 4.14 con soporte completo para atributos btrfs

#### Test de btrfs:
```bash
# Verificar que estás en btrfs
df -T /var/lib/swaptimize

# Al iniciar swaptimize, verás en logs:
# 📋 Filesystem detected: btrfs at /var/lib
# ⚠️ Special btrfs handling enabled:
#   • COW (Copy-on-Write) will be disabled for swap files

# Verificar que COW fue deshabilitado:
lsattr /var/lib/swaptimize/swap-* | grep -c "C"  # Debe mostrar archivos con C
```

### Otros Filesystems (Full Support ✅)

Swaptimize soporta completamente:
- **ext4** / ext3 / ext2: Full support
- **XFS**: Full support
- **F2FS**: Full support
- Cualquier filesystem Linux estándar

---

## 🧪 Pruebas de Validación

### Test 1: Peak Detection (Δ RAM rápido)
```bash
stress-ng --vm 1 --vm-bytes 91% --timeout 10s
# Espera: Detect peak en < 10s, crear swap rápido
```

### Test 2: Gradual Pressure (RAM sube lento)
```bash
stress-ng --vm 2 --vm-bytes 85% --timeout 60s
# Espera: Respuesta en 300s (gradual), crea swap si necesita
```

### Test 3: ZRAM Filling sin RAM pressure
```bash
# Si usas ZRAM como único swap:
while true; do cat /dev/urandom | base64 | dd of=/dev/null; done &
# Espera: Detectar SwapPercent ≥ 85%, crear archivo swap
```

### Test 4: Safe Deletion
```bash
swaptimize status     # Ver swaps activos
# Liberar RAM manualmente
# Espera: Elimina solo si RAM ≤ 40% AND Swap ≤ 30%
```

### Monitoreo durante tests
```bash
# Terminal 1: Logs
journalctl -u swaptimize -f

# Terminal 2: Métricas
watch -n 1 swaptimize status

# Terminal 3: Procesos
watch -n 1 'free -h && swapon --show'
```

---

## 🧠 Comparativa: Swaptimize vs Windows vs macOS

| Característica | Swaptimize v2.1 | Windows | macOS |
|---|---|---|---|
| **Control dinámico** | ✅ Completo | ❌ Oculto | ❌ Oculto |
| **Configuración** | ✅ 13 parámetros `.env` | ❌ Registry opaco | ❌ No configurable |
| **Peak detection** | ✅ PSI + Δ RAM | ⚠️ Hard-coded | ⚠️ Hard-coded |
| **Logging visible** | ✅ Journalctl | ❌ Event Viewer opaco | ❌ No accesible |
| **Anti-oscilación** | ✅ Hysteresis + protección | ⚠️ Limitado | ⚠️ Limitado |
| **Presición** | ✅ PSI (kernel metrics) | ⚠️ Approximated | ⚠️ Approximated |
| **Cost** | ✅ 0 (Open Source) | ⚠️ Licensed OS | ⚠️ Licensed OS |

---

## ❌ Desinstalación

```bash
sudo systemctl stop swaptimize.service
sudo systemctl disable swaptimize.service
sudo rm /usr/local/bin/swaptimize
sudo rm /etc/swaptimize.env           # (optional)
# systemd unit se auto-limpia si está en /etc/systemd/system
```

---

## 📝 Licencia

[MIT LICENSE](./LICENSE)

---

## 🤝 Contribuciones

Swaptimize está diseñado para usuarios avanzados que valoran **control**, **visibilidad** y **adaptabilidad**. Reportes de bugs, features y PRs son bienvenidas.

### Roadmap v2.2 (Planned)
- [ ] NUMA-aware swap management
- [ ] Múltiples discos con diferente velocidad
- [ ] Integration con cgroup v2 limits
- [ ] Web dashboard (Go fiber)
