package ingestion

import (
	"AegisGeo/internal/models"
	"aegisgeo/internal/models"
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

type CwaRainClient struct {
	apiURL string
	token  string
}

func NewCwaRainClient(apiURL string, token string) *CwaRainClient {
	return &CwaRainClient{
		apiURL: apiURL,
		token:  token,
	}
}

func (c *CwaRainClient) GetName() string {
	return "CWA-RainStation"
}

type cwaRainStationResponse struct {
	Records struct {
		Station []struct {
			StationName string `json:"StationName"`
			StationId   string `json:"StationId"`
			GeoInfo     struct {
				Coordinates []struct {
					CoordinateName   string  `json:"CoordinateName"`
					StationLongitude float64 `json:"StationLongitude"`
					StationLatitude  float64 `json:"StationLatitude"`
				} `json:"Coordinates"`
				CountyName string `json:"CountyName"`
				TownName   string `json:"TownName"`
			} `json:"GeoInfo"`
			RainfallElement struct {
				Now struct {
					Precipitation float64 `json:"Precipitation"`
				} `json:"Now"`
			} `json:"RainfallElement"`
			ObsTime struct {
				DateTime string `json:"DateTime"`
			} `json:"ObsTime"`
		} `json:"Station"`
	} `json:"records"`
}

func (c *CwaRainClient) FetchLatest() ([]models.Event, error) {
	url := fmt.Sprintf("%s?Authorization=%s&format=JSON", c.apiURL, c.token)
	resp, err := http.Get(url)
	if err != nil {
		return nil, fmt.Errorf("Fail to create request: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("%v Server return fail code: %d", c.GetName(), resp.StatusCode)
	}

	var raw cwaRainStationResponse
	if err := json.NewDecoder(resp.Body).Decode(&raw); err != nil {
		return nil, err
	}

	events := make([]models.Event, 0) // We don't know the accurate row count.
	for _, s := range raw.Records.Station {
		currentRain := s.RainfallElement.Now.Precipitation
		t, _ := time.Parse("2006-01-02 15:04:06", s.ObsTime.DateTime)
		alertLevel := "None"

		if currentRain >= 40.0 {
			alertLevel = "Heavy Rain Advisory"

			if currentRain >= 100 {
				alertLevel = "Extremely Heavy Rain Warning"
			}
		}
		standardEvent := models.Event{
			ID:        fmt.Sprintf("CWA-RAIN-%s", s.StationId),
			Source:    "CWA",
			Type:      "Rain",
			Title:     fmt.Sprintf("%s: Precipitation %.1f mm", s.StationName, currentRain),
			Magnitude: currentRain,
			Depth:     0.0,
			Timestamp: t,
			Country:   "TW",
			Location:  fmt.Sprintf("%s%s", s.GeoInfo.CountyName, s.GeoInfo.TownName),
			Latitude:  s.GeoInfo.Coordinates[0].StationLatitude,
			Longitude: s.GeoInfo.Coordinates[0].StationLongitude,
			Details: map[string]any{
				"Warning": alertLevel,
			},
		}
		events = append(events, standardEvent)
	}
	return events, nil
}
