package database

import (
	"aegisgeo/internal/models"
	"context"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

type PostgresDB struct {
	Pool *pgxpool.Pool
}

// Initialize connection pool
func NewPostgresDB(connStr string) (*PostgresDB, error) {
	// prevent connection overtime
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	
}