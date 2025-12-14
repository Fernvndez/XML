package database

import (
	"fmt"
	"time"

	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
)

// NewPostgresConnection cria uma nova conexão com o PostgreSQL
func NewPostgresConnection(dsn string, maxConnections int) (*sqlx.DB, error) {
	db, err := sqlx.Connect("postgres", dsn)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to database: %w", err)
	}

	// Configura o pool de conexões
	db.SetMaxOpenConns(maxConnections)
	db.SetMaxIdleConns(maxConnections / 5)
	db.SetConnMaxLifetime(time.Hour)
	db.SetConnMaxIdleTime(10 * time.Minute)

	// Testa a conexão
	if err := db.Ping(); err != nil {
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	return db, nil
}