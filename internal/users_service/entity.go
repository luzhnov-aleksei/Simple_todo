package users_service

// UserRequest - структура, представляющая тело запроса
type UserRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
}
