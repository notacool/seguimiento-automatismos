# Instalación de Podman - Guía Rápida

## ⚠️ Importante

**NO ejecutes el script con `sudo`**. El script pedirá tu contraseña cuando sea necesario.

**Si el proyecto está en un dispositivo externo (USB/externo con FAT32):**
- Los permisos de ejecución no funcionan en sistemas de archivos FAT32
- **SIEMPRE usa `bash` explícitamente** para ejecutar el script

## Desde el directorio raíz del proyecto:

```bash
# ✅ CORRECTO (siempre funciona)
bash scripts/setup-podman.sh

# ❌ NO funciona en FAT32 (permisos no se preservan)
./scripts/setup-podman.sh
chmod +x scripts/setup-podman.sh  # No tiene efecto en FAT32
```

## Si estás en el directorio scripts/:

```bash
# Volver al directorio raíz
cd ..

# Ejecutar el script
bash scripts/setup-podman.sh
```

## Verificar estado antes de instalar:

```bash
# Ver qué falta configurar
bash scripts/check-podman.sh
```

## Después de instalar:

```bash
# Recargar shell
source ~/.zshrc

# Verificar instalación
bash scripts/check-podman.sh

# Ejecutar tests E2E
make test-e2e
```

