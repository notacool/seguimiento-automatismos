#!/bin/bash
# Setup script for CLI development environment

set -e

echo "🔧 Configurando CLI de Automatizaciones..."

# Check Python installation
if ! command -v python3 &> /dev/null; then
    echo "❌ Error: Python3 no está instalado"
    exit 1
fi

echo "✓ Python3 encontrado: $(python3 --version)"

# Create virtual environment
echo "📦 Creando entorno virtual..."
if [ ! -d "venv" ]; then
    python3 -m venv venv
    echo "✓ Entorno virtual creado"
else
    echo "✓ Entorno virtual ya existe"
fi

# Activate virtual environment
echo "🔌 Activando entorno virtual..."
source venv/bin/activate

# Upgrade pip
echo "⬆️  Actualizando pip..."
pip install --upgrade pip

# Install dependencies
echo "📥 Instalando dependencias..."
pip install -r requirements.txt

echo ""
echo "✅ ¡CLI configurado correctamente!"
echo ""
echo "Para usar el CLI:"
echo "  1. Activar entorno virtual:"
echo "     source venv/bin/activate"
echo ""
echo "  2. Ejecutar CLI:"
echo "     python main.py --help"
echo "     python main.py health"
echo "     python main.py task list"
echo ""
echo "Para desactivar el entorno virtual:"
echo "  deactivate"
