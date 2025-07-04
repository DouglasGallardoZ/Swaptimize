#!/bin/bash

# Configuración
SWAP_DIR="/swapfiles"
SWAP_SIZE="4G"
SWAP_SIZE_BYTES=$((4 * 1024 * 1024))  # 4GB en bytes
MAX_SWAP_FILES=4
SWAP_THRESHOLD_HIGH=85
SWAP_THRESHOLD_LOW=40
SLEEP_INTERVAL=30
MIN_FREE_SPACE=$((5 * 1024 * 1024))  # 5GB en bytes

# Crear el directorio de archivos de swap si no existe
mkdir -p $SWAP_DIR

# Función para verificar el espacio en disco
check_disk_space() {
    local free_space=$(df --output=avail / | tail -n 1)
    echo $free_space
}

# Función para verificar el uso actual de swap en porcentaje
check_swap_usage() {
    local swap_total=$(grep SwapTotal /proc/meminfo | awk '{print $2}')
    local swap_free=$(grep SwapFree /proc/meminfo | awk '{print $2}')
    local swap_used=$((swap_total - swap_free))
    
    if [ $swap_total -eq 0 ]; then
        echo 0
    else
        local swap_percentage=$((swap_used * 100 / swap_total))
        echo $swap_percentage
    fi
}

# Función para crear un nuevo archivo de swap
create_swap() {
    local swap_number=$1
    local swap_file="$SWAP_DIR/swapfile${swap_number}"

    if [ ! -f "$swap_file" ]; then
        fallocate -l $SWAP_SIZE $swap_file
        chmod 600 $swap_file
        mkswap $swap_file
        swapon $swap_file
        echo "$(date): Created and enabled $swap_file" >> /var/log/manage_swap.log
    fi
}

# Función para eliminar el archivo de swap más antiguo
remove_oldest_swap() {
    local oldest_swap_file=$(ls -tr $SWAP_DIR | head -n 1)
    if [ -n "$oldest_swap_file" ]; then
        swapoff "$SWAP_DIR/$oldest_swap_file"
        rm "$SWAP_DIR/$oldest_swap_file"
        echo "$(date): Removed $SWAP_DIR/$oldest_swap_file" >> /var/log/manage_swap.log
    fi
}

# Función para eliminar todos los archivos de swap
remove_all_swap() {
    for swap_file in $SWAP_DIR/swapfile*; do
        swapoff $swap_file
        rm $swap_file
        echo "$(date): Removed $swap_file" >> /var/log/manage_swap.log
    done
}

# Verificar uso de swap y gestionar archivos de swap
case "$1" in
    start)
        # Limpiar archivos de swap al inicio (opcional)
        remove_all_swap
        while true; do
            swap_percentage=$(check_swap_usage)
            free_space=$(check_disk_space)

            if [ $swap_percentage -ge $SWAP_THRESHOLD_HIGH ]; then
                swap_files_count=$(ls $SWAP_DIR | wc -l)

                if [ $swap_files_count -lt $MAX_SWAP_FILES ] && [ $free_space -gt $((SWAP_SIZE_BYTES + MIN_FREE_SPACE)) ]; then
                    create_swap $((swap_files_count + 1))
                else
                    echo "$(date): Not enough space or maximum swap files reached. Free space: $free_space bytes. space $free_space prueba $((SWAP_SIZE_BYTES + MIN_FREE_SPACE))" >> /var/log/manage_swap.log
                fi
            fi

            if [ $swap_percentage -le $SWAP_THRESHOLD_LOW ]; then
                swap_files_count=$(ls $SWAP_DIR | wc -l)

                if [ $swap_files_count -gt 0 ]; then
                    remove_oldest_swap
                fi
            fi

            sleep $SLEEP_INTERVAL
        done
        ;;
    stop)
        # Eliminar todos los archivos de swap al detener el servicio
        remove_all_swap
        ;;
    *)
        echo "Uso: $0 {start|stop}"
        exit 1
esac

exit 0

