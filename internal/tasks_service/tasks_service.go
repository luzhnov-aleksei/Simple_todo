package tasks_service

import (
	"encoding/json"
	"github.com/gofiber/fiber/v2"
	"github.com/jackc/pgx/v5"
	"go.uber.org/zap"
	"simple-service/internal/dto"
	"simple-service/internal/repo"
	"simple-service/pkg/validator"
	"strconv"
)

type TasksService interface {
	GetTask(ctx *fiber.Ctx) error
	GetAllTasks(ctx *fiber.Ctx) error
	CreateTask(ctx *fiber.Ctx) error
	UpdateTask(ctx *fiber.Ctx) error
	DeleteTask(ctx *fiber.Ctx) error
	GetAllTasksFromUser(ctx *fiber.Ctx) error
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

	// Проверяем существование UserID
	exist, err := s.repo.CheckUserExists(ctx.Context(), req.UserID)
	if err != nil {
		s.log.Error("Failed to insert task", zap.Error(err))
		return dto.InternalServerError(ctx)
	}
	if exist != true {
		s.log.Error("User not found", zap.Error(err))
		return dto.NotFoundError(ctx)
	}

	task := repo.Task{
		UserID:      req.UserID,
		Title:       req.Title,
		Description: req.Description,
	}

	id, err := s.repo.CreateTask(ctx.Context(), task)
	if err != nil {
		s.log.Error("Failed to insert task", zap.Error(err))
		return dto.InternalServerError(ctx)
	}

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

	// Проверяем существование taskID
	exist, err := s.repo.CheckTaskExists(ctx.Context(), taskID)
	if err != nil {
		s.log.Error("Failed to insert task", zap.Error(err))
		return dto.InternalServerError(ctx)
	}
	if exist != true {
		s.log.Error("Task not found", zap.Error(err))
		return dto.NotFoundError(ctx)
	}

	task, err := s.repo.GetTask(ctx.Context(), taskID)

	if err != nil {
		s.log.Error("Failed to get task", zap.Error(err))
		return dto.InternalServerError(ctx)
	}

	if err == pgx.ErrNoRows {
		s.log.Error("Task not found", zap.Error(err))
		return dto.NotFoundError(ctx)
	}

	response := dto.Response{
		Status: "success",
		Data:   task,
	}

	return ctx.Status(fiber.StatusOK).JSON(response)
}

func (s *service) GetAllTasks(ctx *fiber.Ctx) error {

	allTasks, err := s.repo.GetAllTasks(ctx.Context())
	if err != nil {
		s.log.Error("Failed to get all tasks", zap.Error(err))
		return dto.InternalServerError(ctx)
	}

	response := dto.Response{
		Status: "success",
		Data:   allTasks,
	}

	return ctx.Status(fiber.StatusOK).JSON(response)
}

func (s *service) UpdateTask(ctx *fiber.Ctx) error {
	var req TaskUpdateRequest

	if err := json.Unmarshal(ctx.Body(), &req); err != nil {
		s.log.Error("Invalid request body", zap.Error(err))
		return dto.BadResponseError(ctx, dto.FieldBadFormat, "Invalid request body")
	}

	taskID, err := strconv.Atoi(ctx.Params("id"))
	if err != nil {
		s.log.Error("Failed to parse int", zap.Error(err))
		return dto.BadResponseError(ctx, dto.FieldBadFormat, "ID must be only number")
	}

	if vErr := validator.Validate(ctx.Context(), req); vErr != nil {
		return dto.BadResponseError(ctx, dto.FieldIncorrect, vErr.Error())
	}

	// Проверяем существование taskID
	exist, err := s.repo.CheckTaskExists(ctx.Context(), taskID)
	if err != nil {
		s.log.Error("Failed to insert task", zap.Error(err))
		return dto.InternalServerError(ctx)
	}
	if exist != true {
		s.log.Error("Task not found", zap.Error(err))
		return dto.NotFoundError(ctx)
	}

	// Вставка задачи в БД через репозиторий
	task := repo.TaskUpdate{
		ID:          taskID,
		Title:       req.Title,
		Description: req.Description,
		Status:      req.Status,
	}
	rErr := s.repo.UpdateTask(ctx.Context(), task)
	if rErr != nil {
		s.log.Error("Failed to update task", zap.Error(rErr))
		return dto.InternalServerError(ctx)
	}

	if rErr == pgx.ErrNoRows {
		s.log.Error("Task not found", zap.Error(err))
		return dto.NotFoundError(ctx)
	}

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

	// Проверяем существование taskID
	exist, err := s.repo.CheckTaskExists(ctx.Context(), taskID)
	if err != nil {
		s.log.Error("Failed to insert task", zap.Error(err))
		return dto.InternalServerError(ctx)
	}
	if exist != true {
		s.log.Error("Task not found", zap.Error(err))
		return dto.NotFoundError(ctx)
	}

	dErr := s.repo.DeleteTask(ctx.Context(), taskID)
	if dErr != nil {
		s.log.Error("Failed to delete task", zap.Error(dErr))
		return dto.BadResponseError(ctx, dto.FieldBadFormat, "ID must be only number")
	}

	if dErr == pgx.ErrNoRows {
		s.log.Error("Task not found", zap.Error(err))
		return dto.NotFoundError(ctx)
	}

	response := dto.Response{
		Status: "success",
	}
	return ctx.Status(fiber.StatusOK).JSON(response)
}

func (s *service) GetAllTasksFromUser(ctx *fiber.Ctx) error {
	username := ctx.Query("username")
	if username == "" {
		s.log.Error("Failed to get username")
		return dto.NotFoundError(ctx)
	}

	allTasks, err := s.repo.GetAllTaskFromUser(ctx.Context(), username)
	if err != nil {
		s.log.Error("Failed to get all tasks from user", zap.Error(err))
		return dto.InternalServerError(ctx)
	}

	response := dto.Response{
		Status: "success",
		Data:   allTasks,
	}

	return ctx.Status(fiber.StatusOK).JSON(response)
}
