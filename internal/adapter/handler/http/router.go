package http

import (
	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/grupoapi/proces-log/internal/adapter/repository/postgres"
	"github.com/grupoapi/proces-log/internal/domain/service"
	subtaskUsecase "github.com/grupoapi/proces-log/internal/usecase/subtask"
	taskUsecase "github.com/grupoapi/proces-log/internal/usecase/task"
)

// SetupRouter configura y retorna el router con todas las rutas
func SetupRouter(db *pgxpool.Pool, ginMode string) *gin.Engine {
	gin.SetMode(ginMode)

	router := gin.New()

	// Middleware
	router.Use(gin.Logger())
	router.Use(gin.Recovery())

	// Inicializar repositorios
	taskRepo := postgres.NewTaskRepository(db)
	subtaskRepo := postgres.NewSubtaskRepository(db)

	// Inicializar servicios de dominio
	stateMachine := service.NewStateMachine()

	// Inicializar casos de uso de tareas
	createTaskUseCase := taskUsecase.NewCreateTaskUseCase(taskRepo)
	getTaskUseCase := taskUsecase.NewGetTaskUseCase(taskRepo)
	listTasksUseCase := taskUsecase.NewListTasksUseCase(taskRepo)
	updateTaskUseCase := taskUsecase.NewUpdateTaskUseCase(taskRepo, subtaskRepo, stateMachine)

	// Inicializar casos de uso de subtareas
	updateSubtaskUseCase := subtaskUsecase.NewUpdateSubtaskUseCase(subtaskRepo, taskRepo, stateMachine)
	deleteSubtaskUseCase := subtaskUsecase.NewDeleteSubtaskUseCase(subtaskRepo)

	// Inicializar handlers
	healthHandler := NewHealthHandler(db)
	taskHandler := NewTaskHandler(createTaskUseCase, getTaskUseCase, listTasksUseCase, updateTaskUseCase)
	subtaskHandler := NewSubtaskHandler(updateSubtaskUseCase, deleteSubtaskUseCase)

	// Health check endpoint
	router.GET("/health", healthHandler.Check)

	// Task endpoints
	router.POST("/Automatizacion", taskHandler.Create)
	router.PUT("/Automatizacion", taskHandler.Update)
	router.GET("/Automatizacion/:uuid", taskHandler.Get)
	router.GET("/AutomatizacionListado", taskHandler.List)

	// Subtask endpoints
	router.PUT("/Subtask/:uuid", subtaskHandler.Update)
	router.DELETE("/Subtask/:uuid", subtaskHandler.Delete)

	return router
}
