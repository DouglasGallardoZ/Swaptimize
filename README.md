## 📦 Swaptimize

**Gestor dinámico de memoria swap para estaciones de trabajo Linux.**  
**Swaptimize** es un servicio dinámico para sistemas Linux que optimiza el uso de memoria swap en tiempo real. A diferencia de soluciones tradicionales con particiones fijas, Swaptimize monitorea constantemente el estado del sistema y crea o elimina archivos de swap según los umbrales definidos, brindando flexibilidad, estabilidad y control total.

> ✅ Ideal para estaciones de trabajo, entornos de ciencia de datos, desarrollo backend y sistemas con recursos limitados.

---

### 🚀 Características

- ✅ Demonio modular escrito en Go
- 📊 Monitoreo continuo de RAM, swap y disco
- 🔁 Creación y eliminación automática de swap
- 🧹 Limpieza de archivos swap residuales al inicio
- ⚙️ Configuración vía archivo `.env`
- 🖥️ Instalación como servicio `systemd`
- 💻 CLI amigable con comandos `run`, `status`, `clean`

---

### 📋 Requisitos

- Linux con `systemd`
- Go ≥ 1.22 instalado para compilar
- Acceso a `sudo` o privilegios de root para `swapon` / `swapoff`

---

### ⚙️ Instalación rápida

```bash
sudo make install
```

Esto:

- Compila el binario
- Crea `/etc/swaptimize.env` con valores por defecto
- Instala y habilita el servicio systemd
- Inicia el demonio automáticamente en segundo plano

---

### 🧪 Comandos disponibles

```bash
swaptimize run       # Inicia el demonio
swaptimize status    # Muestra métricas actuales
swaptimize clean     # Elimina swaps activos
```

> Solo `run` requiere privilegios (`sudo`) porque accede al sistema de swap.

---

### 📁 Archivo de configuración (`/etc/manage_swap.env`)

```ini
SWAP_SLEEP_INTERVAL=30         # Intervalo normal de chequeo (segundos)
SWAP_EMERGENCY_INTERVAL=10     # Intervalo rápido si swap ≥ 90%
SWAP_THRESHOLD_HIGH=85         # Umbral superior para crear swap (%)
SWAP_THRESHOLD_LOW=40          # Umbral inferior para eliminar swap (%)
SWAP_SIZE=4096                 # Tamaño de cada archivo swap (MB)
MAX_SWAP_FILES=4               # Cantidad máxima simultánea de archivos swap
```

Puedes ajustarlo para laptops, servidores o entornos de prueba.

---

### 📜 Logs y observabilidad

Swaptimize usa `journalctl` para capturar la salida del demonio:

```bash
journalctl -u swaptimize.service -f
```

> No genera archivos de log adicionales, lo que simplifica la rotación y mantiene compatibilidad con herramientas modernas.

---

### ❌ Desinstalación

```bash
sudo make uninstall
```

Esto:

- Detiene y elimina el servicio
- Elimina el binario en `/usr/local/bin`
- Te pregunta si quieres borrar el archivo `.env`

---

### 📬 Contribuciones

Este proyecto nace para desarrolladores que usan Linux en estaciones de trabajo y quieren una gestión de swap inteligente, liviana y fácil de mantener.  
Contribuciones, ideas y sugerencias son bienvenidas.
