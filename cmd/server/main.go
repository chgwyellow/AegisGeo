package main

import (
	"AegisGeo/internal/ingestion"
	"AegisGeo/internal/store"
	"fmt"
	"sync"
)

func main() {
	fmt.Println("AegisGeo is starting...")

	// Initialize store, get memory address
	cache := store.NewMemoryCache()

	// Create Clients
	clients := []ingestion.IngestionClient{
		ingestion.NewCwaClient("https://api.cwa.gov.tw/fake", "your-token-here"),
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
				fmt.Printf("[%s Engine] Success: %s\n", c.GetName(), event.ID)
			}
		}()
	}

	// Wait for goroutines finish their work
	wg.Wait()
	fmt.Println("\nAll data has been recorded")

	// Get all data
	allEvents := cache.GetAll()
	fmt.Printf("There are %d events in the database!\n", len(allEvents))

	for _, e := range allEvents {
		fmt.Printf("   - [%s] Type: %s, Magnitude: %.1f\n", e.ID, e.Type, e.Magnitude)
	}
}
