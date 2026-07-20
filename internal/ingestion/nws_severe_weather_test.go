package ingestion

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestNwsSevereWeatherClientFetchLatestTransformsRawResponseToEvents(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{
		"features": [
			{
			"properties": {
				"id": "urn:oid:2.49.0.1.840.0.test-alert-001",
				"event": "Tornado Warning",
				"headline": "Tornado Warning issued for Oklahoma County",
				"description": "A severe thunderstorm capable of producing a tornado was located near Oklahoma City.\nTake shelter immediately.",
				"onset": "2026-07-20T09:30:00Z",
				"areaDesc": "Oklahoma County"
				}
			}
		]
		}`))
	}))
	defer server.Close()

	client := NewNwsSevereWeatherClient(server.URL, "test@gmail.com")

	events, err := client.FetchLatest()
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if len(events) != 1 {
		t.Fatalf("expected 1 event, got %d", len(events))
	}

	event := events[0]

	if event.ID != "NWS-SEV-2.49.0.1.840.0.test-alert-001" {
		t.Errorf("expected ID %q, got %q", "NWS-SEV-2.49.0.1.840.0.test-alert-001", event.ID)
	}

	if event.Details["event_type"] != "Tornado Warning" {
		t.Errorf("expected type %v, got %v", "Tornado Warning", event.Details["event_type"])
	}

	if event.Title != "Tornado Warning issued for Oklahoma County" {
		t.Errorf("expected title %v, got %v", "Tornado Warning issued for Oklahoma County", event.Title)
	}

	wantTime, err := time.Parse("2006-01-02T15:04:05Z", "2026-07-20T09:30:00Z")
	if err != nil {
		t.Fatalf("failed to prepare expected timestamp: %v", err)
	}

	if !event.Timestamp.Equal(wantTime) {
		t.Errorf("expected Timestamp %v, got %v", wantTime, event.Timestamp)
	}

	if event.Location != "Oklahoma County" {
		t.Fatalf("expected location %v, got %v", "Oklahoma County", event.Location)
	}

	if event.Details["storm_description"] != "A severe thunderstorm capable of producing a tornado was located near Oklahoma City. Take shelter immediately." {
		t.Fatalf("expected description %v, got %v", "A severe thunderstorm capable of producing a tornado was located near Oklahoma City.\nTake shelter immediately.", event.Details["storm_description"])
	}
}
