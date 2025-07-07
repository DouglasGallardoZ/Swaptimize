#!/bin/bash

set -e

echo "🛠️ Instalando Swaptimize..."

# 1. Compilar binario
echo "🔨 Compilando binario..."
go build -o swaptimize ./main.go

# 2. Copiar binario al sistema
echo "📁 Copiando a /usr/local/bin/"
sudo cp swaptimize /usr/local/bin/swaptimize
sudo chmod +x /usr/local/bin/swaptimize

# 3. Crear archivo .env si no existe
ENV_PATH="/etc/manage_swap.env"
if [ ! -f "$ENV_PATH" ]; then
    echo "🧬 Creando $ENV_PATH con valores por defecto..."
    sudo tee "$ENV_PATH" > /dev/null <<EOF
SWAP_SLEEP_INTERVAL=30
SWAP_EMERGENCY_INTERVAL=10
SWAP_THRESHOLD_HIGH=85
SWAP_THRESHOLD_LOW=40
SWAP_SIZE=4096
MAX_SWAP_FILES=4
EOF
else
    echo "🔍 Archivo de configuración ya existe: $ENV_PATH"
fi

# 4. Instalar systemd service
echo "⚙️ Instalando servicio systemd..."
sudo tee /etc/systemd/system/swaptimize.service > /dev/null <<EOF
[Unit]
Description=Swaptimize Daemon - Gestor Dinámico de Swap
After=network.target

[Service]
ExecStart=/usr/local/bin/swaptimize run
EnvironmentFile=/etc/manage_swap.env
Restart=always
RestartSec=5
User=root

[Install]
WantedBy=multi-user.target
EOF

# 5. Recargar systemd y activar servicio
echo "🚀 Activando servicio..."
sudo systemctl daemon-reexec
sudo systemctl daemon-reload
sudo systemctl enable swaptimize.service
sudo systemctl start swaptimize.service

echo "✅ Instalación completa. Estado del servicio:"
