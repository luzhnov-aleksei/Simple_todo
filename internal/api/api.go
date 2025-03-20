package api

import (
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"

	"simple-service/internal/api/middleware"
	"simple-service/internal/tasks_service"
	"simple-service/internal/users_service"
)

// Routers - структура для хранения зависимостей роутов
type Routers struct {
	TasksService tasks_service.TasksService
	UsersService users_service.UsersService
}

// NewRouters - конструктор для настройки API
func NewRouters(r *Routers, token string) *fiber.App {
	app := fiber.New()

	// Настройка CORS (разрешенные методы, заголовки, авторизация)
	app.Use(cors.New(cors.Config{
		AllowMethods:  "GET, POST, PUT, DELETE",
		AllowHeaders:  "Accept, Authorization, Content-Type, X-CSRF-Token, X-REQUEST-ID",
		ExposeHeaders: "Link",
		MaxAge:        300,
	}))

	// Группа маршрутов с авторизацией
	apiGroup := app.Group("/v1", middleware.Authorization(token))

	// Роут для создания задачи
	apiGroup.Post("/task", r.TasksService.CreateTask)

	// Роут для получения задачи по id
	apiGroup.Get("/task/:id", r.TasksService.GetTask)

	// Роут для получения всех задач
	apiGroup.Get("/tasks", r.TasksService.GetAllTasks)

	// Роут для обновления задачи
	apiGroup.Put("/task/:id", r.TasksService.UpdateTask)

	// Роут для удаления задачи
	apiGroup.Delete("/task/:id", r.TasksService.DeleteTask)

	//--------------------------------------------------------
	// Роут для создания пользователя
	apiGroup.Post("/user", r.UsersService.CreateUser)

	// Роут для получения пользователя по id
	apiGroup.Get("/user/:id", r.UsersService.GetUser)

	// Роут для получения всех пользователей
	apiGroup.Get("/users", r.UsersService.GetAllUsers)

	// Роут для обновления данных пользователя
	apiGroup.Put("/user/:id", r.UsersService.UpdateUser)

	// Роут для удаления пользователя
	apiGroup.Delete("/user/:id", r.UsersService.DeleteUser)

	return app
}
