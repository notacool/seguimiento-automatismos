# Changelog

Todos los cambios notables de este proyecto serán documentados en este archivo.

El formato está basado en [Keep a Changelog](https://keepachangelog.com/es-ES/1.0.0/),
y este proyecto adhiere a [Semantic Versioning](https://semver.org/lang/es/).

## [Unreleased]

### Agregado
- Workflow de versionado automático al crear tags
- Scripts de pre-commit para validación antes de commits
- Rama `develop` para flujo de trabajo Git Flow
- CHANGELOG.md para seguimiento de versiones

## [1.0.1] - 2024-01-XX

### Agregado
- Workflow de GitHub Actions para versionado automático
- Actualización automática de versiones en:
  - `api/openapi/spec.yaml`
  - `api/openapi-generator-config.json`
  - `scripts/cli/main.py`

## [1.0.0] - 2024-01-XX

### Agregado
- API REST para gestión de automatizaciones
- Gestión de tareas y subtareas
- Máquina de estados con validación de transiciones
- Auditoría de cambios (created_by, updated_by)
- Soft delete con limpieza automática tras 30 días
- Filtrado y paginación
- Manejo de errores según RFC 7807
- CLI Python para consultas
- Tests unitarios, de integración y E2E
- Docker y Docker Compose para despliegue
- Documentación completa del proyecto

[Unreleased]: https://github.com/notacool/seguimiento-automatismos/compare/v1.0.1...HEAD
[1.0.1]: https://github.com/notacool/seguimiento-automatismos/compare/v1.0.0...v1.0.1
[1.0.0]: https://github.com/notacool/seguimiento-automatismos/releases/tag/v1.0.0

