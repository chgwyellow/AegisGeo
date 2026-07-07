package database

import (
	"AegisGeo/internal/models"
	"context"
	"encoding/json"
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

	// Check if there is any data whose time is less then 60 seconds
	// And physical distance is less than 50 km
	// And they have the same type
	collisionQuery := `
		SELECT id, source FROM geo_events
		WHERE event_timestamp BETWEEN $1::TIMESTAMPTZ - INTERVAL '60 SECOND' AND $1::TIMESTAMPTZ + INTERVAL '60 SECOND'
		  AND ST_DistanceSphere(geom, ST_SetSRID(ST_MakePoint($2, $3), 4326)) <= 50000
		  AND event_type = $4
		LIMIT 1;
	`
	// ST_MakePoint create a matrix point for long and lat
	// ST_SetSRID provide the earth physical model, 4326 is WGS84 coordinate system
	// ST_DistanceSphere calculates true distance of two coordinates, the unit is meter

	var existingID, existingSource string
	// If there is more than one row return, err is nil
	// Which means there are duplicate data
	err := db.Pool.QueryRow(ctx, collisionQuery, e.Timestamp, e.Longitude, e.Latitude, e.Type).Scan(&existingID, &existingSource)

	if err == nil {
		if (existingSource == "CWA" || existingSource == "JMA") && e.Source == "USGS" {
			fmt.Printf("Eliminate duplicated events：USGS-%s and %s-%s \n", e.ID, existingSource, existingID)
			return nil
		}
	}

	// ON CONFLICT (id) DO UPDATE
	query := `
		INSERT INTO geo_events (id, source, event_type, title, magnitude, depth, event_timestamp, country, location, longitude, latitude, geom, details)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, ST_SetSRID(ST_MakePoint($12, $13), 4326), $14)
		ON CONFLICT (id, event_type) 
		DO UPDATE SET 
			magnitude = EXCLUDED.magnitude,
			depth = EXCLUDED.depth,
			event_timestamp = EXCLUDED.event_timestamp,
			title = EXCLUDED.title,
			location = EXCLUDED.location,
			longitude = EXCLUDED.longitude,
			latitude = EXCLUDED.latitude,
			geom = EXCLUDED.geom,
			details = EXCLUDED.details;
	`
	// Serialization
	detailsJSON, err := json.Marshal(e.Details)
	if err != nil {
		detailsJSON = []byte("{}")
	}

	_, err = db.Pool.Exec(ctx, query,
		e.ID,
		e.Source,
		e.Type,
		e.Title,
		e.Magnitude,
		e.Depth,
		e.Timestamp,
		e.Country,
		e.Location,
		e.Longitude,
		e.Latitude,
		e.Longitude,
		e.Latitude,
		detailsJSON,
	)
	return err
}
