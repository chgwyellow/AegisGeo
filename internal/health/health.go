package health

import "AegisGeo/internal/ingestion"

type HealthResult struct {
	EventCount int
}

func BuildHealthResult(client ingestion.IngestionClient) HealthResult {
	events, _ := client.FetchLatest()

	return HealthResult{
		EventCount: len(events),
	}
}
