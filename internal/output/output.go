package output

import (
	"encoding/json"
	"fmt"
	"io"
	"math"
	"time"

	"github.com/ktsu2i/tenki/internal/forecast"
	"github.com/ktsu2i/tenki/internal/geocode"
)

type ViewMode string

const (
	ViewSummary ViewMode = "summary"
	ViewDaily   ViewMode = "daily"
	ViewHourly  ViewMode = "hourly"
)

type Report struct {
	Location geocode.Location  `json:"location"`
	Mode     ViewMode          `json:"mode"`
	Current  forecast.Current  `json:"current"`
	Daily    []forecast.Daily  `json:"daily"`
	Hourly   []forecast.Hourly `json:"hourly"`
}

func WriteJSON(w io.Writer, report Report) error {
	encoder := json.NewEncoder(w)
	encoder.SetIndent("", "  ")
	return encoder.Encode(report)
}

func WriteText(w io.Writer, report Report) error {
	switch report.Mode {
	case ViewDaily:
		return writeDaily(w, report)
	case ViewHourly:
		return writeHourly(w, report)
	default:
		return writeSummary(w, report)
	}
}

func writeSummary(w io.Writer, report Report) error {
	if _, err := fmt.Fprintln(w, formatLocation(report.Location)); err != nil {
		return err
	}
	if _, err := fmt.Fprintf(w, "Now: %s, %s\n", formatTemperature(report.Current.Temperature), report.Current.Weather); err != nil {
		return err
	}
	if len(report.Daily) > 0 {
		today := report.Daily[0]
		if _, err := fmt.Fprintf(w, "Today: %s / %s, rain %d%%\n", formatTemperature(today.TemperatureMin), formatTemperature(today.TemperatureMax), today.PrecipitationMax); err != nil {
			return err
		}
	}
	if len(report.Daily) <= 1 {
		return nil
	}
	if _, err := fmt.Fprintln(w); err != nil {
		return err
	}
	for _, day := range report.Daily[1:] {
		if _, err := fmt.Fprintf(w, "%-3s  %-13s %s / %s\n", formatDay(day.Date), day.Weather, formatTemperature(day.TemperatureMax), formatTemperature(day.TemperatureMin)); err != nil {
			return err
		}
	}
	return nil
}

func writeDaily(w io.Writer, report Report) error {
	if _, err := fmt.Fprintf(w, "%s\n\n", formatLocation(report.Location)); err != nil {
		return err
	}
	for _, day := range report.Daily {
		if _, err := fmt.Fprintf(w, "%-3s  %-13s %s / %s  %d%%\n", formatDay(day.Date), day.Weather, formatTemperature(day.TemperatureMax), formatTemperature(day.TemperatureMin), day.PrecipitationMax); err != nil {
			return err
		}
	}
	return nil
}

func writeHourly(w io.Writer, report Report) error {
	if _, err := fmt.Fprintf(w, "%s\n\n", formatLocation(report.Location)); err != nil {
		return err
	}
	for _, hour := range report.Hourly {
		if _, err := fmt.Fprintf(w, "%s  %s  %-13s %d%%\n", formatHour(hour.Time), formatTemperature(hour.Temperature), hour.Weather, hour.Precipitation); err != nil {
			return err
		}
	}
	return nil
}

func formatLocation(location geocode.Location) string {
	if location.Country == "" {
		return location.Name
	}
	return fmt.Sprintf("%s, %s", location.Name, location.Country)
}

func formatTemperature(value float64) string {
	return fmt.Sprintf("%.0fC", math.Round(value))
}

func formatDay(value string) string {
	parsed, err := time.Parse("2006-01-02", value)
	if err != nil {
		return value
	}
	return parsed.Format("Mon")
}

func formatHour(value string) string {
	for _, layout := range []string{"2006-01-02T15:04", time.RFC3339} {
		parsed, err := time.Parse(layout, value)
		if err == nil {
			return parsed.Format("15:04")
		}
	}
	return value
}
