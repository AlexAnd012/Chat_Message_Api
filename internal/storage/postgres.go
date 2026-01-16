package storage

import (
	"context"
	"database/sql"
	"fmt"
	"os"
	"time"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

// возвращаем
// *gorm.DB объект, с которым будем работать
// *sql.DB для настройки пула, ping и Close() (закроем через defer в main.go)

func OpenPostgres(ctx context.Context) (*gorm.DB, *sql.DB, error) {

	// читаем DSN из окружения + дефолтное значение
	dsn := os.Getenv("DATABASE_DSN")
	if dsn == "" {
		dsn = "postgres://postgres:postgres@db:5432/chatdb?sslmode=disable"
	}

	// открываем GORM соединение
	gdb, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		return nil, nil, fmt.Errorf("open postgres gorm: %w", err)
	}

	// получаем *sql.DB из GORM
	sqlDB, err := gdb.DB()
	if err != nil {
		return nil, nil, fmt.Errorf("get sql.DB: %w", err)
	}

	// настройки пула
	sqlDB.SetMaxOpenConns(10)
	sqlDB.SetMaxIdleConns(5)
	sqlDB.SetConnMaxLifetime(10 * time.Minute)
	sqlDB.SetConnMaxIdleTime(5 * time.Minute)

	// ping базы с таймаутом
	pingCtx, cancel := context.WithTimeout(ctx, 3*time.Second)
	defer cancel()

	if err := sqlDB.PingContext(pingCtx); err != nil {
		_ = sqlDB.Close()
		return nil, nil, fmt.Errorf("ping postgres: %w", err)
	}

	return gdb, sqlDB, nil
}
