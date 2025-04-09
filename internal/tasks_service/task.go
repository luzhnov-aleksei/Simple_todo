package tasks_service

// Task - структура, соответствующая таблице tasks
type Task struct {
	ID          int    `json:"id"`
	Title       string `json:"title"`
	Description string `json:"description"`
	Status      string `json:"status"`
}

type TaskUpdateRequest struct {
	Title       string `json:"title" validate:"required,min=3"`
	Description string `json:"description"`
	Status      string `json:"status" validate:"required,oneof=todo in_progress done"`
}
