package main

import (
	"AegisGeo/internal/ingestion"
	"AegisGeo/internal/store"
	"fmt"
	"log"
	"os"
	"sync"

	"github.com/joho/godotenv"
)

func main() {
	fmt.Println("AegisGeo is starting...")

	// Load .env
	err := godotenv.Load()
	if err != nil {
		log.Fatalf("Fail to load .env: %s", err)
	}

	// Read env variables
	cwaURL := os.Getenv("CWA_API_URL")
	cwaToken := os.Getenv("CWA_TOKEN")
	usgsURL := os.Getenv("USGS_API_URL")

	if cwaURL == "" || cwaToken == "" || usgsURL == "" {
		log.Fatal("Do not set CWA_API_URL or CWA_TOKEN yet!")
	}

	// Initialize store, get memory address
	cache := store.NewMemoryCache()

	// Create Clients
	clients := []ingestion.IngestionClient{
		ingestion.NewCwaClient(cwaURL, cwaToken),
		ingestion.NewUsgsClient(usgsURL),
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
				cache.Set(event)
				fmt.Printf("[%s Engine] Added: %v\n", c.GetName(), event.ID)
			}
		}()
	}

	// Wait for goroutines finish their work
	wg.Wait()
	fmt.Println("\nAll data has been recorded")

	// Get all data
	allEvents := cache.GetAll()
	fmt.Printf("There are %d event(s) in the database!\n", len(allEvents))

	for _, e := range allEvents {
		localTimeStr := e.Timestamp.Format("2006-01-02 15:04:06 (MST)")
		fmt.Printf("   - [%s] Type: %s, Magnitude: %.1f, Time: %v, Location: %s\n", e.ID, e.Type, e.Magnitude, localTimeStr, e.Location)
	}

}
