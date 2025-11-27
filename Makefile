.PHONY: help build test test-coverage run docker-up docker-down docker-logs docker-build migrate-up migrate-down migrate-create lint fmt clean deps generate-server generate-client generate-all

help: ## Mostrar ayuda
	@echo "Comandos disponibles:"
	@echo "  make build             - Compilar aplicación"
	@echo "  make test              - Ejecutar tests"
	@echo "  make test-coverage     - Ver cobertura de tests"
	@echo "  make run               - Ejecutar aplicación localmente"
	@echo "  make docker-up         - Levantar servicios con Docker Compose"
	@echo "  make docker-down       - Bajar servicios Docker Compose"
	@echo "  make docker-logs       - Ver logs de Docker Compose"
	@echo "  make docker-build      - Construir imagen Docker"
	@echo "  make migrate-up        - Ejecutar migraciones"
	@echo "  make migrate-down      - Revertir migraciones"
	@echo "  make migrate-create    - Crear nueva migración (usar NAME=nombre)"
	@echo "  make generate-server   - Generar código servidor Go desde OpenAPI"
	@echo "  make generate-client   - Generar cliente Python desde OpenAPI"
	@echo "  make generate-all      - Generar servidor Go y cliente Python"
	@echo "  make lint              - Ejecutar linter"
	@echo "  make fmt               - Formatear código"
	@echo "  make clean             - Limpiar archivos generados"
	@echo "  make deps              - Descargar dependencias"

build: ## Compilar aplicación
	go build -o bin/api.exe ./cmd/api

test: ## Ejecutar tests
	go test -v -race ./...

test-coverage: ## Ver cobertura de tests
	go test -v -race -coverprofile=coverage.out ./...
	go tool cover -html=coverage.out

run: ## Ejecutar aplicación localmente
	go run ./cmd/api

docker-build: ## Construir imagen Docker
	docker build -t grupoapi-proces-log:latest -f deployments/docker/Dockerfile .

docker-up: ## Levantar servicios con Docker Compose
	docker-compose -f deployments/docker/docker-compose.yml up -d

docker-down: ## Bajar servicios Docker Compose
	docker-compose -f deployments/docker/docker-compose.yml down

docker-logs: ## Ver logs de Docker Compose
	docker-compose -f deployments/docker/docker-compose.yml logs -f

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

fmt-check: ## Verificar formato sin modificar
	gofumpt -l .

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

cli-deps: ## Instalar dependencias del CLI Python
	cd scripts/cli && pip install -r requirements.txt

cli-build-windows: ## Generar ejecutable CLI para Windows
	cd scripts/cli && pyinstaller --onefile --name automatizacion-cli-windows.exe main.py

cli-build-linux: ## Generar ejecutable CLI para Linux
	cd scripts/cli && pyinstaller --onefile --name automatizacion-cli-linux main.py

.DEFAULT_GOAL := help
