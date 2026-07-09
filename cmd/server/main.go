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
	cwaURL := os.Getenv("CWA_EQK_URL")
	cwaRainURL := os.Getenv("CWA_RAIN_URL")
	cwaToken := os.Getenv("CWA_TOKEN")
	usgsURL := os.Getenv("USGS_API_URL")
	jmaURL := os.Getenv("JMA_API_URL")
	noaaURL := os.Getenv("NOAA_API_URL")
	nwsURL := os.Getenv("NWS_API_URL")
	volURL := os.Getenv("VOLCANO_API_URL")
	email := os.Getenv("Email")

	if cwaURL == "" || cwaToken == "" || usgsURL == "" || jmaURL == "" || noaaURL == "" || cwaRainURL == "" || nwsURL == "" || email == "" || volURL == "" {
		log.Fatal("Get wrong in environment setting!")
	}

	// Initialize store, get memory address
	cache := store.NewMemoryCache()

	// Create Clients
	clients := []ingestion.IngestionClient{
		ingestion.NewCwaClient(cwaURL, cwaToken),
		ingestion.NewUsgsClient(usgsURL),
		ingestion.NewJmaClient(jmaURL),
		ingestion.NewTsunamiClient(noaaURL),
		ingestion.NewCwaRainClient(cwaRainURL, cwaToken),
		ingestion.NewNwsSevereWeatherClient(nwsURL, email),
		ingestion.NewVolcanoClient(volURL),
	}
	fmt.Println("Start Ingestion Cycle...")

	fmt.Println("==================================================")
	fmt.Printf("AegisGeo Telemetry Cycle Started at %v\n", time.Now().Format("2006-01-02 15:04:06"))
	fmt.Println("==================================================")

	var wg sync.WaitGroup
	wg.Add(len(clients))

	for _, client := range clients {
		c := client

		go func() {
			defer wg.Done()

			events, err := c.FetchLatest()
			if err != nil {
				fmt.Printf("[✗] [%s Engine] failed: %v\n", c.GetName(), err)
				return
			}
			for _, event := range events {
				cache.Set(event)

				ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
				err := db.SaveEvent(ctx, event)
				cancel()
				if err != nil {
					fmt.Printf("[✗] [%s Engine] DB Write Error ID: %s, error: %v\n", c.GetName(), event.ID, err)
				}
			}
			fmt.Printf("[✓] [%s Engine] successfully processed %d events\n", c.GetName(), len(events))
		}()
	}

	wg.Wait()
	fmt.Println("--------------------------------------------------")
	allEvents := cache.GetAll()
	fmt.Printf("Telemetry Sync Completed. Total active events in Cache: %d\n", len(allEvents))

	limit := min(len(allEvents), 5)
	if limit > 0 {
		fmt.Println("\nLatest 5 Anomaly Events:")
		for i := range limit {
			e := allEvents[i]
			localTimeStr := e.Timestamp.Format("2006-01-02 15:04:06 (MST)")
			fmt.Printf("   %d. [%s] Type: %-13s Magnitude: %-5.1f Time: %v | Location: %s\n",
				i+1, e.ID, e.Type, e.Magnitude, localTimeStr, e.Location)
		}
	}
	fmt.Println("==================================================")
}
