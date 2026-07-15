package api

import (
	"AegisGeo/internal/database"
	"AegisGeo/internal/models"
	"encoding/json"
	"net/http"
	"time"
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

		q := r.URL.Query()
		eventType := q.Get("type")
		startStr := q.Get("start")
		endStr := q.Get("end")

		// Check the correct format
		var start, end time.Time
		var err error
		const layout = "2006-01-02"

		loc := time.FixedZone("CST", 8*60*60)

		if startStr != "" {
			start, err = time.ParseInLocation(layout, startStr, loc)
			if err != nil {
				http.Error(w, "Invalid start date format, expected YYYY-MM-DD", http.StatusBadRequest)
				return
			}
		} else {
			start = time.Now().In(loc).AddDate(0, 0, -30)
		}

		if endStr != "" {
			end, err = time.ParseInLocation(layout, endStr, loc)
			if err != nil {
				http.Error(w, "Invalid end date format, expected YYYY-MM-DD", http.StatusBadRequest)
				return
			}
			end = end.AddDate(0, 0, 1)
		} else {
			end = time.Now().In(loc)
		}

		var events []models.EventSummary
		if eventType != "" {
			events, err = db.GetEventsByType(r.Context(), eventType, start, end, 20)
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

		// * prevent nil slice
		if events == nil {
			events = []models.EventSummary{}
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
