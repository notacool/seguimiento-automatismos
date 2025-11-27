# Configuración con Archivo .env

El proyecto soporta configuración mediante un archivo `.env` para personalizar el comportamiento sin modificar el código.

## Crear Archivo .env

```bash
# Copiar desde el ejemplo
cp .env.example .env

# Editar según tus necesidades
nano .env  # o tu editor preferido
```

## Variables Disponibles

### CONTAINER_RUNTIME

Configura qué runtime de contenedores usar para los comandos del Makefile.

**Valores posibles:**
- `docker`: Usa Docker explícitamente
- `podman`: Usa Podman explícitamente
- `auto`: Detecta automáticamente (default)
  - Si Docker está instalado, usa Docker
  - Si solo Podman está instalado, usa Podman
  - Si ambos están instalados, Docker tiene prioridad

**Ejemplo:**
```bash
CONTAINER_RUNTIME=podman
```

**Verificar detección:**
```bash
make detect-container-runtime
```

### DATABASE_URL

URL de conexión a la base de datos PostgreSQL.

**Formato:**
```
postgres://usuario:contraseña@host:puerto/nombre_bd?sslmode=disable
```

**Ejemplo:**
```bash
DATABASE_URL=postgres://user:password@localhost:5432/dbname?sslmode=disable
```

### Variables de API

```bash
API_PORT=8080
API_HOST=localhost
```

### Variables de Tests

```bash
TEST_DB_HOST=localhost
TEST_DB_PORT=5432
TEST_DB_USER=test_user
TEST_DB_PASSWORD=test_password
TEST_DB_NAME=test_db
```

## Uso en Makefile

El Makefile carga automáticamente el archivo `.env` si existe. Las variables se pueden sobrescribir desde la línea de comandos:

```bash
# Usar Podman explícitamente (ignora .env)
make docker-up CONTAINER_RUNTIME=podman

# Usar Docker explícitamente
make docker-build CONTAINER_RUNTIME=docker
```

## Ejemplos de Configuración

### Configuración con Podman

```bash
# .env
CONTAINER_RUNTIME=podman
DATABASE_URL=postgres://user:pass@localhost:5432/mydb?sslmode=disable
```

**Importante:** Si usas Podman, asegúrate de tener configurado `DOCKER_HOST`:

```bash
export DOCKER_HOST=unix:///run/user/$(id -u)/podman/podman.sock
# O añadir a ~/.zshrc para persistencia
```

### Configuración con Docker

```bash
# .env
CONTAINER_RUNTIME=docker
DATABASE_URL=postgres://user:pass@localhost:5432/mydb?sslmode=disable
```

### Detección Automática

```bash
# .env
CONTAINER_RUNTIME=auto
# El Makefile detectará automáticamente qué runtime usar
```

## Verificar Configuración

```bash
# Ver qué runtime se usará
make detect-container-runtime

# Ver todas las variables cargadas
make help
```

## Notas

- El archivo `.env` está en `.gitignore` y no se versiona
- Usa `env.example` como plantilla
- Las variables de entorno del sistema tienen prioridad sobre `.env`
- Puedes sobrescribir cualquier variable desde la línea de comandos

