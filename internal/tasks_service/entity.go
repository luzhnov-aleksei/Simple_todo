package tasks_service

// TaskRequest - структура, представляющая тело запроса
type TaskRequest struct {
	UserID      int    `json:"user_id"`
	Title       string `json:"title" validate:"required"`
	Description string `json:"description"`
}
