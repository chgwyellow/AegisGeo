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
			Mag     float64 `json:"mag"`
			Place   string  `json:"place"`
			Time    int64   `json:"time"`
			Title   string  `json:"title"`
			Tsunami int     `json:"tsunami"`
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

		detectedCountry := parseCountryFromPlace(f.Properties.Title)
		if detectedCountry == "OCEAN" {
			detectedCountry = parseCountryFromPlace(f.Properties.Place)
		}

		standardEvent := models.Event{
			ID:        fmt.Sprintf("USGS-%s", f.ID),
			Source:    "USGS",
			Type:      "Earthquake",
			Title:     f.Properties.Title,
			Magnitude: f.Properties.Mag,
			Depth:     depth,
			Timestamp: t,
			Country:   detectedCountry,
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

	// USGS places are typically structured as: "[distance] [direction] of [location], [Country or US State]"
	// If a comma is present, try to match the trimmed last part exactly first.
	if idx := strings.LastIndex(placeUpper, ","); idx != -1 {
		lastPart := strings.TrimSpace(placeUpper[idx+1:])

		// 1. Direct exact match in our dictionary (e.g. "CANADA" -> "CA", "PR" -> "PR")
		if isoCode, exists := countryDictionary[lastPart]; exists {
			return isoCode
		}

		// 2. Check if it's a 2-letter US State postal abbreviation (e.g. "CA", "AK", "NV")
		if usStates[lastPart] {
			return "US"
		}

		// 3. Direct exact match for USA indicator
		if lastPart == "USA" || lastPart == "UNITED STATES" {
			return "US"
		}
	}

	// Fallback to substring matching if exact match of the last part is not found
	for keyword, isoCode := range countryDictionary {
		if strings.Contains(placeUpper, keyword) {
			return isoCode
		}
	}

	// Suffix/Indicator checks
	if strings.Contains(placeUpper, "USA") || strings.Contains(placeUpper, "UNITED STATES") {
		return "US"
	}

	// Ocean indicators (e.g., Ridge, Trench, Basin, Ocean, Rise)
	if strings.Contains(placeUpper, "RIDGE") ||
		strings.Contains(placeUpper, "TRENCH") ||
		strings.Contains(placeUpper, "BASIN") ||
		strings.Contains(placeUpper, "OCEAN") ||
		strings.Contains(placeUpper, "RISE") {
		return "OCEAN"
	}

	return "OCEAN"
}
