package health

import (
	"AegisGeo/internal/ingestion"
	"AegisGeo/internal/models"
	"errors"
	"strings"
	"testing"
	"time"
)

type fakeClient struct{}

type failingClient struct{}

type slowClient struct{}

func (f fakeClient) FetchLatest() ([]models.Event, error) {
	// Fetch two data
	return []models.Event{
		{
			ID:        "event-1",
			Timestamp: time.Date(2026, 7, 21, 9, 0, 0, 0, time.UTC),
		},
		{
			ID:        "event-2",
			Timestamp: time.Date(2026, 7, 21, 10, 0, 0, 0, time.UTC),
		},
	}, nil
}

func (f fakeClient) GetName() string {
	return "FakeClient"
}

func (f failingClient) FetchLatest() ([]models.Event, error) {
	// Simulate a failed fetch
	return nil, errors.New("fetch failed") // create an error object
}

func (f failingClient) GetName() string {
	return "FailingClient"
}

func (s slowClient) FetchLatest() ([]models.Event, error) {
	time.Sleep(10 * time.Millisecond)

	return []models.Event{
		{ID: "event-1"},
	}, nil
}

func (s slowClient) GetName() string {
	return "SlowClient"
}

// Test for events number
func TestBuildHealthResultCountsEvents(t *testing.T) {
	client := fakeClient{}

	result := BuildHealthResult(client)

	if result.EventCount != 2 {
		t.Errorf("expected EventCount %d, got %d", 2, result.EventCount)
	}
}

// Test for source name
func TestBuildHealthResultIncludesClientName(t *testing.T) {
	client := fakeClient{}

	result := BuildHealthResult(client)

	if result.Source != "FakeClient" {
		t.Errorf("expected Source %q, got %q", "FakeClient", result.Source)
	}
}

// Test for status code ok
func TestBuildHealthResultReturnsOKStatusWhenFetchSucceeds(t *testing.T) {
	client := fakeClient{}

	result := BuildHealthResult(client)

	if result.Status != "OK" {
		t.Errorf("expected Status %q, got %q", "OK", result.Status)
	}

}

// Test for fail response
func TestBuildHealthResultReturnsFailStatusWhenFetchFails(t *testing.T) {
	client := failingClient{}

	result := BuildHealthResult(client)

	if result.Status != "FAIL" {
		t.Errorf("expected Status %q, got %q", "FAIL", result.Status)
	}
}

// Test for error message
func TestBuildHealthResultIncludesErrorMessageWhenFetchFails(t *testing.T) {
	client := failingClient{}

	result := BuildHealthResult(client)

	if result.Error != "fetch failed" {
		t.Errorf("expected Error %q, got %q", "fetch failed", result.Error)
	}
}

// Test latest data
func TestBuildHealthResultIncludesLatestEventTime(t *testing.T) {
	client := fakeClient{}

	result := BuildHealthResult(client)

	wantTime := time.Date(2026, 7, 21, 10, 0, 0, 0, time.UTC)
	if !result.LatestEventTime.Equal(wantTime) {
		t.Errorf("expected LatestEventTime %v, got %v", wantTime, result.LatestEventTime)
	}
}

// Test fetch duration
func TestBuildHealthResultIncludesDuration(t *testing.T) {
	client := slowClient{}

	result := BuildHealthResult(client)

	if result.Duration <= 0 {
		t.Errorf("expected Duration to be greater than 0, got %v", result.Duration)
	}
}

// Test client numbers
func TestBuildHealthResultsReturnsResultForEachClient(t *testing.T) {
	clients := []ingestion.IngestionClient{
		fakeClient{},
		fakeClient{},
	}

	results := BuildHealthResults(clients)

	if len(results) != 2 {
		t.Errorf("expected 2 results, got %d", len(results))
	}
}

// Test client sequence
func TestBuildHealthResultsPreservesClientOrder(t *testing.T) {
	clients := []ingestion.IngestionClient{
		fakeClient{},
		failingClient{},
	}

	results := BuildHealthResults(clients)

	if len(results) != 2 {
		t.Fatalf("expected 2 results, got %d", len(results))
	}

	if results[0].Source != "FakeClient" {
		t.Errorf("expected %q, got %q", "FakeClient", results[0].Source)
	}

	if results[1].Source != "FailingClient" {
		t.Errorf("expected second Source %q, got %q", "FailingClient", results[1].Source)
	}
}

// Test result formatting
func TestFormatHealthResultsIncludesSourceStatusAndCount(t *testing.T) {
	results := []HealthResult{
		{
			Source:     "USGS",
			Status:     "OK",
			EventCount: 2,
		},
	}

	output := FormatHealthResults(results)

	if !strings.Contains(output, "USGS") {
		t.Errorf("expected output to contain %q, got %q", "USGS", output)
	}

	if !strings.Contains(output, "OK") {
		t.Errorf("expected output to contain %q, got %q", "OK", output)
	}

	if !strings.Contains(output, "2") {
		t.Errorf("expected output to contain %q, got %q", "2", output)
	}

	if !strings.Contains(output, "Source") {
		t.Errorf("expected output to contain %q, got %q", "Source", output)
	}
}
