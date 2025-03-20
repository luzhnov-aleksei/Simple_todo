package users_service

import (
	"encoding/json"
	"github.com/gofiber/fiber/v2"
	"go.uber.org/zap"
	"simple-service/internal/dto"
	"simple-service/internal/repo"
	"simple-service/pkg/validator"
	"strconv"
)

type UsersService interface {
	GetUser(ctx *fiber.Ctx) error
	GetAllUsers(ctx *fiber.Ctx) error
	CreateUser(ctx *fiber.Ctx) error
	UpdateUser(ctx *fiber.Ctx) error
	DeleteUser(ctx *fiber.Ctx) error
}

func NewService(repo repo.Repository, logger *zap.SugaredLogger) UsersService {
	return &service{
		repo: repo,
		log:  logger,
	}
}

type service struct {
	repo repo.Repository
	log  *zap.SugaredLogger
}

func (s *service) CreateUser(ctx *fiber.Ctx) error {
	var req UserRequest

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
	user := repo.UserCreate{
		Username: req.Username,
		Password: req.Password,
	}
	userID, err := s.repo.CreateUser(ctx.Context(), user)
	if err != nil {
		s.log.Error("Failed to insert task", zap.Error(err))
		return dto.InternalServerError(ctx)
	}

	// Формирование ответа
	response := dto.Response{
		Status: "success",
		Data:   map[string]int{"user_id": userID},
	}

	return ctx.Status(fiber.StatusOK).JSON(response)
}

func (s *service) GetUser(ctx *fiber.Ctx) error {
	userID, err := strconv.Atoi(ctx.Params("id"))
	if err != nil {
		s.log.Error("Failed to parse int", zap.Error(err))
		return dto.BadResponseError(ctx, dto.FieldBadFormat, "ID must be only number")
	}
	user, err := s.repo.GetUser(ctx.Context(), userID)
	if err != nil {
		s.log.Error("Failed to get user", zap.Error(err))
		return dto.InternalServerError(ctx)
	}
	// Формирование ответа
	response := dto.Response{
		Status: "success",
		Data:   user,
	}

	return ctx.Status(fiber.StatusOK).JSON(response)
}

func (s *service) GetAllUsers(ctx *fiber.Ctx) error {
	allUsers, err := s.repo.GetAllUsers(ctx.Context())
	if err != nil {
		s.log.Error("Failed to get all users", zap.Error(err))
		return dto.InternalServerError(ctx)
	}

	response := dto.Response{
		Status: "success",
		Data:   allUsers,
	}

	return ctx.Status(fiber.StatusOK).JSON(response)
}

func (s *service) UpdateUser(ctx *fiber.Ctx) error {
	var req UserRequest

	if err := json.Unmarshal(ctx.Body(), &req); err != nil {
		s.log.Error("Invalid request body", zap.Error(err))
		return dto.BadResponseError(ctx, dto.FieldBadFormat, "Invalid request body")
	}

	id, err := strconv.Atoi(ctx.Params("id"))
	if err != nil {
		s.log.Error("Failed to parse int", zap.Error(err))
		return dto.BadResponseError(ctx, dto.FieldBadFormat, "ID must be only number")
	}
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
	user := repo.UserUpdate{
		Username: req.Username,
		Password: req.Password,
		ID:       id,
	}

	rErr := s.repo.UpdateUser(ctx.Context(), user)
	if rErr != nil {
		s.log.Error("Failed to update user", zap.Error(rErr))
		return dto.InternalServerError(ctx)
	}

	// Формирование ответа
	response := dto.Response{
		Status: "success",
		Data:   map[string]int{"user_id": id},
	}

	return ctx.Status(fiber.StatusOK).JSON(response)
}

func (s *service) DeleteUser(ctx *fiber.Ctx) error {

	userID, err := strconv.Atoi(ctx.Params("id"))
	if err != nil {
		s.log.Error("Failed to parse int", zap.Error(err))
		return dto.BadResponseError(ctx, dto.FieldBadFormat, "ID must be only number")
	}

	dErr := s.repo.DeleteUser(ctx.Context(), userID)
	if dErr != nil {
		s.log.Error("Failed to delete user", zap.Error(dErr))
		return dto.BadResponseError(ctx, dto.FieldBadFormat, "ID must be only number")
	}
	response := dto.Response{
		Status: "success",
	}
	return ctx.Status(fiber.StatusOK).JSON(response)
}
