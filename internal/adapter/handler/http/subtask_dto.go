package http

// UpdateSubtaskRequest representa el request para actualizar una subtarea
type UpdateSubtaskRequest struct {
	Name      *string `json:"name,omitempty"`
	State     *string `json:"state,omitempty"`
	UpdatedBy string  `json:"updated_by" binding:"required"`
}

// DeleteSubtaskRequest representa el request para eliminar una subtarea
type DeleteSubtaskRequest struct {
	DeletedBy string `json:"deleted_by" binding:"required"`
}
