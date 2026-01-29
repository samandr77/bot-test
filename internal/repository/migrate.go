package repository

import (
	"database/sql"
	"embed"
	"fmt"
	"log/slog"

	_ "github.com/lib/pq"
	"github.com/pressly/goose/v3"
)

//go:embed migrations/*.sql
var embedMigrations embed.FS

func RunMigrations(dsn string) error {
	db, openErr := sql.Open("postgres", dsn)
	if openErr != nil {
		return fmt.Errorf("ошибка открытия соединения для миграций: %w", openErr)
	}
	defer func() {
		if closeErr := db.Close(); closeErr != nil {
			slog.Error("Failed to close database during migration", "error", closeErr)
		}
	}()

	goose.SetBaseFS(embedMigrations)

	if dialectErr := goose.SetDialect("postgres"); dialectErr != nil {
		return fmt.Errorf("ошибка установки диалекта goose: %w", dialectErr)
	}

	if upErr := goose.Up(db, "migrations"); upErr != nil {
		return fmt.Errorf("ошибка применения миграций: %w", upErr)
	}

	slog.Info("Миграции успешно применены")
	return nil
}
