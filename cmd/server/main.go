package main

import (
	"AegisGeo/internal/database"
	"AegisGeo/internal/ingestion"
	"AegisGeo/internal/store"
	"context"
	"fmt"
	"log"
	"os"
	"sync"
	"time"

	"github.com/joho/godotenv"
)

func main() {
	fmt.Println("AegisGeo is starting...")

	// Load .env
	err := godotenv.Load()
	if err != nil {
		log.Fatalf("Fail to load .env: %s", err)
	}

	// Initialize DB
	dbURL := os.Getenv("DATABASE_URL")
	db, err := database.NewPostgresDB(dbURL)
	if err != nil {
		log.Fatalf("Database built failed: %v", err)
	}
	defer db.Close()
	fmt.Println("PostgreSQL connection is online.")

	// Read env variables
	cwaURL := os.Getenv("CWA_API_URL")
	cwaToken := os.Getenv("CWA_TOKEN")
	usgsURL := os.Getenv("USGS_API_URL")
	jmaURL := os.Getenv("JMA_API_URL")

	if cwaURL == "" || cwaToken == "" || usgsURL == "" || jmaURL == "" {
		log.Fatal("Get wrong in environment setting!")
	}

	// Initialize store, get memory address
	cache := store.NewMemoryCache()

	// Create Clients
	clients := []ingestion.IngestionClient{
		ingestion.NewCwaClient(cwaURL, cwaToken),
		ingestion.NewUsgsClient(usgsURL),
		ingestion.NewJmaClient(jmaURL),
	}

	// Prepare Wait Group
	var wg sync.WaitGroup
	wg.Add(len(clients)) // According to source amount

	for _, client := range clients {
		c := client

		go func() {
			defer wg.Done()
			fmt.Printf("Start [%s Engine]...\n", c.GetName())

			events, err := c.FetchLatest()
			if err != nil {
				fmt.Printf("[%s Engine] failed: %v\n", c.GetName(), err)
				return
			}
			for _, event := range events {
				cache.Set(event) // To memory
				fmt.Printf("[%s Engine] Added: %v\n", c.GetName(), event.ID)

				// To PostgresSQL
				ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
				err := db.SaveEvent(ctx, event)
				cancel()
				if err != nil {
					fmt.Printf("[Fail to write to DB] ID: %s, error: %v\n", event.ID, err)
				}
			}
		}()
	}

	// Wait for goroutines finish their work
	wg.Wait()
	fmt.Println("\nAll data has been recorded")

	// Get all data
	fmt.Printf("There are %d event(s) in the memory!\n", len(cache.GetAll()))

	for _, e := range cache.GetAll() {
		localTimeStr := e.Timestamp.Format("2006-01-02 15:04:06 (MST)")
		fmt.Printf("   - [%s] Type: %s, Magnitude: %.1f, Time: %v, Location: %s\n", e.ID, e.Type, e.Magnitude, localTimeStr, e.Location)
	}

}
