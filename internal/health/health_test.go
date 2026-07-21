package health

import (
	"AegisGeo/internal/models"
	"testing"
)

type fakeClient struct {}

func (f fakeClient) FetchLatest() ([]models.Event, error) {
	return []models.Event{
		{ID: "event-1"},
		{ID: "event-2"},
	}, nil
}

func (f fakeClient) GetName() string {
	return "FakeClient"
}

func TestBuildHealthResultCountsEvents(t *testing.T) {
	client := fakeClient{}

	result := BuildHealthResult(client)

	if result.EventCount != 2 {
		t.Errorf("expected EventCount %d, got %d", 2, result.EventCount)
	}
}
