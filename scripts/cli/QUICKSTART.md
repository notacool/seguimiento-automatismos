# Quick Start - CLI de Automatizaciones

Guía rápida para empezar a usar el CLI en menos de 2 minutos.

## 1. Instalación Rápida

```bash
# Desde la raíz del proyecto
cd scripts/cli
./setup.sh
```

Esto creará un entorno virtual e instalará todas las dependencias automáticamente.

## 2. Activar el Entorno Virtual

```bash
source venv/bin/activate
```

Verás `(venv)` al inicio de tu terminal indicando que el entorno está activo.

## 3. Verificar Conectividad

Asegúrate de que la API esté corriendo:

```bash
# Desde la raíz del proyecto (en otra terminal)
make docker-up
```

Luego verifica que el CLI puede conectarse:

```bash
python main.py health
```

Deberías ver:

```
┏━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━┓
┃        Health Check             ┃
┡━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━┩
│ Service Status: HEALTHY         │
│ Database: OK                    │
│ Timestamp: 2025-11-27T...       │
└─────────────────────────────────┘
```

## 4. Comandos Básicos

### Listar todas las tareas

```bash
python main.py task list
```

### Filtrar tareas por estado

```bash
python main.py task list --state IN_PROGRESS
```

### Buscar tareas por nombre

```bash
python main.py task list --name "Backup"
```

### Ver detalles de una tarea específica

```bash
# Obtener el ID de una tarea del listado anterior
python main.py task get <task-id>
```

### Ver subtareas de una tarea

```bash
python main.py task subtasks <task-id>
```

## 5. Opciones Útiles

### Cambiar URL de la API

```bash
python main.py --url http://production:8080 health
python main.py --url http://production:8080 task list
```

### Salida en JSON (útil para scripts)

```bash
python main.py task list --json
python main.py task get <task-id> --json
```

### Paginación

```bash
# Ver página 2 con 50 resultados por página
python main.py task list --page 2 --limit 50
```

## 6. Uso Avanzado con Make

Desde la raíz del proyecto (sin necesidad de activar venv):

```bash
# Verificar salud
make cli-run CMD="health"

# Listar tareas
make cli-run CMD="task list"

# Filtrar por estado
make cli-run CMD="task list --state COMPLETED"
```

## 7. Generar Ejecutable (Opcional)

Para distribuir el CLI sin necesidad de Python:

```bash
make cli-build-linux
# El ejecutable estará en scripts/cli/dist/automatizacion-cli-linux

# Usar el ejecutable
./dist/automatizacion-cli-linux health
./dist/automatizacion-cli-linux task list
```

## Ayuda

Para ver ayuda de cualquier comando:

```bash
python main.py --help
python main.py task --help
python main.py task list --help
```

## Desactivar el Entorno Virtual

Cuando termines:

```bash
deactivate
```

## Resumen de Estados

- **PENDING**: Pendiente de iniciar (amarillo)
- **IN_PROGRESS**: En ejecución (azul)
- **COMPLETED**: Completada exitosamente (verde)
- **FAILED**: Fallida (rojo)
- **CANCELLED**: Cancelada (magenta)

## Troubleshooting

### Error: "Module not found"

Asegúrate de haber activado el venv:

```bash
source venv/bin/activate
```

### Error: "Connection refused"

La API no está corriendo. Iníciala con:

```bash
make docker-up
```

### Error: "Task not found"

El UUID de la tarea no existe o fue eliminada (soft delete).

---

Para más detalles, consulta el [README.md](README.md) completo.
