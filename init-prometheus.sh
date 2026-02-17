#!/bin/bash
# Script para generar prometheus.yml dinámicamente con la IP del host

set -e

# Cargar variables del .env
if [ -f ".env.docker" ]; then
    export $(grep -v '^#' .env.docker | xargs)
else
    echo "❌ .env.docker no encontrado"
    exit 1
fi

# Usar IP del host si no está definida
HOST_IP=${HOST_IP:-192.168.1.14}

echo "📝 Generando prometheus.yml con HOST_IP=$HOST_IP"

# Generar prometheus.yml
cat > prometheus.yml <<EOF
global:
  scrape_interval: ${PROMETHEUS_SCRAPE_INTERVAL:-30s}
  evaluation_interval: 30s
  external_labels:
    monitor: 'swaptimize-monitor'

alerting:
  alertmanagers:
    - static_configs:
        - targets: []

rule_files:
  - 'alert_rules.yml'

scrape_configs:
  - job_name: 'swaptimize'
    static_configs:
      - targets: ['${HOST_IP}:9100']  # Swaptimize metrics endpoint (Host IP)
    scrape_interval: 30s
    scrape_timeout: 10s
    metrics_path: '/metrics'
    scheme: http

  - job_name: 'prometheus'
    static_configs:
      - targets: ['localhost:9090']
EOF

echo "✅ prometheus.yml generado correctamente"
