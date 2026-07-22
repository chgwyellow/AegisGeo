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
}

func BuildHealthResult(client ingestion.IngestionClient) HealthResult {
	events, err := client.FetchLatest()

	latestEventTime := time.Time{}

	for _, event := range events {
		if event.Timestamp.After(latestEventTime) {
			latestEventTime = event.Timestamp
		}
	}

	if err != nil {
		return HealthResult{
			EventCount: 0,
			Source:     client.GetName(),
			Status:     "FAIL",
			Error:      err.Error(), // convert error object to string
		}
	}

	return HealthResult{
		EventCount:      len(events),
		Source:          client.GetName(),
		Status:          "OK",
		Error:           "",
		LatestEventTime: latestEventTime,
	}
}
