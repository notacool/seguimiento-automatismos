# Tests E2E sin Contenedores

## Alternativa: PostgreSQL Local

Si prefieres no usar contenedores (Docker/Podman), puedes ejecutar los tests E2E usando PostgreSQL instalado localmente.

## Requisitos

1. **PostgreSQL instalado y corriendo**
   ```bash
   # En Arch Linux
   sudo pacman -S postgresql
   sudo systemctl enable postgresql
   sudo systemctl start postgresql
   ```

2. **Crear usuario de prueba (opcional)**
   ```bash
   sudo -u postgres createuser test_user
   sudo -u postgres psql -c "ALTER USER test_user WITH PASSWORD 'test_password';"
   ```

## Configuración

### Opción 1: Variables de Entorno (Recomendado)

```bash
export TEST_DB_HOST=localhost
export TEST_DB_PORT=5432
export TEST_DB_USER=postgres
export TEST_DB_PASSWORD=postgres
export TEST_DB_MAIN=postgres
```

### Opción 2: Valores por Defecto

Si no configuras variables de entorno, se usarán estos valores:
- Host: `localhost`
- Port: `5432`
- User: `postgres`
- Password: `postgres`
- Main DB: `postgres`

## Ejecutar Tests

```bash
# Ejecutar tests E2E con PostgreSQL local
make test-e2e-local

# O directamente
go test -v -tags=!container ./test/e2e/...

# Ejecutar un test específico
go test -v -tags=!container ./test/e2e/... -run TestE2E_TaskCompleteLifecycle_Local
```

## Cómo Funciona

1. **Crea una base de datos temporal** con nombre único (`test_<timestamp>_<pid>`)
2. **Ejecuta los tests** contra esa base de datos
3. **Limpia automáticamente** la base de datos al finalizar (incluso si fallan)

## Ventajas

- ✅ No requiere Docker/Podman
- ✅ Más rápido (sin overhead de contenedores)
- ✅ Funciona en cualquier sistema con PostgreSQL
- ✅ Limpieza automática de bases de datos temporales

## Desventajas

- ⚠️ Requiere PostgreSQL instalado localmente
- ⚠️ Puede dejar bases de datos si el proceso se interrumpe bruscamente
- ⚠️ Menos aislado que contenedores (comparte el mismo PostgreSQL)

## Troubleshooting

### Error: "No se pudo conectar a PostgreSQL"

**Verificar que PostgreSQL esté corriendo:**
```bash
sudo systemctl status postgresql
sudo systemctl start postgresql
```

**Verificar conexión:**
```bash
psql -h localhost -U postgres -d postgres
```

### Error: "permission denied"

**Asegúrate de tener permisos para crear bases de datos:**
```bash
# Como superusuario postgres
sudo -u postgres psql -c "ALTER USER postgres WITH SUPERUSER;"
```

### Limpiar bases de datos temporales manualmente

Si los tests se interrumpen, puedes limpiar manualmente:
```sql
-- Conectarse a PostgreSQL
psql -U postgres

-- Listar bases de datos temporales
SELECT datname FROM pg_database WHERE datname LIKE 'test_%';

-- Eliminar bases de datos temporales antiguas
DROP DATABASE IF EXISTS test_<nombre>;
```

## Comparación de Opciones

| Característica | Contenedores | PostgreSQL Local |
|----------------|--------------|------------------|
| Requisitos | Docker/Podman | PostgreSQL instalado |
| Aislamiento | Alto | Medio |
| Velocidad | Media | Alta |
| Limpieza | Automática | Automática (con cleanup) |
| Portabilidad | Alta | Media |
| Configuración | Media | Baja |

## Recomendación

- **Desarrollo local:** Usa PostgreSQL local (`test-e2e-local`) - más rápido y simple
- **CI/CD:** Usa contenedores (`test-e2e`) - más aislado y reproducible
- **Equipos sin Docker:** Usa PostgreSQL local

