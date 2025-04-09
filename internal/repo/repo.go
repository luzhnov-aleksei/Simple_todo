package repo

import (
	"context"
	"fmt"
	"github.com/jackc/pgx/v5"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/pkg/errors"

	"simple-service/internal/config"
)

// Слой репозитория, здесь должны быть все методы, связанные с базой данных

// SQL-запрос на вставку задачи
const (
	insertTaskQuery  = `INSERT INTO tasks (user_id, title, description) VALUES ($1, $2, $3) RETURNING id`
	getTaskQuery     = `SELECT id, user_id, title, description, status, created_at FROM tasks WHERE id=($1)`
	getAllTasksQuery = `SELECT id, user_id, title,description, status, created_at FROM tasks`
	updateTaskQuery  = `UPDATE tasks SET title = $1, description = $2, status = $3 WHERE id = $4`
	deleteTaskQuery  = `DELETE FROM tasks WHERE id = $1`
	checkTaskQuery   = "SELECT EXISTS(SELECT 1 FROM tasks WHERE id = $1)"

	insertUserQuery    = `INSERT INTO users (username, password) VALUES ($1, $2) RETURNING id`
	getUserQuery       = `SELECT id, username, password, created_at FROM users WHERE id=($1)`
	getAllUsersQuery   = `SELECT id, username, password, created_at FROM users`
	updateUserQuery    = `UPDATE users SET username = $1, password = $2  WHERE id = $3`
	deleteUserQuery    = `DELETE FROM users WHERE id = $1`
	checkUserQuery     = "SELECT EXISTS(SELECT 1 FROM users WHERE id = $1)"
	getAllTaskFromUser = "SELECT t.* FROM users AS u LEFT JOIN tasks AS t ON t.user_id = u.id WHERE u.username = $1;"
)

type repository struct {
	pool *pgxpool.Pool
}

// Repository - интерфейс с методом создания задачи
type Repository interface {
	// Задачи
	CreateTask(ctx context.Context, task Task) (int, error)
	GetTask(ctx context.Context, taskID int) (*TaskView, error)
	GetAllTasks(ctx context.Context) (*[]TaskView, error)
	UpdateTask(ctx context.Context, task TaskUpdate) error
	DeleteTask(ctx context.Context, id int) error
	CheckTaskExists(ctx context.Context, userID int) (bool, error)
	GetAllTaskFromUser(ctx context.Context, username string) (*[]TaskView, error)

	// Пользователи
	CreateUser(ctx context.Context, user UserCreate) (int, error)
	GetUser(ctx context.Context, userID int) (*UserView, error)
	GetAllUsers(ctx context.Context) (*[]UserView, error)
	UpdateUser(ctx context.Context, user UserUpdate) error
	DeleteUser(ctx context.Context, id int) error
	CheckUserExists(ctx context.Context, userID int) (bool, error)
}

// NewRepository - создание нового экземпляра репозитория с подключением к PostgreSQL
func NewRepository(ctx context.Context, cfg config.PostgreSQL) (Repository, error) {
	// Формируем строку подключения
	connString := fmt.Sprintf(
		`user=%s password=%s host=%s port=%d dbname=%s sslmode=%s 
        pool_max_conns=%d pool_max_conn_lifetime=%s pool_max_conn_idle_time=%s`,
		cfg.User,
		cfg.Password,
		cfg.Host,
		cfg.Port,
		cfg.Name,
		cfg.SSLMode,
		cfg.PoolMaxConns,
		cfg.PoolMaxConnLifetime.String(),
		cfg.PoolMaxConnIdleTime.String(),
	)

	// Парсим конфигурацию подключения
	config, err := pgxpool.ParseConfig(connString)
	if err != nil {
		return nil, errors.Wrap(err, "failed to parse PostgreSQL config")
	}

	// Оптимизация выполнения запросов (кеширование запросов)
	config.ConnConfig.DefaultQueryExecMode = pgx.QueryExecModeCacheDescribe

	// Создаём пул соединений с базой данных
	pool, err := pgxpool.NewWithConfig(ctx, config)
	if err != nil {
		return nil, errors.Wrap(err, "failed to create PostgreSQL connection pool")
	}

	return &repository{pool}, nil
}

// CreateTask - вставка новой задачи в таблицу tasks
func (r *repository) CreateTask(ctx context.Context, task Task) (int, error) {
	var id int
	err := r.pool.QueryRow(ctx, insertTaskQuery, task.UserID, task.Title, task.Description).Scan(&id)
	if err != nil {
		return 0, errors.Wrap(err, "failed to insert task")
	}
	return id, nil
}

func (r *repository) GetTask(ctx context.Context, taskID int) (*TaskView, error) {
	var task TaskView
	err := r.pool.QueryRow(ctx, getTaskQuery, taskID).Scan(
		&task.ID, &task.UserID, &task.Title, &task.Description, &task.Status, &task.CreatedAt)
	if err != nil {
		return nil, errors.Wrap(err, "failed to get task")
	}
	return &task, nil
}

func (r *repository) GetAllTasks(ctx context.Context) (*[]TaskView, error) {
	var tasks []TaskView
	rows, err := r.pool.Query(ctx, getAllTasksQuery)
	if err != nil {
		return nil, errors.Wrap(err, "failed to get all tasks")
	}
	for rows.Next() {
		var task TaskView
		// Сканируем значения из строки в поля структуры User
		if err := rows.Scan(
			&task.ID, &task.UserID, &task.Title, &task.Description, &task.Status, &task.CreatedAt); err != nil {
			return nil, errors.Wrap(err, "failed to get all users")
		}
		tasks = append(tasks, task)
	}

	return &tasks, nil
}

func (r *repository) UpdateTask(ctx context.Context, task TaskUpdate) error {
	_, err := r.pool.Query(ctx, updateTaskQuery, task.Title, task.Description, task.Status, task.ID)
	if err != nil {
		return errors.Wrap(err, "failed to update task")
	}
	return nil
}

func (r *repository) DeleteTask(ctx context.Context, id int) error {
	_, err := r.pool.Query(ctx, deleteTaskQuery, id)
	if err != nil {
		return errors.Wrap(err, "failed to delete task")
	}
	return nil
}

func (r *repository) CheckTaskExists(ctx context.Context, userID int) (bool, error) {
	var exists bool
	err := r.pool.QueryRow(ctx, checkTaskQuery, userID).Scan(&exists)
	if err != nil {
		return false, err
	}
	return exists, nil
}

// Users ------------------------------------

func (r *repository) CreateUser(ctx context.Context, user UserCreate) (int, error) {
	var id int
	err := r.pool.QueryRow(ctx, insertUserQuery, user.Username, user.Password).Scan(&id)
	if err != nil {
		return 0, errors.Wrap(err, "failed to insert user")
	}
	return id, nil
}

func (r *repository) GetUser(ctx context.Context, userID int) (*UserView, error) {
	var user UserView
	err := r.pool.QueryRow(ctx, getUserQuery, userID).Scan(&user.Username, &user.Password)
	if err != nil {
		return nil, errors.Wrap(err, "failed to get user")
	}
	return &user, nil
}

func (r *repository) GetAllUsers(ctx context.Context) (*[]UserView, error) {
	var users []UserView
	rows, err := r.pool.Query(ctx, getAllUsersQuery)
	if err != nil {
		return nil, errors.Wrap(err, "failed to get all users")
	}
	for rows.Next() {
		var user UserView
		// Сканируем значения из строки в поля структуры User
		if err := rows.Scan(&user.ID, &user.Username, &user.Password, &user.CreatedAt); err != nil {
			return nil, errors.Wrap(err, "failed to get all users")
		}
		users = append(users, user)
	}

	return &users, nil
}

func (r *repository) UpdateUser(ctx context.Context, user UserUpdate) error {
	_, err := r.pool.Query(ctx, updateUserQuery, user.Username, user.Password, user.ID)
	if err != nil {
		return errors.Wrap(err, "failed to update user")
	}
	return nil
}

func (r *repository) DeleteUser(ctx context.Context, id int) error {
	_, err := r.pool.Query(ctx, deleteUserQuery, id)
	if err != nil {
		return errors.Wrap(err, "failed to delete user")
	}
	return nil
}

func (r *repository) CheckUserExists(ctx context.Context, userID int) (bool, error) {
	var exists bool
	err := r.pool.QueryRow(ctx, checkUserQuery, userID).Scan(&exists)
	if err != nil {
		return false, err
	}
	return exists, nil
}

func (r *repository) GetAllTaskFromUser(ctx context.Context, username string) (*[]TaskView, error) {
	var tasks []TaskView
	rows, err := r.pool.Query(ctx, getAllTaskFromUser, username)
	if err != nil {
		return nil, errors.Wrap(err, "failed to get all tasks from user")
	}
	for rows.Next() {
		var task TaskView
		// Сканируем значения из строки в поля структуры User
		if err := rows.Scan(
			&task.ID, &task.UserID, &task.Title, &task.Description, &task.Status, &task.CreatedAt); err != nil {
			return nil, errors.Wrap(err, "failed to get all tasks from user")
		}
		tasks = append(tasks, task)
	}

	return &tasks, nil
}
