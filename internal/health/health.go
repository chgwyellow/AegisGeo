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

	fmt.Fprintf(&builder, "%-18s %-6s %-5s %-10s\n", "Source", "Status", "Count", "Duration")
	builder.WriteString(strings.Repeat("-", 39))
	builder.WriteString("\n")

	for _, r := range results {
		fmt.Fprintf(&builder, "%-18s %-6s %-5d %-10d\n", r.Source, r.Status, r.EventCount, r.Duration)
	}

	return builder.String() // builder is not a string type
}
