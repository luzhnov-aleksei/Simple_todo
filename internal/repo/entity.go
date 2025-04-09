package repo

import "time"

// User - структура, соответствующая таблице users
type UserCreate struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

type UserView struct {
	ID        int       `json:"id"`
	Username  string    `json:"username"`
	Password  string    `json:"password"`
	CreatedAt time.Time `json:"created_at"`
}

type UserUpdate struct {
	ID       int    `json:"id"`
	Username string `json:"username"`
	Password string `json:"password"`
}

// Task - структура, соответствующая таблице tasks
type Task struct {
	UserID      int    `json:"user_id"`
	Title       string `json:"title"`
	Description string `json:"description"`
}

type TaskUpdate struct {
	ID          int    `json:"id"`
	Title       string `json:"title"`
	Description string `json:"description"`
	Status      string `json:"status"`
}

type TaskView struct {
	ID          int       `json:"id"`
	UserID      int       `json:"user_id"`
	Title       string    `json:"title"`
	Description string    `json:"description"`
	Status      string    `json:"status"`
	CreatedAt   time.Time `json:"created_at"`
}
