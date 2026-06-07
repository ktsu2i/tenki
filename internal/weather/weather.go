package weather

import (
	"context"
	"fmt"
	"strings"

	"github.com/ktsu2i/tenki/internal/forecast"
	"github.com/ktsu2i/tenki/internal/geocode"
)

type ViewMode string

const (
	ViewSummary ViewMode = "summary"
	ViewDaily   ViewMode = "daily"
	ViewHourly  ViewMode = "hourly"
)

type Request struct {
	LocationName string
	Mode         ViewMode
	Days         int
	Hours        int
}

type Report struct {
	Location geocode.Location  `json:"location"`
	Mode     ViewMode          `json:"mode"`
	Current  forecast.Current  `json:"current"`
	Daily    []forecast.Daily  `json:"daily"`
	Hourly   []forecast.Hourly `json:"hourly"`
}

type Geocoder interface {
	Search(context.Context, string) (geocode.Location, error)
}

type Forecaster interface {
	Get(context.Context, forecast.Request) (forecast.Forecast, error)
}

type Service struct {
	Geocoder   Geocoder
	Forecaster Forecaster
}

func (s Service) Report(ctx context.Context, request Request) (Report, error) {
	locationName := strings.TrimSpace(request.LocationName)
	if locationName == "" {
		return Report{}, fmt.Errorf("location is required")
	}
	if s.Geocoder == nil {
		return Report{}, fmt.Errorf("geocoder is required")
	}
	if s.Forecaster == nil {
		return Report{}, fmt.Errorf("forecaster is required")
	}

	location, err := s.Geocoder.Search(ctx, locationName)
	if err != nil {
		return Report{}, err
	}

	result, err := s.Forecaster.Get(ctx, forecast.Request{
		Latitude:  location.Latitude,
		Longitude: location.Longitude,
		Days:      request.Days,
		Hours:     request.Hours,
	})
	if err != nil {
		return Report{}, err
	}

	return Report{
		Location: location,
		Mode:     request.Mode,
		Current:  result.Current,
		Daily:    result.Daily,
		Hourly:   result.Hourly,
	}, nil
}
