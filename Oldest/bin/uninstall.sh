#!/bin/bash

set -euo pipefail

GREEN='\033[1;32m'
RED='\033[1;31m'
NC='\033[0m'

log() {
  echo -e "${GREEN}✔ $1${NC}"
}

error() {
  echo -e "${RED}✖ $1${NC}"
  exit 1
}

# Confirmación de desinstalación
read -p "¿Estás seguro de que quieres desinstalar Manage-Swap? [s/N]: " confirm
if [[ ! "$confirm" =~ ^[Ss]$ ]]; then
  echo "❎ Operación cancelada."
  exit 0
fi

# Detener y deshabilitar el servicio
if systemctl list-units --full -all | grep -q "manage-swap.service"; then
  systemctl stop manage-swap.service || true
  systemctl disable manage-swap.service || true
  log "Servicio detenido y deshabilitado"
fi

# Eliminar archivos instalados
rm -f /usr/local/bin/manage_swap.sh
rm -f /etc/systemd/system/manage-swap.service
rm -f /etc/logrotate.d/manage_swap
log "Archivos del servicio eliminados"

# Preguntar si desea conservar el archivo de configuración
if [ -f /etc/manage_swap.env ]; then
  read -p "¿Deseas eliminar el archivo de configuración '/etc/manage_swap.env'? [s/N]: " delenv
  if [[ "$delenv" =~ ^[Ss]$ ]]; then
    rm -f /etc/manage_swap.env
    log "Archivo de configuración eliminado"
  else
    echo "ℹ Archivo de configuración conservado."
  fi
fi

# Recargar systemd
systemctl daemon-reload
log "Systemd recargado"

echo -e "${GREEN}✅ Desinstalación completada correctamente.${NC}"
