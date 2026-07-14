package api

import (
	"AegisGeo/internal/database"
	"AegisGeo/internal/models"
	"encoding/json"
	"net/http"
)

// EventsHandler handles the request to fetch recent disaster events.
func EventsHandler(db *database.PostgresDB, key string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Catch "X-API-KEY" from request
		providedKey := r.Header.Get("X-API-KEY")

		if providedKey != key {
			http.Error(w, "Unauthorized: Access Denied", http.StatusUnauthorized)
			return
		}

		eventType := r.URL.Query().Get("type")
		var events []models.EventSummary
		var err error

		if eventType != "" {
			events, err = db.GetEventsByType(r.Context(), eventType, 20)
			if err != nil {
				http.Error(w, "Failed to fetch data from database", http.StatusInternalServerError)
				return
			}
		} else {
			// Fetch summaries from the database with a limit of 20
			events, err = db.GetEventSummaries(r.Context(), 20)
			if err != nil {
				http.Error(w, "Failed to fetch data from database", http.StatusInternalServerError)
				return
			}
		}

		// Set response header to JSON
		w.Header().Set("Content-Type", "application/json")

		// Encode the slice to JSON and write to response
		if err := json.NewEncoder(w).Encode(events); err != nil {
			http.Error(w, "Failed to encode response", http.StatusInternalServerError)
			return
		}
	}
}

// StatusHandler checks if the database connection is alive.
func StatusHandler(db *database.PostgresDB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		err := db.Pool.Ping(r.Context())
		if err != nil {
			http.Error(w, "Database is down", http.StatusInternalServerError)
			return
		}

		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK"))
	}
}
