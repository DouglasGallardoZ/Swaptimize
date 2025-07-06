
# 🚀 Swaptimize

**Swaptimize** es un servicio dinámico para sistemas Linux que optimiza el uso de memoria swap en tiempo real. A diferencia de soluciones tradicionales con particiones fijas, Swaptimize monitorea constantemente el estado del sistema y crea o elimina archivos de swap según los umbrales definidos, brindando flexibilidad, estabilidad y control total.

> ✅ Ideal para estaciones de trabajo, entornos de ciencia de datos, desarrollo backend y sistemas con recursos limitados.

---

## ⚙️ Características

- 📊 Gestión de swap basada en uso de memoria y espacio en disco
- 🔄 Creación y eliminación automática de archivos swap
- 📁 Configuración flexible vía archivo `.env`
- 🧵 Integración con `systemd` como servicio persistente
- 🪵 Registro de operaciones con rotación de logs automática (`logrotate`)
- 💡 Totalmente auditable, extensible y fácil de modificar

> 🔐 Compatible actualmente con sistemas de archivos **ext4**

---

## 🏗️ Estructura del Proyecto

```
Swaptimize/
├── bin/
│   ├── manage_swap.sh
│   ├── install.sh
│   └── uninstall.sh
├── config/
│   ├── manage_swap.env.example
│   └── manage_swap.logrotate
├── systemd/
│   └── manage-swap.service
├── README.md
├── LICENSE
└── .gitignore
```

---

## 🚀 Instalación

Ejecuta el script de instalación:

```bash
sudo ./bin/install.sh
```

Este script:

- Copia `manage_swap.sh` a `/usr/local/bin/`
- Instala el servicio `systemd` en `/etc/systemd/system/manage-swap.service`
- Copia la plantilla de entorno a `/etc/manage_swap.env` (si no existe)
- Configura rotación de logs con `logrotate`
- Recarga `systemd` y habilita el servicio al arranque

---

## 🔧 Configuración vía `.env`

Puedes personalizar parámetros creando/modificando el archivo:

```dotenv
# /etc/manage_swap.env
SWAP_SLEEP_INTERVAL=30           # Intervalo entre inspecciones (segundos)
SWAP_THRESHOLD_HIGH=85           # % uso de swap que activa creación
SWAP_THRESHOLD_LOW=40            # % uso para permitir eliminación
MAX_SWAP_FILES=4                 # Límite de archivos simultáneos
```

> Si alguna variable no está definida, se utiliza un valor por defecto.

También puedes usar la plantilla de ejemplo:

```bash
sudo cp config/manage_swap.env /etc/manage_swap.env
```

---

## 🌀 Uso del Servicio

```bash
sudo systemctl status manage-swap.service   # Estado del servicio
```

Se recomienda reiniciar el sistema después de la instalación inicial.

---

## 🪵 Logs y mantenimiento

Todas las operaciones se registran en:

```
/var/log/manage_swap.log
```

Para evitar crecimiento excesivo, se incluye una política `logrotate`:

```conf
/var/log/manage_swap.log {
    daily
    rotate 7
    compress
    missingok
    notifempty
    create 644 root root
    su root root
}
```

Puedes forzar rotación manual con:

```bash
sudo logrotate -f /etc/logrotate.d/manage_swap
```

---

## ❌ Desinstalación

Para eliminar completamente Swaptimize:

```bash
sudo ./bin/uninstall.sh
```

Esto:

- Detiene y deshabilita el servicio
- Elimina los archivos del sistema (`bin`, `systemd`, `logrotate`)
- Pregunta si deseas conservar la configuración

---

## 🧩 Compatibilidad

- ✅ Ubuntu 22.04 / 24.04 Desktop
- ✅ Fedora Workstation con ext4
- ⚠️ Requiere sistemas de archivos **ext4**
- ❌ No compatible con WSL2 o swap por zram/zswap

---

## 📘 Licencia

Este proyecto está licenciado bajo la [MIT License](LICENSE)

---