package dummy

type LoginRequest struct {
	Role string `json:"role" binding:"required"`
}

type ErrorResponse struct {
	Message string `json:"message"`
}
