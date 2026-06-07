package weather

import (
	"context"
	"testing"

	"github.com/ktsu2i/tenki/internal/forecast"
	"github.com/ktsu2i/tenki/internal/geocode"
)

func TestServiceReportResolvesLocationAndGetsForecast(t *testing.T) {
	geocoder := &fakeGeocoder{
		location: geocode.Location{
			Name:      "Tokyo",
			Country:   "Japan",
			Latitude:  35.6895,
			Longitude: 139.6917,
			Timezone:  "Asia/Tokyo",
		},
	}
	forecaster := &fakeForecaster{
		forecast: forecast.Forecast{
			Current: forecast.Current{Temperature: 22, Weather: "Partly cloudy"},
			Daily:   []forecast.Daily{{Date: "2026-06-05", Weather: "Clear"}},
			Hourly:  []forecast.Hourly{{Time: "2026-06-05T12:00", Weather: "Partly cloudy"}},
		},
	}
	service := Service{Geocoder: geocoder, Forecaster: forecaster}

	report, err := service.Report(context.Background(), Request{
		LocationName: " tokyo ",
		Mode:         ViewHourly,
		Days:         3,
		Hours:        12,
	})
	if err != nil {
		t.Fatalf("Report returned error: %v", err)
	}

	if geocoder.name != "tokyo" {
		t.Fatalf("geocoder name = %q, want tokyo", geocoder.name)
	}
	if forecaster.request.Latitude != 35.6895 || forecaster.request.Longitude != 139.6917 {
		t.Fatalf("forecast location = %+v, want Tokyo coordinates", forecaster.request)
	}
	if forecaster.request.Days != 3 || forecaster.request.Hours != 12 {
		t.Fatalf("forecast request = %+v, want days=3 hours=12", forecaster.request)
	}
	if report.Location.Name != "Tokyo" || report.Mode != ViewHourly || report.Current.Weather != "Partly cloudy" {
		t.Fatalf("report = %+v, want hourly Tokyo report", report)
	}
}

func TestServiceReportRequiresLocation(t *testing.T) {
	service := Service{Geocoder: &fakeGeocoder{}, Forecaster: &fakeForecaster{}}

	if _, err := service.Report(context.Background(), Request{LocationName: " "}); err == nil {
		t.Fatal("Report returned nil error, want location error")
	}
}

type fakeGeocoder struct {
	location geocode.Location
	name     string
}

func (g *fakeGeocoder) Search(ctx context.Context, name string) (geocode.Location, error) {
	g.name = name
	return g.location, nil
}

type fakeForecaster struct {
	forecast forecast.Forecast
	request  forecast.Request
}

func (f *fakeForecaster) Get(ctx context.Context, request forecast.Request) (forecast.Forecast, error) {
	f.request = request
	return f.forecast, nil
}
