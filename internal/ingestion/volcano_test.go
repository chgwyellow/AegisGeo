package ingestion

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

func TestVolcanoClientFetchLatestTransformsRawResponseToEvents(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/XML")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{
			<rss version="2.0">
			<channel>
				<title>USGS Volcano Updates</title>
				<description>Recent volcano activity updates</description>
				<link>https://volcanoes.usgs.gov/</link>
				<pubDate>Mon, 20 Jul 2026 09:00:00 +0000</pubDate>
				<item>
					<author>USGS Volcano Hazards Program</author>
					<title>
						Kilauea Volcano Activity Update
					</title>
					<description>Kilauea volcano remains at advisory level with ongoing monitoring.</description>
					<link>https://volcanoes.usgs.gov/volcanoes/kilauea/status.html</link>
					<guid>kilauea-update-001</guid>
					<pubDate>Mon, 20 Jul 2026 08:30:00 +0000</pubDate>
				</item>
			</channel>
		</rss>
		}`))
	}))
	defer server.Close()

	client := NewVolcanoClient(server.URL)

	events, err := client.FetchLatest()
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if len(events) != 1 {
		t.Fatalf("expected 1 event, got %d", len(events))
	}

	event := events[0]

	if event.ID != "USGS-VOL-kilauea-update-001" {
		t.Errorf("expected ID %q, got %q", "USGS-VOL-kilauea-update-001", event.ID)
	}

	if event.Title != "Kilauea Volcano Activity Update" {
		t.Errorf("expected Title %q, got %q", "Kilauea Volcano Activity Update", event.Title)
	}

	if event.Location != "Kilauea Volcano Activity Update" {
		t.Errorf("expected Location %q, got %q", "Kilauea Volcano Activity Update", event.Location)
	}

	if event.Details["author"] != "USGS Volcano Hazards Program" {
		t.Errorf("expected Author %q, got %q", "USGS Volcano Hazards Program", event.Details["author"])
	}

	expected_description := "Kilauea volcano remains at advisory level with ongoing monitoring."
	if event.Details["description"] != expected_description {
		t.Errorf("expected Description %q, got %q", expected_description, event.Details["description"])
	}

	expected_link := "https://volcanoes.usgs.gov/volcanoes/kilauea/status.html"
	if event.Details["url_link"] != expected_link {
		t.Errorf("expected Link %q, got %q", expected_link, event.Details["url_link"])
	}

	wantTime, err := time.Parse(time.RFC1123Z, strings.TrimSpace("Mon, 20 Jul 2026 08:30:00 +0000"))
	if err != nil {
		t.Errorf("Fail to pass %v", "Mon, 20 Jul 2026 08:30:00 +0000")
	}

	if !event.Timestamp.Equal(wantTime) {
		t.Errorf("expected Timestamp %v, got %v", wantTime, event.Timestamp)
	}
}
