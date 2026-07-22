package health

import (
	"AegisGeo/internal/models"
	"errors"
	"testing"
)

type fakeClient struct{}

type failingClient struct{}

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

func (f failingClient) FetchLatest() ([]models.Event, error) {
	// Simulate a failed fetch
	return nil, errors.New("fetch failed")
}

func (f failingClient) GetName() string {
	return "FailingClient"
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
