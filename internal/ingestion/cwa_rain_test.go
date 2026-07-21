package ingestion

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestCwaRainClientFetchLatestTransformsRawResponseToEvents(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{
		"records": {
			"Station": [
			{
				"StationName": "Taipei Station",
				"StationId": "466920",
				"GeoInfo": {
				"Coordinates": [
					{
					"CoordinateName": "WGS84",
					"StationLongitude": "121.5149",
					"StationLatitude": "25.0375"
					}
				],
				"CountyName": "Taipei City",
				"TownName": "Zhongzheng District"
				},
				"RainfallElement": {
				"Now": {
					"Precipitation": "45.5"
				}
				},
				"ObsTime": {
				"DateTime": "2026-07-21T09:30:00+08:00"
				}
			}
			]
		}
		}`))
	}))
	defer server.Close()

	client := NewCwaRainClient(server.URL, "fake_token")

	events, err := client.FetchLatest()
	if err != nil {
		t.Fatalf("expected no err, got %v", err)
	}

	if len(events) != 1 {
		t.Fatalf("expected 1 event, got %v events", len(events))
	}

	event := events[0]

	if event.ID != "CWA-RAIN-466920" {
		t.Errorf("expected ID %v, got %v", "CWA-RAIN-466920", event.ID)
	}

	if event.Title != "Taipei Station" {
		t.Errorf("expected Title %v, got %v", "Taipei Station", event.Title)
	}

	if event.Magnitude != 45.5 {
		t.Errorf("expected Magnitude %v, got %v", 45.5, event.Magnitude)
	}

	if event.Location != "Taipei CityZhongzheng District" {
		t.Errorf("expected Location %v, got %v", "Taipei CityZhongzheng District", event.Location)
	}

	if event.Latitude != 25.0375 {
		t.Errorf("expected Latitude %v, got %v", 25.0375, event.Latitude)
	}

	if event.Longitude != 121.5149 {
		t.Errorf("expected Longitude %v, got %v", 121.5149, event.Longitude)
	}

	wantTime, err := time.Parse(time.RFC3339, "2026-07-21T09:30:00+08:00")
	if err != nil {
		t.Fatalf("failed to prepare expected timestamp: %v", err)
	}

	if !event.Timestamp.Equal(wantTime) {
		t.Errorf("expected Time %v, got %v", wantTime, event.Timestamp)
	}

	if event.Details["Warning"] != "Heavy Rain Advisory" {
		t.Errorf("expected Warning %v, got %v", "Heavy Rain Advisory", event.Details["Warning"])
	}
}
