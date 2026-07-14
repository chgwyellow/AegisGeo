package database

import (
	"AegisGeo/internal/models"
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type PostgresDB struct {
	Pool *pgxpool.Pool
}

// Initialize connection pool
func NewPostgresDB(connStr string) (*PostgresDB, error) {
	// prevent connection overtime
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
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
			details = EXCLUDED.details,
			country = EXCLUDED.country;
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

// SaveEvents saves a slice of events using database transactions and batching for high performance
func (db *PostgresDB) SaveEvents(ctx context.Context, events []models.Event) error {
	if len(events) == 0 {
		return nil
	}

	tx, err := db.Pool.Begin(ctx)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %v", err)
	}
	defer tx.Rollback(ctx)

	var insertEvents []models.Event

	// 1. Process collision checks sequentially (only for Earthquake types since count is small)
	for _, e := range events {
		if e.Type == "Earthquake" {
			collisionQuery := `
				SELECT id, source FROM geo_events
				WHERE event_timestamp BETWEEN $1::TIMESTAMPTZ - INTERVAL '60 SECOND' AND $1::TIMESTAMPTZ + INTERVAL '60 SECOND'
				  AND ST_DistanceSphere(geom, ST_SetSRID(ST_MakePoint($2, $3), 4326)) <= 50000
				  AND event_type = $4
				LIMIT 1;
			`
			var existingID, existingSource string
			err := tx.QueryRow(ctx, collisionQuery, e.Timestamp, e.Longitude, e.Latitude, e.Type).Scan(&existingID, &existingSource)
			if err == nil {
				if (existingSource == "CWA" || existingSource == "JMA") && e.Source == "USGS" {
					fmt.Printf("Eliminate duplicated events：USGS-%s and %s-%s \n", e.ID, existingSource, existingID)
					continue
				}
			}
		}
		insertEvents = append(insertEvents, e)
	}

	if len(insertEvents) == 0 {
		return nil
	}

	// 2. Queue all inserts in a batch
	batch := &pgx.Batch{}
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
			details = EXCLUDED.details,
			country = EXCLUDED.country;
	`

	for _, e := range insertEvents {
		detailsJSON, err := json.Marshal(e.Details)
		if err != nil {
			detailsJSON = []byte("{}")
		}

		batch.Queue(query,
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
	}

	// 3. Execute the batch inside the transaction
	br := tx.SendBatch(ctx, batch)
	for i := 0; i < len(insertEvents); i++ {
		_, err := br.Exec()
		if err != nil {
			br.Close()
			return fmt.Errorf("batch execution error at index %d: %v", i, err)
		}
	}

	err = br.Close()
	if err != nil {
		return fmt.Errorf("failed to close batch results: %v", err)
	}

	// 4. Commit transaction
	return tx.Commit(ctx)
}

// Get all type data
func (db *PostgresDB) GetEventSummaries(ctx context.Context, limit int) ([]models.EventSummary, error) {
	query := `
		SELECT id, title, source, event_type, magnitude, depth, event_timestamp, country, location
		FROM geo_events
		ORDER BY event_timestamp DESC
		LIMIT $1
	`

	rows, err := db.Pool.Query(ctx, query, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var summaries []models.EventSummary
	for rows.Next() {
		var s models.EventSummary
		err := rows.Scan(
			&s.ID, &s.Title, &s.Source, &s.Type, &s.Magnitude, &s.Depth, &s.Timestamp, &s.Country, &s.Location,
		)
		if err != nil {
			return nil, err
		}
		summaries = append(summaries, s)
	}
	return summaries, nil
}

// Get single type data
func (db *PostgresDB) GetEventsByType(ctx context.Context, eventType string, limit int) ([]models.EventSummary, error) {
	query := `
		SELECT id, title, source, event_type, magnitude, depth, event_timestamp, country, location
		FROM geo_events
		WHERE event_type = $1
		ORDER BY event_timestamp DESC, magnitude DESC
		LIMIT $2
	`

	rows, err := db.Pool.Query(ctx, query, eventType, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var summaries []models.EventSummary
	for rows.Next() {
		var s models.EventSummary
		err := rows.Scan(
			&s.ID, &s.Title, &s.Source, &s.Type, &s.Magnitude, &s.Depth, &s.Timestamp, &s.Country, &s.Location,
		)
		if err != nil {
			return nil, err
		}
		summaries = append(summaries, s)
	}
	return summaries, nil
}
