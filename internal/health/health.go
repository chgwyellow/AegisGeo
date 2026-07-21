package health

import "AegisGeo/internal/ingestion"

type HealthResult struct {
	EventCount int
	Source     string
	Status     string
}

func BuildHealthResult(client ingestion.IngestionClient) HealthResult {
	events, _ := client.FetchLatest()

	return HealthResult{
		EventCount: len(events),
		Source:     client.GetName(),
		Status:     "OK",
	}
}
