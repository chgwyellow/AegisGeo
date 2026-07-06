// CWA data ingestion
package ingestion

import (
	"AegisGeo/internal/models"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
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
			No             int64  `json:"EarthquakeNo"`
			ReportContent  string `json:"ReportContent"`
			EarthquakeInfo struct {
				OriginTime string  `json:"OriginTime"`
				Depth      float64 `json:"FocalDepth"`
				Epicenter  struct {
					Location  string  `json:"Location"`
					Latitude  float64 `json:"EpicenterLatitude"`
					Longitude float64 `json:"EpicenterLongitude"`
				}
				EarthquakeMagnitude struct {
					Magnitude float64 `json:"MagnitudeValue"`
				} `json:"EarthquakeMagnitude"`
			} `json:"EarthquakeInfo"`
		} `json:"Earthquake"`
	} `json:"records"`
}

func (c *CwaClient) FetchLatest() ([]models.Event, error) {
	// Create HTTP GET request
	req, err := http.NewRequest("GET", c.apiURL, nil)
	if err != nil {
		return nil, fmt.Errorf("Fail to create request: %v", err)
	}

	// Add token to Header
	req.Header.Add("Authorization", c.token)

	// Launch HTTP Request
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("Fail to connect %v Server: %v", c.GetName(), err)
	}

	// Release socket to prevent resource run out
	defer resp.Body.Close()

	// Check status code
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("%v Server return fail code: %d", c.GetName(), resp.StatusCode)
	}

	// Load data flow
	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("Load %v response body fail: %v", c.GetName(), err)
	}

	// parsing original data
	var raw cwaRawResponse
	err = json.Unmarshal(bodyBytes, &raw) // parsing data and store to pointed v
	if err != nil {
		return nil, fmt.Errorf("Fail to parse CWA data: %v", err)
	}

	// Transform raw data to models.Event type
	events := make([]models.Event, 0, len(raw.Records.Earthquake))

	for _, eq := range raw.Records.Earthquake {
		// Transform string to time type
		// Time layout doesn't need to know the format formula
		// Using Go's Birthday and time, 2006-01-02 15:04:06
		t, err := time.Parse(time.RFC3339, eq.EarthquakeInfo.OriginTime)

		if err != nil {
			fmt.Printf("Can not parse time '%s': %v\n", eq.EarthquakeInfo.OriginTime, err)
			continue
		}

		standardEvent := models.Event{
			ID:        fmt.Sprintf("CWA-%d", eq.No),
			Source:    "CWA",
			Type:      "Earthquake",
			Title:     eq.ReportContent,
			Magnitude: eq.EarthquakeInfo.EarthquakeMagnitude.Magnitude,
			Depth:     eq.EarthquakeInfo.Depth,
			Timestamp: t,
			Country:   "TW",
			Location:  eq.EarthquakeInfo.Epicenter.Location,
			Latitude:  eq.EarthquakeInfo.Epicenter.Latitude,
			Longitude: eq.EarthquakeInfo.Epicenter.Longitude,
		}

		events = append(events, standardEvent)
	}
	return events, nil
}
