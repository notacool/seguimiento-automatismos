package integration

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/stretchr/testify/require"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"
)

// PostgresContainer encapsula un contenedor PostgreSQL para tests.
type PostgresContainer struct {
	Container  testcontainers.Container
	ConnString string
	Pool       *pgxpool.Pool
}

// SetupPostgresContainer inicia un contenedor PostgreSQL para tests.
// Soporta tanto Docker como Podman automáticamente (testcontainers-go detecta automáticamente).
func SetupPostgresContainer(ctx context.Context, t *testing.T) *PostgresContainer {
	t.Helper()

	req := testcontainers.ContainerRequest{
		Image:        "postgres:16-alpine",
		ExposedPorts: []string{"5432/tcp"},
		Env: map[string]string{
			"POSTGRES_DB":       "test_db",
			"POSTGRES_USER":     "test_user",
			"POSTGRES_PASSWORD": "test_password",
		},
		WaitingFor: wait.ForAll(
			wait.ForLog("database system is ready to accept connections"),
			wait.ForListeningPort("5432/tcp"),
		).WithDeadline(60 * time.Second),
	}

	container, err := testcontainers.GenericContainer(ctx, testcontainers.GenericContainerRequest{
		ContainerRequest: req,
		Started:          true,
	})
	require.NoError(t, err)

	// Obtener host y puerto
	host, err := container.Host(ctx)
	require.NoError(t, err)

	port, err := container.MappedPort(ctx, "5432")
	require.NoError(t, err)

	// Crear connection string
	connString := fmt.Sprintf(
		"postgres://test_user:test_password@%s:%s/test_db?sslmode=disable",
		host,
		port.Port(),
	)

	// Crear pool de conexiones
	poolConfig, err := pgxpool.ParseConfig(connString)
	require.NoError(t, err)

	poolConfig.MaxConns = 10
	poolConfig.MinConns = 2

	pool, err := pgxpool.NewWithConfig(ctx, poolConfig)
	require.NoError(t, err)

	// Verificar conexión
	err = pool.Ping(ctx)
	require.NoError(t, err)

	return &PostgresContainer{
		Container:  container,
		ConnString: connString,
		Pool:       pool,
	}
}

// Teardown limpia el contenedor.
func (pc *PostgresContainer) Teardown(ctx context.Context, t *testing.T) {
	t.Helper()

	if pc.Pool != nil {
		pc.Pool.Close()
	}

	if pc.Container != nil {
		err := pc.Container.Terminate(ctx)
		require.NoError(t, err)
	}
}

// ExecuteSQL ejecuta un script SQL.
func (pc *PostgresContainer) ExecuteSQL(ctx context.Context, t *testing.T, sql string) {
	t.Helper()

	_, err := pc.Pool.Exec(ctx, sql)
	require.NoError(t, err)
}

// CreateTasksTable crea la tabla tasks.
func (pc *PostgresContainer) CreateTasksTable(ctx context.Context, t *testing.T) {
	t.Helper()

	sql := `
		CREATE TABLE IF NOT EXISTS tasks (
			id UUID PRIMARY KEY,
			name VARCHAR(256) NOT NULL,
			state VARCHAR(20) NOT NULL,
			created_by VARCHAR(256) NOT NULL,
			updated_by VARCHAR(256),
			deleted_by VARCHAR(256),
			start_date TIMESTAMP,
			end_date TIMESTAMP,
			created_at TIMESTAMP NOT NULL DEFAULT NOW(),
			updated_at TIMESTAMP NOT NULL DEFAULT NOW(),
			deleted_at TIMESTAMP,
			CONSTRAINT name_valid CHECK (name ~ '^[a-zA-Z0-9 _-]+$'),
			CONSTRAINT state_valid CHECK (state IN ('PENDING', 'IN_PROGRESS', 'COMPLETED', 'FAILED', 'CANCELLED'))
		);

		CREATE INDEX IF NOT EXISTS idx_tasks_state ON tasks(state) WHERE deleted_at IS NULL;
		CREATE INDEX IF NOT EXISTS idx_tasks_name ON tasks(name) WHERE deleted_at IS NULL;
		CREATE INDEX IF NOT EXISTS idx_tasks_created_at ON tasks(created_at) WHERE deleted_at IS NULL;
		CREATE INDEX IF NOT EXISTS idx_tasks_deleted_at ON tasks(deleted_at) WHERE deleted_at IS NOT NULL;
	`

	pc.ExecuteSQL(ctx, t, sql)
}

// CreateSubtasksTable crea la tabla subtasks.
func (pc *PostgresContainer) CreateSubtasksTable(ctx context.Context, t *testing.T) {
	t.Helper()

	sql := `
		CREATE TABLE IF NOT EXISTS subtasks (
			id UUID PRIMARY KEY,
			task_id UUID NOT NULL,
			name VARCHAR(256) NOT NULL,
			state VARCHAR(20) NOT NULL,
			created_by VARCHAR(256) NOT NULL,
			updated_by VARCHAR(256),
			deleted_by VARCHAR(256),
			start_date TIMESTAMP,
			end_date TIMESTAMP,
			created_at TIMESTAMP NOT NULL DEFAULT NOW(),
			updated_at TIMESTAMP NOT NULL DEFAULT NOW(),
			deleted_at TIMESTAMP,
			CONSTRAINT fk_task FOREIGN KEY (task_id) REFERENCES tasks(id) ON DELETE CASCADE,
			CONSTRAINT name_valid CHECK (name ~ '^[a-zA-Z0-9 _-]+$'),
			CONSTRAINT state_valid CHECK (state IN ('PENDING', 'IN_PROGRESS', 'COMPLETED', 'FAILED', 'CANCELLED'))
		);

		CREATE INDEX IF NOT EXISTS idx_subtasks_task_id ON subtasks(task_id) WHERE deleted_at IS NULL;
		CREATE INDEX IF NOT EXISTS idx_subtasks_state ON subtasks(state) WHERE deleted_at IS NULL;
		CREATE INDEX IF NOT EXISTS idx_subtasks_deleted_at ON subtasks(deleted_at) WHERE deleted_at IS NOT NULL;
	`

	pc.ExecuteSQL(ctx, t, sql)
}

// TruncateTables limpia todas las tablas.
func (pc *PostgresContainer) TruncateTables(ctx context.Context, t *testing.T) {
	t.Helper()

	sql := `
		TRUNCATE TABLE subtasks CASCADE;
		TRUNCATE TABLE tasks CASCADE;
	`

	pc.ExecuteSQL(ctx, t, sql)
}
