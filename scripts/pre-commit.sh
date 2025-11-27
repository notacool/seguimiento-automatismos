#!/bin/bash
# Pre-commit hook para ejecutar formateador, linter y tests unitarios
# Uso: Copiar este archivo a .git/hooks/pre-commit o ejecutarlo manualmente antes de commitear

set -e

echo "🔍 Ejecutando validaciones pre-commit..."

# Colores para output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# Función para mostrar errores
error() {
    echo -e "${RED}❌ $1${NC}" >&2
    exit 1
}

# Función para mostrar éxito
success() {
    echo -e "${GREEN}✅ $1${NC}"
}

# Función para mostrar advertencia
warning() {
    echo -e "${YELLOW}⚠️  $1${NC}"
}

# 1. Verificar que gofumpt está instalado
if ! command -v gofumpt &> /dev/null; then
    warning "gofumpt no está instalado. Instalando..."
    go install mvdan.cc/gofumpt@latest
fi

# 2. Formatear código
echo ""
echo "📝 Formateando código con gofumpt..."
if gofumpt -l . | grep -q .; then
    warning "Archivos sin formatear detectados. Formateando automáticamente..."
    gofumpt -w .
    error "Archivos formateados. Por favor, revisa los cambios y vuelve a intentar el commit."
else
    success "Código correctamente formateado"
fi

# 3. Verificar imports
echo ""
echo "📦 Verificando imports con goimports..."
if ! command -v goimports &> /dev/null; then
    warning "goimports no está instalado. Instalando..."
    go install golang.org/x/tools/cmd/goimports@latest
fi

# Verificar si hay imports sin ordenar
if goimports -l . | grep -q .; then
    warning "Imports sin ordenar detectados. Ordenando automáticamente..."
    goimports -w -local github.com/grupoapi/proces-log .
    error "Imports ordenados. Por favor, revisa los cambios y vuelve a intentar el commit."
else
    success "Imports correctamente ordenados"
fi

# 4. Ejecutar linter
echo ""
echo "🔍 Ejecutando linter (golangci-lint)..."
if ! command -v golangci-lint &> /dev/null; then
    warning "golangci-lint no está instalado."
    warning "Instala desde: https://golangci-lint.run/usage/install/"
    warning "O ejecuta: go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest"
    warning "Continuando sin linter..."
else
    if golangci-lint run --timeout=5m; then
        success "Linter: sin errores"
    else
        error "Linter encontró errores. Por favor, corrígelos antes de commitear."
    fi
fi

# 5. Ejecutar tests unitarios
echo ""
echo "🧪 Ejecutando tests unitarios..."
if go test -v -race -short ./internal/... ./test/helpers/...; then
    success "Tests unitarios: todos pasaron"
else
    error "Tests unitarios fallaron. Por favor, corrígelos antes de commitear."
fi

echo ""
success "✨ Todas las validaciones pasaron. Listo para commitear."

