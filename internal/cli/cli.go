package cli

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/alecthomas/kong"
	"github.com/ktsu2i/tenki/internal/forecast"
	"github.com/ktsu2i/tenki/internal/geocode"
	"github.com/ktsu2i/tenki/internal/output"
	"github.com/ktsu2i/tenki/internal/weather"
)

const (
	appName        = "tenki"
	defaultDays    = 3
	defaultHours   = 12
	maxDailyDays   = 7
	maxHourlyHours = 24
)

type CLI struct {
	Version kong.VersionFlag `name:"version" short:"v" help:"Show version and exit."`

	Daily  bool `help:"Show daily forecast."`
	Hourly bool `help:"Show hourly forecast."`
	Days   *int `help:"Number of daily forecast days." placeholder:"N"`
	Hours  *int `help:"Number of hourly forecast hours." placeholder:"N"`
	JSON   bool `name:"json" help:"Print JSON output."`

	Location string `arg:"" name:"location" help:"Location name to resolve."`
}

type runContext struct {
	ctx     context.Context
	stdout  io.Writer
	service reporter
}

type reporter interface {
	Report(context.Context, weather.Request) (weather.Report, error)
}

// Main is the process-oriented entry point used by cmd/tenki.
func Main(args []string, stdout, stderr io.Writer, version string) int {
	if err := Run(args, stdout, stderr, version); err != nil {
		fmt.Fprintln(stderr, err)
		return 1
	}
	return 0
}

// Run parses CLI arguments and executes the selected command behavior.
func Run(args []string, stdout, stderr io.Writer, version string) error {
	httpClient := &http.Client{Timeout: 10 * time.Second}
	service := weather.Service{
		Geocoder:   geocode.NewClient(httpClient),
		Forecaster: forecast.NewClient(httpClient),
	}
	return runWithService(args, stdout, stderr, version, context.Background(), service)
}

func runWithService(args []string, stdout, stderr io.Writer, version string, baseContext context.Context, service reporter) error {
	var cli CLI

	parser, err := kong.New(
		&cli,
		kong.Name(appName),
		kong.Description("Show weather forecasts from Open-Meteo."),
		kong.UsageOnError(),
		kong.Vars{"version": version},
		kong.Writers(stdout, stderr),
	)
	if err != nil {
		return err
	}

	ctx, err := parser.Parse(args)
	if err != nil {
		return err
	}

	return ctx.Run(&runContext{
		ctx:     baseContext,
		stdout:  stdout,
		service: service,
	})
}

func (c *CLI) Run(ctx *runContext) error {
	mode, days, hours, err := c.resolveView()
	if err != nil {
		return err
	}

	report, err := ctx.service.Report(ctx.ctx, weather.Request{
		LocationName: c.Location,
		Mode:         mode,
		Days:         days,
		Hours:        hours,
	})
	if err != nil {
		return err
	}

	if c.JSON {
		return output.WriteJSON(ctx.stdout, report)
	}

	return output.WriteText(ctx.stdout, report)
}

func (c *CLI) resolveView() (weather.ViewMode, int, int, error) {
	if c.Daily && c.Hourly {
		return "", 0, 0, fmt.Errorf("--daily and --hourly cannot be used together")
	}
	if c.Days != nil && c.Hours != nil {
		return "", 0, 0, fmt.Errorf("--days and --hours cannot be used together")
	}
	if c.Daily && c.Hours != nil {
		return "", 0, 0, fmt.Errorf("--daily and --hours cannot be used together")
	}
	if c.Hourly && c.Days != nil {
		return "", 0, 0, fmt.Errorf("--hourly and --days cannot be used together")
	}

	days := defaultDays
	if c.Days != nil {
		days = *c.Days
	}
	if days < 1 || days > maxDailyDays {
		return "", 0, 0, fmt.Errorf("--days must be between 1 and %d", maxDailyDays)
	}

	hours := defaultHours
	if c.Hours != nil {
		hours = *c.Hours
	}
	if hours < 1 || hours > maxHourlyHours {
		return "", 0, 0, fmt.Errorf("--hours must be between 1 and %d", maxHourlyHours)
	}

	switch {
	case c.Hourly || c.Hours != nil:
		return weather.ViewHourly, days, hours, nil
	case c.Daily || c.Days != nil:
		return weather.ViewDaily, days, hours, nil
	default:
		return weather.ViewSummary, days, hours, nil
	}
}
