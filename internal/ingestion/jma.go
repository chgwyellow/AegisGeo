package ingestion

import (
	"AegisGeo/internal/models"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
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
