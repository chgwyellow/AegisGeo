package ingestion

import (
	"AegisGeo/internal/models"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

type TsunamiClient struct {
	apiURL string
}

func NewTsunamiClient(apiURL string) *TsunamiClient {
	return &TsunamiClient{
		apiURL: apiURL,
	}
}

func (t *TsunamiClient) GetName() string {
	return "NOAA-Tsunami"
}

type noaaTsunamiResponse struct {
	Items []struct {
		ID                   int     `json:"id"`
		Year                 int     `json:"year"`
		Month                int     `json:"month"`
		Day                  int     `json:"day"`
		Hour                 int     `json:"hour"`
		Minute               int     `json:"minute"`
		MaxWaterHeight       float64 `json:"maxWaterHeight"`
		TsunamiIntensity     float64 `json:"tsunamiIntensity"`
		Country              string  `json:"country"`
		LocationName         string  `json:"locationName"`
		Latitude             float64 `json:"latitude"`
		Longitude            float64 `json:"longitude"`
		AssociatedEarthquake int     `json:"earthquakeId"`
	} `json:"items"`
}

func (t *TsunamiClient) FetchLatest() ([]models.Event, error) {
	resp, err := http.Get(t.apiURL)
	if err != nil {
		return nil, fmt.Errorf("NOAA Connection Fail: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("NOAA Server responses false status code: %d", resp.StatusCode)
	}

	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("Fail to load NOAA Json: %v", err)
	}

	var raw noaaTsunamiResponse
	err = json.Unmarshal(bodyBytes, &raw)
	if err != nil {
		return nil, fmt.Errorf("Fail to parse NOAA Json: %v", err)
	}

	events := make([]models.Event, 0, len(raw.Items))
	for _, item := range raw.Items {
		eventTime := time.Date(item.Year, time.Month(item.Month), item.Day, item.Hour, item.Minute, 0, 0, time.UTC)

		standardEvent := models.Event{
			ID:        fmt.Sprintf("NOAA-TSU-%d", item.ID),
			Source:    "NOAA",
			Type:      "Tsunami",
			Title:     fmt.Sprintf("Tsunami Monitor, Max Water Height %.2fm", item.MaxWaterHeight),
			Magnitude: item.MaxWaterHeight,
			Depth:     0.0,
			Timestamp: eventTime,
			Country:   item.Country,
			Location:  item.LocationName,
			Latitude:  item.Latitude,
			Longitude: item.Longitude,
			Details: map[string]any{
				"tsunami_intensity":     item.TsunamiIntensity,
				"associated_earthquake": item.AssociatedEarthquake,
			},
		}

		events = append(events, standardEvent)
	}
	return events, nil
}
