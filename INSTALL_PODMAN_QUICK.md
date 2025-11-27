# Instalación Rápida de Podman

## Estado Actual
- ❌ Podman no está instalado
- ❌ Socket de Podman no configurado
- ❌ DOCKER_HOST no configurado
- ✅ PATH incluye ~/.local/bin

## Pasos para Instalar

### 1. Ejecutar Script de Instalación

```bash
bash scripts/setup-podman.sh
```

**Nota:** El script pedirá tu contraseña de sudo para:
- Instalar Podman
- Configurar subuid/subgid

### 2. O Instalación Manual

Si prefieres hacerlo manualmente:

```bash
# 1. Instalar Podman y shadow (necesario para rootless)
sudo pacman -S podman podman-compose shadow

# 2. Configurar rootless
echo "$(id -u):100000:65536" | sudo tee -a /etc/subuid
echo "$(id -g):100000:65536" | sudo tee -a /etc/subgid

# 3. Iniciar socket de Podman
systemctl --user enable --now podman.socket

# 4. Crear alias de Docker
mkdir -p ~/.local/bin
ln -s $(which podman) ~/.local/bin/docker

# 5. Configurar DOCKER_HOST (añadir a ~/.zshrc)
echo 'export DOCKER_HOST=unix:///run/user/$(id -u)/podman/podman.sock' >> ~/.zshrc

# 6. Recargar shell
source ~/.zshrc
```

### 3. Verificar Instalación

```bash
# Verificar Podman
podman --version

# Verificar socket
ls -la /run/user/$(id -u)/podman/podman.sock

# Verificar DOCKER_HOST
echo $DOCKER_HOST
# Debe mostrar: unix:///run/user/1000/podman/podman.sock

# Probar con hello-world
podman run hello-world
```

### 4. Ejecutar Tests E2E

Una vez configurado:

```bash
make test-e2e
```

## Solución de Problemas

Si después de instalar sigues viendo el error `checked path: $XDG_RUNTIME_DIR`:

1. Verifica que el socket exista:
   ```bash
   ls -la /run/user/$(id -u)/podman/podman.sock
   ```

2. Si no existe, inicia el servicio:
   ```bash
   systemctl --user start podman.socket
   ```

3. Verifica DOCKER_HOST:
   ```bash
   echo $DOCKER_HOST
   # Si está vacío, recarga tu shell: source ~/.zshrc
   ```

