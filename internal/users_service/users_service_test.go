package users_service

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/jackc/pgx/v5"
	"github.com/pkg/errors"
	"net/http"
	"simple-service/internal/api/middleware"
	"simple-service/internal/repo/mocks"
	"testing"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"go.uber.org/zap"
	"simple-service/internal/dto"
	"simple-service/internal/repo"
)

func TestCreateUser(t *testing.T) {
	// Создаем мок репозитория
	mockRepo := mocks.NewRepository(t)
	logger := zap.NewNop().Sugar() // Без вывода логов
	// Создаем экземпляр сервиса с мок-репозиторием
	s := NewService(mockRepo, logger)

	setupApp := func() *fiber.App {
		app := fiber.New()
		api := app.Group("/v1", middleware.Authorization("token"))
		api.Post("/user", s.CreateUser)
		return app
	}

	t.Run("успешное создание пользователя", func(t *testing.T) {
		// Инициализируем Fiber-контекст
		app := setupApp()

		user := UserRequest{
			Username: "testuser",
			Password: "password123",
		}
		body, _ := json.Marshal(user)

		// Ожидаем, что вызов метода `CreateUser` в репозитории вернёт ID = 1
		mockRepo.On("CreateUser", mock.Anything, repo.UserCreate{
			Username: user.Username,
			Password: user.Password,
		}).Return(1, nil).Once()

		// Отправляем запрос
		req, err := http.NewRequest("POST", "/v1/user", bytes.NewReader(body))
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

	t.Run("ошибка при создании пользователя в репозитории", func(t *testing.T) {
		// Инициализируем Fiber-контекст
		app := setupApp()

		user := UserRequest{
			Username: "testuser",
			Password: "password123",
		}
		body, _ := json.Marshal(user)

		// Мокируем ошибку при создании пользователя
		mockRepo.On("CreateUser", mock.Anything, repo.UserCreate{
			Username: user.Username,
			Password: user.Password,
		}).Return(0, errors.New("internal error")).Once()

		// Отправляем запрос
		req, err := http.NewRequest("POST", "/v1/user", bytes.NewReader(body))
		assert.NoError(t, err)
		req.Header.Set("Content-Type", "application/json")

		// Выполняем запрос
		resp, err := app.Test(req)
		assert.NoError(t, err)
		assert.Equal(t, fiber.StatusInternalServerError, resp.StatusCode)

		var response dto.Response
		json.NewDecoder(resp.Body).Decode(&response)
		assert.Equal(t, "error", response.Status)

		// Проверяем вызов мок-методов
		mockRepo.AssertExpectations(t)
	})
}

func TestGetUser(t *testing.T) {
	// Создаем мок репозитория
	mockRepo := mocks.NewRepository(t)
	logger := zap.NewNop().Sugar() // Без вывода логов
	// Создаем экземпляр сервиса с мок-репозиторием
	s := NewService(mockRepo, logger)

	setupApp := func() *fiber.App {
		app := fiber.New()
		api := app.Group("/v1", middleware.Authorization("token"))
		api.Get("/user/:id", s.GetUser)
		return app
	}

	t.Run("успешное получение пользователя", func(t *testing.T) {
		// Инициализируем Fiber-контекст
		app := setupApp()

		userID := 1
		// Возвращаем указатель на UserView
		mockRepo.On("GetUser", mock.Anything, userID).Return(&repo.UserView{
			ID:        userID,
			Username:  "testuser",
			Password:  "password123",
			CreatedAt: time.Now(),
		}, nil).Once()

		// Отправляем запрос
		req, err := http.NewRequest("GET", "/v1/user/1", nil)
		assert.NoError(t, err)

		// Выполняем запрос
		resp, err := app.Test(req)
		assert.NoError(t, err)
		assert.Equal(t, fiber.StatusOK, resp.StatusCode)

		// Проверяем ответ
		var response dto.Response
		errDecoder := json.NewDecoder(resp.Body).Decode(&response)
		assert.NoError(t, errDecoder)

		// Проверяем, что ответ содержит правильные данные
		assert.Equal(t, "success", response.Status)
		assert.NotNil(t, response.Data)
		assert.Equal(t, float64(userID), response.Data.(map[string]interface{})["id"]) // Здесь мы ожидаем, что в данных будет id
		assert.Equal(t, "testuser", response.Data.(map[string]interface{})["username"])

		// Проверяем вызов мок-методов
		mockRepo.AssertExpectations(t)
	})

	t.Run("ошибка при разборе ID", func(t *testing.T) {
		// Инициализируем Fiber-контекст
		app := setupApp()

		// Отправляем запрос с некорректным ID
		req, err := http.NewRequest("GET", "/v1/user/invalid_id", nil)
		assert.NoError(t, err)

		// Выполняем запрос
		resp, err := app.Test(req)
		assert.NoError(t, err)
		assert.Equal(t, fiber.StatusBadRequest, resp.StatusCode)

		// Проверяем ответ
		var response dto.Response
		json.NewDecoder(resp.Body).Decode(&response)
		assert.Equal(t, "error", response.Status)
	})

	t.Run("пользователь не найден", func(t *testing.T) {
		// Инициализируем Fiber-контекст
		app := setupApp()

		userID := 1
		mockRepo.On("GetUser", mock.Anything, userID).Return(&repo.UserView{}, pgx.ErrNoRows).Once()

		// Отправляем запрос
		req, err := http.NewRequest("GET", "/v1/user/1", nil)
		assert.NoError(t, err)

		// Выполняем запрос
		resp, err := app.Test(req)
		assert.NoError(t, err)
		assert.Equal(t, fiber.StatusInternalServerError, resp.StatusCode)

		// Проверяем ответ
		var response dto.Response
		json.NewDecoder(resp.Body).Decode(&response)
		assert.Equal(t, "error", response.Status)

		// Проверяем вызов мок-методов
		mockRepo.AssertExpectations(t)
	})

	t.Run("внутренняя ошибка при получении пользователя", func(t *testing.T) {
		// Инициализируем Fiber-контекст
		app := setupApp()

		userID := 1
		// Возвращаем указатель на пустой UserView и имитируем ошибку
		mockRepo.On("GetUser", mock.Anything, userID).Return(nil, errors.New("internal error")).Once()

		// Отправляем запрос
		req, err := http.NewRequest("GET", "/v1/user/1", nil)
		assert.NoError(t, err)

		// Выполняем запрос
		resp, err := app.Test(req)
		assert.NoError(t, err)
		assert.Equal(t, fiber.StatusInternalServerError, resp.StatusCode)

		// Проверяем ответ
		var response dto.Response
		json.NewDecoder(resp.Body).Decode(&response)
		assert.Equal(t, "error", response.Status)

		// Проверяем вызов мок-методов
		mockRepo.AssertExpectations(t)
	})

}

func TestGetAllUsers(t *testing.T) {
	mockRepo := mocks.NewRepository(t)
	logger := zap.NewNop().Sugar()
	s := NewService(mockRepo, logger)

	setupApp := func() *fiber.App {
		app := fiber.New()
		api := app.Group("/v1", middleware.Authorization("token"))
		api.Get("/users", s.GetAllUsers)
		return app
	}

	t.Run("успешное получение всех пользователей", func(t *testing.T) {
		app := setupApp()

		expectedUsers := &[]repo.UserView{
			{ID: 1, Username: "user1", Password: "password1", CreatedAt: time.Now()},
			{ID: 2, Username: "user2", Password: "password2", CreatedAt: time.Now()},
		}

		mockRepo.On("GetAllUsers", mock.Anything).Return(expectedUsers, nil).Once()

		req, err := http.NewRequest("GET", "/v1/users", nil)
		assert.NoError(t, err)

		resp, err := app.Test(req)
		assert.NoError(t, err)
		assert.Equal(t, fiber.StatusOK, resp.StatusCode)

		var response dto.Response
		err = json.NewDecoder(resp.Body).Decode(&response)
		assert.NoError(t, err)
		assert.Equal(t, "success", response.Status)

		users, ok := response.Data.([]interface{})
		assert.True(t, ok)
		assert.Len(t, users, 2)

		mockRepo.AssertExpectations(t)
	})

	t.Run("ошибка при получении пользователей", func(t *testing.T) {
		app := setupApp()

		mockRepo.On("GetAllUsers", mock.Anything).Return(nil, errors.New("db error")).Once()

		req, err := http.NewRequest("GET", "/v1/users", nil)
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

func TestUpdateUser(t *testing.T) {
	mockRepo := mocks.NewRepository(t)
	logger := zap.NewNop().Sugar()
	s := NewService(mockRepo, logger)

	setupApp := func() *fiber.App {
		app := fiber.New()
		api := app.Group("/v1", middleware.Authorization("token"))
		api.Put("/user/:id", s.UpdateUser)
		return app
	}

	t.Run("успешное обновление пользователя", func(t *testing.T) {
		app := setupApp()

		userID := 1
		reqBody := UserRequest{
			Username: "UpdatedUsername",
			Password: "UpdatedPassword",
		}
		body, _ := json.Marshal(reqBody)

		mockRepo.On("UpdateUser", mock.Anything, repo.UserUpdate{
			ID:       userID,
			Username: reqBody.Username,
			Password: reqBody.Password,
		}).Return(nil).Once()

		req, _ := http.NewRequest("PUT", "/v1/user/1", bytes.NewReader(body))
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

		reqBody := UserRequest{
			Username: "UpdatedUsername",
			Password: "UpdatedPassword",
		}
		body, _ := json.Marshal(reqBody)

		req, _ := http.NewRequest("PUT", "/v1/user/abc", bytes.NewReader(body))
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

		req, _ := http.NewRequest("PUT", "/v1/user/1", bytes.NewReader(body))
		req.Header.Set("Content-Type", "application/json")

		resp, err := app.Test(req)
		assert.NoError(t, err)
		assert.Equal(t, fiber.StatusBadRequest, resp.StatusCode)

		var response dto.Response
		json.NewDecoder(resp.Body).Decode(&response)
		assert.Equal(t, "error", response.Status)
	})

	t.Run("ошибка_валидации", func(t *testing.T) {
		app := setupApp()

		// Подготавливаем некорректное тело запроса (например, пустые поля)
		reqBody := UserRequest{
			Username: "", // Пустое имя пользователя
			Password: "", // Пустой пароль
		}
		body, _ := json.Marshal(reqBody)

		// Не ожидаем вызов метода UpdateUser на репозитории, так как ошибка валидации должна происходить раньше
		mockRepo.On("UpdateUser", mock.Anything, mock.Anything).Return(errors.New("unexpected call")).Once()

		req, _ := http.NewRequest("PUT", "/v1/user/1", bytes.NewReader(body))
		req.Header.Set("Content-Type", "application/json")

		resp, err := app.Test(req)
		assert.NoError(t, err)
		assert.Equal(t, fiber.StatusInternalServerError, resp.StatusCode)

		var response dto.Response
		err = json.NewDecoder(resp.Body).Decode(&response)
		assert.NoError(t, err)
		assert.Equal(t, "error", response.Status)

		mockRepo.AssertExpectations(t)
	})

	t.Run("ошибка при обновлении пользователя", func(t *testing.T) {
		app := setupApp()

		userID := 1
		reqBody := UserRequest{
			Username: "UpdatedUsername",
			Password: "UpdatedPassword",
		}
		body, _ := json.Marshal(reqBody)

		mockRepo.On("UpdateUser", mock.Anything, repo.UserUpdate{
			ID:       userID,
			Username: reqBody.Username,
			Password: reqBody.Password,
		}).Return(errors.New("db error")).Once()

		req, _ := http.NewRequest("PUT", "/v1/user/1", bytes.NewReader(body))
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

func TestDeleteUser(t *testing.T) {
	// Создаем мок репозитория
	mockRepo := mocks.NewRepository(t)
	logger := zap.NewNop().Sugar() // Без вывода логов
	// Создаем экземпляр сервиса с мок-репозиторием
	s := NewService(mockRepo, logger)

	setupApp := func() *fiber.App {
		app := fiber.New()
		api := app.Group("/v1", middleware.Authorization("token"))
		api.Delete("/user/:id", s.DeleteUser)
		return app
	}

	t.Run("неверный формат ID", func(t *testing.T) {
		app := setupApp()

		// Отправляем запрос с некорректным ID (не число)
		req, err := http.NewRequest("DELETE", "/v1/user/abc", nil)
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

	t.Run("пользователь не найден", func(t *testing.T) {
		app := setupApp()

		userID := 1
		// Настроим мок для метода DeleteUser
		mockRepo.On("DeleteUser", mock.Anything, userID).Return(pgx.ErrNoRows).Once()

		// Отправляем запрос
		req, err := http.NewRequest("DELETE", fmt.Sprintf("/v1/user/%d", userID), nil)
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

	t.Run("ошибка в репозитории", func(t *testing.T) {
		app := setupApp()

		userID := 1
		// Настроим мок для метода DeleteUser
		mockRepo.On("DeleteUser", mock.Anything, userID).Return(errors.New("database error")).Once()

		// Отправляем запрос
		req, err := http.NewRequest("DELETE", fmt.Sprintf("/v1/user/%d", userID), nil)
		assert.NoError(t, err)

		// Выполняем запрос
		resp, err := app.Test(req)
		assert.NoError(t, err)

		// Проверяем, что статус 400 (Bad Request)
		assert.Equal(t, fiber.StatusBadRequest, resp.StatusCode)

		// Проверяем ответ
		var response dto.Response
		json.NewDecoder(resp.Body).Decode(&response)
		assert.Equal(t, "error", response.Status)

		// Проверяем вызов мок-методов
		mockRepo.AssertExpectations(t)
	})

	t.Run("успешное удаление пользователя", func(t *testing.T) {
		app := setupApp()

		userID := 1
		// Настроим мок для метода DeleteUser
		mockRepo.On("DeleteUser", mock.Anything, userID).Return(nil).Once()

		// Отправляем запрос
		req, err := http.NewRequest("DELETE", fmt.Sprintf("/v1/user/%d", userID), nil)
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
