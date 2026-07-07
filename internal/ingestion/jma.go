package ingestion

import (
	"AegisGeo/internal/models"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

type jmaClient struct {
	apiURL string
}

func NewJmaClient(apiURL string) *jmaClient {
	return &jmaClient{
		apiURL: apiURL,
	}
}

func (j *jmaClient) GetName() string {
	return "JMA"
}

type jmaRawResponse []struct {
	ID         string `json:"id"`
	Time       string `json:"time"`
	Earthquake struct {
		OriginTime string `json:"time"`
		Hypocenter struct {
			Name      string  `json:"name"`
			Depth     float64 `json:"depth"`
			Latitude  float64 `json:"latitude"`
			Longitude float64 `json:"longitude"`
			Magnitude float64 `json:"magnitude"`
		} `json:"hypocenter"`
	} `json:"earthquake"`
}

func (j *jmaClient) FetchLatest() ([]models.Event, error) {
	resp, err := http.Get(j.apiURL)
	if err != nil {
		return nil, fmt.Errorf("JAM connection failed: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("JAM server rejected to connect: %d", resp.StatusCode)
	}

	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("Loaded JAM fail: %v", err)
	}

	var raw jmaRawResponse
	err = json.Unmarshal(bodyBytes, &raw)
	if err != nil {
		return nil, fmt.Errorf("Fail to parse JMA Json: %v", err)
	}

	events := make([]models.Event, 0, len(raw))
	for _, item := range raw {
		if item.Earthquake.Hypocenter.Name == "" {
			continue
		}
		timeStrWithZone := fmt.Sprintf("%s +0900", item.Earthquake.OriginTime)
		t, err := time.Parse("2006/01/02 15:04:05 -0700", timeStrWithZone)
		if err != nil {
			t = time.Now()
		}

		standardEvent := models.Event{
			ID:        fmt.Sprintf("JMA-%s", item.ID),
			Source:    "JMA",
			Type:      "Earthquake",
			Title:     item.Earthquake.Hypocenter.Name,
			Magnitude: item.Earthquake.Hypocenter.Magnitude,
			Depth:     item.Earthquake.Hypocenter.Depth,
			Timestamp: t,
			Country:   "JP",
			Location:  item.Earthquake.Hypocenter.Name,
			Longitude: item.Earthquake.Hypocenter.Longitude,
			Latitude:  item.Earthquake.Hypocenter.Latitude,
		}

		events = append(events, standardEvent)
	}
	return events, nil
}
