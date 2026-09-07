package repository

import (
	"context"
	"database/sql"
	"fmt"
	"net/url"
	"time"

	_ "github.com/jackc/pgx/v5/stdlib" // замена lib/pq на pgx
)

type PostgresConfig struct {
	Host     string
	Port     string
	User     string
	Password string
	DBName   string
	SSLMode  string

	MaxOpenConns    int
	MaxIdleConns    int
	ConnMaxLifetime time.Duration
	ConnMaxIdleTime time.Duration
}

func NewPostgresDB(cfg PostgresConfig) (*sql.DB, error) {
	// защита от сбоев при наличии спецсимволов в пароле
	dsn := url.URL{
		Scheme: "postgres",
		User:   url.UserPassword(cfg.User, cfg.Password),
		Host:   fmt.Sprintf("%s:%s", cfg.Host, cfg.Port),
		Path:   cfg.DBName,
	}

	q := dsn.Query()
	q.Set("sslmode", cfg.SSLMode)
	dsn.RawQuery = q.Encode()

	db, err := sql.Open("pgx", dsn.String())
	if err != nil {
		return nil, fmt.Errorf("failed to open db connection: %w", err)
	}

	// настройка connection pool для предотвращения падений БД под нагрузкой
	maxOpen := 25
	if cfg.MaxOpenConns > 0 {
		maxOpen = cfg.MaxOpenConns
	}
	db.SetMaxOpenConns(maxOpen)

	maxIdle := 10
	if cfg.MaxIdleConns > 0 {
		maxIdle = cfg.MaxIdleConns
	}
	db.SetMaxIdleConns(maxIdle)

	connMaxLifetime := 15 * time.Minute
	if cfg.ConnMaxLifetime > 0 {
		connMaxLifetime = cfg.ConnMaxLifetime
	}
	db.SetConnMaxLifetime(connMaxLifetime)

	connMaxIdleTime := 5 * time.Minute
	if cfg.ConnMaxIdleTime > 0 {
		connMaxIdleTime = cfg.ConnMaxIdleTime
	}
	db.SetConnMaxIdleTime(connMaxIdleTime)

	// ping с таймаутом, чтобы сервис не зависал навечно при недоступной БД
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := db.PingContext(ctx); err != nil {
		_ = db.Close()
		return nil, fmt.Errorf("failed to ping db: %w", err)
	}

	return db, nil
}
