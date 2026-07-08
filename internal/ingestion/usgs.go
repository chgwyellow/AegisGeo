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

// Geographic keyword dict
var countryDictionary = map[string]string{
	// TW
	"TAIWAN": "TW",
	"JIUFEN": "TW",

	// US State
	"ALASKA":         "US",
	"HAWAII":         "US",
	"COLORADO":       "US",
	"CALIFORNIA":     "US",
	"NEW MEXICO":     "US",
	"VIRGIN ISLANDS": "US",
	"OKLAHOMA":       "US",

	// Countries
	"TIMOR LESTE": "TL",
	"MAURITIUS":   "MU",
	"PHILIPPINES": "PH",
	"JAPAN":       "JP",
	"CHILE":       "CL",
	"TONGA":       "TO",
	"FIJI":        "FJ",
	"COSTA RICA":  "CR",
	"INDONESIA":   "ID",
	"CHINA":       "CN",
	"RUSSIA":      "RU",
	"NICARAGUA":   "NI",
	"PAKISTAN":    "PK",
	"PUERTO RICO": "PR",
	"TAJIKISTAN":  "TJ",
	"GREECE":      "GR",
	"NEW ZEALAND": "NZ",
	"ARGENTINA":   "AR",
	"VENEZUELA":   "VE",
	"ETHIOPIA":    "ET",
}

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

	// Iterate dict to match keyword
	for keyword, isoCode := range countryDictionary {
		if strings.Contains(placeUpper, keyword) {
			return isoCode
		}
	}

	// USA states
	if strings.Contains(placeUpper, ", CA") ||
		strings.Contains(placeUpper, ", AK") ||
		strings.Contains(placeUpper, ", HI") ||
		strings.Contains(placeUpper, ", CO") ||
		strings.Contains(placeUpper, ", NV") ||
		strings.Contains(placeUpper, "USA") {
		return "US"
	}

	// Ocean character (e.g. Ridge, Trench, Ocean )
	if strings.Contains(placeUpper, "RIDGE") ||
		strings.Contains(placeUpper, "TRENCH") ||
		strings.Contains(placeUpper, "BASIN") ||
		strings.Contains(placeUpper, "OCEAN") {
		return "OCEAN"
	}

	return "OCEAN"
}
