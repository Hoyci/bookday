// Package geocoder provides clients for converting addresses into geographic coordinates.
package geocoder

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strconv"
	"time"

	"github.com/hoyci/bookday/internal/routing"
)

type NominatimResult struct {
	Lat string `json:"lat"`
	Lon string `json:"lon"`
}

type nominatimClient struct {
	httpClient *http.Client
	baseURL    string
	userAgent  string
}

func NewNominatimClient(appName, appVersion string) routing.Geocoder {
	return &nominatimClient{
		httpClient: &http.Client{Timeout: 10 * time.Second},
		baseURL:    "https://nominatim.openstreetmap.org/search",
		userAgent:  fmt.Sprintf("%s/%s", appName, appVersion),
	}
}

func (c *nominatimClient) Geocode(ctx context.Context, address string) (float64, float64, error) {
	fullURL, err := url.Parse(c.baseURL)
	if err != nil {
		return 0, 0, fmt.Errorf("failed to parse base URL: %w", err)
	}
	params := url.Values{}
	params.Add("q", address)
	params.Add("format", "json")
	params.Add("limit", "1")
	fullURL.RawQuery = params.Encode()

	req, err := http.NewRequestWithContext(ctx, "GET", fullURL.String(), nil)
	if err != nil {
		return 0, 0, fmt.Errorf("failed to create geocoding request: %w", err)
	}
	req.Header.Set("User-Agent", c.userAgent)

	time.Sleep(1 * time.Second)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return 0, 0, fmt.Errorf("failed to execute geocoding request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return 0, 0, fmt.Errorf("nominatim API returned non-200 status: %d", resp.StatusCode)
	}

	var results []NominatimResult
	if err := json.NewDecoder(resp.Body).Decode(&results); err != nil {
		return 0, 0, fmt.Errorf("failed to decode nominatim response: %w", err)
	}

	if len(results) == 0 {
		return 0, 0, fmt.Errorf("no geocoding results found for address: %s", address)
	}

	lat, err := strconv.ParseFloat(results[0].Lat, 64)
	if err != nil {
		return 0, 0, fmt.Errorf("failed to parse latitude: %w", err)
	}
	lon, err := strconv.ParseFloat(results[0].Lon, 64)
	if err != nil {
		return 0, 0, fmt.Errorf("failed to parse longitude: %w", err)
	}

	return lat, lon, nil
}
