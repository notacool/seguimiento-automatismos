.PHONY: help build test test-unit test-integration test-e2e test-all test-coverage run docker-up docker-down docker-logs docker-build migrate-up migrate-down migrate-create lint fmt fmt-check clean deps detect-container-runtime

# Cargar variables de entorno desde .env si existe
-include .env
export

# Detectar runtime de contenedores (docker, podman, o auto)
CONTAINER_RUNTIME ?= auto

# Función para detectar el comando de contenedores
define detect_docker_cmd
$(shell \
	if [ "$(CONTAINER_RUNTIME)" = "podman" ]; then \
		echo "podman"; \
	elif [ "$(CONTAINER_RUNTIME)" = "docker" ]; then \
		echo "docker"; \
	elif [ "$(CONTAINER_RUNTIME)" = "auto" ]; then \
		if command -v podman >/dev/null 2>&1 && ! command -v docker >/dev/null 2>&1; then \
			echo "podman"; \
		elif command -v docker >/dev/null 2>&1; then \
			echo "docker"; \
		elif command -v podman >/dev/null 2>&1; then \
			echo "podman"; \
		else \
			echo "docker"; \
		fi; \
	else \
		echo "docker"; \
	fi \
)
endef

DOCKER_CMD := $(call detect_docker_cmd)
COMPOSE_CMD := $(shell if [ "$(DOCKER_CMD)" = "podman" ]; then echo "podman-compose"; else echo "docker-compose"; fi)

help: ## Mostrar ayuda
	@echo "Comandos disponibles:"
	@echo ""
	@echo "Desarrollo:"
	@echo "  make build             - Compilar aplicación"
	@echo "  make run               - Ejecutar aplicación localmente"
	@echo "  make deps              - Descargar dependencias Go"
	@echo "  make fmt               - Formatear código"
	@echo "  make lint              - Ejecutar linter"
	@echo ""
	@echo "Tests:"
	@echo "  make test              - Ejecutar tests unitarios"
	@echo "  make test-unit         - Ejecutar solo tests unitarios"
	@echo "  make test-integration  - Ejecutar tests de integración"
	@echo "  make test-e2e          - Ejecutar tests End-to-End (con contenedores)"
	@echo "  make test-e2e-local    - Ejecutar tests E2E con PostgreSQL local (sin contenedores)"
	@echo "  make test-all          - Ejecutar todos los tests"
	@echo "  make test-coverage     - Ver cobertura de tests"
	@echo ""
	@echo "Contenedores:"
	@echo "  make docker-up         - Levantar servicios con Docker/Podman Compose"
	@echo "  make docker-down       - Bajar servicios Docker/Podman Compose"
	@echo "  make docker-logs       - Ver logs de Docker/Podman Compose"
	@echo "  make docker-build      - Construir imagen Docker/Podman"
	@echo "  make detect-container-runtime - Detectar runtime de contenedores"
	@echo ""
	@echo "  Configurar en .env: CONTAINER_RUNTIME=docker|podman|auto (default: auto)"
	@echo ""
	@echo "Base de Datos:"
	@echo "  make migrate-up        - Ejecutar migraciones"
	@echo "  make migrate-down      - Revertir migraciones"
	@echo "  make migrate-create    - Crear nueva migración (usar NAME=nombre)"
	@echo ""
	@echo "Generación de Código:"
	@echo "  make generate-server   - Generar código servidor Go desde OpenAPI"
	@echo "  make generate-client   - Generar cliente Python desde OpenAPI"
	@echo "  make generate-all      - Generar servidor Go y cliente Python"
	@echo ""
	@echo "CLI Python (Consulta):"
	@echo "  make cli-setup         - Crear entorno virtual para CLI"
	@echo "  make cli-install       - Configurar venv e instalar dependencias"
	@echo "  make cli-deps          - Instalar dependencias (requiere venv activado)"
	@echo "  make cli-run           - Ejecutar CLI (ej: make cli-run CMD=\"task list\")"
	@echo "  make cli-test          - Ejecutar tests del CLI"
	@echo "  make cli-test-coverage - Ejecutar tests con reporte de cobertura"
	@echo "  make cli-build-windows - Generar ejecutable Windows"
	@echo "  make cli-build-linux   - Generar ejecutable Linux"
	@echo ""
	@echo "Otros:"
	@echo "  make clean             - Limpiar archivos generados"

build: ## Compilar aplicación
	go build -o bin/api.exe ./cmd/api

test: test-unit ## Ejecutar tests unitarios (alias de test-unit)

test-unit: ## Ejecutar solo tests unitarios (excluyendo integración y E2E)
	go test -v -race -short ./internal/... ./test/helpers/...

test-integration: ## Ejecutar tests de integración con PostgreSQL
	go test -v -race -tags=integration ./test/integration/...

test-e2e: ## Ejecutar tests End-to-End (requiere Docker/Podman)
	@if [ "$(DOCKER_CMD)" = "podman" ]; then \
		bash scripts/start-podman-service.sh || exit 1; \
		echo "🧪 Ejecutando tests E2E..."; \
		env DOCKER_HOST=$${DOCKER_HOST:-unix:///run/user/$$(id -u)/podman/podman.sock} \
		    TESTCONTAINERS_RYUK_DISABLED=$${TESTCONTAINERS_RYUK_DISABLED:-true} \
		    go test -v -race -tags=container ./test/e2e/...; \
	else \
		echo "🧪 Ejecutando tests E2E..."; \
		go test -v -race -tags=container ./test/e2e/...; \
	fi

test-e2e-local: ## Ejecutar tests E2E con PostgreSQL local (sin contenedores)
	@echo "Asegúrate de tener PostgreSQL instalado y corriendo"
	@echo "Variables de entorno opcionales: TEST_DB_HOST, TEST_DB_PORT, TEST_DB_USER, TEST_DB_PASSWORD"
	go test -v -race -tags=!container ./test/e2e/...

test-all: ## Ejecutar todos los tests (unitarios, integración y E2E)
	go test -v -race ./...

test-coverage: ## Ver cobertura de tests
	go test -v -race -coverprofile=coverage.out ./...
	go tool cover -html=coverage.out

test-coverage-html: ## Generar reporte HTML de cobertura
	go test -v -race -coverprofile=coverage.out -covermode=atomic ./...
	go tool cover -html=coverage.out -o coverage.html
	@echo "Reporte de cobertura generado en coverage.html"

run: ## Ejecutar aplicación localmente
	go run ./cmd/api

detect-container-runtime: ## Detectar y mostrar el runtime de contenedores que se usará
	@echo "Runtime detectado: $(DOCKER_CMD)"
	@echo "Compose detectado: $(COMPOSE_CMD)"
	@echo "Configuración: CONTAINER_RUNTIME=$(CONTAINER_RUNTIME)"
	@if [ "$(DOCKER_CMD)" = "podman" ]; then \
		echo "✓ Usando Podman"; \
		if [ -z "$$DOCKER_HOST" ]; then \
			echo "⚠️  DOCKER_HOST no está configurado. Para Podman, ejecuta:"; \
			echo "   export DOCKER_HOST=unix:///run/user/$$(id -u)/podman/podman.sock"; \
		fi; \
	else \
		echo "✓ Usando Docker"; \
	fi

docker-build: ## Construir imagen Docker/Podman
	@echo "Usando: $(DOCKER_CMD)"
	$(DOCKER_CMD) build -t grupoapi-proces-log:latest -f deployments/docker/Dockerfile .

docker-up: ## Levantar servicios con Docker/Podman Compose
	@echo "Usando: $(COMPOSE_CMD)"
	$(COMPOSE_CMD) -f deployments/docker/docker-compose.yml up -d

docker-down: ## Bajar servicios Docker/Podman Compose
	@echo "Usando: $(COMPOSE_CMD)"
	$(COMPOSE_CMD) -f deployments/docker/docker-compose.yml down

docker-logs: ## Ver logs de Docker/Podman Compose
	@echo "Usando: $(COMPOSE_CMD)"
	$(COMPOSE_CMD) -f deployments/docker/docker-compose.yml logs -f

migrate-up: ## Ejecutar migraciones
	migrate -path internal/adapter/repository/postgres/migrations -database "$(DATABASE_URL)" up

migrate-down: ## Revertir migraciones
	migrate -path internal/adapter/repository/postgres/migrations -database "$(DATABASE_URL)" down

migrate-create: ## Crear nueva migración (uso: make migrate-create NAME=create_tasks)
	migrate create -ext sql -dir internal/adapter/repository/postgres/migrations -seq $(NAME)

lint: ## Ejecutar linter
	golangci-lint run ./...

fmt: ## Formatear código
	gofumpt -l -w .
	goimports -w -local github.com/grupoapi/proces-log .

fmt-check: ## Verificar formato sin modificar
	gofumpt -l .
	goimports -l -local github.com/grupoapi/proces-log .

clean: ## Limpiar archivos generados
	rm -rf bin/
	rm -f coverage.out

deps: ## Descargar dependencias
	go mod download
	go mod tidy

generate-server: ## Generar código servidor Go desde OpenAPI spec
	@echo "Generando código servidor Go con oapi-codegen..."
	@mkdir -p internal/adapter/handler/http/generated
	oapi-codegen -config api/oapi-codegen.yaml api/openapi/spec.yaml > internal/adapter/handler/http/generated/api.gen.go
	@echo "✓ Código servidor generado en internal/adapter/handler/http/generated/api.gen.go"

generate-client: ## Generar cliente Python desde OpenAPI spec
	@echo "Generando cliente Python con openapi-generator..."
	@echo "IMPORTANTE: Asegúrate de tener Docker instalado o instala openapi-generator-cli:"
	@echo "  npm install @openapitools/openapi-generator-cli -g"
	@echo "  o"
	@echo "  pip install openapi-generator-cli"
	@echo ""
	@echo "Comando para generar cliente Python:"
	@echo "  openapi-generator-cli generate -i api/openapi/spec.yaml -g python -o generated/python-client -c api/openapi-generator-config.json"
	@echo ""
	@echo "O con Docker:"
	@echo "  docker run --rm -v $(PWD):/local openapitools/openapi-generator-cli generate \\"
	@echo "    -i /local/api/openapi/spec.yaml \\"
	@echo "    -g python \\"
	@echo "    -o /local/generated/python-client \\"
	@echo "    -c /local/api/openapi-generator-config.json"

generate-all: generate-server ## Generar servidor Go y cliente Python
	@echo ""
	@echo "✓ Generación de código servidor completada"
	@echo "ℹ Para generar el cliente Python, ejecuta: make generate-client"

cli-setup: ## Configurar entorno virtual del CLI Python
	@echo "Configurando entorno virtual para CLI..."
	cd scripts/cli && python3 -m venv venv || python -m venv venv
	@echo "✓ Entorno virtual creado"
	@echo "Para activar el entorno virtual:"
	@echo "  source scripts/cli/venv/bin/activate  (Linux/Mac)"
	@echo "  scripts\\cli\\venv\\Scripts\\activate     (Windows)"

cli-deps: ## Instalar dependencias del CLI Python (requiere venv activado)
	@echo "Instalando dependencias..."
	cd scripts/cli && pip install -r requirements.txt
	@echo "✓ Dependencias instaladas"

cli-install: cli-setup ## Configurar venv e instalar dependencias
	@echo "Instalando dependencias en el venv..."
	cd scripts/cli && source venv/bin/activate && pip install --upgrade pip && pip install -r requirements.txt
	@echo "✓ CLI configurado correctamente"
	@echo ""
	@echo "Para usar el CLI:"
	@echo "  1. Activar venv: source scripts/cli/venv/bin/activate"
	@echo "  2. Ejecutar: python scripts/cli/main.py --help"

cli-run: ## Ejecutar CLI (ejemplo: make cli-run CMD="task list")
	cd scripts/cli && source venv/bin/activate && python main.py $(CMD)

cli-test: ## Ejecutar tests del CLI
	@echo "Ejecutando tests del CLI..."
	cd scripts/cli && source venv/bin/activate && pip install -q -r requirements-dev.txt && pytest -v

cli-test-coverage: ## Ejecutar tests con reporte de cobertura
	@echo "Ejecutando tests con cobertura..."
	cd scripts/cli && source venv/bin/activate && pip install -q -r requirements-dev.txt && pytest --cov=. --cov-report=html --cov-report=term
	@echo "✓ Reporte de cobertura generado en scripts/cli/htmlcov/index.html"

cli-test-watch: ## Ejecutar tests en modo watch (requiere pytest-watch)
	cd scripts/cli && source venv/bin/activate && pip install -q pytest-watch && ptw

cli-build-windows: ## Generar ejecutable CLI para Windows
	cd scripts/cli && source venv/bin/activate && pyinstaller --onefile --name automatizacion-cli-windows.exe main.py

cli-build-linux: ## Generar ejecutable CLI para Linux
	cd scripts/cli && source venv/bin/activate && pyinstaller --onefile --name automatizacion-cli-linux main.py

.DEFAULT_GOAL := help
