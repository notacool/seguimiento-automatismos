package http

import (
	"errors"
	"net/http"
	"strings"

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
		// Verificar si el error contiene alguna referencia a errores conocidos
		errStr := err.Error()
		if strings.Contains(errStr, entity.ErrInvalidName.Error()) ||
			strings.Contains(errStr, entity.ErrInvalidStateTransition.Error()) ||
			strings.Contains(errStr, entity.ErrInconsistentParentChildState.Error()) ||
			strings.Contains(errStr, entity.ErrMissingRequiredFields.Error()) ||
			strings.Contains(errStr, entity.ErrTaskNotFound.Error()) ||
			strings.Contains(errStr, entity.ErrSubtaskNotFound.Error()) {
			// Intentar mapear basándose en el contenido del mensaje
			if strings.Contains(errStr, "name") && strings.Contains(errStr, "invalid") {
				pd.Type = "https://api.grupoapi.com/problems/invalid-name"
				pd.Title = "Invalid Task Name"
				pd.Status = http.StatusBadRequest
			} else if strings.Contains(errStr, "state transition") {
				pd.Type = "https://api.grupoapi.com/problems/invalid-state-transition"
				pd.Title = "Invalid State Transition"
				pd.Status = http.StatusBadRequest
			} else if strings.Contains(errStr, "not found") {
				if strings.Contains(errStr, "task") {
					pd.Type = "https://api.grupoapi.com/problems/task-not-found"
					pd.Title = "Task Not Found"
					pd.Status = http.StatusNotFound
				} else if strings.Contains(errStr, "subtask") {
					pd.Type = "https://api.grupoapi.com/problems/subtask-not-found"
					pd.Title = "Subtask Not Found"
					pd.Status = http.StatusNotFound
				} else {
					pd.Type = "https://api.grupoapi.com/problems/internal-error"
					pd.Title = "Internal Server Error"
					pd.Status = http.StatusInternalServerError
				}
			} else {
				pd.Type = "https://api.grupoapi.com/problems/internal-error"
				pd.Title = "Internal Server Error"
				pd.Status = http.StatusInternalServerError
			}
			pd.Detail = errStr
		} else {
			pd.Type = "https://api.grupoapi.com/problems/internal-error"
			pd.Title = "Internal Server Error"
			pd.Status = http.StatusInternalServerError
			pd.Detail = "An unexpected error occurred. Please contact support if the problem persists."
		}
	}

	c.JSON(pd.Status, pd)
}
