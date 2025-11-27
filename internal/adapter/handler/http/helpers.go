package http

import (
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"github.com/grupoapi/proces-log/internal/domain/entity"
)

// parseUUIDOrError parsea un UUID desde un string y mapea el error si falla.
// Retorna el UUID y un booleano indicando si el parsing fue exitoso.
// Si falla, ya se ha enviado la respuesta HTTP al cliente.
func parseUUIDOrError(c *gin.Context, uuidStr string, notFoundErr error) (uuid.UUID, bool) {
	id, err := ParseUUID(uuidStr)
	if err != nil {
		MapErrorToProblemDetails(c, notFoundErr)
		return uuid.Nil, false
	}
	return id, true
}

// parseStateOrError parsea un estado opcional desde un string y mapea el error si falla.
// Retorna el estado y un booleano indicando si el parsing fue exitoso.
// Si falla, ya se ha enviado la respuesta HTTP al cliente.
// Si stateStr es nil, retorna nil y true (estado opcional).
func parseStateOrError(c *gin.Context, stateStr *string) (*entity.State, bool) {
	if stateStr == nil {
		return nil, true
	}

	parsedState, err := ParseState(*stateStr)
	if err != nil {
		MapErrorToProblemDetails(c, err)
		return nil, false
	}
	return &parsedState, true
}
