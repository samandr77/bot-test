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
	db, err := sql.Open("postgres", dsn)
	if err != nil {
		return fmt.Errorf("ошибка открытия соединения для миграций: %w", err)
	}
	defer db.Close()

	goose.SetBaseFS(embedMigrations)

	if err := goose.SetDialect("postgres"); err != nil {
		return fmt.Errorf("ошибка установки диалекта goose: %w", err)
	}

	if err := goose.Up(db, "migrations"); err != nil {
		return fmt.Errorf("ошибка применения миграций: %w", err)
	}

	slog.Info("Миграции успешно применены")
	return nil
}
