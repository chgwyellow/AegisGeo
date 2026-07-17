package ingestion

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

// Test normal json data transform to Events type
func TestJmaClientFetchLatestTransformsRawResponseToEvents(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`[
			{
				"id": "20260716143000",
				"time": "2026/07/16 14:31:00",
				"earthquake": {
					"time": "2026/07/16 14:30:00",
					"hypocenter": {
						"name": "Ishikawa Prefecture",
						"depth": 10,
						"latitude": 37.5,
						"longitude": 137.2,
						"magnitude": 5.1
					}
				}
			}
		]`))
	}))
	defer server.Close()

	client := NewJmaClient(server.URL)

	events, err := client.FetchLatest()
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if len(events) != 1 {
		t.Fatalf("expected 1 event, got %d", len(events))
	}

	event := events[0]

	if event.ID != "JMA-20260716143000" {
		t.Errorf("expected ID %q, got %q", "JMA-20260716143000", event.ID)
	}

	if event.Source != "JMA" {
		t.Errorf("expected Source %q, got %q", "JMA", event.Source)
	}

	if event.Type != "Earthquake" {
		t.Errorf("expected Type %q, got %q", "Earthquake", event.Type)
	}

	if event.Title != "Ishikawa Prefecture" {
		t.Errorf("expected Title %q, got %q", "Ishikawa Prefecture", event.Title)
	}

	if event.Magnitude != 5.1 {
		t.Errorf("expected Magnitude %v, got %v", 5.1, event.Magnitude)
	}

	if event.Depth != 10 {
		t.Errorf("expected Depth %v, got %v", 10.0, event.Depth)
	}

	if event.Country != "JP" {
		t.Errorf("expected Country %q, got %q", "JP", event.Country)
	}

	if event.Location != "Ishikawa Prefecture" {
		t.Errorf("expected Location %q, got %q", "Ishikawa Prefecture", event.Location)
	}

	if event.Latitude != 37.5 {
		t.Errorf("expected Latitude %v, got %v", 37.5, event.Latitude)
	}

	if event.Longitude != 137.2 {
		t.Errorf("expected Longitude %v, got %v", 137.2, event.Longitude)
	}

	wantTime, err := time.Parse("2006/01/02 15:04:05 -0700", "2026/07/16 14:30:00 +0900")
	if err != nil {
		t.Fatalf("failed to prepare expected timestamp: %v", err)
	}

	if !event.Timestamp.Equal(wantTime) {
		t.Errorf("expected Timestamp %v, got %v", wantTime, event.Timestamp)
	}
}

// Test the empty name, which will jump out the loop
func TestJmaClientFetchLatestSkipsEventWhenHypocenterNameIsEmpty(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`[
			{
				"id": "20260716143000",
				"time": "2026/07/16 14:31:00",
				"earthquake": {
					"time": "2026/07/16 14:30:00",
					"hypocenter": {
						"name": "",
						"depth": 10,
						"latitude": 37.5,
						"longitude": 137.2,
						"magnitude": 5.1
					}
				}
			}
		]`))
	}))
	defer server.Close()

	client := NewJmaClient(server.URL)

	events, err := client.FetchLatest()
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if len(events) != 0 {
		t.Fatalf("expected 0 events, got %d", len(events))
	}
}
