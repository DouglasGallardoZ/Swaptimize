#!/bin/bash
# Script de diagnóstico para problemas de swap en Swaptimize

echo "🔍 DIAGNÓSTICO DE SWAP - Swaptimize"
echo "===================================="
echo ""

# 1. Verificar filesystem
echo "1️⃣ Filesystem Detectado:"
df -T /var/lib/
echo ""

# 2. Verificar permisos del directorio
echo "2️⃣ Permisos del directorio /var/lib/swaptimize:"
ls -ld /var/lib/swaptimize/ 2>/dev/null || echo "❌ Directorio no existe"
echo ""

# 3. Ver archivo swap existente
echo "3️⃣ Archivos swap existentes:"
ls -lh /var/lib/swaptimize/swap-* 2>/dev/null || echo "❌ No hay archivos"
echo ""

# 4. Verificar atributos btrfs
echo "4️⃣ Atributos (COW status):"
lsattr /var/lib/swaptimize/ 2>/dev/null || echo "⚠️  lsattr no disponible"
echo ""

# 5. Verificar estado de swap activo
echo "5️⃣ Swaps activos en el sistema:"
swapon --show 2>/dev/null || echo "❌ swapon no disponible"
echo ""

# 6. Test manual de creación
echo "6️⃣ TEST MANUAL de creación de swap:"
TEST_FILE="/var/lib/swaptimize/test-swap-manual"

# Limpiar si existe
sudo rm -f $TEST_FILE 2>/dev/null

# 1. Crear archivo
echo "  a) Creando archivo..."
sudo fallocate -l 256M $TEST_FILE 2>&1 && echo "      ✓ fallocate OK" || echo "      ❌ fallocate falló"

# 2. Permisos
echo "  b) Estableciendo permisos 0600..."
sudo chmod 0600 $TEST_FILE 2>&1 && echo "      ✓ chmod OK" || echo "      ❌ chmod falló"

# 3. mkswap
echo "  c) Ejecutando mkswap..."
sudo mkswap $TEST_FILE 2>&1 | head -3
if [ ${PIPESTATUS[0]} -eq 0 ]; then
    echo "      ✓ mkswap OK"
else
    echo "      ❌ mkswap falló"
fi

# 4. swapon
echo "  d) Ejecutando swapon..."
sudo swapon $TEST_FILE 2>&1
if [ ${PIPESTATUS[0]} -eq 0 ]; then
    echo "      ✓ swapon OK"
    sudo swapoff $TEST_FILE 2>/dev/null
    sudo rm -f $TEST_FILE
    echo "      (Limpiado)"
else
    echo "      ❌ swapon FALLÓ - VER MENSAAJE ARRIBA"
fi
echo ""

# 7. Verificar SELinux
echo "7️⃣ SELinux status:"
getenforce 2>/dev/null || echo "⚠️  SELinux no disponible"
echo ""

# 8. Verificar límites del sistema
echo "8️⃣ Límites de archivos/procesos:"
ulimit -n
echo ""

# 9. Ver logs del kernel si falló
echo "9️⃣ Últimos errores del kernel:"
dmesg | tail -5 | grep -i "swap\|memory" || echo "(sin errores de swap/memoria)"
echo ""

echo "📝 FIN DEL DIAGNÓSTICO"
