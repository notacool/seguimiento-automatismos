# Pre-commit hook para ejecutar formateador, linter y tests unitarios
# Uso: Ejecutar este script antes de commitear en PowerShell

$ErrorActionPreference = "Stop"

Write-Host "🔍 Ejecutando validaciones pre-commit..." -ForegroundColor Cyan

# Función para mostrar errores
function Show-Error {
    param([string]$Message)
    Write-Host "❌ $Message" -ForegroundColor Red
    exit 1
}

# Función para mostrar éxito
function Show-Success {
    param([string]$Message)
    Write-Host "✅ $Message" -ForegroundColor Green
}

# Función para mostrar advertencia
function Show-Warning {
    param([string]$Message)
    Write-Host "⚠️  $Message" -ForegroundColor Yellow
}

# 1. Verificar que go está disponible
if (-not (Get-Command go -ErrorAction SilentlyContinue)) {
    Show-Error "Go no está instalado o no está en el PATH"
}

# 2. Verificar que gofumpt está instalado
Write-Host ""
Write-Host "📝 Verificando formateador (gofumpt)..." -ForegroundColor Cyan
$gofumptPath = Get-Command gofumpt -ErrorAction SilentlyContinue
if (-not $gofumptPath) {
    Show-Warning "gofumpt no está instalado. Instalando..."
    go install mvdan.cc/gofumpt@latest
    if ($LASTEXITCODE -ne 0) {
        Show-Error "Error al instalar gofumpt"
    }
}

# Formatear código
Write-Host "Formateando código con gofumpt..." -ForegroundColor Cyan
$unformatted = gofumpt -l . 2>&1
if ($unformatted) {
    Show-Warning "Archivos sin formatear detectados. Formateando automáticamente..."
    gofumpt -w .
    if ($LASTEXITCODE -ne 0) {
        Show-Error "Error al formatear código"
    }
    Show-Error "Archivos formateados. Por favor, revisa los cambios y vuelve a intentar el commit."
} else {
    Show-Success "Código correctamente formateado"
}

# 3. Verificar imports
Write-Host ""
Write-Host "📦 Verificando imports con goimports..." -ForegroundColor Cyan
$goimportsPath = Get-Command goimports -ErrorAction SilentlyContinue
if (-not $goimportsPath) {
    Show-Warning "goimports no está instalado. Instalando..."
    go install golang.org/x/tools/cmd/goimports@latest
    if ($LASTEXITCODE -ne 0) {
        Show-Error "Error al instalar goimports"
    }
}

# Verificar si hay imports sin ordenar
$unsorted = goimports -l . 2>&1
if ($unsorted) {
    Show-Warning "Imports sin ordenar detectados. Ordenando automáticamente..."
    goimports -w -local github.com/grupoapi/proces-log .
    if ($LASTEXITCODE -ne 0) {
        Show-Error "Error al ordenar imports"
    }
    Show-Error "Imports ordenados. Por favor, revisa los cambios y vuelve a intentar el commit."
} else {
    Show-Success "Imports correctamente ordenados"
}

# 4. Ejecutar linter
Write-Host ""
Write-Host "🔍 Ejecutando linter (golangci-lint)..." -ForegroundColor Cyan
$golangciLintPath = Get-Command golangci-lint -ErrorAction SilentlyContinue
if (-not $golangciLintPath) {
    Show-Warning "golangci-lint no está instalado."
    Show-Warning "Instala desde: https://golangci-lint.run/usage/install/"
    Show-Warning "O ejecuta: go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest"
    Show-Warning "Continuando sin linter..."
} else {
    golangci-lint run --timeout=5m
    if ($LASTEXITCODE -ne 0) {
        Show-Error "Linter encontró errores. Por favor, corrígelos antes de commitear."
    }
    Show-Success "Linter: sin errores"
}

# 5. Ejecutar tests unitarios
Write-Host ""
Write-Host "🧪 Ejecutando tests unitarios..." -ForegroundColor Cyan
go test -v -race -short ./internal/... ./test/helpers/...
if ($LASTEXITCODE -ne 0) {
    Show-Error "Tests unitarios fallaron. Por favor, corrígelos antes de commitear."
}
Show-Success "Tests unitarios: todos pasaron"

Write-Host ""
Show-Success "✨ Todas las validaciones pasaron. Listo para commitear."

