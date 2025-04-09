package tasks_service

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/jackc/pgx/v5"
	"github.com/pkg/errors"
	"io"
	"net/http"
	"simple-service/internal/api/middleware"
	"simple-service/internal/repo"
	"simple-service/internal/repo/mocks"
	"testing"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"go.uber.org/zap"
	"simple-service/internal/dto"
)

// TestCreateTask - тестирование метода CreateTask
func TestCreateTask(t *testing.T) {
	// Создаем мок репозитория

	mockRepo := mocks.NewRepository(t)
	logger := zap.NewNop().Sugar() // Без вывода логов
	// Создаем экземпляр сервиса с мок-репозиторием
	s := NewTaskService(mockRepo, logger)

	setupApp := func() *fiber.App {
		app := fiber.New()
		api := app.Group("/v1", middleware.Authorization("token"))
		api.Post("/task", s.CreateTask)
		return app
	}

	t.Run("успешное создание задачи", func(t *testing.T) {
		// Инициализируем Fiber-контекст
		app := setupApp()

		task := TaskRequest{
			UserID:      1,
			Title:       "Test Task",
			Description: "Test Description",
		}
		body, _ := json.Marshal(task)

		mockRepo.On("CheckUserExists", mock.Anything, task.UserID).Return(true, nil).Once()

		// Ожидаем, что вызов метода `CreateTask` в репозитории вернёт ID = 1
		mockRepo.On("CreateTask", mock.Anything, repo.Task{
			UserID:      task.UserID,
			Title:       task.Title,
			Description: task.Description,
		}).Return(1, nil).Once()

		// Отправляем запрос
		req, err := http.NewRequest("POST", "/v1/task", bytes.NewReader(body))
		assert.NoError(t, err)
		req.Header.Set("Content-Type", "application/json")

		// Выполняем запрос
		resp, err := app.Test(req)
		assert.NoError(t, err)
		assert.Equal(t, fiber.StatusOK, resp.StatusCode)

		// Проверяем ответ
		var response dto.Response
		errDecoder := json.NewDecoder(resp.Body).Decode(&response)
		assert.NoError(t, errDecoder)
		assert.Equal(t, "success", response.Status)
		// Проверяем вызов мок-методов
		mockRepo.AssertExpectations(t)
	})

	t.Run("ошибка валидации входных данных", func(t *testing.T) {
		// Инициализируем Fiber-контекст
		app := setupApp()

		body := []byte(`{}`) // Пустое тело, `title` обязателен

		req, err := http.NewRequest("POST", "/v1/task", bytes.NewReader(body))
		assert.NoError(t, err)
		req.Header.Set("Content-Type", "application/json")

		resp, err := app.Test(req)
		assert.NoError(t, err)
		assert.Equal(t, fiber.StatusBadRequest, resp.StatusCode)

		var response dto.Response
		json.NewDecoder(resp.Body).Decode(&response)
		assert.Equal(t, "error", response.Status)
		assert.Equal(t, "Field is required: TaskRequest.Title", response.Error.Desc)
	})

	t.Run("ID пользователя отсутствует в запросе", func(t *testing.T) {

		// Инициализируем Fiber-контекст
		app := setupApp()

		task := TaskRequest{
			Title:       "Test Task",
			Description: "Test Description",
		}
		body, _ := json.Marshal(task)

		mockRepo.On("CheckUserExists", mock.Anything, task.UserID).Return(false, nil).Once()

		req, err := http.NewRequest("POST", "/v1/task", bytes.NewReader(body))
		req.Header.Set("Content-Type", "application/json")

		resp, err := app.Test(req)
		assert.NoError(t, err)
		assert.NoError(t, err)
		assert.Equal(t, fiber.StatusNotFound, resp.StatusCode)

		var response dto.Response
		json.NewDecoder(resp.Body).Decode(&response)
		assert.Equal(t, "error", response.Status)
		assert.Equal(t, "Not found", response.Error.Desc)

		mockRepo.AssertExpectations(t)
	})
}

// TestGetTask - тестирование метода GetTask
func TestGetTask(t *testing.T) {
	// Создаем мок репозитория

	mockRepo := mocks.NewRepository(t)
	logger := zap.NewNop().Sugar() // Без вывода логов
	// Создаем экземпляр сервиса с мок-репозиторием
	s := NewTaskService(mockRepo, logger)

	setupApp := func() *fiber.App {
		app := fiber.New()
		api := app.Group("/v1", middleware.Authorization("token"))
		api.Get("/task/:id", s.GetTask)
		return app
	}

	t.Run("успешное создание задачи", func(t *testing.T) {
		// Инициализируем Fiber-контекст
		app := setupApp()

		// Отправляем запрос
		req, err := http.NewRequest("GET", "/v1/task/asd", nil)
		assert.NoError(t, err)

		// Выполняем запрос
		resp, err := app.Test(req)
		assert.NoError(t, err)
		assert.Equal(t, fiber.StatusBadRequest, resp.StatusCode)

		// Проверяем ответ
		var response dto.Response
		errDecoder := json.NewDecoder(resp.Body).Decode(&response)
		assert.NoError(t, errDecoder)
		assert.Equal(t, "error", response.Status)
		assert.Equal(t, "ID must be only number", response.Error.Desc)

		// Проверяем вызов мок-методов
		mockRepo.AssertExpectations(t)
	})

	t.Run("невалидный ID задачи", func(t *testing.T) {
		app := setupApp()

		req, err := http.NewRequest("GET", "/v1/task/abc", nil)
		assert.NoError(t, err)

		resp, err := app.Test(req)
		assert.NoError(t, err)
		assert.Equal(t, fiber.StatusBadRequest, resp.StatusCode)

		var response dto.Response
		json.NewDecoder(resp.Body).Decode(&response)
		assert.Equal(t, "error", response.Status)
		assert.Equal(t, "ID must be only number", response.Error.Desc)
	})

	t.Run("задача не существует", func(t *testing.T) {
		app := setupApp()

		taskID := 99

		mockRepo.On("CheckTaskExists", mock.Anything, taskID).Return(false, nil).Once()

		req, err := http.NewRequest("GET", fmt.Sprintf("/v1/task/%d", taskID), nil)
		assert.NoError(t, err)

		resp, err := app.Test(req)
		assert.NoError(t, err)
		assert.Equal(t, fiber.StatusNotFound, resp.StatusCode)

		var response dto.Response
		json.NewDecoder(resp.Body).Decode(&response)
		assert.Equal(t, "error", response.Status)
		assert.Equal(t, "Not found", response.Error.Desc)

		mockRepo.AssertExpectations(t)
	})

	t.Run("ошибка при получении задачи", func(t *testing.T) {
		app := setupApp()

		taskID := 5

		mockRepo.On("CheckTaskExists", mock.Anything, taskID).Return(true, nil).Once()
		mockRepo.On("GetTask", mock.Anything, taskID).Return((*repo.TaskView)(nil), pgx.ErrNoRows).Once()

		req, err := http.NewRequest("GET", fmt.Sprintf("/v1/task/%d", taskID), nil)
		assert.NoError(t, err)

		resp, err := app.Test(req)
		assert.NoError(t, err)
		assert.Equal(t, fiber.StatusInternalServerError, resp.StatusCode)

		var response dto.Response
		json.NewDecoder(resp.Body).Decode(&response)
		assert.Equal(t, "error", response.Status)

		mockRepo.AssertExpectations(t)
	})
}

func TestGetAllTasks(t *testing.T) {
	mockRepo := mocks.NewRepository(t)
	logger := zap.NewNop().Sugar()
	s := NewTaskService(mockRepo, logger)

	setupApp := func() *fiber.App {
		app := fiber.New()
		api := app.Group("/v1", middleware.Authorization("token"))
		api.Get("/tasks", s.GetAllTasks)
		return app
	}

	t.Run("успешное получение всех задач", func(t *testing.T) {
		app := setupApp()

		expectedTasks := &[]repo.TaskView{
			{ID: 1, UserID: 1, Title: "Task 1", Description: "Description 1"},
			{ID: 2, UserID: 2, Title: "Task 2", Description: "Description 2"},
		}

		mockRepo.On("GetAllTasks", mock.Anything).Return(expectedTasks, nil).Once()

		req, err := http.NewRequest("GET", "/v1/tasks", nil)
		assert.NoError(t, err)

		resp, err := app.Test(req)
		assert.NoError(t, err)
		assert.Equal(t, fiber.StatusOK, resp.StatusCode)

		var response dto.Response
		err = json.NewDecoder(resp.Body).Decode(&response)
		assert.NoError(t, err)
		assert.Equal(t, "success", response.Status)

		tasks, ok := response.Data.([]interface{})
		assert.True(t, ok)
		assert.Len(t, tasks, 2)

		mockRepo.AssertExpectations(t)
	})

	t.Run("ошибка при получении задач", func(t *testing.T) {
		app := setupApp()

		mockRepo.On("GetAllTasks", mock.Anything).Return(nil, errors.New("db error")).Once()

		req, err := http.NewRequest("GET", "/v1/tasks", nil)
		assert.NoError(t, err)

		resp, err := app.Test(req)
		assert.NoError(t, err)
		assert.Equal(t, fiber.StatusInternalServerError, resp.StatusCode)

		var response dto.Response
		err = json.NewDecoder(resp.Body).Decode(&response)
		assert.NoError(t, err)
		assert.Equal(t, "error", response.Status)

		mockRepo.AssertExpectations(t)
	})
}

func TestUpdateTask(t *testing.T) {
	mockRepo := mocks.NewRepository(t)
	logger := zap.NewNop().Sugar()
	s := NewTaskService(mockRepo, logger)

	setupApp := func() *fiber.App {
		app := fiber.New()
		api := app.Group("/v1", middleware.Authorization("token"))
		api.Put("/task/:id", s.UpdateTask)
		return app
	}

	t.Run("успешное обновление задачи", func(t *testing.T) {
		app := setupApp()

		taskID := 1
		reqBody := TaskUpdateRequest{
			Title:       "Updated Title",
			Description: "Updated Desc",
			Status:      "done",
		}
		body, _ := json.Marshal(reqBody)

		mockRepo.On("CheckTaskExists", mock.Anything, taskID).Return(true, nil).Once()
		mockRepo.On("UpdateTask", mock.Anything, repo.TaskUpdate{
			ID:          taskID,
			Title:       reqBody.Title,
			Description: reqBody.Description,
			Status:      reqBody.Status,
		}).Return(nil).Once()

		req, _ := http.NewRequest("PUT", "/v1/task/1", bytes.NewReader(body))
		req.Header.Set("Content-Type", "application/json")

		resp, err := app.Test(req)
		assert.NoError(t, err)
		assert.Equal(t, fiber.StatusOK, resp.StatusCode)

		var response dto.Response
		err = json.NewDecoder(resp.Body).Decode(&response)
		assert.NoError(t, err)
		assert.Equal(t, "success", response.Status)

		mockRepo.AssertExpectations(t)
	})

	t.Run("невалидный id", func(t *testing.T) {
		app := setupApp()

		reqBody := TaskUpdateRequest{
			Title:       "Updated Title",
			Description: "Updated Desc",
			Status:      "done",
		}
		body, _ := json.Marshal(reqBody)

		req, _ := http.NewRequest("PUT", "/v1/task/abc", bytes.NewReader(body))
		req.Header.Set("Content-Type", "application/json")

		resp, err := app.Test(req)
		assert.NoError(t, err)
		assert.Equal(t, fiber.StatusBadRequest, resp.StatusCode)

		var response dto.Response
		json.NewDecoder(resp.Body).Decode(&response)
		assert.Equal(t, "error", response.Status)
	})

	t.Run("ошибка в теле запроса", func(t *testing.T) {
		app := setupApp()

		body := []byte(`{invalid json}`)

		req, _ := http.NewRequest("PUT", "/v1/task/1", bytes.NewReader(body))
		req.Header.Set("Content-Type", "application/json")

		resp, err := app.Test(req)
		assert.NoError(t, err)
		assert.Equal(t, fiber.StatusBadRequest, resp.StatusCode)

		var response dto.Response
		json.NewDecoder(resp.Body).Decode(&response)
		assert.Equal(t, "error", response.Status)
	})

	t.Run("ошибка валидации", func(t *testing.T) {
		app := setupApp()

		body := []byte(`{}`) // отсутствует обязательное поле title

		req, _ := http.NewRequest("PUT", "/v1/task/1", bytes.NewReader(body))
		req.Header.Set("Content-Type", "application/json")

		resp, err := app.Test(req)
		assert.NoError(t, err)
		assert.Equal(t, fiber.StatusBadRequest, resp.StatusCode)

		var response dto.Response
		json.NewDecoder(resp.Body).Decode(&response)
		assert.Equal(t, "error", response.Status)
	})

	t.Run("ошибка при обновлении задачи", func(t *testing.T) {
		app := setupApp()

		taskID := 1
		reqBody := TaskUpdateRequest{
			Title:       "Title",
			Description: "Desc",
			Status:      "in_progress",
		}
		body, _ := json.Marshal(reqBody)

		mockRepo.On("CheckTaskExists", mock.Anything, taskID).Return(true, nil).Once()
		mockRepo.On("UpdateTask", mock.Anything, repo.TaskUpdate{
			ID:          taskID,
			Title:       reqBody.Title,
			Description: reqBody.Description,
			Status:      reqBody.Status,
		}).Return(errors.New("db error")).Once()

		req, _ := http.NewRequest("PUT", "/v1/task/1", bytes.NewReader(body))
		req.Header.Set("Content-Type", "application/json")

		resp, err := app.Test(req)
		assert.NoError(t, err)
		assert.Equal(t, fiber.StatusInternalServerError, resp.StatusCode)

		var response dto.Response
		json.NewDecoder(resp.Body).Decode(&response)
		assert.Equal(t, "error", response.Status)
		mockRepo.AssertExpectations(t)
	})
}

func TestDeleteTask(t *testing.T) {
	// Создаем мок репозитория
	mockRepo := mocks.NewRepository(t)
	logger := zap.NewNop().Sugar() // Без вывода логов
	// Создаем экземпляр сервиса с мок-репозиторием
	s := NewTaskService(mockRepo, logger)

	setupApp := func() *fiber.App {
		app := fiber.New()
		api := app.Group("/v1", middleware.Authorization("token"))
		api.Delete("/task/:id", s.DeleteTask)
		return app
	}

	t.Run("неверный формат ID", func(t *testing.T) {
		app := setupApp()

		// Отправляем запрос с некорректным ID (не число)
		req, err := http.NewRequest("DELETE", "/v1/task/abc", nil)
		assert.NoError(t, err)

		// Выполняем запрос
		resp, err := app.Test(req)
		assert.NoError(t, err)
		assert.Equal(t, fiber.StatusBadRequest, resp.StatusCode)

		// Проверяем ответ
		var response dto.Response
		errDecoder := json.NewDecoder(resp.Body).Decode(&response)
		assert.NoError(t, errDecoder)
		assert.Equal(t, "error", response.Status)

		// Проверяем вызов мок-методов
		mockRepo.AssertExpectations(t)
	})

	t.Run("задача не найдена", func(t *testing.T) {
		app := setupApp()

		taskID := 1
		// Настроим мок для метода CheckTaskExists
		mockRepo.On("CheckTaskExists", mock.Anything, taskID).Return(false, nil).Once()

		// Отправляем запрос
		req, err := http.NewRequest("DELETE", fmt.Sprintf("/v1/task/%d", taskID), nil)
		assert.NoError(t, err)

		// Выполняем запрос
		resp, err := app.Test(req)
		assert.NoError(t, err)

		// Проверяем, что статус 404 (Not Found)
		assert.Equal(t, fiber.StatusNotFound, resp.StatusCode)

		// Проверяем ответ
		var response dto.Response
		json.NewDecoder(resp.Body).Decode(&response)
		assert.Equal(t, "error", response.Status)

		// Проверяем вызов мок-методов
		mockRepo.AssertExpectations(t)
	})

	t.Run("ошибка в репозитории", func(t *testing.T) {
		app := setupApp()

		taskID := 1
		// Настроим мок для метода CheckTaskExists
		mockRepo.On("CheckTaskExists", mock.Anything, taskID).Return(true, nil).Once()
		mockRepo.On("DeleteTask", mock.Anything, taskID).Return(pgx.ErrNoRows).Once()

		// Отправляем запрос
		req, err := http.NewRequest("DELETE", fmt.Sprintf("/v1/task/%d", taskID), nil)
		assert.NoError(t, err)

		// Выполняем запрос
		resp, err := app.Test(req)
		assert.NoError(t, err)

		// Проверяем, что статус 404 (Not Found)
		assert.Equal(t, fiber.StatusBadRequest, resp.StatusCode)

		// Проверяем ответ
		var response dto.Response
		json.NewDecoder(resp.Body).Decode(&response)
		assert.Equal(t, "error", response.Status)

		// Проверяем вызов мок-методов
		mockRepo.AssertExpectations(t)
	})

	t.Run("успешное удаление задачи", func(t *testing.T) {
		app := setupApp()

		taskID := 1
		// Настроим мок для метода CheckTaskExists
		mockRepo.On("CheckTaskExists", mock.Anything, taskID).Return(true, nil).Once()
		mockRepo.On("DeleteTask", mock.Anything, taskID).Return(nil).Once()

		// Отправляем запрос
		req, err := http.NewRequest("DELETE", fmt.Sprintf("/v1/task/%d", taskID), nil)
		assert.NoError(t, err)

		// Выполняем запрос
		resp, err := app.Test(req)
		assert.NoError(t, err)

		// Проверяем, что статус 200 (OK)
		assert.Equal(t, fiber.StatusOK, resp.StatusCode)

		// Проверяем ответ
		var response dto.Response
		json.NewDecoder(resp.Body).Decode(&response)
		assert.Equal(t, "success", response.Status)

		// Проверяем вызов мок-методов
		mockRepo.AssertExpectations(t)
	})
}

func TestGetAllTasksFromUser(t *testing.T) {
	mockRepo := mocks.NewRepository(t)
	logger := zap.NewNop().Sugar()
	s := NewTaskService(mockRepo, logger)

	setupApp := func() *fiber.App {
		app := fiber.New()
		api := app.Group("/v1", middleware.Authorization("token"))
		api.Get("/tasks", s.GetAllTasksFromUser)
		return app
	}

	t.Run("успешное получение задач", func(t *testing.T) {
		app := setupApp()

		username := "user123"
		expectedTasks := []repo.TaskView{
			{ID: 1, UserID: 1, Title: "Task 1", Description: "Description 1", Status: "new", CreatedAt: time.Now()},
			{ID: 2, UserID: 1, Title: "Task 2", Description: "Description 2", Status: "in progress", CreatedAt: time.Now()},
		}

		mockRepo.On("GetAllTaskFromUser", mock.Anything, username).Return(&expectedTasks, nil).Once()

		req, err := http.NewRequest("GET", fmt.Sprintf("/v1/tasks?username=%s", username), nil)
		assert.NoError(t, err)

		resp, err := app.Test(req)
		assert.NoError(t, err)
		defer resp.Body.Close()
		assert.Equal(t, fiber.StatusOK, resp.StatusCode)

		// Читаем тело ответа
		bodyBytes, err := io.ReadAll(resp.Body)
		assert.NoError(t, err)

		// Декодируем в структуру с правильным типом данных
		var result struct {
			Status string          `json:"status"`
			Data   []repo.TaskView `json:"data"`
		}
		err = json.Unmarshal(bodyBytes, &result)
		assert.NoError(t, err)

		// Проверяем статус
		assert.Equal(t, "success", result.Status)

		// Проверяем данные
		assert.Equal(t, len(expectedTasks), len(result.Data))

		// Дополнительные проверки полей
		if len(result.Data) > 0 {
			assert.Equal(t, expectedTasks[0].Title, result.Data[0].Title)
			assert.Equal(t, expectedTasks[0].Status, result.Data[0].Status)
		}

		mockRepo.AssertExpectations(t)
	})

	t.Run("ошибка при получении задач", func(t *testing.T) {
		app := setupApp()

		username := "user123"
		mockRepo.On("GetAllTaskFromUser", mock.Anything, username).Return((*[]repo.TaskView)(nil), fmt.Errorf("error")).Once()

		req, err := http.NewRequest("GET", fmt.Sprintf("/v1/tasks?username=%s", username), nil)
		assert.NoError(t, err)

		resp, err := app.Test(req)
		assert.NoError(t, err)
		defer resp.Body.Close()
		assert.Equal(t, fiber.StatusInternalServerError, resp.StatusCode)

		var response dto.Response
		err = json.NewDecoder(resp.Body).Decode(&response)
		assert.NoError(t, err)
		assert.Equal(t, "error", response.Status)

		mockRepo.AssertExpectations(t)
	})
}
