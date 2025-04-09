package users_service

import "time"

// User - структура, соответствующая таблице user
type User struct {
	ID        int       `json:"id"`
	Username  string    `json:"username"`
	Password  string    `json:"password"`
	CreatedAt time.Time `json:"created_at"`
}
