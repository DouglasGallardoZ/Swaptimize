Tu README está claro, profesional y bien estructurado, Douglas 👌. Solo ajustaré algunos puntos menores para alinearlo con tu implementación actual y destacar funcionalidades que no deben pasarse por alto — como la lógica adaptativa, la detección de swap inicial, y la comparativa con otros sistemas operativos. Aquí tienes la versión mejorada:

---

## 📦 Swaptimize

**Gestor dinámico de memoria swap para estaciones de trabajo Linux.**  
Swaptimize es un daemon modular escrito en Go que optimiza el uso de la swap del sistema en tiempo real. A diferencia de enfoques tradicionales basados en particiones fijas, Swaptimize monitorea el estado del sistema, crea y elimina archivos de intercambio según umbrales configurables, ofreciendo flexibilidad, estabilidad y control total.

> ✅ Ideal para estaciones de trabajo, entornos de ciencia de datos, servidores backend y sistemas con recursos limitados.

---

### 🚀 Características

- ⚙️ Demonio eficiente en Go, sin procesos externos
- 📊 Monitoreo continuo de RAM, swap y espacio en disco
- 🔁 Creación/eliminación dinámica de archivos swap
- 📈 Lógica adaptativa: reduce el intervalo si hay presión alta
- 🧹 Limpieza automática de swap huérfanos al iniciar
- 🧠 Detecta ausencia de swap inicial y la crea automáticamente
- 🖥️ Integración con `systemd`
- 🔧 Configuración vía archivo `.env`
- 💻 CLI modular con comandos `run`, `status`, `clean`

---

### 📋 Requisitos

- Sistema Linux con `systemd`
- Go ≥ 1.22 instalado para compilación local
- Acceso `sudo` para uso de `swapon` / `swapoff`

---

### ⚙️ Instalación

```bash
sudo make install
```

Esto:

- Compila el binario
- Crea `/etc/manage_swap.env` con valores por defecto
- Instala y habilita el servicio systemd
- Activa el demonio como servicio en segundo plano

---

### 💻 CLI disponible

```bash
swaptimize run       # Ejecuta el daemon (directo o por systemd)
swaptimize status    # Muestra métricas actuales del sistema
swaptimize clean     # Elimina archivos swap activos
```

> El subcomando `run` requiere privilegios, los demás pueden ejecutarse como usuario.

---

### 📁 Configuración (`/etc/swaptimize.env`)

```ini
SWAP_SLEEP_INTERVAL=30         # Intervalo de chequeo normal (segundos)
SWAP_EMERGENCY_INTERVAL=10     # Intervalo si Swap ≥ 90%
SWAP_THRESHOLD_HIGH=85         # Umbral para crear swap (%)
SWAP_THRESHOLD_LOW=40          # Umbral para eliminar swap (%)
SWAP_SIZE=4096                 # Tamaño de cada archivo swap (MB)
MAX_SWAP_FILES=4               # Número máximo simultáneo de swap files
```

> Swaptimize se adapta a laptops, estaciones de trabajo y entornos de contenedores.

---

### 📜 Logging y observabilidad

El servicio utiliza `journalctl` para registrar eventos y métricas:

```bash
journalctl -u swaptimize.service -f
```

> No genera archivos de log adicionales, lo que simplifica la rotación y auditoría del sistema.

---

### 🧪 Pruebas de estrés recomendadas

Para validar el comportamiento bajo presión de RAM:

```bash
stress-ng --vm 2 --vm-bytes 85% --timeout 2m
```

Luego consulta el estado del daemon:

```bash
swaptimize status
journalctl -u swaptimize.service
```

---

### 🧠 Comparativa con otros sistemas operativos

| Sistema       | Control del swap | Visibilidad | Adaptación dinámica | Configuración |
|---------------|------------------|-------------|----------------------|----------------|
| **Swaptimize**| ✅ Total         | ✅ CLI/logs | ✅ Por presión RAM   | ✅ `.env`       |
| **Windows**   | ❌ Oculto        | ⚠️ Parcial  | ❌ No configurable   | ❌              |
| **macOS**     | ❌ Oculto        | ❌ No accesible | ⚠️ Interna         | ❌              |

> Swaptimize entrega un nivel de control que ni Windows ni macOS ofrecen al usuario avanzado.

---

### ❌ Desinstalación

```bash
sudo make uninstall
```

Esto:

- Detiene y elimina el servicio
- Borra el binario en `/usr/local/bin`
- Te pregunta si deseas eliminar el archivo `.env`

---

### 📝 Licencia

Este proyecto está bajo la [Licencia MIT](./LICENSE).

---

### 📬 Contribuciones

Swaptimize está diseñado para desarrolladores que valoran rendimiento, transparencia y control. Ideas, sugerencias y mejoras son bienvenidas para seguir expandiendo la herramienta.