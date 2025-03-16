package service

// Task - структура, соответствующая таблице tasks
type Task struct {
	Title       string `json:"title"`
	Description string `json:"description"`
	Status      string `json:"status"`
}
