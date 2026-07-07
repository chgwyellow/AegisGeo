package database

import (
	"AegisGeo/internal/models"
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

	pool, err := pgxpool.New(ctx, connStr)
	if err != nil {
		return nil, fmt.Errorf("Can not establish database connection pool: %v", err)
	}

	// Physical test
	err = pool.Ping(ctx)
	if err != nil {
		return nil, fmt.Errorf("Database Ping failed: %v", err)
	}

	return &PostgresDB{Pool: pool}, nil
}

// Close pool
func (db *PostgresDB) Close() {
	if db.Pool != nil {
		db.Pool.Close()
	}
}

// Save event via SQL Upsert
func (db *PostgresDB) SaveEvent(ctx context.Context, e models.Event) error {
	// ON CONFLICT (id) DO UPDATE
	query := `
		INSERT INTO geo_events (id, source, event_type, title, magnitude, depth, event_timestamp, country, location, latitude, longitude)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)
		ON CONFLICT (id) 
		DO UPDATE SET 
			magnitude = EXCLUDED.magnitude,
			depth = EXCLUDED.depth,
			event_timestamp = EXCLUDED.event_timestamp,
			title = EXCLUDED.title,
			location = EXCLUDED.location,
			latitude = EXCLUDED.latitude,
			longitude = EXCLUDED.longitude;
	`
	_, err := db.Pool.Exec(ctx, query,
		e.ID,
		e.Source,
		e.Type,
		e.Title,
		e.Magnitude,
		e.Depth,
		e.Timestamp,
		e.Country,
		e.Location,
		e.Latitude,
		e.Longitude,
	)
	return err
}
