package ingestion

import (
	"AegisGeo/internal/models"
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
)

type NwsSevereWeatherClient struct {
	apiURL string
	Email  string
}

func NewNwsSevereWeatherClient(apiURL string, email string) *NwsSevereWeatherClient {
	return &NwsSevereWeatherClient{
		apiURL: apiURL,
		Email:  email,
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
	now := time.Now().UTC()
	todayMidnight := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, time.UTC)

	startTimeStr := todayMidnight.Format("2006-01-02T15:04:05Z")

	baseURL := n.apiURL
	if idx := strings.Index(baseURL, "/alerts"); idx != -1 {
		baseURL = baseURL[:idx]
	}
	dynamicURL := fmt.Sprintf("%s/alerts?start=%s&event=Tornado%%20Watch,Tornado%%20Warning,Severe%%20Thunderstorm%%20Watch,Severe%%20Thunderstorm%%20Warning",
		baseURL, startTimeStr)

	req, err := http.NewRequest("GET", dynamicURL, nil)
	if err != nil {
		return nil, fmt.Errorf("Fail to sent request: %v", err)
	}
	userAgent := fmt.Sprintf("AegisGeoSevereWeatherEngine/1.0 (contact:%s)", n.Email)
	req.Header.Set("User-Agent", userAgent)

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

		cleanID := p.ID
		idParts := strings.SplitN(p.ID, ":", 3)
		if len(idParts) == 3 {
			cleanID = idParts[2]
		}

		standardEvent := models.Event{
			ID:        fmt.Sprintf("NWS-SEV-%s", cleanID),
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
				"storm_description": strings.Join(strings.Fields(p.Description), " "),
				"original_urn":      p.ID,
			},
		}
		events = append(events, standardEvent)
	}
	return events, nil
}
