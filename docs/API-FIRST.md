# API-First Development Workflow

Este proyecto sigue el enfoque **API-First**, donde la especificación OpenAPI es la fuente de verdad y se genera código automáticamente desde ella.

## Filosofía API-First

1. **Diseño primero**: La API se diseña en [api/openapi/spec.yaml](../api/openapi/spec.yaml) antes de escribir código
2. **Contrato único**: La especificación OpenAPI es el contrato entre frontend y backend
3. **Generación automática**: El código se genera desde la especificación
4. **Consistencia garantizada**: No hay desviación entre documentación y código

## Herramientas Configuradas

### Servidor Go - oapi-codegen

**Herramienta**: [oapi-codegen](https://github.com/oapi-codegen/oapi-codegen)

**Instalación**:
```bash
go install github.com/oapi-codegen/oapi-codegen/v2/cmd/oapi-codegen@latest
```

**Configuración**: [api/oapi-codegen.yaml](../api/oapi-codegen.yaml)

**Características**:
- Genera interfaces de servidor para Gin
- Genera modelos de datos (structs Go)
- Valida parámetros automáticamente
- Soporta tipos UUID nativos
- Genera spec embebida para Swagger UI

**Comando**:
```bash
make generate-server
```

Esto genera: `internal/adapter/handler/http/generated/api.gen.go`

### Cliente Python - openapi-generator

**Herramienta**: [OpenAPI Generator](https://openapi-generator.tech/)

**Instalación** (elegir una opción):

1. **NPM** (recomendado):
```bash
npm install @openapitools/openapi-generator-cli -g
```

2. **pip**:
```bash
pip install openapi-generator-cli
```

3. **Docker** (sin instalación):
```bash
# Ver comando completo en make generate-client
```

**Configuración**: [api/openapi-generator-config.json](../api/openapi-generator-config.json)

**Características**:
- Genera cliente Python completo con urllib3
- Incluye modelos tipados (Pydantic/dataclasses)
- Manejo automático de autenticación
- Serialización/deserialización automática
- Tests de ejemplo

**Comando**:
```bash
make generate-client
```

Esto genera: `generated/python-client/`

## Workflow de Desarrollo

### 1. Diseñar API

Editar [api/openapi/spec.yaml](../api/openapi/spec.yaml):

```yaml
paths:
  /nuevo-endpoint:
    post:
      summary: Crear nuevo recurso
      operationId: crearRecurso
      requestBody:
        content:
          application/json:
            schema:
              $ref: '#/components/schemas/NuevoRecurso'
      responses:
        '201':
          description: Recurso creado
          content:
            application/json:
              schema:
                $ref: '#/components/schemas/Recurso'
```

### 2. Generar Código

```bash
# Generar código servidor Go
make generate-server

# Generar cliente Python (opcional)
make generate-client
```

### 3. Implementar Handlers

El código generado crea **interfaces** que debemos implementar:

```go
// internal/adapter/handler/http/generated/api.gen.go (generado)
type ServerInterface interface {
    CrearRecurso(c *gin.Context)
    // ... otros endpoints
}

// internal/adapter/handler/http/handlers.go (implementar manualmente)
type APIHandlers struct {
    db *pgxpool.Pool
}

func (h *APIHandlers) CrearRecurso(c *gin.Context) {
    // Implementación del endpoint
}
```

### 4. Registrar en Router

```go
// internal/adapter/handler/http/router.go
import "internal/adapter/handler/http/generated"

func SetupRouter(db *pgxpool.Pool, mode string) *gin.Engine {
    r := gin.New()

    // Crear handlers
    handlers := &APIHandlers{db: db}

    // Registrar con código generado
    generated.RegisterHandlers(r, handlers)

    return r
}
```

### 5. Probar con Cliente Python

```python
from automatizacion_client import ApiClient, Configuration
from automatizacion_client.api import AutomatizacionesApi

# Configurar cliente
config = Configuration(host="http://localhost:8080")
client = ApiClient(config)
api = AutomatizacionesApi(client)

# Usar API
task = api.crear_automatizacion(
    create_task_request={
        "name": "Mi Tarea",
        "created_by": "Equipo Test"
    }
)
print(f"Tarea creada: {task.id}")
```

## Ventajas del API-First

### ✅ Desarrollo Paralelo
- Frontend y Backend pueden trabajar simultáneamente
- El contrato API está definido desde el día 1
- Mocks automáticos desde la especificación

### ✅ Documentación Siempre Actualizada
- Swagger UI generado automáticamente
- Documentación en sync con el código
- Ejemplos de uso incluidos

### ✅ Validación Automática
- Tipos verificados en tiempo de compilación (Go)
- Validación de parámetros automática
- Mensajes de error consistentes (RFC 7807)

### ✅ Testing Facilitado
- Clientes de prueba generados automáticamente
- Schemas para validación de contratos
- Ejemplos de requests/responses

### ✅ Evolución Controlada
- Cambios en spec → regenerar código
- Breaking changes detectados inmediatamente
- Versionado de API explícito

## Comandos Útiles

```bash
# Ver ayuda completa
make help

# Generar solo servidor
make generate-server

# Ver comando para generar cliente
make generate-client

# Generar ambos
make generate-all

# Validar especificación OpenAPI
npx @apidevtools/swagger-cli validate api/openapi/spec.yaml

# Ver Swagger UI local
# (Agregar Swagger UI handler al router)
```

## Estructura de Archivos

```
.
├── api/
│   ├── openapi/
│   │   └── spec.yaml                     # ✏️ EDITAR: Especificación OpenAPI
│   ├── oapi-codegen.yaml                 # Configuración generador Go
│   └── openapi-generator-config.json     # Configuración generador Python
├── internal/
│   └── adapter/
│       └── handler/
│           └── http/
│               ├── generated/            # 🤖 GENERADO: No editar manualmente
│               │   └── api.gen.go
│               ├── handlers.go           # ✏️ IMPLEMENTAR: Lógica de negocio
│               └── router.go             # ✏️ CONFIGURAR: Routing
└── generated/
    └── python-client/                    # 🤖 GENERADO: Cliente Python
        ├── automatizacion_client/
        ├── docs/
        └── README.md
```

## Reglas de Oro

1. **NUNCA editar archivos generados manualmente**
   - Los cambios se perderán en la próxima generación

2. **La especificación OpenAPI es la fuente de verdad**
   - Cualquier cambio en API debe reflejarse primero en spec.yaml

3. **Regenerar después de cada cambio en spec**
   - `make generate-all` después de modificar spec.yaml

4. **Versionar la especificación, no el código generado**
   - spec.yaml va en Git
   - `internal/adapter/handler/http/generated/` está en .gitignore
   - `generated/` está en .gitignore

5. **Validar la spec antes de generar**
   - Usar herramientas de validación OpenAPI
   - Revisar que ejemplos sean correctos

## Troubleshooting

### Error: "oapi-codegen: command not found"

```bash
# Instalar oapi-codegen
go install github.com/oapi-codegen/oapi-codegen/v2/cmd/oapi-codegen@latest

# Verificar que $GOPATH/bin esté en PATH
export PATH=$PATH:$(go env GOPATH)/bin
```

### Error: "openapi-generator-cli: command not found"

```bash
# Opción 1: NPM
npm install @openapitools/openapi-generator-cli -g

# Opción 2: Docker (sin instalación)
make generate-client  # Ver comando Docker en output
```

### Los handlers generados no compilan

1. Verificar que spec.yaml sea válido:
```bash
npx @apidevtools/swagger-cli validate api/openapi/spec.yaml
```

2. Regenerar con última versión de oapi-codegen:
```bash
go install github.com/oapi-codegen/oapi-codegen/v2/cmd/oapi-codegen@latest
make generate-server
```

## Referencias

- [OpenAPI Specification 3.0](https://swagger.io/specification/)
- [oapi-codegen Documentation](https://github.com/oapi-codegen/oapi-codegen)
- [OpenAPI Generator Docs](https://openapi-generator.tech/docs/generators/python)
- [RFC 7807 - Problem Details](https://www.rfc-editor.org/rfc/rfc7807)
