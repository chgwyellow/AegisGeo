package ingestion

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestCwaClientFetchLatestTransformsRawResponseToEvents(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{
		"records": {
			"Earthquake": [
			{
				"EarthquakeNo": 20260720001,
				"ReportContent": "07/20 09:15 Hualien County earthquake, magnitude 5.3",
				"EarthquakeInfo": {
				"OriginTime": "2026-07-20T09:15:00+08:00",
				"FocalDepth": 18.5,
				"Epicenter": {
					"Location": "Hualien County",
					"EpicenterLatitude": 23.98,
					"EpicenterLongitude": 121.61
				},
				"EarthquakeMagnitude": {
					"MagnitudeValue": 5.3
				}
				}
			}
			]
		}
	}`))
	}))
	defer server.Close()

	client := NewCwaClient(server.URL, "test_token")

	events, err := client.FetchLatest()
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if len(events) != 1 {
		t.Fatalf("expected 1 event, got %d", len(events))
	}

	event := events[0]

	if event.ID != "CWA-20260720001" {
		t.Errorf("expected ID %q, got %q", "CWA-20260720001", event.ID)
	}

	if event.Title != "07/20 09:15 Hualien County earthquake, magnitude 5.3" {
		t.Errorf("expected Title %q, got %q", "07/20 09:15 Hualien County earthquake, magnitude 5.3", event.Title)
	}

	if event.Magnitude != 5.3 {
		t.Errorf("expected Magnitude %v, got %v", 5.1, event.Magnitude)
	}

	if event.Depth != 18.5 {
		t.Errorf("expected Depth %v, got %v", 18.5, event.Depth)
	}

	if event.Location != "Hualien County" {
		t.Errorf("expected Location %v, got %v", "Hualien County", event.Location)
	}

	if event.Latitude != 23.98 {
		t.Errorf("expected Latitude %v, got %v", 23.98, event.Latitude)
	}

	if event.Longitude != 121.61 {
		t.Errorf("expected Longitude %v, got %v", 121.61, event.Longitude)
	}

	wantTime, err := time.Parse(time.RFC3339, "2026-07-20T09:15:00+08:00")
	if err != nil {
		t.Fatalf("failed to prepare expected timestamp: %v", err)
	}

	if !event.Timestamp.Equal(wantTime) {
		t.Errorf("expected Timestamp %v, got %v", wantTime, event.Timestamp)
	}
}
