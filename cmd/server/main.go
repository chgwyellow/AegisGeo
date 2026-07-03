package main

import (
	"AegisGeo/internal/models"
	"AegisGeo/internal/store"
	"fmt"
	"sync"
	"time"
)

func main() {
	fmt.Println("AegisGeo is starting...")

	// Initialize store, get memory address
	cache := store.NewMemoryCache()

	// Prepare Wait Group
	var wg sync.WaitGroup
	wg.Add(3) // 3 pipelines

	// 1st pipeline: CWA
	go func() {
		defer wg.Done()
		for i := 1; i <= 5; i++ {
			event := models.Event{
				ID:        fmt.Sprintf("CWA-%d", i),
				Source:    "CWA",
				Type:      "Earthquake",
				Magnitude: 4.5 + float64(i)*0.2,
				Timestamp: time.Now(),
			}
			cache.Set(event)
			fmt.Printf("[CWA] Written Event: %s\n", event.ID)
			time.Sleep(100 * time.Millisecond)
		}
	}()

	// 2nd pipeline: USA
	go func() {
		defer wg.Done()
		for i := 1; i <= 5; i++ {
			event := models.Event{
				ID:        fmt.Sprintf("USGS-%d", i),
				Source:    "USGS",
				Type:      "Tsunami",
				Magnitude: 0.0,
				Timestamp: time.Now(),
			}
			cache.Set(event)
			fmt.Printf("[USGS] Written Event: %s\n", event.ID)
			time.Sleep(150 * time.Millisecond)
		}
	}()

	// 3rd pipeline: Japan
	go func() {
		defer wg.Done()
		for i := 1; i <= 5; i++ {
			event := models.Event{
				ID:        fmt.Sprintf("JMA-%d", i),
				Source:    "JMA",
				Type:      "Volcano",
				Magnitude: 3.8 + float64(i)*0.1,
				Timestamp: time.Now(),
			}
			cache.Set(event)
			fmt.Printf("[JAM] Written Event: %s\n", event.ID)
			time.Sleep(500 * time.Millisecond)
		}
	}()

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
