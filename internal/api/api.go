package api

import (
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"

	"simple-service/internal/api/middleware"
	"simple-service/internal/service"
)

// Routers - структура для хранения зависимостей роутов
type Routers struct {
	Service service.Service
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
	apiGroup.Post("/tasks", r.Service.CreateTask)

	// Роут для получения задачи по id
	apiGroup.Get("/task/:id", r.Service.GetTask)

	// Роут для получения всех задач
	apiGroup.Get("/tasks", r.Service.GetAllTasks)

	// Роут для обновления задачи
	apiGroup.Put("/task/:id", r.Service.UpdateTask)

	// Роут для удаления задачи
	apiGroup.Delete("/task/:id", r.Service.DeleteTask)

	return app
}
