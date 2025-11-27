package http

import (
	"context"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"
)

// HealthHandler maneja el endpoint de health check
type HealthHandler struct {
	db *pgxpool.Pool
}

// NewHealthHandler crea una nueva instancia de HealthHandler
func NewHealthHandler(db *pgxpool.Pool) *HealthHandler {
	return &HealthHandler{db: db}
}

// HealthResponse estructura de respuesta del health check
type HealthResponse struct {
	Status    string    `json:"status"`
	Database  string    `json:"database"`
	Timestamp time.Time `json:"timestamp"`
}

// Check verifica el estado del servicio y la conexión a base de datos
func (h *HealthHandler) Check(c *gin.Context) {
	ctx, cancel := context.WithTimeout(c.Request.Context(), 3*time.Second)
	defer cancel()

	// Verificar conexión a base de datos
	dbStatus := "ok"
	if err := h.db.Ping(ctx); err != nil {
		dbStatus = "error"
		c.JSON(http.StatusServiceUnavailable, HealthResponse{
			Status:    "unhealthy",
			Database:  dbStatus,
			Timestamp: time.Now(),
		})
		return
	}

	c.JSON(http.StatusOK, HealthResponse{
		Status:    "healthy",
		Database:  dbStatus,
		Timestamp: time.Now(),
	})
}
