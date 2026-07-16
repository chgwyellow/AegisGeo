package api

import (
	"testing"
	"time"
)

func TestDateParsing(t *testing.T) {
	loc := time.FixedZone("CST", 8*60*60)
	layout := "2006-01-02"

	// Define test case: input & expected result
	tests := []struct {
		input   string
		wantErr bool
	}{
		{"2026-07-16", false},
		{"2026/07/16", true},
		{"not-a-date", true},
	}

	for _, tt := range tests {
		_, err := time.ParseInLocation(layout, tt.input, loc)

		if (err != nil) != tt.wantErr {
			t.Errorf("For %q, expected error is %v, but the result is %v", tt.input, tt.wantErr, err)
		}
	}
}
