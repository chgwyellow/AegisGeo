package ingestion

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

// Test normal json data transform to Events type
func TestUsgsClientFetchLatestTransformsRawResponseToEvents(t *testing.T) {
	// Create fake server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{
		"features": [
				{
					"id": "test-earthquake-001",
					"properties": {
						"mag": 5.6,
						"place": "10 km S of Hualien City, Taiwan",
						"time": 1784160000000,
						"title": "M 5.6 - 10 km S of Hualien City, Taiwan",
						"tsunami": 0
					},
					"geometry": {
						"coordinates": [121.6, 23.9, 12.3]
					}
				}
			]
		}`,
		))
	}))
	defer server.Close()

	client := NewUsgsClient(server.URL)

	events, err := client.FetchLatest()
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if len(events) != 1 {
		t.Fatalf("expected 1 event, got %d", len(events))
	}

	event := events[0]

	if event.ID != "USGS-test-earthquake-001" {
		t.Errorf("expected ID %q, got %q", "USGS-test-earthquake-001", event.ID)
	}

	if event.Source != "USGS" {
		t.Errorf("expected Source %q, got %q", "USGS", event.Source)
	}

	if event.Type != "Earthquake" {
		t.Errorf("expected Type %q, got %q", "Earthquake", event.Type)
	}

	if event.Title != "M 5.6 - 10 km S of Hualien City, Taiwan" {
		t.Errorf("expected Title %q, got %q", "M 5.6 - 10 km S of Hualien City, Taiwan", event.Title)
	}

	if event.Magnitude != 5.6 {
		t.Errorf("expected Magnitude %v, got %v", 5.6, event.Magnitude)
	}

	if event.Location != "10 km S of Hualien City, Taiwan" {
		t.Errorf("expected Location %q, got %q", "10 km S of Hualien City, Taiwan", event.Location)
	}

	if event.Longitude != 121.6 {
		t.Errorf("expected Longitude %v, got %v", 121.6, event.Longitude)
	}

	if event.Latitude != 23.9 {
		t.Errorf("expected Latitude %v, got %v", 23.9, event.Latitude)
	}

	if event.Depth != 12.3 {
		t.Errorf("expected Depth %v, got %v", 12.3, event.Depth)
	}

	wantTime := time.UnixMilli(1784160000000)
	if !event.Timestamp.Equal(wantTime) {
		t.Errorf("expected Timestamp %v, got %v", wantTime, event.Timestamp)
	}
}

// Test got fail status code
func TestUsgsClientFetchLatestReturnsErrorWhenServerStatusIsNotOK(t *testing.T) {
	// Test got 500 status code
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer server.Close()

	client := NewUsgsClient(server.URL)

	_, err := client.FetchLatest()

	if err == nil {
		t.Fatal("expected error, got nil")
	}
}

// Test broken invalid data
func TestUsgsClientFetchLatestReturnsErrorWhenJSONIsInvalid(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"features": [`))
	}))
	defer server.Close()

	client := NewUsgsClient(server.URL)

	_, err := client.FetchLatest()

	if err == nil {
		t.Fatal("expected error, got nil")
	}
}

// The default depth should be 0 which coordinates has only two elements
func TestUsgsClientFetchLatestDefaultsDepthToZeroWhenDepthIsMissing(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{
		"features": [
				{
					"id": "test-earthquake-001",
					"properties": {
						"mag": 5.6,
						"place": "10 km S of Hualien City, Taiwan",
						"time": 1784160000000,
						"title": "M 5.6 - 10 km S of Hualien City, Taiwan",
						"tsunami": 0
					},
					"geometry": {
						"coordinates": [121.6, 23.9]
					}
				}
			]
		}`,
		))
	}))
	defer server.Close()

	client := NewUsgsClient(server.URL)

	events, err := client.FetchLatest()

	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	event := events[0]

	if event.Depth != 0 {
		t.Fatal("expected 0, got nil")
	}
}
