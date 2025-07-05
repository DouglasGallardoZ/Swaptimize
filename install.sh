#!/bin/bash

# Copia el script principal
sudo cp ./manage_swap.sh /usr/local/bin/
sudo chmod +x /usr/local/bin/manage_swap.sh

# Copia el archivo de servicio systemd
sudo cp ./manage-swap.service /etc/systemd/system/

# Instala archivo de entorno si no existe aún
if [ ! -f /etc/manage_swap.env ]; then
    sudo cp ./manage_swap.env /etc/manage_swap.env
fi

# Instala política de logrotate
sudo cp ./manage_swap.logrotate /etc/logrotate.d/manage_swap

# Recarga systemd y habilita el servicio
sudo systemctl daemon-reload
sudo systemctl enable manage-swap.service

echo '✅ Instalación realizada con éxito.'