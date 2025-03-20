package tasks_service

import (
	"encoding/json"
	"github.com/gofiber/fiber/v2"
	"go.uber.org/zap"
	"simple-service/internal/dto"
	"simple-service/internal/repo"
	"simple-service/pkg/validator"
	"strconv"
	"sync"
)

var (
	tasks = make(map[int]Task)
	id    = 1
	mu    sync.RWMutex
)

type TasksService interface {
	GetTask(ctx *fiber.Ctx) error
	GetAllTasks(ctx *fiber.Ctx) error
	CreateTask(ctx *fiber.Ctx) error
	UpdateTask(ctx *fiber.Ctx) error
	DeleteTask(ctx *fiber.Ctx) error
}

func NewService(repo repo.Repository, logger *zap.SugaredLogger) TasksService {
	return &service{
		repo: repo,
		log:  logger,
	}
}

type service struct {
	repo repo.Repository
	log  *zap.SugaredLogger
}

func (s *service) CreateTask(ctx *fiber.Ctx) error {
	var req TaskRequest

	if err := json.Unmarshal(ctx.Body(), &req); err != nil {
		s.log.Error("Invalid request body", zap.Error(err))
		return dto.BadResponseError(ctx, dto.FieldBadFormat, "Invalid request body")
	}

	if vErr := validator.Validate(ctx.Context(), req); vErr != nil {
		return dto.BadResponseError(ctx, dto.FieldIncorrect, vErr.Error())
	}

	mu.Lock()
	tasks[id] = Task{
		ID:          id,
		Title:       req.Title,
		Description: req.Description,
		Status:      "new",
	}
	id++
	mu.Unlock()

	response := dto.Response{
		Status: "success",
		Data:   map[string]int{"task_id": id - 1},
	}

	return ctx.Status(fiber.StatusOK).JSON(response)
}

func (s *service) GetTask(ctx *fiber.Ctx) error {
	taskID, err := strconv.Atoi(ctx.Params("id"))
	if err != nil {
		s.log.Error("Failed to parse int", zap.Error(err))
		return dto.BadResponseError(ctx, dto.FieldBadFormat, "ID must be only number")
	}

	mu.RLock()
	value, exists := tasks[taskID]
	mu.RUnlock()

	if !exists {
		s.log.Error("Failed to get task", zap.Error(err))
		return dto.NotFoundError(ctx)
	}

	response := dto.Response{
		Status: "success",
		Data:   value,
	}

	return ctx.Status(fiber.StatusOK).JSON(response)
}

func (s *service) GetAllTasks(ctx *fiber.Ctx) error {
	mu.RLock()
	allTasks := make([]Task, 0, len(tasks))
	for _, task := range tasks {
		allTasks = append(allTasks, task)
	}
	mu.RUnlock()

	response := dto.Response{
		Status: "success",
		Data:   allTasks,
	}

	return ctx.Status(fiber.StatusOK).JSON(response)
}

func (s *service) UpdateTask(ctx *fiber.Ctx) error {
	taskID, err := strconv.Atoi(ctx.Params("id"))
	if err != nil {
		s.log.Error("Failed to parse int", zap.Error(err))
		return dto.BadResponseError(ctx, dto.FieldBadFormat, "ID must be only number")
	}

	mu.Lock()
	_, exists := tasks[taskID]
	if !exists {
		mu.Unlock()
		s.log.Error("Failed to update task", zap.Error(err))
		return dto.NotFoundError(ctx)
	}

	var req TaskRequest
	if err := json.Unmarshal(ctx.Body(), &req); err != nil {
		mu.Unlock()
		s.log.Error("Invalid request body", zap.Error(err))
		return dto.BadResponseError(ctx, dto.FieldBadFormat, "Invalid request body")
	}

	if vErr := validator.Validate(ctx.Context(), req); vErr != nil {
		mu.Unlock()
		return dto.BadResponseError(ctx, dto.FieldIncorrect, vErr.Error())
	}

	tasks[taskID] = Task{
		ID:          taskID,
		Title:       req.Title,
		Description: req.Description,
		Status:      "done",
	}
	mu.Unlock()

	response := dto.Response{
		Status: "success",
		Data:   map[string]int{"task_id": taskID},
	}
	return ctx.Status(fiber.StatusOK).JSON(response)
}

func (s *service) DeleteTask(ctx *fiber.Ctx) error {
	taskID, err := strconv.Atoi(ctx.Params("id"))
	if err != nil {
		s.log.Error("Failed to parse int", zap.Error(err))
		return dto.BadResponseError(ctx, dto.FieldBadFormat, "ID must be only number")
	}

	mu.Lock()
	_, exists := tasks[taskID]
	if !exists {
		mu.Unlock()
		s.log.Error("Failed to delete task", zap.Error(err))
		return dto.NotFoundError(ctx)
	}

	delete(tasks, taskID)
	mu.Unlock()

	response := dto.Response{
		Status: "success",
		Data:   map[string]int{"task_id": taskID},
	}
	return ctx.Status(fiber.StatusOK).JSON(response)
}
