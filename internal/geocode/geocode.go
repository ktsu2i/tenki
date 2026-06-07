package geocode

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strings"
)

const defaultEndpoint = "https://geocoding-api.open-meteo.com/v1/search"

type HTTPClient interface {
	Do(*http.Request) (*http.Response, error)
}

type Client struct {
	Endpoint   string
	HTTPClient HTTPClient
}

type Location struct {
	Name      string  `json:"name"`
	Country   string  `json:"country,omitempty"`
	Latitude  float64 `json:"latitude"`
	Longitude float64 `json:"longitude"`
	Timezone  string  `json:"timezone,omitempty"`
}

func NewClient(httpClient HTTPClient) *Client {
	return &Client{
		Endpoint:   defaultEndpoint,
		HTTPClient: httpClient,
	}
}

func (c *Client) Search(ctx context.Context, name string) (Location, error) {
	name = strings.TrimSpace(name)
	if name == "" {
		return Location{}, fmt.Errorf("location is required")
	}

	endpoint := c.Endpoint
	if endpoint == "" {
		endpoint = defaultEndpoint
	}

	values := url.Values{}
	values.Set("name", name)
	values.Set("count", "1")
	values.Set("language", "en")

	reqURL := endpoint + "?" + values.Encode()
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, reqURL, nil)
	if err != nil {
		return Location{}, err
	}

	httpClient := c.HTTPClient
	if httpClient == nil {
		httpClient = http.DefaultClient
	}

	resp, err := httpClient.Do(req)
	if err != nil {
		return Location{}, fmt.Errorf("geocoding API request failed: %w", err)
	}
	defer resp.Body.Close()

	var body searchResponse
	if err := json.NewDecoder(resp.Body).Decode(&body); err != nil {
		return Location{}, fmt.Errorf("geocoding API response is invalid: %w", err)
	}
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		if body.Reason != "" {
			return Location{}, fmt.Errorf("geocoding API request failed: %s", body.Reason)
		}
		return Location{}, fmt.Errorf("geocoding API request failed: HTTP %d", resp.StatusCode)
	}
	if body.Error {
		if body.Reason != "" {
			return Location{}, fmt.Errorf("geocoding API request failed: %s", body.Reason)
		}
		return Location{}, fmt.Errorf("geocoding API request failed")
	}
	if len(body.Results) == 0 {
		return Location{}, fmt.Errorf("location not found: %s", name)
	}

	result := body.Results[0]
	return Location{
		Name:      result.Name,
		Country:   result.Country,
		Latitude:  result.Latitude,
		Longitude: result.Longitude,
		Timezone:  result.Timezone,
	}, nil
}

type searchResponse struct {
	Results []searchResult `json:"results"`
	Error   bool           `json:"error"`
	Reason  string         `json:"reason"`
}

type searchResult struct {
	Name      string  `json:"name"`
	Country   string  `json:"country"`
	Latitude  float64 `json:"latitude"`
	Longitude float64 `json:"longitude"`
	Timezone  string  `json:"timezone"`
}
