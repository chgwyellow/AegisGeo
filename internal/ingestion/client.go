package ingestion

import (
	"AegisGeo/internal/models"
)

type IngestionClient interface {
	// Catch the latest event and transfer to Event slice
	FetchLatest() ([]models.Event, error)
	// Return the source
	GetName() string
}
