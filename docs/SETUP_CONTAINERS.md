# Configuración de Contenedores para Tests E2E

## Opciones Disponibles

### 1. Podman (Recomendado para Arch Linux)

**Instalación:**
```bash
sudo pacman -S podman podman-compose
```

**Configuración inicial:**
```bash
# Configurar subuid y subgid para rootless
echo "$(id -u):100000:65536" | sudo tee -a /etc/subuid
echo "$(id -g):100000:65536" | sudo tee -a /etc/subgid

# Iniciar servicio de usuario (opcional, para rootless)
systemctl --user enable --now podman.socket
```

**Configurar testcontainers para Podman:**
```bash
# Solo necesitas configurar DOCKER_HOST (el alias es opcional)
export DOCKER_HOST=unix:///run/user/$(id -u)/podman/podman.sock

# Añadir a ~/.zshrc para persistencia:
echo 'export DOCKER_HOST=unix:///run/user/$(id -u)/podman/podman.sock' >> ~/.zshrc
```

**Nota:** El alias de Docker (`docker -> podman`) NO es necesario. testcontainers-go detecta Podman automáticamente a través de `DOCKER_HOST`.

**Verificar instalación:**
```bash
podman --version
podman run hello-world
```

### 2. Docker (Alternativa)

**Instalación:**
```bash
sudo pacman -S docker docker-compose
sudo systemctl enable docker
sudo systemctl start docker
sudo usermod -aG docker $USER
# Cerrar sesión y volver a entrar, o:
newgrp docker
```

**Verificar instalación:**
```bash
docker --version
docker run hello-world
```

## Ejecutar Tests E2E

Una vez configurado Podman o Docker:

```bash
# Ejecutar todos los tests E2E
make test-e2e

# O directamente
go test -v ./test/e2e/...

# Ejecutar un test específico
go test -v ./test/e2e/... -run TestE2E_TaskCompleteLifecycle
```

## Troubleshooting

### Error: "Cannot connect to Docker daemon"

**Con Podman:**
```bash
# Verificar que el socket esté disponible
ls -la /run/user/$(id -u)/podman/podman.sock

# Si no existe, iniciar el servicio
systemctl --user start podman.socket

# Configurar variable de entorno
export DOCKER_HOST=unix:///run/user/$(id -u)/podman/podman.sock
```

**Con Docker:**
```bash
# Verificar que el servicio esté corriendo
sudo systemctl status docker

# Iniciar si no está corriendo
sudo systemctl start docker

# Verificar permisos
groups | grep docker
```

### Error: "permission denied"

**Con Podman (rootless):**
- Asegúrate de haber configurado subuid/subgid
- Verifica que el socket tenga permisos correctos

**Con Docker:**
- Añade tu usuario al grupo docker: `sudo usermod -aG docker $USER`
- Cierra sesión y vuelve a entrar

## Notas

- **testcontainers-go** detecta automáticamente si está disponible Docker o Podman
- Si ambos están instalados, Docker tiene prioridad
- Los tests E2E crean contenedores temporales que se limpian automáticamente
- Los contenedores se eliminan incluso si los tests fallan

