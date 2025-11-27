package http

import (
	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"
)

// SetupRouter configura y retorna el router con todas las rutas
func SetupRouter(db *pgxpool.Pool, ginMode string) *gin.Engine {
	gin.SetMode(ginMode)

	router := gin.New()

	// Middleware
	router.Use(gin.Logger())
	router.Use(gin.Recovery())

	// Health check endpoint
	healthHandler := NewHealthHandler(db)
	router.GET("/health", healthHandler.Check)

	// API v1 routes (se añadirán después)
	// v1 := router.Group("/api/v1")

	return router
}
