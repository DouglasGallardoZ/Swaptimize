# Variables
BIN_NAME = swaptimize
SRC = ./main.go
ENV_PATH = /etc/manage_swap.env
SERVICE_PATH = /etc/systemd/system/swaptimize.service

.PHONY: all build install uninstall clean service env

# Compila el binario principal
build:
	@echo "🔨 Compilando $(BIN_NAME)..."
	go build -o $(BIN_NAME) $(SRC)

# Instala todo en el sistema
install: build service env
	@echo "📁 Copiando binario a /usr/local/bin/"
	sudo cp $(BIN_NAME) /usr/local/bin/$(BIN_NAME)
	sudo chmod +x /usr/local/bin/$(BIN_NAME)
	@echo "🚀 Recargando y activando servicio..."
	sudo systemctl daemon-reexec
	sudo systemctl daemon-reload
	sudo systemctl enable swaptimize.service
	sudo systemctl start swaptimize.service
	@echo "✅ Instalación completa."

# Instala el archivo systemd
service:
	@echo "⚙️ Instalando servicio systemd..."
	echo "[Unit]"                              | sudo tee $(SERVICE_PATH)
	echo "Description=Swaptimize Daemon"     | sudo tee -a $(SERVICE_PATH)
	echo "After=network.target"              | sudo tee -a $(SERVICE_PATH)
	echo ""                                  | sudo tee -a $(SERVICE_PATH)
	echo "[Service]"                         | sudo tee -a $(SERVICE_PATH)
	echo "ExecStart=/usr/local/bin/swaptimize run" | sudo tee -a $(SERVICE_PATH)
	echo "EnvironmentFile=$(ENV_PATH)"       | sudo tee -a $(SERVICE_PATH)
	echo "Restart=always"                    | sudo tee -a $(SERVICE_PATH)
	echo "RestartSec=5"                      | sudo tee -a $(SERVICE_PATH)
	echo "User=root"                         | sudo tee -a $(SERVICE_PATH)
	echo ""                                  | sudo tee -a $(SERVICE_PATH)
	echo "[Install]"                         | sudo tee -a $(SERVICE_PATH)
	echo "WantedBy=multi-user.target"        | sudo tee -a $(SERVICE_PATH)

# Crea el archivo de configuración .env si no existe
env:
	@echo "📦 Validando archivo $(ENV_PATH)..."
	@if [ ! -f $(ENV_PATH) ]; then \
        echo "🧬 Creando archivo con valores por defecto..."; \
        sudo tee $(ENV_PATH) > /dev/null <<EOF \
SWAP_SLEEP_INTERVAL=30 \
SWAP_EMERGENCY_INTERVAL=10 \
SWAP_THRESHOLD_HIGH=85 \
SWAP_THRESHOLD_LOW=40 \
SWAP_SIZE=4096 \
MAX_SWAP_FILES=4 \
EOF \
    else \
        echo "✔️ Archivo ya existe."; \
    fi

# Elimina binario, servicio y configuración
uninstall:
	@echo "🧹 Desinstalando Swaptimize..."
	-sudo systemctl stop swaptimize.service
	-sudo systemctl disable swaptimize.service
	-sudo rm -f $(SERVICE_PATH)
	-sudo rm -f /usr/local/bin/$(BIN_NAME)
	@echo "¿Eliminar archivo de configuración ($(ENV_PATH))? [s/N]"
	@read RESP; \
    if [ "$$RESP" = "s" ] || [ "$$RESP" = "S" ]; then \
        sudo rm -f $(ENV_PATH); \
        echo "🧽 Configuración eliminada."; \
    else \
        echo "📦 Configuración preservada."; \
    fi
	@sudo systemctl daemon-reexec
	@sudo systemctl daemon-reload
	@echo "✅ Swaptimize desinstalado correctamente."

# Elimina binario local
clean:
	@echo "🧹 Limpiando binario local..."
	rm -f $(BIN_NAME)
