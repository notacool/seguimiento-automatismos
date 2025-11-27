#!/bin/bash
# Script para instalar y configurar Podman en Arch Linux
# Para usar con testcontainers-go en los tests E2E

set -e

echo "🐳 Configurando Podman para tests E2E..."
echo ""

# Verificar si ya está instalado
if command -v podman &> /dev/null; then
    echo "✓ Podman ya está instalado: $(podman --version)"
else
    echo "📦 Instalando Podman y dependencias..."
    sudo pacman -S --needed podman podman-compose shadow
    echo "✓ Podman instalado"
fi

# Verificar shadow (necesario para newuidmap/newgidmap en rootless)
if ! command -v newuidmap &> /dev/null || ! command -v newgidmap &> /dev/null; then
    echo "📦 Instalando shadow (requerido para Podman rootless)..."
    sudo pacman -S --needed shadow
    echo "✓ shadow instalado"
fi

echo ""
echo "🔧 Configurando Podman para rootless..."

# Configurar subuid y subgid si no existen
USER_ID=$(id -u)
GROUP_ID=$(id -g)

if ! grep -q "^${USER_ID}:" /etc/subuid 2>/dev/null; then
    echo "Configurando subuid..."
    echo "${USER_ID}:100000:65536" | sudo tee -a /etc/subuid
    echo "✓ subuid configurado"
else
    echo "✓ subuid ya configurado"
fi

if ! grep -q "^${GROUP_ID}:" /etc/subgid 2>/dev/null; then
    echo "Configurando subgid..."
    echo "${GROUP_ID}:100000:65536" | sudo tee -a /etc/subgid
    echo "✓ subgid configurado"
else
    echo "✓ subgid ya configurado"
fi

echo ""
echo "🔌 Configurando socket de Podman..."

# Iniciar y habilitar socket de Podman para usuario
systemctl --user enable --now podman.socket 2>/dev/null || {
    echo "⚠️  No se pudo iniciar podman.socket automáticamente"
    echo "   Ejecuta manualmente: systemctl --user enable --now podman.socket"
}

echo ""
echo "🔗 Configurando para testcontainers-go..."

# NOTA: El alias de Docker es OPCIONAL
# testcontainers-go detecta Podman automáticamente a través de DOCKER_HOST
# Solo creamos el alias si el usuario lo necesita para scripts que usan 'docker' directamente

# Crear directorio local/bin si no existe (por si se necesita el alias)
mkdir -p ~/.local/bin

# Crear alias de Docker si no existe (OPCIONAL)
if [ ! -f ~/.local/bin/docker ] || [ ! -L ~/.local/bin/docker ]; then
    DOCKER_PATH=$(which podman)
    if [ -n "$DOCKER_PATH" ]; then
        ln -sf "$DOCKER_PATH" ~/.local/bin/docker
        echo "✓ Alias de Docker creado (opcional): ~/.local/bin/docker -> $DOCKER_PATH"
        echo "  ℹ️  Este alias NO es necesario para testcontainers-go si DOCKER_HOST está configurado"
    else
        echo "⚠️  No se encontró podman en PATH"
    fi
else
    echo "✓ Alias de Docker ya existe (opcional)"
fi

# Añadir ~/.local/bin al PATH si no está
if [[ ":$PATH:" != *":$HOME/.local/bin:"* ]]; then
    echo ""
    echo "📝 Añadiendo ~/.local/bin al PATH..."
    
    # Detectar shell
    if [ -n "$ZSH_VERSION" ]; then
        SHELL_RC="$HOME/.zshrc"
    elif [ -n "$BASH_VERSION" ]; then
        SHELL_RC="$HOME/.bashrc"
    else
        SHELL_RC="$HOME/.profile"
    fi
    
    if ! grep -q '\.local/bin' "$SHELL_RC" 2>/dev/null; then
        echo 'export PATH="$HOME/.local/bin:$PATH"' >> "$SHELL_RC"
        echo "✓ PATH actualizado en $SHELL_RC"
        echo "  Ejecuta: source $SHELL_RC"
    else
        echo "✓ PATH ya está configurado en $SHELL_RC"
    fi
fi

echo ""
echo "🌐 Configurando variable de entorno DOCKER_HOST..."

# Configurar DOCKER_HOST para testcontainers
DOCKER_HOST="unix:///run/user/${USER_ID}/podman/podman.sock"

if [ -n "$ZSH_VERSION" ]; then
    SHELL_RC="$HOME/.zshrc"
elif [ -n "$BASH_VERSION" ]; then
    SHELL_RC="$HOME/.bashrc"
else
    SHELL_RC="$HOME/.profile"
fi

if ! grep -q "DOCKER_HOST.*podman" "$SHELL_RC" 2>/dev/null; then
    echo "export DOCKER_HOST=${DOCKER_HOST}" >> "$SHELL_RC"
    echo "✓ DOCKER_HOST configurado en $SHELL_RC"
    echo "  Ejecuta: source $SHELL_RC"
else
    echo "✓ DOCKER_HOST ya está configurado"
fi

# Exportar para la sesión actual
export DOCKER_HOST="${DOCKER_HOST}"
export PATH="$HOME/.local/bin:$PATH"

echo ""
echo "✅ Verificando instalación..."

# Verificar Podman
if podman --version &> /dev/null; then
    echo "✓ Podman funciona: $(podman --version)"
else
    echo "❌ Error: Podman no funciona"
    exit 1
fi

# Verificar socket
if [ -S "/run/user/${USER_ID}/podman/podman.sock" ]; then
    echo "✓ Socket de Podman disponible"
else
    echo "⚠️  Socket de Podman no encontrado. Iniciando servicio..."
    systemctl --user start podman.socket
    sleep 2
    if [ -S "/run/user/${USER_ID}/podman/podman.sock" ]; then
        echo "✓ Socket de Podman ahora disponible"
    else
        echo "❌ Error: No se pudo iniciar el socket de Podman"
        echo "   Ejecuta manualmente: systemctl --user enable --now podman.socket"
    fi
fi

# Probar con hello-world
echo ""
echo "🧪 Probando Podman con hello-world..."
if podman run --rm hello-world &> /dev/null; then
    echo "✓ Podman funciona correctamente"
else
    echo "⚠️  Advertencia: No se pudo ejecutar hello-world (puede ser normal si no hay imágenes)"
    echo "   Esto no impide que testcontainers funcione"
fi

echo ""
echo "🎉 Configuración completada!"
echo ""
echo "📋 Próximos pasos:"
echo "   1. Recarga tu shell: source $SHELL_RC"
echo "   2. O cierra y abre una nueva terminal"
echo "   3. Verifica: podman --version"
echo "   4. Ejecuta tests E2E: make test-e2e"
echo ""
echo "💡 Nota: Si tienes problemas, verifica:"
echo "   - systemctl --user status podman.socket"
echo "   - ls -la /run/user/${USER_ID}/podman/podman.sock"
echo "   - echo \$DOCKER_HOST"

