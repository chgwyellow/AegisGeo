package ingestion

import (
	"AegisGeo/internal/models"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
)

type UsgsClient struct {
	apiURL string
}

func NewUsgsClient(apiURL string) *UsgsClient {
	return &UsgsClient{
		apiURL: apiURL,
	}
}

func (u *UsgsClient) GetName() string {
	return "USGS"
}

type usgsRawResponse struct {
	Features []struct {
		ID         string `json:"id"`
		Properties struct {
			Mag   float64 `json:"mag"`
			Place string  `json:"place"`
			Time  int64   `json:"time"`
			Title string  `json:"title"`
		} `json:"properties"`
		Geometry struct {
			Coordinates []float64 `json:"coordinates"` //[long, lat, depth]
		} `json:"geometry"`
	} `json:"features"`
}

func (u *UsgsClient) FetchLatest() ([]models.Event, error) {
	// Request without authorization
	resp, err := http.Get(u.apiURL)
	if err != nil {
		return nil, fmt.Errorf("USGS Connection Fail: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("USGS Server responses false status code: %d", resp.StatusCode)
	}

	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("Fail to load USGS Json: %v", err)
	}

	var raw usgsRawResponse
	err = json.Unmarshal(bodyBytes, &raw)
	if err != nil {
		return nil, fmt.Errorf("Fail to parse USGS Json: %v", err)
	}

	events := make([]models.Event, 0, len(raw.Features))
	for _, f := range raw.Features {
		t := time.UnixMilli(f.Properties.Time)

		// Long & Lat
		var depth float64
		if len(f.Geometry.Coordinates) >= 3 {
			depth = f.Geometry.Coordinates[2]
		}

		standardEvent := models.Event{
			ID:        fmt.Sprintf("USGS-%s", f.ID),
			Source:    "USGS",
			Type:      "Earthquake",
			Title:     f.Properties.Title,
			Magnitude: f.Properties.Mag,
			Depth:     depth,
			Timestamp: t,
			Country:   parseCountryFromPlace(f.Properties.Place),
			Location:  f.Properties.Place,
			Latitude:  f.Geometry.Coordinates[1],
			Longitude: f.Geometry.Coordinates[0],
		}

		events = append(events, standardEvent)
	}
	return events, nil
}

// Transfer Country name
func parseCountryFromPlace(place string) string {
	if place == "" {
		return "UNKNOWN"
	}

	placeUpper := strings.ToUpper(place)

	if strings.HasSuffix(placeUpper, ", CA") ||
		strings.HasSuffix(placeUpper, ", AK") ||
		strings.HasSuffix(placeUpper, ", NV") ||
		strings.HasSuffix(placeUpper, "USA") {
		return "US"
	}
	if strings.Contains(placeUpper, "TAIWAN") {
		return "TW"
	}
	if strings.Contains(placeUpper, "CHINA") {
		return "CN"
	}
	if strings.Contains(placeUpper, "FIJI") {
		return "FJ"
	}
	if strings.Contains(placeUpper, "CHILE") {
		return "CL"
	}
	if strings.Contains(placeUpper, "INDONESIA") {
		return "ID"
	}

	return "OCEAN"
}
