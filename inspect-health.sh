#!/bin/bash

# Script para inspeccionar la salud y configuración de Swaptimize vía HTTP
# Uso: ./inspect-health.sh [filter]

HEALTH_URL="http://localhost:9100/health"
FILTER="${1:-.}"  # Default: mostrar todo

echo "🔍 Swaptimize Health & Configuration Inspector"
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
echo ""

# Colores
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Verificar si el endpoint está disponible
# if ! command -v curl &> /dev/null; then
#     echo -e "${RED}❌ curl no está instalado${NC}"
#     exit 1
# fi

# Obtener respuesta
RESPONSE=$(curl -s -w "\n%{http_code}" "$HEALTH_URL" 2>&1)
HTTP_CODE=$(echo "$RESPONSE" | tail -n1)
BODY=$(echo "$RESPONSE" | head -n-1)

# Verificar status HTTP
if [ "$HTTP_CODE" != "200" ]; then
    echo -e "${RED}❌ Error HTTP $HTTP_CODE${NC}"
    echo ""
    echo "Posibles causas:"
    echo "  1. El daemon no está corriendo: sudo systemctl start swaptimize"
    echo "  2. El puerto 9100 no está escuchando"
    echo "  3. El network interface está cerrado"
    echo ""
    echo "Verificar:"
    echo "  sudo systemctl status swaptimize"
    echo "  sudo netstat -tlnp | grep 9100"
    exit 1
fi

# Verificar si es JSON válido
if ! echo "$BODY" | jq . > /dev/null 2>&1; then
    echo -e "${RED}❌ La respuesta no es JSON válido${NC}"
    echo "Respuesta cruda:"
    echo "$BODY"
    exit 1
fi

# Si se especifica un filtro, aplicarlo
if [ "$FILTER" != "." ]; then
    echo -e "${BLUE}🔎 Filtrando por: $FILTER${NC}"
    echo ""
    echo "$BODY" | jq "$FILTER"
    exit 0
fi

# Salida formateada por defecto
echo -e "${GREEN}✅ Daemon Status${NC}"
STATUS=$(echo "$BODY" | jq -r '.status')
UPTIME=$(echo "$BODY" | jq -r '.uptime_seconds')
TIMESTAMP=$(echo "$BODY" | jq -r '.timestamp')

echo "  Status: $STATUS"
echo "  Uptime: $(printf '%.0f' $UPTIME)s ($(printf '%d' $((UPTIME / 3600)))h $(printf '%d' $(((UPTIME % 3600) / 60)))m)"
echo "  Timestamp: $TIMESTAMP"
echo ""

echo -e "${GREEN}📍 Environment Detected${NC}"
ENV_TYPE=$(echo "$BODY" | jq -r '.configuration.environment.type')
ENV_DESC=$(echo "$BODY" | jq -r '.configuration.environment.description')

echo "  Type: $ENV_TYPE"
echo "  Description: $ENV_DESC"
echo ""

echo -e "${GREEN}⚙️  Profile & Thresholds${NC}"
PROFILE=$(echo "$BODY" | jq -r '.configuration.profile.name')
MODE=$(echo "$BODY" | jq -r '.configuration.profile.mode')
THRESHOLD_HIGH=$(echo "$BODY" | jq -r '.configuration.profile.threshold_high_percent')
THRESHOLD_LOW=$(echo "$BODY" | jq -r '.configuration.profile.threshold_low_percent')
MAX_FILES=$(echo "$BODY" | jq -r '.configuration.profile.max_swap_files')
SWAP_SIZE=$(echo "$BODY" | jq -r '.configuration.profile.swap_size_mb')
ALLOW_CREATION=$(echo "$BODY" | jq -r '.configuration.profile.allow_creation')

echo "  Profile: $PROFILE"
echo "  Mode: $MODE"
echo "  Threshold High: ${THRESHOLD_HIGH}%"
echo "  Threshold Low: ${THRESHOLD_LOW}%"
echo "  Max Swap Files: $MAX_FILES"
echo "  Swap Size: ${SWAP_SIZE} MB"
echo "  Allow Creation: $ALLOW_CREATION"
echo ""

echo -e "${GREEN}📊 Current Metrics${NC}"
MEM_PERCENT=$(echo "$BODY" | jq -r '.metrics.memory_usage_percent')
SWAP_PERCENT=$(echo "$BODY" | jq -r '.metrics.swap_usage_percent')
FILES_CREATED=$(echo "$BODY" | jq -r '.metrics.swap_files_created')
FILES_DELETED=$(echo "$BODY" | jq -r '.metrics.swap_files_deleted')
OPS_TOTAL=$(echo "$BODY" | jq -r '.metrics.operations_total')
ERRORS=$(echo "$BODY" | jq -r '.metrics.errors_total')
RETRIES=$(echo "$BODY" | jq -r '.metrics.retry_attempts_total')

echo "  Memory Usage: ${MEM_PERCENT}%"
echo "  Swap Usage: ${SWAP_PERCENT}%"
echo "  Swap Files Created: $FILES_CREATED"
echo "  Swap Files Deleted: $FILES_DELETED"
echo "  Total Operations: $OPS_TOTAL"
echo "  Total Errors: $ERRORS"
echo "  Total Retries: $RETRIES"
echo ""

# Alertas
if [ "$STATUS" != "healthy" ]; then
    echo -e "${YELLOW}⚠️  Warning: Status is $STATUS${NC}"
fi

if [ "$(echo "$ERRORS" | head -c 1)" != "0" ]; then
    echo -e "${YELLOW}⚠️  Warning: ${ERRORS} errors recorded${NC}"
fi

echo ""
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
echo ""
echo "Comandos útiles:"
echo "  • Ver JSON completo:     curl $HEALTH_URL | jq"
echo "  • Solo configuración:    curl $HEALTH_URL | jq '.configuration'"
echo "  • Solo perfil:           curl $HEALTH_URL | jq '.configuration.profile'"
echo "  • Solo environment:      curl $HEALTH_URL | jq '.configuration.environment'"
echo "  • Solo métricas:         curl $HEALTH_URL | jq '.metrics'"
echo "  • Ver logs:              sudo journalctl -u swaptimize -f"
echo "  • Ver status:            sudo systemctl status swaptimize"
echo ""
