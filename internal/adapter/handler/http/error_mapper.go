package http

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/grupoapi/proces-log/internal/domain/entity"
)

// ProblemDetails representa un error según RFC 7807
type ProblemDetails struct {
	Type     string `json:"type"`
	Title    string `json:"title"`
	Status   int    `json:"status"`
	Detail   string `json:"detail,omitempty"`
	Instance string `json:"instance,omitempty"`
}

// MapErrorToProblemDetails mapea errores de dominio a RFC 7807 Problem Details
// Usa errors.Is() para detectar errores envueltos, eliminando la necesidad de
// lógica frágil basada en strings.Contains()
func MapErrorToProblemDetails(c *gin.Context, err error) {
	var pd ProblemDetails
	pd.Instance = c.Request.URL.Path

	switch {
	case errors.Is(err, entity.ErrInvalidName):
		pd.Type = "https://api.grupoapi.com/problems/invalid-name"
		pd.Title = "Invalid Task Name"
		pd.Status = http.StatusBadRequest
		pd.Detail = err.Error()

	case errors.Is(err, entity.ErrInvalidStateTransition):
		pd.Type = "https://api.grupoapi.com/problems/invalid-state-transition"
		pd.Title = "Invalid State Transition"
		pd.Status = http.StatusBadRequest
		pd.Detail = err.Error()

	case errors.Is(err, entity.ErrInconsistentParentChildState):
		pd.Type = "https://api.grupoapi.com/problems/inconsistent-parent-child-state"
		pd.Title = "Inconsistent Parent-Child State"
		pd.Status = http.StatusBadRequest
		pd.Detail = err.Error()

	case errors.Is(err, entity.ErrMissingRequiredFields):
		pd.Type = "https://api.grupoapi.com/problems/missing-required-fields"
		pd.Title = "Missing Required Fields"
		pd.Status = http.StatusBadRequest
		pd.Detail = err.Error()

	case errors.Is(err, entity.ErrTaskNotFound):
		pd.Type = "https://api.grupoapi.com/problems/task-not-found"
		pd.Title = "Task Not Found"
		pd.Status = http.StatusNotFound
		pd.Detail = err.Error()

	case errors.Is(err, entity.ErrSubtaskNotFound):
		pd.Type = "https://api.grupoapi.com/problems/subtask-not-found"
		pd.Title = "Subtask Not Found"
		pd.Status = http.StatusNotFound
		pd.Detail = err.Error()

	case errors.Is(err, entity.ErrDatabaseUnavailable):
		pd.Type = "https://api.grupoapi.com/problems/database-unavailable"
		pd.Title = "Database Unavailable"
		pd.Status = http.StatusServiceUnavailable
		pd.Detail = "Cannot connect to database. Service is temporarily unavailable."

	case errors.Is(err, entity.ErrDatabaseError):
		pd.Type = "https://api.grupoapi.com/problems/database-error"
		pd.Title = "Database Error"
		pd.Status = http.StatusInternalServerError
		pd.Detail = "An unexpected database error occurred. Please try again later."

	default:
		// Error desconocido: devolver error genérico
		// errors.Is() ya maneja el unwrapping automáticamente, por lo que
		// si llegamos aquí, el error no es uno de los errores conocidos del dominio
		pd.Type = "https://api.grupoapi.com/problems/internal-error"
		pd.Title = "Internal Server Error"
		pd.Status = http.StatusInternalServerError
		pd.Detail = "An unexpected error occurred. Please contact support if the problem persists."
	}

	c.JSON(pd.Status, pd)
}
