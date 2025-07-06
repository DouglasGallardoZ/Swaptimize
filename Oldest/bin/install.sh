#!/bin/bash

set -euo pipefail

# Colores para salida más clara
GREEN='\033[1;32m'
RED='\033[1;31m'
NC='\033[0m' # Sin color

log() {
  echo -e "${GREEN}✔ $1${NC}"
}

error() {
  echo -e "${RED}✖ $1${NC}"
  exit 1
}

# Verifica si se ejecuta con sudo
if [ "$EUID" -ne 0 ]; then
  error "Este script debe ejecutarse con privilegios de superusuario (sudo)."
fi

# Verifica si los archivos existen antes de copiarlos
[ -f ./manage_swap.sh ] || error "Archivo 'manage_swap.sh' no encontrado."
[ -f ./manage-swap.service ] || error "Archivo 'manage-swap.service' no encontrado."
[ -f ./manage_swap.logrotate ] || error "Archivo 'manage_swap.logrotate' no encontrado."
[ -f ./manage_swap.env ] || log "Nota: no se encontró 'manage_swap.env'; no se instalará archivo de entorno."

# Instala el script principal
cp ./manage_swap.sh /usr/local/bin/ || error "No se pudo copiar 'manage_swap.sh'"
chmod +x /usr/local/bin/manage_swap.sh
log "Script principal instalado en /usr/local/bin/"

# Instala unidad systemd
cp ./manage-swap.service /etc/systemd/system/ || error "No se pudo copiar el archivo de servicio"
log "Servicio systemd instalado"

# Instala archivo de entorno si no existe aún
if [ -f ./manage_swap.env ] && [ ! -f /etc/manage_swap.env ]; then
  cp ./manage_swap.env /etc/manage_swap.env
  log "Archivo de entorno instalado en /etc/manage_swap.env"
fi

# Instala política de logrotate
cp ./manage_swap.logrotate /etc/logrotate.d/manage_swap || error "No se pudo instalar logrotate"
log "Política de logrotate aplicada"

# Recarga y habilita systemd
systemctl daemon-reload
systemctl enable manage-swap.service || error "No se pudo habilitar el servicio"

log "✅ Instalación completada correctamente."
