#!/bin/bash
# Script para completar la configuración de Podman
# Solo configura lo que falta (sin reinstalar)
# NOTA: El alias de Docker es OPCIONAL - testcontainers-go puede usar Podman directamente con DOCKER_HOST

set -e

echo "🔧 Completando configuración de Podman..."
echo ""

USER_ID=$(id -u)
GROUP_ID=$(id -g)

# Configurar subgid si falta
if ! grep -q "^${GROUP_ID}:" /etc/subgid 2>/dev/null; then
    echo "📝 Configurando subgid (requiere sudo)..."
    echo "${GROUP_ID}:100000:65536" | sudo tee -a /etc/subgid
    echo "✅ subgid configurado"
else
    echo "✅ subgid ya está configurado"
fi

# Configurar DOCKER_HOST en ~/.zshrc
DOCKER_HOST_VALUE="unix:///run/user/${USER_ID}/podman/podman.sock"

if [ -f ~/.zshrc ]; then
    if ! grep -q "DOCKER_HOST.*podman" ~/.zshrc 2>/dev/null; then
        echo ""
        echo "📝 Configurando DOCKER_HOST en ~/.zshrc..."
        echo "export DOCKER_HOST=${DOCKER_HOST_VALUE}" >> ~/.zshrc
        echo "✅ DOCKER_HOST añadido a ~/.zshrc"
    else
        echo "✅ DOCKER_HOST ya está en ~/.zshrc"
    fi
else
    echo "⚠️  ~/.zshrc no existe, creándolo..."
    echo "export DOCKER_HOST=${DOCKER_HOST_VALUE}" > ~/.zshrc
    echo "✅ ~/.zshrc creado con DOCKER_HOST"
fi

# Exportar para la sesión actual
export DOCKER_HOST="${DOCKER_HOST_VALUE}"

echo ""
echo "✅ Configuración completada!"
echo ""
echo "📋 Próximos pasos:"
echo "   1. Recarga tu shell: source ~/.zshrc"
echo "   2. O cierra y abre una nueva terminal"
echo "   3. Verifica: bash scripts/check-podman.sh"
echo "   4. Ejecuta tests: make test-e2e"
echo ""

