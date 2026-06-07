package forecast

import (
	"context"
	"io"
	"net/http"
	"strings"
	"testing"

	"github.com/ktsu2i/tenki/internal/geocode"
)

func TestGetRequestsForecastAndNormalizesResponse(t *testing.T) {
	httpClient := &fakeHTTPClient{
		statusCode: http.StatusOK,
		body: `{
			"current":{"time":"2026-06-05T12:00","temperature_2m":22.4,"weather_code":2},
			"daily":{
				"time":["2026-06-05","2026-06-06"],
				"weather_code":[0,61],
				"temperature_2m_max":[24.1,21.2],
				"temperature_2m_min":[17.2,16.8],
				"precipitation_probability_max":[20,70]
			},
			"hourly":{
				"time":["2026-06-05T12:00","2026-06-05T13:00"],
				"temperature_2m":[22.4,23.0],
				"weather_code":[2,0],
				"precipitation_probability":[10,0]
			}
		}`,
	}

	client := NewClient(httpClient)
	client.Endpoint = "https://example.test/forecast"

	result, err := client.Get(context.Background(), Request{
		Location: geocode.Location{Latitude: 35.6895, Longitude: 139.6917},
		Days:     2,
		Hours:    2,
	})
	if err != nil {
		t.Fatalf("Get returned error: %v", err)
	}
	query := httpClient.request.URL.Query()
	for key, want := range map[string]string{
		"latitude":       "35.6895",
		"longitude":      "139.6917",
		"current":        "temperature_2m,weather_code",
		"daily":          "weather_code,temperature_2m_max,temperature_2m_min,precipitation_probability_max",
		"hourly":         "temperature_2m,weather_code,precipitation_probability",
		"timezone":       "auto",
		"forecast_days":  "2",
		"forecast_hours": "2",
	} {
		if got := query.Get(key); got != want {
			t.Fatalf("%s query = %q, want %q", key, got, want)
		}
	}
	if result.Current.Weather != "Partly cloudy" {
		t.Fatalf("current weather = %q, want Partly cloudy", result.Current.Weather)
	}
	if len(result.Daily) != 2 || result.Daily[1].Weather != "Light rain" {
		t.Fatalf("daily = %+v, want two normalized days", result.Daily)
	}
	if len(result.Hourly) != 2 || result.Hourly[0].Precipitation != 10 {
		t.Fatalf("hourly = %+v, want two normalized hours", result.Hourly)
	}
}

func TestGetReportsAPIError(t *testing.T) {
	client := NewClient(&fakeHTTPClient{
		statusCode: http.StatusBadRequest,
		body:       `{"error":true,"reason":"bad request"}`,
	})
	client.Endpoint = "https://example.test/forecast"

	if _, err := client.Get(context.Background(), Request{}); err == nil {
		t.Fatal("Get returned nil error, want API error")
	}
}

type fakeHTTPClient struct {
	statusCode int
	body       string
	request    *http.Request
}

func (c *fakeHTTPClient) Do(request *http.Request) (*http.Response, error) {
	c.request = request
	return &http.Response{
		StatusCode: c.statusCode,
		Body:       io.NopCloser(strings.NewReader(c.body)),
		Header:     make(http.Header),
		Request:    request,
	}, nil
}
