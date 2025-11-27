# Troubleshooting Podman para Tests E2E

## Error: "newuidmap: write to uid_map failed: Invalid argument"

Este error ocurre cuando Podman no puede configurar el namespace de usuario después de configurar `subuid`/`subgid`.

### Causas Comunes

1. **Paquete `shadow` no instalado** - Proporciona las herramientas `newuidmap` y `newgidmap` necesarias
2. **Entradas duplicadas** en `/etc/subuid` y `/etc/subgid` (por nombre de usuario y por UID numérico)
3. **Permisos incorrectos** en `/usr/bin/newuidmap` y `/usr/bin/newgidmap` (falta el bit setuid)
4. **Kernel no ha recargado los mapeos** después de configurar `subuid`/`subgid`

### Verificar si shadow está instalado

```bash
# Verificar que newuidmap y newgidmap existen
which newuidmap newgidmap

# Si no existen, instalar shadow
sudo pacman -S shadow
```

### Solución Rápida (Automática)

Ejecuta el script de corrección:

```bash
bash scripts/fix-podman-rootless.sh
```

Este script:
- ✅ Elimina entradas duplicadas en `/etc/subuid` y `/etc/subgid`
- ✅ Corrige permisos de `newuidmap` y `newgidmap` (activa setuid)
- ✅ Verifica la configuración final
- ✅ Inicia el socket de Podman si no está activo

**Después de ejecutar el script:**

1. **Cierra sesión completamente** (no solo cierres la terminal)
2. Vuelve a entrar
3. Verifica que Podman funcione:
   ```bash
   podman run --rm hello-world
   ```
4. Ejecuta los tests E2E:
   ```bash
   make test-e2e
   ```

### Solución Manual

Si prefieres hacerlo manualmente:

1. **Eliminar entradas duplicadas:**
   ```bash
   # Eliminar entradas por nombre de usuario (dejar solo las numéricas)
   sudo sed -i '/^tu_usuario:/d' /etc/subuid
   sudo sed -i '/^tu_usuario:/d' /etc/subgid
   ```

2. **Corregir permisos de binarios:**
   ```bash
   sudo chmod u+s /usr/bin/newuidmap /usr/bin/newgidmap
   ```

3. **Cerrar sesión completamente** y volver a entrar

### Verificación

Después de reiniciar, verifica:

```bash
# Verificar que Podman funciona
podman run hello-world

# Verificar socket
ls -la /run/user/$(id -u)/podman/podman.sock

# Verificar DOCKER_HOST
echo $DOCKER_HOST
# Debe mostrar: unix:///run/user/1000/podman/podman.sock

# Ejecutar tests
make test-e2e
```

### Soluciones Alternativas (Temporales)

Si no puedes reiniciar la sesión ahora:

1. **Usar tests E2E locales** (sin contenedores):
   ```bash
   # Requiere PostgreSQL instalado localmente
   sudo pacman -S postgresql
   sudo systemctl start postgresql
   make test-e2e-local
   ```

2. **Usar Docker en lugar de Podman** (si está instalado):
   ```bash
   # En .env o desde línea de comandos
   make test-e2e CONTAINER_RUNTIME=docker
   ```

### Configuración Adicional para testcontainers-go

Si después de reiniciar sigues teniendo problemas, añade estas variables de entorno:

```bash
# Deshabilitar Ryuk (puede causar problemas con Podman rootless)
export TESTCONTAINERS_RYUK_DISABLED=true

# Asegurar DOCKER_HOST
export DOCKER_HOST=unix:///run/user/$(id -u)/podman/podman.sock

# Añadir a ~/.zshrc para persistencia
echo 'export TESTCONTAINERS_RYUK_DISABLED=true' >> ~/.zshrc
echo 'export DOCKER_HOST=unix:///run/user/$(id -u)/podman/podman.sock' >> ~/.zshrc
source ~/.zshrc
```

### Referencias

- [Podman Rootless Setup](https://github.com/containers/podman/blob/main/docs/tutorials/rootless_tutorial.md)
- [testcontainers-go Podman Support](https://golang.testcontainers.org/system_requirements/using_podman/)

