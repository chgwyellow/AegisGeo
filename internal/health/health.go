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

	line := strings.Repeat("-", 104)

	OK := 0
	FAIL := 0

	fmt.Fprintf(
		&builder,
		"AegisGeo Data Health Check\nGenerated at: %v\n",
		time.Now().Format("2006-01-02 15:04:05 MST"),
	)

	builder.WriteString(line)
	builder.WriteString("\n")
	fmt.Fprintf(
		&builder,
		"%-20s %-6s %-6s %-25s %-12s %-30s\n",
		"Source",
		"Status",
		"Count",
		"Latest Event Time",
		"Duration",
		"Error",
	)
	builder.WriteString(line)
	builder.WriteString("\n")

	for _, r := range results {
		if r.Status == "OK" {
			OK += 1
		} else {
			FAIL += 1
		}

		latest := "-"
		if !r.LatestEventTime.IsZero() {
			latest = r.LatestEventTime.Format("2006-01-02 15:04:05 MST")
		}

		errText := "-"
		if r.Error != "" {
			errText = r.Error
		}

		fmt.Fprintf(
			&builder,
			"%-20s %-6s %-6d %-25s %-12s %-30s\n",
			r.Source,
			r.Status,
			r.EventCount,
			latest,
			r.Duration.String(),
			errText,
		)
	}
	fmt.Fprintf(&builder, "Summary: %d sources checked, %d OK, %d FAIL\n", len(results), OK, FAIL)
	builder.WriteString(line)

	return builder.String() // builder is not a string type
}
