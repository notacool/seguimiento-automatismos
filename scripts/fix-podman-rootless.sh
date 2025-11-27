#!/bin/bash
# Script para corregir problemas de Podman rootless después del reinicio
# Soluciona: entradas duplicadas en subuid/subgid y permisos de newuidmap/newgidmap

set -e

echo "🔧 Corrigiendo configuración de Podman rootless..."
echo ""

USER_ID=$(id -u)
GROUP_ID=$(id -g)
USERNAME=$(id -un)

# 1. Eliminar entradas duplicadas en /etc/subuid
echo "📝 Limpiando entradas duplicadas en /etc/subuid..."
if grep -q "^${USERNAME}:" /etc/subuid 2>/dev/null; then
    echo "   Eliminando entrada por nombre de usuario: ${USERNAME}"
    sudo sed -i "/^${USERNAME}:/d" /etc/subuid
fi

# Verificar que existe la entrada numérica
if ! grep -q "^${USER_ID}:" /etc/subuid 2>/dev/null; then
    echo "   ⚠️  No se encontró entrada numérica, añadiendo..."
    echo "${USER_ID}:100000:65536" | sudo tee -a /etc/subuid
else
    echo "   ✅ Entrada numérica correcta: $(grep "^${USER_ID}:" /etc/subuid)"
fi

# 2. Eliminar entradas duplicadas en /etc/subgid
echo ""
echo "📝 Limpiando entradas duplicadas en /etc/subgid..."
if grep -q "^${USERNAME}:" /etc/subgid 2>/dev/null; then
    echo "   Eliminando entrada por nombre de usuario: ${USERNAME}"
    sudo sed -i "/^${USERNAME}:/d" /etc/subgid
fi

# Verificar que existe la entrada numérica
if ! grep -q "^${GROUP_ID}:" /etc/subgid 2>/dev/null; then
    echo "   ⚠️  No se encontró entrada numérica, añadiendo..."
    echo "${GROUP_ID}:100000:65536" | sudo tee -a /etc/subgid
else
    echo "   ✅ Entrada numérica correcta: $(grep "^${GROUP_ID}:" /etc/subgid)"
fi

# 3. Corregir permisos de newuidmap y newgidmap (activar setuid)
echo ""
echo "🔐 Corrigiendo permisos de newuidmap/newgidmap..."
NEWUIDMAP_PERMS=$(stat -c "%a" /usr/bin/newuidmap 2>/dev/null || echo "000")
NEWGIDMAP_PERMS=$(stat -c "%a" /usr/bin/newgidmap 2>/dev/null || echo "000")

if [[ ! "$NEWUIDMAP_PERMS" =~ ^[46]755$ ]]; then
    echo "   Activando setuid en /usr/bin/newuidmap..."
    sudo chmod u+s /usr/bin/newuidmap
    echo "   ✅ Permisos actualizados: $(stat -c "%a" /usr/bin/newuidmap)"
else
    echo "   ✅ Permisos de newuidmap correctos: $NEWUIDMAP_PERMS"
fi

if [[ ! "$NEWGIDMAP_PERMS" =~ ^[46]755$ ]]; then
    echo "   Activando setuid en /usr/bin/newgidmap..."
    sudo chmod u+s /usr/bin/newgidmap
    echo "   ✅ Permisos actualizados: $(stat -c "%a" /usr/bin/newgidmap)"
else
    echo "   ✅ Permisos de newgidmap correctos: $NEWGIDMAP_PERMS"
fi

# 4. Verificar configuración final
echo ""
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
echo "✅ Verificación final:"
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
echo ""

echo "📋 /etc/subuid:"
grep "^${USER_ID}:" /etc/subuid || echo "   ❌ No se encontró entrada"
echo ""

echo "📋 /etc/subgid:"
grep "^${GROUP_ID}:" /etc/subgid || echo "   ❌ No se encontró entrada"
echo ""

echo "🔐 Permisos de binarios:"
stat -c "   %a %n" /usr/bin/newuidmap /usr/bin/newgidmap
echo ""

# 5. Verificar socket de Podman
echo "🔌 Estado del socket de Podman:"
if systemctl --user is-active --quiet podman.socket 2>/dev/null; then
    echo "   ✅ Socket activo"
else
    echo "   ⚠️  Socket no activo, iniciando..."
    systemctl --user enable --now podman.socket 2>/dev/null || {
        echo "   ❌ No se pudo iniciar el socket"
    }
fi
echo ""

# 6. Advertencia sobre reinicio de sesión
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
echo "⚠️  IMPORTANTE:"
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
echo ""
echo "Los cambios en /etc/subuid y /etc/subgid requieren que el kernel"
echo "recargue los mapeos de usuario. Esto solo ocurre cuando:"
echo ""
echo "  1. Cierras sesión completamente (no solo la terminal)"
echo "  2. O reinicias el sistema"
echo ""
echo "Después de cerrar sesión y volver a entrar, prueba:"
echo ""
echo "  podman run --rm hello-world"
echo ""
echo "Si el problema persiste después de cerrar sesión, ejecuta:"
echo ""
echo "  bash scripts/check-podman.sh"
echo ""

