// CWA data ingestion
package ingestion

import (
	"AegisGeo/internal/models"
	"encoding/json"
	"fmt"
	"time"
)

type CwaClient struct {
	apiURL string
	token  string // key
}

// Create Cwa Client
func NewCwaClient(apiURL, token string) *CwaClient {
	return &CwaClient{
		apiURL: apiURL,
		token:  token,
	}
}

func (c *CwaClient) GetName() string {
	return "CWA"
}

// CWA raw json structure
type cwaRawResponse struct {
	Records struct {
		Earthquake []struct {
			No             string `json:"EarthquakeNo"`
			ReportContent  string `json:"ReportContent"`
			OriginTime     string `json:"OriginTime"`
			EarthquakeInfo struct {
				Magnitude float64 `json:"EarthquakeMagnitude"`
				Depth     float64 `json:"Depth"`
			} `json:"EarthquakeInfo"`
		} `json:"Earthquake"`
	} `json:"Records"`
}

func (c *CwaClient) FetchLatest() ([]models.Event, error) {
	fakeCwaJson := `{
		"Records": {
			"Earthquake": [
				{
					"EarthquakeNo": "115001",
					"ReportContent": "臺灣東部海域顯著有感地震",
					"OriginTime": "2026-07-06 10:00:00",
					"EarthquakeInfo": {
						"EarthquakeMagnitude": 5.8,
						"Depth": 15.4
					}
				}
			]
		}
	}`

	// parsing original data
	var raw cwaRawResponse
	err := json.Unmarshal([]byte(fakeCwaJson), &raw) // parsing data and store to pointed v
	if err != nil {
		return nil, fmt.Errorf("Fail to parse CWA data: %v", err)
	}

	// Transform raw data to models.Event type
	events := make([]models.Event, 0, len(raw.Records.Earthquake))

	for _, eq := range raw.Records.Earthquake {
		t, _ := time.Parse("2026-01-02 15:04:06", eq.OriginTime)

		standardEvent := models.Event{
			ID:        fmt.Sprintf("CWA-%s", eq.No),
			Source:    "CWA",
			Type:      "Earthquake",
			Title:     eq.ReportContent,
			Magnitude: eq.EarthquakeInfo.Magnitude,
			Depth:     eq.EarthquakeInfo.Depth,
			Timestamp: t,
			Country:   "TW",
		}

		events = append(events, standardEvent)
	}
	return events, nil
}
