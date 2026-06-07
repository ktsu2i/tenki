package cli

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/alecthomas/kong"
	"github.com/ktsu2i/tenki/internal/forecast"
	"github.com/ktsu2i/tenki/internal/geocode"
	"github.com/ktsu2i/tenki/internal/output"
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
	ctx        context.Context
	stdout     io.Writer
	geocoder   geocoder
	forecaster forecaster
}

type geocoder interface {
	Search(context.Context, string) (geocode.Location, error)
}

type forecaster interface {
	Get(context.Context, forecast.Request) (forecast.Forecast, error)
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
	return runWithClients(args, stdout, stderr, version, context.Background(), geocode.NewClient(httpClient), forecast.NewClient(httpClient))
}

func runWithClients(args []string, stdout, stderr io.Writer, version string, baseContext context.Context, geocoder geocoder, forecaster forecaster) error {
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
		ctx:        baseContext,
		stdout:     stdout,
		geocoder:   geocoder,
		forecaster: forecaster,
	})
}

func (c *CLI) Run(ctx *runContext) error {
	location := strings.TrimSpace(c.Location)
	if location == "" {
		return fmt.Errorf("location is required")
	}

	mode, days, hours, err := c.resolveView()
	if err != nil {
		return err
	}

	resolved, err := ctx.geocoder.Search(ctx.ctx, location)
	if err != nil {
		return err
	}

	forecastResult, err := ctx.forecaster.Get(ctx.ctx, forecast.Request{
		Location: resolved,
		Days:     days,
		Hours:    hours,
	})
	if err != nil {
		return err
	}

	report := output.Report{
		Location: resolved,
		Mode:     mode,
		Current:  forecastResult.Current,
		Daily:    forecastResult.Daily,
		Hourly:   forecastResult.Hourly,
	}

	if c.JSON {
		return output.WriteJSON(ctx.stdout, report)
	}

	return output.WriteText(ctx.stdout, report)
}

func (c *CLI) resolveView() (output.ViewMode, int, int, error) {
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
		return output.ViewHourly, days, hours, nil
	case c.Daily || c.Days != nil:
		return output.ViewDaily, days, hours, nil
	default:
		return output.ViewSummary, days, hours, nil
	}
}
