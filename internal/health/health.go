package health

import (
	"AegisGeo/internal/ingestion"
	"fmt"
	"strings"
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

func FormatHealthResults(results []HealthResult) string {
	var builder strings.Builder // a container which can be written string

	fmt.Fprintf(&builder, "%-10s %-4s %-5s\n", "Source", "Status", "Counts")

	for _, r := range results {
		builder.WriteString(r.Source)
		builder.WriteString(" ")
		builder.WriteString(r.Status)
		builder.WriteString(" ")
		fmt.Fprintf(&builder, "%d", r.EventCount)
		builder.WriteString("\n")
	}

	return builder.String() // builder is not a string type
}
