# Guía de Desarrollo en GrupoAPI

## 1. Gestión de Issues, Pull Requests y Milestones

### 1.1 Issues (Incidencias)

Los Issues establecen el objetivo de cada desarrollo y deben estar bien documentados. Son la base de la gestión de tareas y objetivos.

Clasificación recomendada de Issues:

- **Feature [FEATURE]**: Nueva funcionalidad o mejora.
- **Bugfix [BUG]**: Corrección de errores. Debe incluir pasos para reproducir y datos relevantes.
- **Documentación [DOC]**: Mejoras o correcciones en la documentación o archivos derivados (ejemplo: docker compose, scripts, etc).
- **Refactor [REF]**: Mejoras en la legibilidad o estructura del código sin cambiar funcionalidad.
- **Consulta al desarrollador**: Pregunta o solicitud del cliente que puede derivar en una nueva Issue.

Ejemplo de Issue:

```markdown
Título: [BUG] Error al iniciar sesión
Descripción: Al intentar iniciar sesión con usuario válido, la aplicación muestra error 500. Pasos para reproducir: ...
```


### 1.2 PR

Los Pull Requests (PR) deben estar vinculados a un Issue y se recomienda crearlos en modo draft, con un primer commit que describa el objetivo.
Nomenclatura recomendada para PR: `[TIPO] Breve descripción del objetivo`

Flujo de trabajo recomendado:

1. Crear rama desde `develop` para Features y mejoras.
2. Crear rama desde `main` solo para Bugfix críticos en producción.
3. Los PR se hacen contra `develop` por defecto. Solo los Bugfix urgentes se mergean directo a `main` y luego se actualiza `develop` mediante PR.
4. No se debe mergear a `main` hasta que el Milestone esté finalizado.

### 1.3 Milestones

Para agrupar los issues en entregables usaremos los Milestones. Con un Milestone agrupamos las Issues y los PR para definir nuestros objetivos.
Ejemplo:

Ejemplo de Milestone:

```markdown
Milestone: "Entrega inicial"
Incluye: Issues y PR necesarios para el primer entregable funcional al usuario.
```


### 1.4 Labels


Se recomienda usar Labels para clasificar y priorizar los Issues:

- frontend: indica si el feature o bugfix es parte del frontend de la aplicación.
- backend: indica si el feature o bugfix es parte del backend de la aplicación.
- scope change: indica que el feature es un cambio de alcance del proyecto.
- documentation: indica que se requiere documentar algo en específico.
- critical: indica que es un issue crítico y tiene prioridad.
- no prioritario: indica que el issue no es prioritario, ya que es una mejora no esperada (no usar por defecto).
- scripts: para establecer algún cambio en un script de apoyo, sería para documentación.
- linter: cuando el Issue se genera por errores de linter no solucionados.
- analysis required: necesita revisión y análisis para el desarrollo.
- dependant: esta issue depende de otra.
- blocked: issue bloqueada hasta que se cumpla una condición.

Ejemplo de uso de Labels:
`[critical][backend][blocked]`


## 2. Documentación

La documentación debe ser clara, actualizada y accesible. Si el archivo readme se hace demasiado gande, se recomienda la siguiente estructura:

- **README.MD**: Información básica de la aplicación, cómo desplegarla y su funcionamiento.
- **README_DEVELOP.md**: Lenguajes de programación y frameworks utilizados, decisiones de arquitectura, patrones y cómo abordar el desarrollo desde cero.
- **README_TESTING.md**: Explicación de cómo probar la aplicación, incluyendo test unitarios, de integración y E2E.
- **README_AYUDAS.md**: Explicación de los scripts de apoyo en el desarrollo y despliegue.

siendo los 3 ultimos opcionales.

Ejemplo de sección en README.md:

```markdown
## Instalación
1. Clonar el repositorio
2. Instalar dependencias
3. Configurar variables de entorno
4. Ejecutar la aplicación
```

Se pueden crear README específicos por módulos o funciones y vincularlos al README.md principal.

### 2.2 SCRIPTS

Se recomienda tener una carpeta `scripts/` con subcarpetas por tipo (ejemplo: `deploy/`, `db/`, `test/`). Cada subcarpeta debe incluir un README explicativo.
Los scripts deben estar disponibles para Linux y PowerShell. Opcionalmente, se pueden crear para Make.

**Scripts recomendados:**

- Pre-commit: Ejecuta formateadores y linter antes de cada commit.
- Instalación de dependencias: Automatiza la instalación de dependencias para el desarrollo.

Si es necesario, documentar scripts de base de datos: uno para la definición y cambios mínimos, y otro para datos precargados en pruebas.
Todo lo adicional es opcional, pero debe estar documentado si se comparte con el equipo.

### 2.3 Docker

Si es viable, crear un `Dockerfile` para construir la imagen y un `docker-compose.yml` para lanzar la aplicación.

Ejemplo de estructura mínima:

```dockerfile
# Dockerfile básico
FROM node:18
WORKDIR /app
COPY . .
RUN npm install
CMD ["npm", "start"]
```

```yaml
# docker-compose.yml básico
version: '3'
services:
	app:
		build: .
		ports:
			- "3000:3000"
```

## 3. Commits y Merges

**Buenas prácticas:**


- **Merge a main:** Solo se mergea contra `main` si es un bugfix de producción. El resto va a `develop` y, cuando funciona, se mergea a `main`.
- **Pre-commit:** Antes de cada commit, ejecutar linter y test unitarios.
- **Merge a develop:** Los merges hacia `develop` requieren PR documentado y se recomienda usar "squash and merge".
- **Merge de develop a main:** Los merges de `develop` a `main` serán merges normales (sin squash). Los merges generando versión de `develop` a `main` serán merges normales y no usarán squash.
- **Mensajes de commit:** Usar mensajes descriptivos y en presente. Ejemplo: `Agrega validación de usuario en login`.


## 4. Linter y Test

**Recomendaciones:**

1. Usar formateador de código antes de ejecutar linter.
2. Los linter deben estar configurados según el lenguaje y documentados en README_TESTING.
3. Mantener y actualizar los tests unitarios y de integración.

Tipos de test:

- Unitarios: dependen del lenguaje y deben mantenerse.
- Integración: no dependen del lenguaje y deben mantenerse con los cambios.

**Ejemplo de flujo de validación:**

1. Ejecutar formateador
2. Ejecutar linter
3. Ejecutar tests
4. Realizar commit si todo es correcto



## 5. Revisión de Código (Code Review)

En la mayoría de los proyectos solo habrá un desarrollador, por lo que los Pull Requests (PR) se utilizan principalmente para documentar los cambios realizados y dejar constancia del trabajo. Se recomienda:

- Describir claramente el objetivo y alcance de cada PR.
- Documentar las decisiones técnicas relevantes.
- Si el equipo crece, seguir buenas prácticas de revisión y solicitar feedback cuando sea posible.

## 6. Convenciones de Nombres

- Ramas: `tipo/descripcion-corta` (ejemplo: `feature/login-social`, `bugfix/error-login`)
- Commits: Mensajes en presente y descriptivos (ejemplo: `Corrige validación de email`)
- Archivos: Nombres en minúsculas y descriptivos, usar guiones para separar palabras.



## 7. Integración Continua (CI/CD)

Se recomienda usar GitHub Actions, GitLab CI o similar para automatizar pruebas y despliegues.
Ejemplo básico de workflow para Node.js:

```yaml
name: CI
on: [push, pull_request]
jobs:
	build:
		runs-on: ubuntu-latest
		steps:
			- uses: actions/checkout@v3
			- name: Instalar dependencias
				run: npm install
			- name: Ejecutar tests
				run: npm test
```




## 8. Seguridad

- No subir credenciales ni secretos al repositorio.
- Usar archivos `.env` y agregarlo a `.gitignore`.
- Mantener dependencias actualizadas y revisar vulnerabilidades.
- Validar entradas de usuario y evitar inyecciones.

Se debe incluir un archivo `.env.example` con las variables necesarias para el entorno, y asegurarse de que `.env` esté en el `.gitignore` para evitar exponer datos sensibles.


Ejemplo de `.env.example`:

```env
DB_HOST=localhost
DB_USER=usuario
DB_PASS=contraseña
SECRET_KEY=tu_clave_secreta
```


Ejemplo de entrada en `.gitignore`:

```
.env
```




## 9. Plantillas para GitHub

Para mejorar la documentación y el seguimiento de cambios, se recomienda crear las siguientes plantillas en el repositorio:

- Las plantillas de Issues se ubican en `.github/ISSUE_TEMPLATE/`.
- La plantilla de Pull Request se ubica en `.github/PULL_REQUEST_TEMPLATE.md`.

### Uso de las plantillas

Al crear un nuevo Issue o Pull Request en GitHub, se podrá seleccionar la plantilla correspondiente. Esto ayuda a mantener la información organizada y facilita el seguimiento de tareas y cambios.

#### Ejemplo de plantilla de Issue (`.github/ISSUE_TEMPLATE/bug_report.md`)

```markdown
---
name: Bug Report
about: Reporta un bug para que sea corregido
title: '[BUG] Breve descripción'
labels: bug
assignees: ''
---

**Descripción**
Describe el problema encontrado.

**Pasos para reproducir**
1. ...
2. ...

**Comportamiento esperado**
...

**Capturas de pantalla**
Si aplica, añade imágenes.

**Información adicional**
Sistema operativo, versión, etc.
```

#### Ejemplo de plantilla de Pull Request (`.github/PULL_REQUEST_TEMPLATE.md`)

```markdown
# Descripción
Breve resumen de los cambios realizados.

# Relacionado con Issue(s)
Closes #XX

# Checklist

- [ ] El código sigue la guía de estilos
- [ ] Se han agregado pruebas
- [ ] La documentación está actualizada
- [ ] No hay credenciales ni secretos expuestos

# Notas adicionales
...
```




## 10. Recursos Útiles

- [Guía de Markdown](https://guides.github.com/features/mastering-markdown/)
- [Documentación oficial de GitHub Actions](https://docs.github.com/en/actions)
- [Guía de buenas prácticas de seguridad en Node.js](https://nodejs.org/en/learn/security/)
