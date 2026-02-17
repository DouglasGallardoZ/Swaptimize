#!/bin/bash

#############################################
# Swaptimize Monitoring Stack Setup Script
# Prometheus + Grafana in Docker
#############################################

set -e

echo "🚀 Swaptimize Monitoring Setup"
echo "================================"

# Colors
GREEN='\033[0;32m'
BLUE='\033[0;34m'
RED='\033[0;31m'
NC='\033[0m' # No Color

# Check Docker
echo -e "${BLUE}[1/5]${NC} Checking Docker installation..."
if ! command -v docker &> /dev/null; then
    echo -e "${RED}❌ Docker not found. Install with: curl -fsSL https://get.docker.com | sh${NC}"
    exit 1
fi

if ! command -v docker-compose &> /dev/null && ! docker compose version &> /dev/null; then
    echo -e "${RED}❌ Docker Compose not found${NC}"
    exit 1
fi

echo -e "${GREEN}✅ Docker found${NC}"

# Check Swaptimize is running
echo -e "${BLUE}[2/5]${NC} Checking Swaptimize daemon..."
if ! curl -s http://localhost:9100/health &> /dev/null; then
    echo -e "${RED}⚠️  Swaptimize not responding on :9100${NC}"
    echo "Please start Swaptimize first:"
    echo "  sudo systemctl start swaptimize"
    exit 1
fi
echo -e "${GREEN}✅ Swaptimize is running${NC}"

# Create directories
echo -e "${BLUE}[3/5]${NC} Creating directories..."
mkdir -p grafana/provisioning/datasources
mkdir -p grafana/provisioning/dashboards
echo -e "${GREEN}✅ Directories created${NC}"

# Check if configs exist
echo -e "${BLUE}[4/5]${NC} Verifying configuration files..."
for file in docker-compose.yml prometheus.yml alert_rules.yml \
            grafana/provisioning/datasources/prometheus.yml \
            grafana/provisioning/dashboards/provisioning.yml \
            grafana/provisioning/dashboards/swaptimize-dashboard.json; do
    if [ ! -f "$file" ]; then
        echo -e "${RED}❌ Missing: $file${NC}"
        exit 1
    fi
done
echo -e "${GREEN}✅ All config files present${NC}"

# Start services
echo -e "${BLUE}[5/5]${NC} Starting Prometheus + Grafana..."
docker-compose up -d

# Wait for services
echo "⏳ Waiting for services to start..."
sleep 5

# Verification
echo ""
echo -e "${GREEN}✅ Setup complete!${NC}"
echo ""
echo "📊 Monitoring URLs:"
echo "  Grafana:         ${BLUE}http://localhost:3000${NC}         (admin/admin)"
echo "  Prometheus:      ${BLUE}http://localhost:9090${NC}"
echo "  Swaptimize:      ${BLUE}http://localhost:9100/metrics${NC}"
echo ""
echo "📈 Next steps:"
echo "  1. Open Grafana: http://localhost:3000"
echo "  2. Login with admin/admin"
echo "  3. Go to Dashboard → Swaptimize v2.0 Monitoring Dashboard"
echo "  4. Watch real-time metrics"
echo ""
echo "🛑 To stop services:"
echo "  docker-compose down"
echo ""
echo "📖 More info:"
echo "  cat MONITORING.md"
