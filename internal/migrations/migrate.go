package db

import (
	"database/sql"
	"embed"
	"github.com/pressly/goose/v3"
)

//go:embed migrations/*.sql
var embedMigrations embed.FS

func RunMigrations(db *sql.DB) error {
	goose.SetDialect("postgres")
	goose.SetBaseFS(embedMigrations)
	return goose.Up(db, "migrations")
}
