package database

import (
	"database/sql"
	"fmt"
	"log"

	"github.com/deepwrite/project-service/internal/config"
	_ "github.com/lib/pq"
)

func NewPostgres(cfg *config.DatabaseConfig) (*sql.DB, error) {
	db, err := sql.Open("postgres", cfg.DSN())
	if err != nil {
		return nil, fmt.Errorf("open database: %w", err)
	}
	if err := db.Ping(); err != nil {
		return nil, fmt.Errorf("ping database: %w", err)
	}
	db.SetMaxOpenConns(25)
	db.SetMaxIdleConns(10)
	log.Println("PostgreSQL connected successfully")
	return db, nil
}
