package tasks_service

import (
	"bytes"
	"encoding/json"
	"net/http"
	"testing"

	"github.com/gofiber/fiber/v2"
	"github.com/stretchr/testify/assert"
	"go.uber.org/zap"
	"simple-service/internal/dto"
)

// TestCreateTask - тестирование метода CreateTask
func TestCreateTask(t *testing.T) {
	// Создаем мок репозитория
	logger := zap.NewNop().Sugar() // Без вывода логов

	// Создаем экземпляр сервиса
	s := NewService(logger)

	// Инициализируем Fiber-контекст
	app := fiber.New()
	app.Post("/tasks", s.CreateTask)

	t.Run("успешное создание задачи", func(t *testing.T) {
		task := TaskRequest{
			Title:       "Test Task",
			Description: "Test Description",
		}
		body, _ := json.Marshal(task)

		// Отправляем запрос
		req, err := http.NewRequest("POST", "/tasks", bytes.NewReader(body))
		assert.NoError(t, err)
		req.Header.Set("Content-Type", "application/json")

		// Выполняем запрос
		resp, err := app.Test(req)
		assert.NoError(t, err)
		assert.Equal(t, fiber.StatusOK, resp.StatusCode)

		// Проверяем ответ
		var response dto.Response
		json.NewDecoder(resp.Body).Decode(&response)
		assert.Equal(t, "success", response.Status)

	})

	t.Run("ошибка валидации входных данных", func(t *testing.T) {
		body := []byte(`{}`) // Пустое тело, `title` обязателен

		req, err := http.NewRequest("POST", "/tasks", bytes.NewReader(body))
		assert.NoError(t, err)
		req.Header.Set("Content-Type", "application/json")

		resp, err := app.Test(req)
		assert.NoError(t, err)
		assert.Equal(t, fiber.StatusBadRequest, resp.StatusCode)

		var response dto.Response
		json.NewDecoder(resp.Body).Decode(&response)
		assert.Equal(t, "error", response.Status)
	})

}
