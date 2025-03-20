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
	insertTaskQuery = `INSERT INTO tasks (title, description) VALUES ($1, $2) RETURNING id`
	getTaskQuery    = `SELECT title, description FROM tasks WHERE id=($1)`

	insertUserQuery  = `INSERT INTO users (username, password) VALUES ($1, $2) RETURNING id`
	getUserQuery     = `SELECT id, username, password, created_at FROM users WHERE id=($1)`
	getAllUsersQuery = `SELECT id, username, password, created_at FROM users`
	updateUserQuery  = `UPDATE users SET username = $1, password = $2 WHERE id = $3`
	deleteUserQuery  = `DELETE FROM users WHERE id = $1`
)

type repository struct {
	pool *pgxpool.Pool
}

// Repository - интерфейс с методом создания задачи
type Repository interface {
	// Задачи
	GetTask(ctx context.Context, taskID int) (*Task, error)
	CreateTask(ctx context.Context, task Task) (int, error)

	// Пользователи
	CreateUser(ctx context.Context, user UserCreate) (int, error)
	GetUser(ctx context.Context, userID int) (*UserView, error)
	GetAllUsers(ctx context.Context) (*[]UserView, error)
	UpdateUser(ctx context.Context, user UserUpdate) error
	DeleteUser(ctx context.Context, id int) error
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
	err := r.pool.QueryRow(ctx, insertTaskQuery, task.Title, task.Description).Scan(&id)
	if err != nil {
		return 0, errors.Wrap(err, "failed to insert task")
	}
	return id, nil
}

func (r *repository) GetTask(ctx context.Context, taskID int) (*Task, error) {
	var task Task
	err := r.pool.QueryRow(ctx, getTaskQuery, taskID).Scan(&task.Title, &task.Description)
	if err != nil {
		return nil, errors.Wrap(err, "failed to get task")
	}
	return &task, nil
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
