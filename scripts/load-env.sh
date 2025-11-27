#!/bin/bash
# Script para cargar variables de entorno desde .env
# Uso: source scripts/load-env.sh

if [ -f .env ]; then
    export $(grep -v '^#' .env | xargs)
    echo "✓ Variables de entorno cargadas desde .env"
else
    echo "ℹ️  Archivo .env no encontrado. Crea uno desde .env.example:"
    echo "   cp .env.example .env"
fi

