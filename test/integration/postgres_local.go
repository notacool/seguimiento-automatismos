package integration

import (
	"context"
	"fmt"
	"os"
	"testing"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/stretchr/testify/require"
)

// PostgresLocal encapsula una conexión a PostgreSQL local para tests.
type PostgresLocal struct {
	ConnString string
	Pool       *pgxpool.Pool
	Database   string // Nombre de la base de datos temporal
}

// SetupPostgresLocal crea una conexión a PostgreSQL local para tests.
// Requiere que PostgreSQL esté instalado y corriendo localmente.
// Crea una base de datos temporal que se limpia después del test.
func SetupPostgresLocal(ctx context.Context, t *testing.T) *PostgresLocal {
	t.Helper()

	// Obtener configuración de conexión desde variables de entorno o usar defaults
	host := getEnv("TEST_DB_HOST", "localhost")
	port := getEnv("TEST_DB_PORT", "5432")
	user := getEnv("TEST_DB_USER", "postgres")
	password := getEnv("TEST_DB_PASSWORD", "postgres")
	mainDB := getEnv("TEST_DB_MAIN", "postgres") // Base de datos para crear la temporal

	// Crear nombre único para la base de datos temporal
	testDBName := fmt.Sprintf("test_%d_%d", time.Now().Unix(), os.Getpid())

	// Conectar a la base de datos principal para crear la temporal
	mainConnString := fmt.Sprintf(
		"postgres://%s:%s@%s:%s/%s?sslmode=disable",
		user, password, host, port, mainDB,
	)

	mainPool, err := pgxpool.New(ctx, mainConnString)
	require.NoError(t, err, "No se pudo conectar a PostgreSQL. Asegúrate de que esté instalado y corriendo.")
	defer mainPool.Close()

	// Crear base de datos temporal
	_, err = mainPool.Exec(ctx, fmt.Sprintf("CREATE DATABASE %s", testDBName))
	require.NoError(t, err, "No se pudo crear la base de datos temporal")

	// Registrar cleanup para eliminar la base de datos al finalizar
	t.Cleanup(func() {
		cleanupCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		// Terminar todas las conexiones a la base de datos temporal
		cleanupPool, err := pgxpool.New(cleanupCtx, mainConnString)
		if err == nil {
			// Terminar conexiones activas
			_, _ = cleanupPool.Exec(cleanupCtx, fmt.Sprintf(
				"SELECT pg_terminate_backend(pid) FROM pg_stat_activity WHERE datname = '%s' AND pid <> pg_backend_pid()",
				testDBName,
			))
			// Eliminar base de datos
			_, _ = cleanupPool.Exec(cleanupCtx, fmt.Sprintf("DROP DATABASE IF EXISTS %s", testDBName))
			cleanupPool.Close()
		}
	})

	// Conectar a la base de datos temporal
	connString := fmt.Sprintf(
		"postgres://%s:%s@%s:%s/%s?sslmode=disable",
		user, password, host, port, testDBName,
	)

	poolConfig, err := pgxpool.ParseConfig(connString)
	require.NoError(t, err)

	poolConfig.MaxConns = 10
	poolConfig.MinConns = 2

	pool, err := pgxpool.NewWithConfig(ctx, poolConfig)
	require.NoError(t, err)

	// Verificar conexión
	err = pool.Ping(ctx)
	require.NoError(t, err)

	return &PostgresLocal{
		ConnString: connString,
		Pool:       pool,
		Database:   testDBName,
	}
}

// Teardown limpia la conexión (la base de datos se elimina automáticamente en Cleanup).
func (pl *PostgresLocal) Teardown(ctx context.Context, t *testing.T) {
	t.Helper()

	if pl.Pool != nil {
		pl.Pool.Close()
	}
}

// ExecuteSQL ejecuta un script SQL.
func (pl *PostgresLocal) ExecuteSQL(ctx context.Context, t *testing.T, sql string) {
	t.Helper()

	_, err := pl.Pool.Exec(ctx, sql)
	require.NoError(t, err)
}

// CreateTasksTable crea la tabla tasks.
func (pl *PostgresLocal) CreateTasksTable(ctx context.Context, t *testing.T) {
	t.Helper()

	sql := `
		CREATE TABLE IF NOT EXISTS tasks (
			id UUID PRIMARY KEY,
			name VARCHAR(256) NOT NULL,
			state VARCHAR(20) NOT NULL,
			created_by VARCHAR(256) NOT NULL,
			updated_by VARCHAR(256),
			start_date TIMESTAMPTZ,
			end_date TIMESTAMPTZ,
			created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
			updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
			deleted_at TIMESTAMPTZ,
			CONSTRAINT name_valid CHECK (name ~ '^[a-zA-Z0-9 _-]+$'),
			CONSTRAINT state_valid CHECK (state IN ('PENDING', 'IN_PROGRESS', 'COMPLETED', 'FAILED', 'CANCELLED'))
		);

		CREATE INDEX IF NOT EXISTS idx_tasks_state ON tasks(state) WHERE deleted_at IS NULL;
		CREATE INDEX IF NOT EXISTS idx_tasks_name ON tasks(name) WHERE deleted_at IS NULL;
		CREATE INDEX IF NOT EXISTS idx_tasks_created_at ON tasks(created_at DESC) WHERE deleted_at IS NULL;
		CREATE INDEX IF NOT EXISTS idx_tasks_deleted_at ON tasks(deleted_at) WHERE deleted_at IS NOT NULL;
	`

	pl.ExecuteSQL(ctx, t, sql)
}

// CreateSubtasksTable crea la tabla subtasks.
func (pl *PostgresLocal) CreateSubtasksTable(ctx context.Context, t *testing.T) {
	t.Helper()

	sql := `
		CREATE TABLE IF NOT EXISTS subtasks (
			id UUID PRIMARY KEY,
			task_id UUID NOT NULL REFERENCES tasks(id) ON DELETE CASCADE,
			name VARCHAR(256) NOT NULL,
			state VARCHAR(20) NOT NULL,
			start_date TIMESTAMPTZ,
			end_date TIMESTAMPTZ,
			created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
			updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
			deleted_at TIMESTAMPTZ,
			CONSTRAINT name_valid CHECK (name ~ '^[a-zA-Z0-9 _-]+$'),
			CONSTRAINT state_valid CHECK (state IN ('PENDING', 'IN_PROGRESS', 'COMPLETED', 'FAILED', 'CANCELLED')),
			CONSTRAINT valid_subtask_dates CHECK (
				(start_date IS NULL OR start_date >= created_at) AND
				(end_date IS NULL OR (start_date IS NOT NULL AND end_date >= start_date))
			),
			CONSTRAINT valid_subtask_final_state_dates CHECK (
				(state IN ('COMPLETED', 'FAILED', 'CANCELLED') AND end_date IS NOT NULL) OR
				(state NOT IN ('COMPLETED', 'FAILED', 'CANCELLED'))
			)
		);

		CREATE INDEX IF NOT EXISTS idx_subtasks_task_id ON subtasks(task_id) WHERE deleted_at IS NULL;
		CREATE INDEX IF NOT EXISTS idx_subtasks_state ON subtasks(state) WHERE deleted_at IS NULL;
		CREATE INDEX IF NOT EXISTS idx_subtasks_created_at ON subtasks(created_at DESC) WHERE deleted_at IS NULL;
		CREATE INDEX IF NOT EXISTS idx_subtasks_deleted_at ON subtasks(deleted_at) WHERE deleted_at IS NOT NULL;
		CREATE INDEX IF NOT EXISTS idx_subtasks_task_state ON subtasks(task_id, state) WHERE deleted_at IS NULL;
	`

	pl.ExecuteSQL(ctx, t, sql)
}

// TruncateTables limpia todas las tablas.
func (pl *PostgresLocal) TruncateTables(ctx context.Context, t *testing.T) {
	t.Helper()

	sql := `
		TRUNCATE TABLE subtasks CASCADE;
		TRUNCATE TABLE tasks CASCADE;
	`

	pl.ExecuteSQL(ctx, t, sql)
}

// getEnv obtiene una variable de entorno o retorna un valor por defecto.
func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}
