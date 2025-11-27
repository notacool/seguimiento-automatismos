#!/bin/bash
# Script para verificar y mostrar comandos necesarios para configurar Podman

echo "🔍 Verificando estado de Podman..."
echo ""

# Verificar Podman
if command -v podman &> /dev/null; then
    echo "✅ Podman instalado: $(podman --version)"
    PODMAN_INSTALLED=true
else
    echo "❌ Podman NO está instalado"
    PODMAN_INSTALLED=false
fi

# Verificar socket
USER_ID=$(id -u)
SOCKET_PATH="/run/user/${USER_ID}/podman/podman.sock"
if [ -S "$SOCKET_PATH" ]; then
    echo "✅ Socket de Podman disponible: $SOCKET_PATH"
    SOCKET_OK=true
else
    echo "❌ Socket de Podman NO encontrado: $SOCKET_PATH"
    SOCKET_OK=false
fi

# Verificar subuid/subgid
GROUP_ID=$(id -g)
if grep -q "^${USER_ID}:" /etc/subuid 2>/dev/null; then
    echo "✅ subuid configurado"
    SUBUID_OK=true
else
    echo "❌ subuid NO configurado"
    SUBUID_OK=false
fi

if grep -q "^${GROUP_ID}:" /etc/subgid 2>/dev/null; then
    echo "✅ subgid configurado"
    SUBGID_OK=true
else
    echo "❌ subgid NO configurado"
    SUBGID_OK=false
fi

# Verificar alias de Docker (OPCIONAL - solo necesario si testcontainers no detecta Podman)
if [ -L ~/.local/bin/docker ] && [ "$(readlink ~/.local/bin/docker)" = "$(which podman)" ]; then
    echo "✅ Alias de Docker configurado (opcional)"
    ALIAS_OK=true
else
    echo "ℹ️  Alias de Docker NO configurado (opcional - solo si testcontainers no detecta Podman)"
    ALIAS_OK=true  # No es crítico, solo opcional
fi

# Verificar DOCKER_HOST
if [ -n "$DOCKER_HOST" ] && [[ "$DOCKER_HOST" == *"podman.sock"* ]]; then
    echo "✅ DOCKER_HOST configurado: $DOCKER_HOST"
    DOCKER_HOST_OK=true
else
    echo "❌ DOCKER_HOST NO configurado o incorrecto"
    DOCKER_HOST_OK=false
fi

echo ""
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
echo "📋 COMANDOS A EJECUTAR:"
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
echo ""

if [ "$PODMAN_INSTALLED" = false ]; then
    echo "1️⃣  Instalar Podman:"
    echo "   sudo pacman -S podman podman-compose"
    echo ""
fi

if [ "$SUBUID_OK" = false ] || [ "$SUBGID_OK" = false ]; then
    echo "2️⃣  Configurar rootless (requiere sudo):"
    echo "   echo \"${USER_ID}:100000:65536\" | sudo tee -a /etc/subuid"
    echo "   echo \"${GROUP_ID}:100000:65536\" | sudo tee -a /etc/subgid"
    echo ""
fi

if [ "$SOCKET_OK" = false ]; then
    echo "3️⃣  Iniciar socket de Podman:"
    echo "   systemctl --user enable --now podman.socket"
    echo ""
fi

if [ "$ALIAS_OK" = false ] && [ -n "$(which podman)" ]; then
    echo "4️⃣  Crear alias de Docker (OPCIONAL - solo si testcontainers no detecta Podman):"
    echo "   mkdir -p ~/.local/bin"
    echo "   ln -s \$(which podman) ~/.local/bin/docker"
    echo "   export PATH=\"\$HOME/.local/bin:\$PATH\""
    echo ""
fi

if [ "$DOCKER_HOST_OK" = false ]; then
    echo "5️⃣  Configurar DOCKER_HOST (añadir a ~/.zshrc):"
    echo "   echo 'export DOCKER_HOST=unix:///run/user/${USER_ID}/podman/podman.sock' >> ~/.zshrc"
    echo "   source ~/.zshrc"
    echo ""
fi

if [ "$PODMAN_INSTALLED" = true ] && [ "$SOCKET_OK" = true ] && [ "$SUBUID_OK" = true ] && [ "$SUBGID_OK" = true ] && [ "$DOCKER_HOST_OK" = true ]; then
    echo "✅ ¡Todo está configurado correctamente!"
    echo ""
    echo "🧪 Puedes ejecutar los tests E2E:"
    echo "   make test-e2e"
else
    echo "⚠️  Ejecuta los comandos marcados arriba según lo que falte."
    echo ""
    echo "💡 Después de ejecutar los comandos, vuelve a ejecutar este script para verificar:"
    echo "   bash scripts/check-podman.sh"
fi

echo ""

