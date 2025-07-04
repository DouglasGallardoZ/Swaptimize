#!/bin/bash
sudo cp ./manage_swap.sh /usr/local/bin/.
sudo cp ./manage-swap.service /etc/systemd/system/.
sudo systemctl daemon-reload
sudo systemctl enable manage-swap.service
