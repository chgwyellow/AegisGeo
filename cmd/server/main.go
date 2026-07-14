package main

import (
	"AegisGeo/internal/api"
	"AegisGeo/internal/database"
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/joho/godotenv"
)

func main() {
	fmt.Println("AegisGeo API Server is starting...")

	// Load .env (optional)
	_ = godotenv.Load()

	// Initialize DB
	dbURL := os.Getenv("DATABASE_URL")
	db, err := database.NewPostgresDB(dbURL)
	if err != nil {
		log.Fatalf("Database initialization failed: %v", err)
	}
	defer db.Close()
	fmt.Println("PostgreSQL connection is online.")

	// Route definition
	http.HandleFunc("/api/events", api.EventsHandler(db))
	http.HandleFunc("/api/status", api.StatusHandler(db))

	log.Println("AegisGeo server is running on :8080...")
	if err := http.ListenAndServe(":8080", nil); err != nil {
		log.Fatal("Server failed:", err)
	}
}
