#!/bin/bash

#############################################
# Validar Perfil Detectado por Swaptimize
#############################################

set -e

# Colors
GREEN='\033[0;32m'
BLUE='\033[0;34m'
YELLOW='\033[1;33m'
RED='\033[0;31m'
NC='\033[0m' # No Color

echo -e "${BLUE}╔════════════════════════════════════════════╗${NC}"
echo -e "${BLUE}║  Validación de Perfil - Swaptimize v2.0   ║${NC}"
echo -e "${BLUE}╚════════════════════════════════════════════╝${NC}"
echo ""

# 1. Check daemon status
echo -e "${BLUE}[1]${NC} Estado del Daemon"
echo "════════════════════════════════════════════"
if systemctl is-active --quiet swaptimize; then
    echo -e "${GREEN}✅ Daemon running${NC}"
else
    echo -e "${RED}❌ Daemon NOT running${NC}"
    echo "Start with: sudo systemctl start swaptimize"
    exit 1
fi
echo ""

# 2. Extract environment from logs
echo -e "${BLUE}[2]${NC} Ambiente Detectado"
echo "════════════════════════════════════════════"
ENV=$(sudo journalctl -u swaptimize -n 200 | grep -i "detected:" | tail -1 || echo "")
if [ -z "$ENV" ]; then
    echo -e "${YELLOW}⚠️  No environment detection found in recent logs${NC}"
    echo "Check full logs: sudo journalctl -u swaptimize -n 500"
else
    echo -e "${GREEN}$ENV${NC}"
fi
echo ""

# 3. Extract profile from logs
echo -e "${BLUE}[3]${NC} Perfil Seleccionado"
echo "════════════════════════════════════════════"
PROFILE=$(sudo journalctl -u swaptimize -n 200 | grep -i "profile selected" | tail -1 || echo "")
if [ -z "$PROFILE" ]; then
    echo -e "${YELLOW}⚠️  No profile selection found${NC}"
else
    echo -e "${GREEN}$PROFILE${NC}"
fi
echo ""

# 4. Extract threshold config
echo -e "${BLUE}[4]${NC} Configuración Adaptada"
echo "════════════════════════════════════════════"
THRESHOLD=$(sudo journalctl -u swaptimize -n 200 | grep -i "thresholdHigh\|MaxSwapFiles" | tail -2 || echo "")
if [ -z "$THRESHOLD" ]; then
    echo -e "${YELLOW}⚠️  No threshold config found${NC}"
else
    echo -e "${GREEN}$THRESHOLD${NC}"
fi
echo ""

# 5. Check health endpoint
echo -e "${BLUE}[5]${NC} Health Check (HTTP)"
echo "════════════════════════════════════════════"
HEALTH=$(curl -s http://localhost:9100/health 2>/dev/null || echo "")
if [ -z "$HEALTH" ]; then
    echo -e "${YELLOW}⚠️  Swaptimize metrics endpoint not responding${NC}"
    echo "Check: curl http://localhost:9100/health"
else
    echo -e "${GREEN}✅ Metrics endpoint responding${NC}"
    echo "$HEALTH" | jq '.' 2>/dev/null || echo "$HEALTH"
fi
echo ""

# 6. System info (for reference)
echo -e "${BLUE}[6]${NC} Info del Sistema"
echo "════════════════════════════════════════════"
echo "Memory Total: $(free -h | grep ^Mem | awk '{print $2}')"
echo "Memory Used:  $(free -h | grep ^Mem | awk '{print $3}')"
echo "Swap Total:   $(free -h | grep ^Swap | awk '{print $2}')"
echo "Swap Used:    $(free -h | grep ^Swap | awk '{print $3}')"
echo ""

# 7. Check ZSWAP status
echo -e "${BLUE}[7]${NC} ZSWAP/ZRAM Status"
echo "════════════════════════════════════════════"
if [ -f "/sys/module/zswap/parameters/enabled" ]; then
    ZSWAP=$(cat /sys/module/zswap/parameters/enabled)
    if [ "$ZSWAP" = "Y" ]; then
        echo -e "${GREEN}✅ ZSWAP: Enabled${NC}"
    else
        echo -e "${YELLOW}⚠️  ZSWAP: Disabled${NC}"
    fi
else
    echo -e "${RED}❌ ZSWAP: Not available${NC}"
fi

if lsblk 2>/dev/null | grep -q zram; then
    echo -e "${GREEN}✅ ZRAM: Detected${NC}"
else
    echo -e "${YELLOW}⚠️  ZRAM: Not available${NC}"
fi
echo ""

# 8. Check current swap files
echo -e "${BLUE}[8]${NC} Archivos Swap Activos"
echo "════════════════════════════════════════════"
swapon --show || echo "No swap files active"
echo ""

# 9. Show full daemon startup logs
echo -e "${BLUE}[9]${NC} Logs Completos de Inicio"
echo "════════════════════════════════════════════"
echo "Últimas 50 líneas (contiene detección de environment/profile):"
echo "---"
sudo journalctl -u swaptimize -n 50 --no-pager | tail -30
echo "---"
echo ""

# 10. Summary
echo -e "${BLUE}╔════════════════════════════════════════════╗${NC}"
echo -e "${BLUE}║  RESUMEN${NC}"
echo -e "${BLUE}╚════════════════════════════════════════════╝${NC}"
echo ""
echo -e "Para validar perfil, busca en logs:"
echo -e "  ${GREEN}• 'Detected: ZSWAP_ONLY' o similar${NC}"
echo -e "  ${GREEN}• 'Profile selected: laptop|workstation|server'${NC}"
echo -e "  ${GREEN}• 'ThresholdHigh:', 'MaxSwapFiles:'${NC}"
echo ""
echo -e "Comando rápido:"
echo -e "  ${YELLOW}sudo journalctl -u swaptimize -n 100 | grep -i profile${NC}"
echo ""
