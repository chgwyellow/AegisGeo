package health

import "AegisGeo/internal/ingestion"

type HealthResult struct {
	EventCount int
	Source     string
	Status     string
}

func BuildHealthResult(client ingestion.IngestionClient) HealthResult {
	events, err := client.FetchLatest()

	if err != nil {
		return HealthResult{
			EventCount: 0,
			Source:     client.GetName(),
			Status:     "FALL",
		}
	}

	return HealthResult{
		EventCount: len(events),
		Source:     client.GetName(),
		Status:     "OK",
	}
}
