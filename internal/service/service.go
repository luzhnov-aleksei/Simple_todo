package service

import (
	"encoding/json"
	"github.com/gofiber/fiber/v2"
	"go.uber.org/zap"
	"simple-service/internal/dto"
	"simple-service/pkg/validator"
	"strconv"
)

// Создание мапы для in-memory хранения
var tasks = make(map[int]Task)
var id = 1

// Слой бизнес-логики. Тут должна быть основная логика сервиса

// Service - интерфейс для бизнес-логики
type Service interface {
	GetTask(ctx *fiber.Ctx) error
	CreateTask(ctx *fiber.Ctx) error
}

type service struct {
	log *zap.SugaredLogger
}

// NewService - конструктор сервиса
func NewService(logger *zap.SugaredLogger) Service {
	return &service{
		log: logger,
	}
}

// CreateTask - обработчик запроса на создание задачи
func (s *service) CreateTask(ctx *fiber.Ctx) error {
	var req TaskRequest

	// Десериализация JSON-запроса
	if err := json.Unmarshal(ctx.Body(), &req); err != nil {
		s.log.Error("Invalid request body", zap.Error(err))
		return dto.BadResponseError(ctx, dto.FieldBadFormat, "Invalid request body")
	}

	// Валидация входных данных
	if vErr := validator.Validate(ctx.Context(), req); vErr != nil {
		return dto.BadResponseError(ctx, dto.FieldIncorrect, vErr.Error())
	}

	tasks[id] = Task{
		Title:       req.Title,
		Description: req.Description,
		Status:      "new",
	}

	// Формирование ответа
	response := dto.Response{
		Status: "success",
		Data:   map[string]int{"task_id": id},
	}
	id++

	return ctx.Status(fiber.StatusOK).JSON(response)
}

func (s *service) GetTask(ctx *fiber.Ctx) error {
	taskID, err := strconv.Atoi(ctx.Params("id"))
	if err != nil {
		s.log.Error("Failed to parse int", zap.Error(err))
		return dto.BadResponseError(ctx, dto.FieldBadFormat, "ID must be only number")
	}
	value, exists := tasks[taskID]
	if exists != true {
		s.log.Error("Failed to get task", zap.Error(err))
		return dto.InternalServerError(ctx)
	}
	// Формирование ответа
	response := dto.Response{
		Status: "success",
		Data:   value,
	}

	return ctx.Status(fiber.StatusOK).JSON(response)
}
