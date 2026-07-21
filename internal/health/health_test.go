package health

import (
	"AegisGeo/internal/models"
	"testing"
)

type fakeClient struct{}

func (f fakeClient) FetchLatest() ([]models.Event, error) {
	// Fetch two data
	return []models.Event{
		{ID: "event-1"},
		{ID: "event-2"},
	}, nil
}

func (f fakeClient) GetName() string {
	return "FakeClient"
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
