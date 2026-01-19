package repo

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"

	"task-manager/internal/logger"
	"task-manager/internal/models"
)

//go:generate mockgen -source=connect.go -destination=../mocks/mock_repo.go -package=mocks

type EventRepo interface {
	CreateEvent(ctx context.Context, e *models.Event) error
	UpdateEvent(ctx context.Context, e *models.Event) error
	DeleteEvent(ctx context.Context, id int64) error
	GetEventsForDay(ctx context.Context, userID int64, date time.Time) ([]models.Event, error)
	GetEventsForRange(ctx context.Context, userID int64, from, to time.Time) ([]models.Event, error)
	Ping(ctx context.Context) error
	Close()
}

type PgxPool interface {
	Exec(context.Context, string, ...interface{}) (pgconn.CommandTag, error)
	QueryRow(context.Context, string, ...interface{}) pgx.Row
	Query(context.Context, string, ...interface{}) (pgx.Rows, error)
	Close()
	Ping(context.Context) error
}

type Repository struct {
	Pool PgxPool
}

func NewRepository(pool PgxPool) *Repository {
	return &Repository{Pool: pool}
}

func Connect(databaseURL string) (*Repository, error) {
	connStr := os.Getenv("DATABASE_URL")
	if connStr == "" {
		connStr = databaseURL
	}

	logger.Printf("Using database URL: %s", connStr)

	// Конфигурация пула соединений
	config, err := pgxpool.ParseConfig(connStr)
	if err != nil {
		return nil, fmt.Errorf("parse config error: %v", err)
	}

	config.MaxConns = 10
	config.MinConns = 2
	config.MaxConnLifetime = time.Hour

	// Подключение с ретраями
	var pool *pgxpool.Pool
	for i := range 10 {
		pool, err = pgxpool.NewWithConfig(context.Background(), config)
		if err == nil {
			err = pool.Ping(context.Background())
			if err == nil {
				logger.Println("Database connection established with pgxpool")
				return &Repository{Pool: pool}, nil
			}
		}
		logger.Printf("Database not available, retrying... (attempt %d/10)", i+1)
		time.Sleep(2 * time.Second)
	}

	return nil, fmt.Errorf("database unavailable after 10 attempts: %v", err)
}

func (r *Repository) Close() {
	r.Pool.Close()
}

func (r *Repository) Ping(ctx context.Context) error {
	return r.Pool.Ping(ctx)
}
