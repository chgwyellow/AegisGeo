package ingestion

import (
	"AegisGeo/internal/models"
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

type NwsSevereWeatherClient struct {
	apiURL string
}

func NewNwsSevereWeatherClient(apiURL string) *NwsSevereWeatherClient {
	return &NwsSevereWeatherClient{
		apiURL: apiURL,
	}
}

func (n *NwsSevereWeatherClient) GetName() string {
	return "NWS-SevereWeather"
}

type nwsAlertResponse struct {
	Features []struct {
		Properties struct {
			ID          string    `json:"id"`
			Event       string    `json:"event"`
			Headline    string    `json:"headline"` // Title
			Description string    `json:"description"`
			Onset       time.Time `json:"onset"` // Alert start time
			AreaDesc    string    `json:"areaDesc"`
		} `json:"properties"`
	} `json:"features"`
}

func (n *NwsSevereWeatherClient) FetchLatest() ([]models.Event, error) {
	req, err := http.NewRequest("GET", n.apiURL, nil)
	if err != nil {
		return nil, fmt.Errorf("Fail to sent request: %v", err)
	}
	req.Header.Set("User-Agent", "AegisGeoSevereWeatherEngine/1.0 (contact@aegisgeo.com)")

	client := &http.Client{} // Create customized client
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("NWS connection failed: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("NWS Server returned error code: %d", resp.StatusCode)
	}

	var logBuffer bytes.Buffer

	teeReader := io.TeeReader(resp.Body, &logBuffer)

	var raw nwsAlertResponse
	if err := json.NewDecoder(teeReader).Decode(&raw); err != nil {
		return nil, fmt.Errorf("NWS JSON decode failed: %v", err)
	}

	events := make([]models.Event, 0, len(raw.Features))

	for _, f := range raw.Features {
		p := f.Properties

		standardEvent := models.Event{
			ID:        fmt.Sprintf("NWS-SEV-%s", p.ID),
			Source:    "NWS",
			Type:      "SevereWeather",
			Title:     p.Headline,
			Magnitude: 0.0,
			Depth:     0.0,
			Timestamp: p.Onset,
			Country:   "US",
			Location:  p.AreaDesc,
			Latitude:  0.0,
			Longitude: 0.0,
			Details: map[string]any{
				"event_type":        p.Event,
				"storm_description": p.Description,
			},
		}
		events = append(events, standardEvent)
	}
	return events, nil
}
