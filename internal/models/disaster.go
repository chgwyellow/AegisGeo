// Define global disaster event type

package models

import "time"

type Event struct {
	ID        string         `json:"id"` // Struct tag, to align the outside area name
	Source    string         `json:"source"`
	Type      string         `json:"type"`
	Title     string         `json:"title"`
	Magnitude float64        `json:"magnitude"`
	Depth     float64        `json:"depth"`
	Timestamp time.Time      `json:"timestamp"`
	Country   string         `json:"country"`
	Location  string         `json:"location"`
	Latitude  float64        `json:"latitude"`
	Longitude float64        `json:"longitude"`
	Details   map[string]any `json:"details"` // Dynamic, JSONB
}
