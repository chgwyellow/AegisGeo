package main

import (
	"AegisGeo/internal/health"
	"AegisGeo/internal/ingestion"
	"fmt"
	"log"
	"os"

	"github.com/joho/godotenv"
)

func main() {
	_ = godotenv.Load()

	cwaURL := os.Getenv("CWA_EQK_URL")
	cwaRainURL := os.Getenv("CWA_RAIN_URL")
	cwaToken := os.Getenv("CWA_TOKEN")
	usgsURL := os.Getenv("USGS_API_URL")
	jmaURL := os.Getenv("JMA_API_URL")
	noaaURL := os.Getenv("NOAA_API_URL")
	nwsURL := os.Getenv("NWS_API_URL")
	volURL := os.Getenv("VOLCANO_API_URL")
	email := os.Getenv("EMAIL")

	if cwaURL == "" || cwaToken == "" || usgsURL == "" || jmaURL == "" || noaaURL == "" || cwaRainURL == "" || nwsURL == "" || email == "" || volURL == "" {
		log.Fatal("Get wrong in environment setting!")
	}

	// Create Clients
	clients := []ingestion.IngestionClient{
		ingestion.NewCwaClient(cwaURL, cwaToken),
		ingestion.NewUsgsClient(usgsURL),
		ingestion.NewJmaClient(jmaURL),
		ingestion.NewTsunamiClient(noaaURL),
		ingestion.NewCwaRainClient(cwaRainURL, cwaToken),
		ingestion.NewNwsSevereWeatherClient(nwsURL, email),
		ingestion.NewVolcanoClient(volURL),
	}

	results := health.BuildHealthResults(clients)

	output := health.FormatHealthResults(results)

	fmt.Println(output)

}
