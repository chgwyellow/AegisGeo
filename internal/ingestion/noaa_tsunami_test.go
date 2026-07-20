package ingestion

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestTsunamiClientFetchLatestTransformsRawResponseToEvents(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{
			"items": [
				{
					"id": 1001,
					"year": 2026,
					"month": 7,
					"day": 16,
					"hour": 10,
					"minute": 30,
					"second": 15,
					"maxWaterHeight": 2.5,
					"tsunamiIntensity": 1.2,
					"country": "Japan",
					"locationName": "Honshu, Japan",
					"latitude": 38.1,
					"longitude": 142.9,
					"eqDepth": 20,
					"eqMagnitude": 7.1,
					"earthquakeEventId": 555,
					"causeCode": 1,
					"deathsTotal": 0,
					"damageMillionsDollarsTotal": 0.5
				}
			]
		}`))
	}))

	defer server.Close()

	client := NewTsunamiClient(server.URL)

	events, err := client.FetchLatest()
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if len(events) != 1 {
		t.Fatalf("expected 1 event, got %d", len(events))
	}

	event := events[0]

	if event.ID != "NOAA-TSU-1001" {
		t.Errorf("expected ID %q, got %q", "NOAA-TSU-1001", event.ID)
	}

	if event.Source != "NOAA" {
		t.Errorf("expected Source %q, got %q", "NOAA", event.Source)
	}

	if event.Type != "Tsunami" {
		t.Errorf("expected Type %q, got %q", "Tsunami", event.Type)
	}

	if event.Title != "Tsunami Monitor, Max Water Height 2.50m" {
		t.Errorf("expected Title %q, got %q", "Tsunami Monitor, Max Water Height 2.50m", event.Title)
	}

	if event.Magnitude != 2.5 {
		t.Errorf("expected Magnitude %v, got %v", 2.5, event.Magnitude)
	}

	if event.Depth != 20 {
		t.Errorf("expected Depth %v, got %v", 20.0, event.Depth)
	}

	if event.Country != "JP" {
		t.Errorf("expected Country %q, got %q", "JP", event.Country)
	}

	if event.Location != "Honshu, Japan" {
		t.Errorf("expected Location %q, got %q", "Honshu, Japan", event.Location)
	}

	if event.Latitude != 38.1 {
		t.Errorf("expected Latitude %v, got %v", 38.1, event.Latitude)
	}

	if event.Longitude != 142.9 {
		t.Errorf("expected Longitude %v, got %v", 142.9, event.Longitude)
	}

	wantTime := time.Date(2026, time.July, 16, 10, 30, 15, 0, time.UTC)
	if !event.Timestamp.Equal(wantTime) {
		t.Errorf("expected Timestamp %v, got %v", wantTime, event.Timestamp)
	}

	if event.Details["tsunami_intensity"] != 1.2 {
		t.Errorf("expected tsunami_intensity %v, got %v", 1.2, event.Details["tsunami_intensity"])
	}

	if event.Details["eq_magnitude"] != 7.1 {
		t.Errorf("expected eq_magnitude %v, got %v", 7.1, event.Details["eq_magnitude"])
	}

	if event.Details["earthquake_event_id"] != 555 {
		t.Errorf("expected earthquake_event_id %v, got %v", 555, event.Details["earthquake_event_id"])
	}

	if event.Details["cause_code"] != 1 {
		t.Errorf("expected cause_code %v, got %v", 1, event.Details["cause_code"])
	}

	if event.Details["deaths_total"] != 0 {
		t.Errorf("expected deaths_total %v, got %v", 0, event.Details["deaths_total"])
	}

	if event.Details["damage_millions_dollars_total"] != 0.5 {
		t.Errorf("expected damage_millions_dollars_total %v, got %v", 0.5, event.Details["damage_millions_dollars_total"])
	}

}
