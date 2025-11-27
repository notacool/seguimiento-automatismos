#!/bin/bash
# Script para iniciar el servicio de Podman para testcontainers

set -e

USER_ID=$(id -u)
SOCKET_PATH="/run/user/${USER_ID}/podman/podman.sock"

# Verificar si el servicio ya está funcionando
if pgrep -f "^[^ ]*podman[^ ]* system service" > /dev/null 2>&1; then
    # Verificar que el servicio realmente responda
    if command -v curl > /dev/null 2>&1; then
        if curl -s --max-time 2 --unix-socket "$SOCKET_PATH" http://d/v3.0.0/libpod/info > /dev/null 2>&1; then
            echo "✅ Servicio de Podman ya está funcionando"
            exit 0
        fi
    else
        # Si no hay curl, asumimos que está funcionando si el proceso existe
        echo "✅ Servicio de Podman detectado (curl no disponible para verificar)"
        exit 0
    fi
fi

# Si llegamos aquí, el servicio no está funcionando o no responde
echo "⚠️  Iniciando servicio de Podman..."

# Matar cualquier proceso anterior
pkill -f "podman.*system service" 2>/dev/null || true
sleep 1

# Iniciar el servicio
podman system service --time=0 > /dev/null 2>&1 &
SERVICE_PID=$!

# Esperar a que el servicio esté listo
for i in {1..10}; do
    sleep 1
    if command -v curl > /dev/null 2>&1; then
        if curl -s --max-time 2 --unix-socket "$SOCKET_PATH" http://d/v3.0.0/libpod/info > /dev/null 2>&1; then
            echo "✅ Servicio de Podman iniciado correctamente (PID: $SERVICE_PID)"
            exit 0
        fi
    else
        # Si no hay curl, esperar un poco y asumir que funciona
        if [ $i -ge 3 ]; then
            echo "✅ Servicio de Podman iniciado (PID: $SERVICE_PID, curl no disponible para verificar)"
            exit 0
        fi
    fi
done

echo "❌ Error: El servicio de Podman no responde después de 10 segundos"
exit 1

