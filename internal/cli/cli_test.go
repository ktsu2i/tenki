package cli

import (
	"bytes"
	"context"
	"encoding/json"
	"strings"
	"testing"

	"github.com/ktsu2i/tenki/internal/forecast"
	"github.com/ktsu2i/tenki/internal/geocode"
)

func TestRunPrintsSummary(t *testing.T) {
	var stdout, stderr bytes.Buffer

	err := runWithClients([]string{"tokyo"}, &stdout, &stderr, "test", context.Background(), fakeGeocoder{}, fakeForecaster{})
	if err != nil {
		t.Fatalf("Run returned error: %v", err)
	}
	if stderr.Len() != 0 {
		t.Fatalf("stderr = %q, want empty", stderr.String())
	}

	out := stdout.String()
	for _, want := range []string{"Tokyo, Japan", "Now: 22C, Partly cloudy", "Today: 17C / 24C, rain 20%", "Sat  Cloudy"} {
		if !strings.Contains(out, want) {
			t.Fatalf("stdout = %q, want it to contain %q", out, want)
		}
	}
}

func TestRunJSON(t *testing.T) {
	var stdout, stderr bytes.Buffer

	err := runWithClients([]string{"tokyo", "--hourly", "--hours", "24", "--json"}, &stdout, &stderr, "test", context.Background(), fakeGeocoder{}, fakeForecaster{})
	if err != nil {
		t.Fatalf("Run returned error: %v", err)
	}
	if stderr.Len() != 0 {
		t.Fatalf("stderr = %q, want empty", stderr.String())
	}

	var got struct {
		Location geocode.Location  `json:"location"`
		Mode     string            `json:"mode"`
		Current  forecast.Current  `json:"current"`
		Hourly   []forecast.Hourly `json:"hourly"`
	}
	if err := json.Unmarshal(stdout.Bytes(), &got); err != nil {
		t.Fatalf("stdout is not JSON: %v\n%s", err, stdout.String())
	}
	if got.Location.Name != "Tokyo" || got.Mode != "hourly" || got.Current.Weather != "Partly cloudy" || len(got.Hourly) != 2 {
		t.Fatalf("decoded JSON = %+v, want hourly report", got)
	}
}

func TestRunRejectsConflictingModes(t *testing.T) {
	var stdout, stderr bytes.Buffer

	err := Run([]string{"tokyo", "--daily", "--hourly"}, &stdout, &stderr, "test")
	if err == nil {
		t.Fatal("Run returned nil error, want conflict error")
	}
	if !strings.Contains(err.Error(), "--daily and --hourly cannot be used together") {
		t.Fatalf("error = %q, want mode conflict", err.Error())
	}
}

func TestRunRejectsInvalidDays(t *testing.T) {
	var stdout, stderr bytes.Buffer

	err := Run([]string{"tokyo", "--days", "0"}, &stdout, &stderr, "test")
	if err == nil {
		t.Fatal("Run returned nil error, want invalid days error")
	}
	if !strings.Contains(err.Error(), "--days must be between 1 and 7") {
		t.Fatalf("error = %q, want days range error", err.Error())
	}
}

func TestRunRequiresLocation(t *testing.T) {
	var stdout, stderr bytes.Buffer

	err := Run(nil, &stdout, &stderr, "test")
	if err == nil {
		t.Fatal("Run returned nil error, want location error")
	}
	if !strings.Contains(err.Error(), "<location>") {
		t.Fatalf("error = %q, want location error", err.Error())
	}
}

type fakeGeocoder struct{}

func (fakeGeocoder) Search(ctx context.Context, name string) (geocode.Location, error) {
	return geocode.Location{
		Name:      "Tokyo",
		Country:   "Japan",
		Latitude:  35.6895,
		Longitude: 139.6917,
		Timezone:  "Asia/Tokyo",
	}, nil
}

type fakeForecaster struct{}

func (fakeForecaster) Get(ctx context.Context, request forecast.Request) (forecast.Forecast, error) {
	return forecast.Forecast{
		Current: forecast.Current{
			Time:        "2026-06-05T12:00",
			Temperature: 22,
			WeatherCode: 2,
			Weather:     "Partly cloudy",
		},
		Daily: []forecast.Daily{
			{
				Date:             "2026-06-05",
				TemperatureMax:   24,
				TemperatureMin:   17,
				WeatherCode:      0,
				Weather:          "Clear",
				PrecipitationMax: 20,
			},
			{
				Date:             "2026-06-06",
				TemperatureMax:   23,
				TemperatureMin:   18,
				WeatherCode:      3,
				Weather:          "Cloudy",
				PrecipitationMax: 40,
			},
		},
		Hourly: []forecast.Hourly{
			{
				Time:          "2026-06-05T12:00",
				Temperature:   22,
				WeatherCode:   2,
				Weather:       "Partly cloudy",
				Precipitation: 10,
			},
			{
				Time:          "2026-06-05T13:00",
				Temperature:   23,
				WeatherCode:   0,
				Weather:       "Clear",
				Precipitation: 0,
			},
		},
	}, nil
}
