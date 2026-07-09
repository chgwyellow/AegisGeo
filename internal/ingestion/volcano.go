package ingestion

import (
	"AegisGeo/internal/models"
	"encoding/xml"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
)

type VolcanoClient struct {
	apiURL string
}

func NewVolcanoClient(apiURL string) *VolcanoClient {
	return &VolcanoClient{
		apiURL: apiURL,
	}
}

func (v *VolcanoClient) GetName() string {
	return "USGS-Volcano"
}

type volcanoRawResponse struct {
	XMLName xml.Name `xml:"rss"`
	Channel struct {
		Title       string `xml:"title"`
		Description string `xml:"description"`
		Link        string `xml:"link"`
		PubDate     string `xml:"pubDate"`
		Items       []struct {
			Author      string `xml:"author"`
			Title       string `xml:"title"`
			Description string `xml:"description"`
			Link        string `xml:"link"`
			GUID        string `xml:"guid"`
			PubDate     string `xml:"pubDate"`
		} `xml:"item"`
	} `xml:"channel"`
}

func (v *VolcanoClient) FetchLatest() ([]models.Event, error) {
	resp, err := http.Get(v.apiURL)
	if err != nil {
		return nil, fmt.Errorf("Volcano connection fail: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("Volcano server error status code: %d", resp.StatusCode)
	}

	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("Fail to read Volcano body: %v", err)
	}

	var raw volcanoRawResponse
	if err := xml.Unmarshal(bodyBytes, &raw); err != nil {
		return nil, fmt.Errorf("Fail to parse Volcano XML: %v", err)
	}

	events := make([]models.Event, 0, len(raw.Channel.Items))

	for _, item := range raw.Channel.Items {
		t, err := time.Parse(time.RFC1123Z, strings.TrimSpace(item.PubDate))
		if err != nil {
			t = time.Now().UTC()
		}
		cleanTitle := strings.Join(strings.Fields(item.Title), " ")

		standardEvent := models.Event{
			ID:        fmt.Sprintf("USGS-VOL-%s", item.GUID),
			Source:    "USGS",
			Type:      "Volcano",
			Title:     cleanTitle,
			Magnitude: 0.0,
			Depth:     0.0,
			Timestamp: t,
			Country:   "US",
			Location:  cleanTitle,
			Latitude:  0.0,
			Longitude: 0.0,
			Details: map[string]any{
				"author":      item.Author,
				"description": item.Description,
				"url_link":    item.Link,
			},
		}
		events = append(events, standardEvent)
	}
	return events, nil
}
