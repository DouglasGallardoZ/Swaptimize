#!/bin/bash
sudo cp ./manage_swap.sh /usr/local/bin/.
sudo cp ./manage-swap.service /etc/systemd/system/.
sudo cp ./manage_swap.env /etc/manage_swap.env
sudo systemctl daemon-reload
sudo systemctl enable manage-swap.service

echo 'Instalación realizada con éxito.'
