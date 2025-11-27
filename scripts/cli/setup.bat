@echo off
REM Setup script for CLI development environment (Windows)

echo Configurando CLI de Automatizaciones...

REM Check Python installation
python --version >nul 2>&1
if %errorlevel% neq 0 (
    echo Error: Python no esta instalado
    exit /b 1
)

echo Python encontrado

REM Create virtual environment
echo Creando entorno virtual...
if not exist "venv" (
    python -m venv venv
    echo Entorno virtual creado
) else (
    echo Entorno virtual ya existe
)

REM Activate virtual environment
echo Activando entorno virtual...
call venv\Scripts\activate.bat

REM Upgrade pip
echo Actualizando pip...
python -m pip install --upgrade pip

REM Install dependencies
echo Instalando dependencias...
pip install -r requirements.txt

echo.
echo CLI configurado correctamente!
echo.
echo Para usar el CLI:
echo   1. Activar entorno virtual:
echo      venv\Scripts\activate.bat
echo.
echo   2. Ejecutar CLI:
echo      python main.py --help
echo      python main.py health
echo      python main.py task list
echo.
echo Para desactivar el entorno virtual:
echo   deactivate
