package health

import (
	"AegisGeo/internal/ingestion"
	"time"
)

type HealthResult struct {
	EventCount      int
	Source          string
	Status          string
	Error           string
	LatestEventTime time.Time
	Duration        time.Duration
}

func BuildHealthResult(client ingestion.IngestionClient) HealthResult {
	start := time.Now()

	events, err := client.FetchLatest()

	duration := time.Since(start)

	latestEventTime := time.Time{}

	for _, event := range events {
		if event.Timestamp.After(latestEventTime) {
			latestEventTime = event.Timestamp
		}
	}

	if err != nil {
		return HealthResult{
			EventCount:      0,
			Source:          client.GetName(),
			Status:          "FAIL",
			Error:           err.Error(), // convert error object to string
			LatestEventTime: latestEventTime,
			Duration:        duration,
		}
	}

	return HealthResult{
		EventCount:      len(events),
		Source:          client.GetName(),
		Status:          "OK",
		Error:           "",
		LatestEventTime: latestEventTime,
		Duration:        duration,
	}
}

func BuildHealthResults(clients []ingestion.IngestionClient) []HealthResult {
	results := make([]HealthResult, 0, len(clients))

	for _, client := range clients {
		result := BuildHealthResult(client)
		results = append(results, result)
	}

	return results
}
