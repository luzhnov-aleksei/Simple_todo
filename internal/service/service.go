package service

import (
	"encoding/json"
	"github.com/gofiber/fiber/v2"
	"go.uber.org/zap"
	"simple-service/internal/dto"
	"simple-service/internal/repo"
	"simple-service/pkg/validator"
	"strconv"
)

// Слой бизнес-логики. Тут должна быть основная логика сервиса

// Service - интерфейс для бизнес-логики
type Service interface {
	GetTask(ctx *fiber.Ctx) error
	CreateTask(ctx *fiber.Ctx) error
}

type service struct {
	repo repo.Repository
	log  *zap.SugaredLogger
}

// NewService - конструктор сервиса
func NewService(repo repo.Repository, logger *zap.SugaredLogger) Service {
	return &service{
		repo: repo,
		log:  logger,
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

	// Вставка задачи в БД через репозиторий
	task := repo.Task{
		Title:       req.Title,
		Description: req.Description,
	}
	taskID, err := s.repo.CreateTask(ctx.Context(), task)
	if err != nil {
		s.log.Error("Failed to insert task", zap.Error(err))
		return dto.InternalServerError(ctx)
	}

	// Формирование ответа
	response := dto.Response{
		Status: "success",
		Data:   map[string]int{"task_id": taskID},
	}

	return ctx.Status(fiber.StatusOK).JSON(response)
}

func (s *service) GetTask(ctx *fiber.Ctx) error {
	taskID, err := strconv.Atoi(ctx.Params("id"))
	if err != nil {
		s.log.Error("Failed to parse int", zap.Error(err))
		return dto.BadResponseError(ctx, dto.FieldBadFormat, "ID must be only number")
	}
	task, err := s.repo.GetTask(ctx.Context(), taskID)
	if err != nil {
		s.log.Error("Failed to get task", zap.Error(err))
		return dto.InternalServerError(ctx)
	}
	// Формирование ответа
	response := dto.Response{
		Status: "success",
		Data:   task,
	}

	return ctx.Status(fiber.StatusOK).JSON(response)
}
