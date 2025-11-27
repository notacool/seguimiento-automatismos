# Guía de Desarrollo

## Herramientas de Desarrollo

### Linting y Formato

Este proyecto utiliza las siguientes herramientas para mantener la calidad del código:

#### golangci-lint (v1.64.8)
Linter integral que ejecuta múltiples linters en paralelo.

**Instalación:**
```bash
# Usando el instalador oficial
curl -sSfL https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh | sh -b $(go env GOPATH)/bin v1.62.2

# O con go install (puede ser más lento)
go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest
```

**Uso:**
```bash
# Ejecutar linter
make lint

# O directamente
golangci-lint run ./...
```

**Configuración:** Ver `.golangci.yml`

**Linters habilitados:**
- `errcheck` - Verifica errores no manejados
- `gosimple` - Simplifica código
- `govet` - Análisis estático de Go
- `ineffassign` - Detecta asignaciones inefectivas
- `staticcheck` - Análisis estático avanzado
- `unused` - Detecta código no utilizado
- `gofmt` - Verifica formato de código
- `gofumpt` - Formato más estricto
- `goimports` - Verifica imports
- `revive` - Linter configurable (reemplazo de golint)
- `stylecheck` - Estilo de código
- `gosec` - Problemas de seguridad
- `misspell` - Errores ortográficos
- `unconvert` - Conversiones innecesarias
- `unparam` - Parámetros no usados
- `dogsled` - Blank identifiers excesivos
- `gocyclo` - Complejidad ciclomática
- `gocognit` - Complejidad cognitiva
- `bodyclose` - Verifica cierre de response bodies
- `gocritic` - Diagnósticos de bugs, performance y estilo
- `errname` - Nombres de errores con sufijo Err
- `errorlint` - Problemas con error wrapping
- `nolintlint` - Directivas nolint mal formadas
- `prealloc` - Slices que podrían pre-alocarse
- `predeclared` - Shadowing de identificadores predeclarados
- `whitespace` - Espacios en blanco

**Linters deshabilitados:**
- `exportloopref` - Deprecado desde Go 1.22
- `godot` - Demasiado estricto para comentarios en español

#### gofumpt (v0.9.2)
Formateador de código Go más estricto que `gofmt`.

**Instalación:**
```bash
go install mvdan.cc/gofumpt@latest
```

**Uso:**
```bash
# Formatear código
make fmt

# Solo verificar (sin modificar)
make fmt-check

# O directamente
gofumpt -l -w .
```

#### goimports
Gestiona imports automáticamente (agrupa, ordena, elimina no usados).

**Instalación:**
```bash
go install golang.org/x/tools/cmd/goimports@latest
```

**Uso:**
```bash
# Se ejecuta automáticamente con make fmt
make fmt

# O directamente
goimports -w -local github.com/grupoapi/proces-log .
```

## Flujo de Trabajo Recomendado

### Antes de hacer commit

```bash
# 1. Formatear código
make fmt

# 2. Ejecutar linter
make lint

# 3. Ejecutar tests
make test

# 4. Verificar cobertura
make test-coverage
```

### Integración Continua

GitHub Actions ejecuta automáticamente:
- Verificación de formato (`gofumpt`)
- Linting con `golangci-lint`
- Tests con coverage
- Build de la aplicación
- Build de imagen Docker

Ver `.github/workflows/ci.yml` para detalles.

## Configuración del Editor

### VS Code

Instalar la extensión oficial de Go y configurar:

```json
{
  "go.lintTool": "golangci-lint",
  "go.lintFlags": ["--fast"],
  "go.formatTool": "gofumpt",
  "editor.formatOnSave": true,
  "go.useLanguageServer": true,
  "[go]": {
    "editor.codeActionsOnSave": {
      "source.organizeImports": true
    }
  }
}
```

### GoLand / IntelliJ IDEA

1. File → Settings → Tools → File Watchers
2. Agregar watcher para gofumpt
3. Tools → golangci-lint → Enable golangci-lint

## Solución de Problemas

### Error: "linter X can't be disabled and enabled at one moment"

Verifica que no tengas el mismo linter en `enable` y `disable` en `.golangci.yml`.

### Error: "version mismatch"

Asegúrate de que la versión en `.golangci.yml` coincida con la instalada:
```bash
golangci-lint --version
```

### Imports mal formateados

Ejecuta:
```bash
goimports -w -local github.com/grupoapi/proces-log .
```

## Estándares de Código

### Comentarios

- Todos los exports deben tener comentario
- Los comentarios pueden estar en español
- Usar `//nolint:lintername // razón` para excepciones justificadas

### Manejo de Errores

- Siempre manejar errores explícitamente
- Usar error wrapping con `fmt.Errorf("context: %w", err)`
- No ignorar errores sin justificación documentada

### Complejidad

- Complejidad ciclomática máxima: 15
- Complejidad cognitiva máxima: 20
- Refactorizar funciones que excedan estos límites

## Referencias

- [golangci-lint](https://golangci-lint.run/)
- [gofumpt](https://github.com/mvdan/gofumpt)
- [Effective Go](https://go.dev/doc/effective_go)
- [Go Code Review Comments](https://github.com/golang/go/wiki/CodeReviewComments)
