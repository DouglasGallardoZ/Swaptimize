## 🧠 Manage-Swap (Gestión dinámica de memoria swap)

**Manage-Swap** es un servicio de sistema para Linux que permite gestionar archivos de swap dinámicamente según el uso de memoria y el espacio disponible en disco. Ideal para entornos de ciencia de datos, desarrollo backend o sistemas con recursos limitados donde se requiere estabilidad bajo cargas variables.

**Nota: Actualmente trabaja unicamente en sistemas de archivo ext4.**

---

### 🚀 Instalación

Ejecuta el script de instalación:

```bash
sudo ./install.sh
```

Este script:

- Copia el script principal `manage_swap.sh` a `/usr/local/bin/`
- Instala el servicio `systemd` como `manage-swap.service`
- Copia la plantilla de configuración a `/etc/manage_swap.env`
- Recarga `systemd` y habilita el servicio para autoarranque

> ✅ La instalación requiere permisos de superusuario.

---

### ⚙️ Configuración vía `.env`

El script lee variables opcionales desde `/etc/manage_swap.env` si está presente. Puedes personalizar el comportamiento sin modificar el código fuente:

```dotenv
# /etc/manage_swap.env
SWAP_SLEEP_INTERVAL=30          # Intervalo de chequeo (segundos)
SWAP_THRESHOLD_HIGH=85          # % de uso de swap que crea nuevo archivo
SWAP_THRESHOLD_LOW=40           # % de uso que permite remover swap
MAX_SWAP_FILES=4                # Límite total de archivos swap
```

> Si alguna variable no está definida, se aplica un valor por defecto seguro.

---

### 🔁 Uso general

Una vez instalado o configurado se debe reiniciar el sistema para que apliquen los cambios:

```bash
sudo systemctl status manage-swap.service  # Verifica estado
```

---

### 🪵 Logs

Todas las operaciones se registran en:

```bash
/var/log/manage_swap.log
```

---

### 🧩 Compatibilidad

Tested en:

- Ubuntu 22.04 o superiores, Fedora Workstation (ext4).
- Compatible unicamente con sistemas de archivo ext4.

---
