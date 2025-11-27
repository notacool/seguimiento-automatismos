# Instalación de Podman para Tests E2E

## Instalación Rápida (Automática)

Ejecuta el script de instalación:

```bash
./scripts/setup-podman.sh
```

Este script:
- ✅ Instala Podman y podman-compose
- ✅ Configura subuid/subgid para rootless
- ✅ Inicia el socket de Podman
- ✅ Crea alias de Docker para testcontainers
- ✅ Configura variables de entorno
- ✅ Verifica la instalación

## Instalación Manual

### 1. Instalar Podman

```bash
sudo pacman -S podman podman-compose shadow
```

**Nota:** El paquete `shadow` es necesario para que Podman funcione en modo rootless, ya que proporciona las herramientas `newuidmap` y `newgidmap` requeridas para el mapeo de usuarios en contenedores.

### 2. Configurar Rootless

```bash
# Configurar subuid y subgid
echo "$(id -u):100000:65536" | sudo tee -a /etc/subuid
echo "$(id -g):100000:65536" | sudo tee -a /etc/subgid
```

### 3. Iniciar Socket de Podman

```bash
# Habilitar e iniciar socket de usuario
systemctl --user enable --now podman.socket

# Verificar que esté corriendo
systemctl --user status podman.socket
```

### 4. Configurar para testcontainers-go

**Solo necesitas configurar DOCKER_HOST** (el alias de Docker es opcional):

```bash
# Añadir a ~/.zshrc o ~/.bashrc
export DOCKER_HOST=unix:///run/user/$(id -u)/podman/podman.sock
```

**Nota sobre el alias de Docker:**
- ❌ **NO es necesario** si configuras `DOCKER_HOST` correctamente
- ✅ testcontainers-go detecta Podman automáticamente a través de `DOCKER_HOST`
- El alias solo es útil si tienes scripts que usan el comando `docker` directamente
- Si quieres crear el alias (opcional):
  ```bash
  mkdir -p ~/.local/bin
  ln -s $(which podman) ~/.local/bin/docker
  export PATH="$HOME/.local/bin:$PATH"
  ```

### 5. Verificar Instalación

```bash
# Verificar Podman
podman --version

# Verificar socket
ls -la /run/user/$(id -u)/podman/podman.sock

# Probar con contenedor de prueba
podman run hello-world
```

## Ejecutar Tests E2E

Una vez configurado:

```bash
# Recargar shell o abrir nueva terminal
source ~/.zshrc  # o ~/.bashrc

# IMPORTANTE: Iniciar el servicio de Podman (necesario para testcontainers)
podman system service --time=0 > /dev/null 2>&1 &

# Configurar variables de entorno
export DOCKER_HOST=unix:///run/user/$(id -u)/podman/podman.sock
export TESTCONTAINERS_RYUK_DISABLED=true

# Ejecutar tests E2E
make test-e2e

# O directamente
go test -v -race -tags=container ./test/e2e/...
```

**Nota:** El servicio de Podman (`podman system service`) debe estar corriendo para que testcontainers-go pueda conectarse. Puedes ejecutarlo en segundo plano o crear un servicio systemd si lo prefieres.

## Troubleshooting

### Error: "Cannot connect to Podman socket"

```bash
# Verificar que el socket exista
ls -la /run/user/$(id -u)/podman/podman.sock

# Si no existe, iniciar servicio
systemctl --user start podman.socket

# Verificar estado
systemctl --user status podman.socket
```

### Error: "permission denied"

```bash
# Verificar subuid/subgid
cat /etc/subuid | grep $(id -u)
cat /etc/subgid | grep $(id -g)

# Si no aparecen, añadirlos manualmente
echo "$(id -u):100000:65536" | sudo tee -a /etc/subuid
echo "$(id -g):100000:65536" | sudo tee -a /etc/subgid
```

### Error: "command not found: docker"

```bash
# Verificar que el alias exista
ls -la ~/.local/bin/docker

# Verificar PATH
echo $PATH | grep -q ".local/bin" && echo "PATH OK" || echo "PATH no incluye .local/bin"

# Añadir al PATH si falta
export PATH="$HOME/.local/bin:$PATH"
# Y añadir a ~/.zshrc o ~/.bashrc para persistencia
```

### Podman funciona pero testcontainers no lo detecta

```bash
# Verificar DOCKER_HOST
echo $DOCKER_HOST

# Debe ser: unix:///run/user/<UID>/podman/podman.sock
# Si no está configurado:
export DOCKER_HOST=unix:///run/user/$(id -u)/podman/podman.sock

# IMPORTANTE: Iniciar el servicio de Podman (necesario para testcontainers)
podman system service --time=0 > /dev/null 2>&1 &

# También configurar:
export TESTCONTAINERS_RYUK_DISABLED=true
```

**Nota:** El socket de Podman (`podman.socket`) y el servicio de Podman (`podman system service`) son diferentes. El socket permite comunicación local, pero testcontainers necesita el servicio para la API compatible con Docker.

## Ventajas de Podman

- ✅ **Sin daemon**: No requiere proceso en segundo plano
- ✅ **Rootless**: Más seguro, no requiere privilegios
- ✅ **Ligero**: Menor consumo de recursos
- ✅ **Compatible**: testcontainers-go lo detecta automáticamente
- ✅ **Nativo en Arch**: Bien integrado en el sistema

## Referencias

- [Podman Documentation](https://docs.podman.io/)
- [testcontainers-go Podman Support](https://golang.testcontainers.org/features/container_drivers/)

