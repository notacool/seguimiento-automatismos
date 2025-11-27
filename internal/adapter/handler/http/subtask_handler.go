package http

import (
	"context"
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/grupoapi/proces-log/internal/domain/entity"
	subtaskUsecase "github.com/grupoapi/proces-log/internal/usecase/subtask"
)

// UpdateSubtaskUseCaseInterface define la interfaz para actualizar subtareas
type UpdateSubtaskUseCaseInterface interface {
	Execute(ctx context.Context, input subtaskUsecase.UpdateSubtaskInput) (*subtaskUsecase.UpdateSubtaskOutput, error)
}

// DeleteSubtaskUseCaseInterface define la interfaz para eliminar subtareas
type DeleteSubtaskUseCaseInterface interface {
	Execute(ctx context.Context, input subtaskUsecase.DeleteSubtaskInput) (*subtaskUsecase.DeleteSubtaskOutput, error)
}

// SubtaskHandler maneja las peticiones HTTP relacionadas con subtareas
type SubtaskHandler struct {
	updateUseCase UpdateSubtaskUseCaseInterface
	deleteUseCase DeleteSubtaskUseCaseInterface
}

// NewSubtaskHandler crea una nueva instancia de SubtaskHandler
func NewSubtaskHandler(
	updateUseCase UpdateSubtaskUseCaseInterface,
	deleteUseCase DeleteSubtaskUseCaseInterface,
) *SubtaskHandler {
	return &SubtaskHandler{
		updateUseCase: updateUseCase,
		deleteUseCase: deleteUseCase,
	}
}

// Update maneja PUT /Subtask/{uuid}
func (h *SubtaskHandler) Update(c *gin.Context) {
	uuidStr := c.Param("uuid")
	subtaskID, ok := parseUUIDOrError(c, uuidStr, entity.ErrSubtaskNotFound)
	if !ok {
		return
	}

	var req UpdateSubtaskRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		MapErrorToProblemDetails(c, entity.ErrMissingRequiredFields)
		return
	}

	// Validar nombre si se proporciona
	if req.Name != nil {
		if err := entity.ValidateName(*req.Name); err != nil {
			MapErrorToProblemDetails(c, err)
			return
		}
	}

	// Parsear estado si se proporciona
	state, ok := parseStateOrError(c, req.State)
	if !ok {
		return
	}

	input := subtaskUsecase.UpdateSubtaskInput{
		ID:        subtaskID,
		Name:      req.Name,
		State:     state,
		UpdatedBy: req.UpdatedBy,
	}

	output, err := h.updateUseCase.Execute(c.Request.Context(), input)
	if err != nil {
		MapErrorToProblemDetails(c, err)
		return
	}

	c.JSON(http.StatusOK, ToSubtaskResponse(output.Subtask))
}

// Delete maneja DELETE /Subtask/{uuid}
func (h *SubtaskHandler) Delete(c *gin.Context) {
	uuidStr := c.Param("uuid")
	subtaskID, ok := parseUUIDOrError(c, uuidStr, entity.ErrSubtaskNotFound)
	if !ok {
		return
	}

	var req DeleteSubtaskRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		MapErrorToProblemDetails(c, entity.ErrMissingRequiredFields)
		return
	}

	input := subtaskUsecase.DeleteSubtaskInput{
		ID:        subtaskID,
		DeletedBy: req.DeletedBy,
	}

	_, err := h.deleteUseCase.Execute(c.Request.Context(), input)
	if err != nil {
		MapErrorToProblemDetails(c, err)
		return
	}

	c.Status(http.StatusNoContent)
}
