# Pre-commit Hooks

Scripts para validar código antes de realizar commits, siguiendo las buenas prácticas de la [Guía de Desarrollo](../Guia_desarrollo.md).

## ¿Qué hacen los scripts?

Los scripts de pre-commit ejecutan automáticamente:

1. **Formateo de código** con `gofumpt`
2. **Ordenamiento de imports** con `goimports`
3. **Linter** con `golangci-lint`
4. **Tests unitarios**

Si alguna validación falla, el commit se bloquea hasta que se corrijan los errores.

## Instalación

### Linux/macOS

```bash
# Hacer el script ejecutable
chmod +x scripts/pre-commit.sh

# Instalar como hook de Git
cp scripts/pre-commit.sh .git/hooks/pre-commit
chmod +x .git/hooks/pre-commit
```

### Windows (PowerShell)

```powershell
# Instalar como hook de Git
Copy-Item scripts/pre-commit.ps1 .git/hooks/pre-commit.ps1
```

**Nota**: En Windows, Git puede requerir configuración adicional para ejecutar scripts PowerShell. Si tienes problemas, ejecuta el script manualmente antes de commitear.

## Ejecución Manual

Si prefieres ejecutar las validaciones manualmente antes de commitear:

### Linux/macOS

```bash
./scripts/pre-commit.sh
```

### Windows (PowerShell)

```powershell
.\scripts\pre-commit.ps1
```

## Requisitos

Los scripts instalarán automáticamente las herramientas necesarias si no están disponibles:

- `gofumpt`: Formateador de código Go
- `goimports`: Gestor de imports
- `golangci-lint`: Linter (opcional, se advierte si no está instalado)

### Instalación Manual de golangci-lint

Si prefieres instalar `golangci-lint` manualmente:

```bash
# Linux/macOS
curl -sSfL https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh | sh -s -- -b $(go env GOPATH)/bin v1.62.2

# Windows (PowerShell)
# Descargar desde: https://github.com/golangci/golangci-lint/releases
# O usar: go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest
```

## Desactivar Temporalmente

Si necesitas hacer un commit sin ejecutar los hooks (no recomendado):

```bash
git commit --no-verify -m "Mensaje del commit"
```

**⚠️ Advertencia**: Solo usa `--no-verify` en casos excepcionales. Los hooks están ahí para mantener la calidad del código.

## Solución de Problemas

### Error: "gofumpt no está instalado"

El script intentará instalarlo automáticamente. Si falla:

```bash
go install mvdan.cc/gofumpt@latest
```

### Error: "golangci-lint no está instalado"

El script continuará sin el linter, pero es recomendable instalarlo:

```bash
# Ver instrucciones en: https://golangci-lint.run/usage/install/
go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest
```

### Error: "Tests fallaron"

Corrige los tests antes de commitear. Los tests unitarios deben pasar siempre.

### El hook no se ejecuta automáticamente

Verifica que el archivo tenga permisos de ejecución:

```bash
# Linux/macOS
ls -l .git/hooks/pre-commit
chmod +x .git/hooks/pre-commit
```

## Integración con Make

También puedes usar los comandos de Make para validaciones individuales:

```bash
# Formatear código
make fmt

# Verificar formato
make fmt-check

# Ejecutar linter
make lint

# Ejecutar tests unitarios
make test-unit
```

