#!/bin/bash

set -e

echo "🧹 Desinstalando Swaptimize..."

# 1. Detener servicio si está activo
echo "⛔ Deteniendo servicio..."
sudo systemctl stop swaptimize.service || echo "Servicio no activo"

# 2. Deshabilitar y eliminar servicio systemd
echo "⚙️ Eliminando servicio systemd..."
sudo systemctl disable swaptimize.service || true
sudo rm -f /etc/systemd/system/swaptimize.service

# 3. Eliminar binario principal
echo "📁 Eliminando binario en /usr/local/bin/"
sudo rm -f /usr/local/bin/swaptimize

# 4. Eliminar archivo de configuración
ENV_PATH="/etc/manage_swap.env"
if [ -f "$ENV_PATH" ]; then
    read -p "❓ ¿Deseas eliminar también la configuración $ENV_PATH? [s/N]: " confirm
    if [[ "$confirm" =~ ^[sS]$ ]]; then
        sudo rm -f "$ENV_PATH"
        echo "🧽 Configuración eliminada."
    else
        echo "📦 Configuración preservada."
    fi
fi

# 5. Recargar systemd
echo "🔁 Recargando systemd..."
sudo systemctl daemon-reexec
sudo systemctl daemon-reload

echo "✅ Swaptimize ha sido desinstalado correctamente."
