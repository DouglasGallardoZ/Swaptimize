### Ajuste del Servicio Systemd

Modifica el archivo del servicio Systemd para reflejar estos cambios:

```ini
sudo nano /etc/systemd/system/manage-swap.service
```

Contenido del archivo de servicio:

```ini
[Unit]
Description=Manage Swap Files Dynamically
After=network.target

[Service]
Type=oneshot
ExecStart=/usr/local/bin/manage_swap.sh start
ExecStop=/usr/local/bin/manage_swap.sh stop
RemainAfterExit=true

[Install]
WantedBy=multi-user.target
```

### Implementación

1. **Guardar y Configurar el Script**:
   
   ```bash
   sudo nano /usr/local/bin/manage_swap.sh
   ```

   Pegar el script ajustado y guardar los cambios.

2. **Dar Permisos de Ejecución**:
   
   ```bash
   sudo chmod +x /usr/local/bin/manage_swap.sh
   ```

3. **Crear o Modificar el Servicio Systemd**:
   
   ```bash
   sudo nano /etc/systemd/system/manage-swap.service
   ```

   Pegar el contenido del archivo de servicio y guardar los cambios.

4. **Recargar Systemd y Habilitar el Servicio**:
   
   ```bash
   sudo systemctl daemon-reload
   sudo systemctl enable manage-swap.service
   ```

5. **Verificar el Estado del Servicio**:
   
   ```bash
   sudo systemctl status manage-swap.service
   ```

Con estos ajustes, el script y el servicio Systemd deberían manejar correctamente la gestión de archivos de swap y asegurar que se limpien todos los archivos al reiniciar o detener el servicio.