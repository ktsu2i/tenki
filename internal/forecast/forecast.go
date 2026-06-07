package forecast

import (
	"context"
	"encoding/json"
	"fmt"
	"math"
	"net/http"
	"net/url"
	"strconv"

	"github.com/ktsu2i/tenki/internal/geocode"
)

const defaultEndpoint = "https://api.open-meteo.com/v1/forecast"

type HTTPClient interface {
	Do(*http.Request) (*http.Response, error)
}

type Client struct {
	Endpoint   string
	HTTPClient HTTPClient
}

type Request struct {
	Location geocode.Location
	Days     int
	Hours    int
}

type Forecast struct {
	Current Current  `json:"current"`
	Daily   []Daily  `json:"daily"`
	Hourly  []Hourly `json:"hourly"`
}

type Current struct {
	Time        string  `json:"time,omitempty"`
	Temperature float64 `json:"temperature"`
	WeatherCode int     `json:"weather_code"`
	Weather     string  `json:"weather"`
}

type Daily struct {
	Date             string  `json:"date"`
	TemperatureMax   float64 `json:"temperature_max"`
	TemperatureMin   float64 `json:"temperature_min"`
	WeatherCode      int     `json:"weather_code"`
	Weather          string  `json:"weather"`
	PrecipitationMax int     `json:"precipitation_probability"`
}

type Hourly struct {
	Time          string  `json:"time"`
	Temperature   float64 `json:"temperature"`
	WeatherCode   int     `json:"weather_code"`
	Weather       string  `json:"weather"`
	Precipitation int     `json:"precipitation_probability"`
}

func NewClient(httpClient HTTPClient) *Client {
	return &Client{
		Endpoint:   defaultEndpoint,
		HTTPClient: httpClient,
	}
}

func (c *Client) Get(ctx context.Context, request Request) (Forecast, error) {
	endpoint := c.Endpoint
	if endpoint == "" {
		endpoint = defaultEndpoint
	}

	values := url.Values{}
	values.Set("latitude", strconv.FormatFloat(request.Location.Latitude, 'f', -1, 64))
	values.Set("longitude", strconv.FormatFloat(request.Location.Longitude, 'f', -1, 64))
	values.Set("current", "temperature_2m,weather_code")
	values.Set("daily", "weather_code,temperature_2m_max,temperature_2m_min,precipitation_probability_max")
	values.Set("hourly", "temperature_2m,weather_code,precipitation_probability")
	values.Set("timezone", "auto")
	if request.Days > 0 {
		values.Set("forecast_days", strconv.Itoa(request.Days))
	}
	if request.Hours > 0 {
		values.Set("forecast_hours", strconv.Itoa(request.Hours))
	}

	reqURL := endpoint + "?" + values.Encode()
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, reqURL, nil)
	if err != nil {
		return Forecast{}, err
	}

	httpClient := c.HTTPClient
	if httpClient == nil {
		httpClient = http.DefaultClient
	}

	resp, err := httpClient.Do(req)
	if err != nil {
		return Forecast{}, fmt.Errorf("forecast API request failed: %w", err)
	}
	defer resp.Body.Close()

	var body forecastResponse
	if err := json.NewDecoder(resp.Body).Decode(&body); err != nil {
		return Forecast{}, fmt.Errorf("forecast API response is invalid: %w", err)
	}
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		if body.Reason != "" {
			return Forecast{}, fmt.Errorf("forecast API request failed: %s", body.Reason)
		}
		return Forecast{}, fmt.Errorf("forecast API request failed: HTTP %d", resp.StatusCode)
	}
	if body.Error {
		if body.Reason != "" {
			return Forecast{}, fmt.Errorf("forecast API request failed: %s", body.Reason)
		}
		return Forecast{}, fmt.Errorf("forecast API request failed")
	}

	return normalize(body, request.Days, request.Hours), nil
}

func normalize(body forecastResponse, days, hours int) Forecast {
	current := Current{
		Time:        body.Current.Time,
		Temperature: body.Current.Temperature,
		WeatherCode: body.Current.WeatherCode,
		Weather:     WeatherDescription(body.Current.WeatherCode),
	}

	dailyCount := minNonNegative(days, len(body.Daily.Time), len(body.Daily.WeatherCode), len(body.Daily.TemperatureMax), len(body.Daily.TemperatureMin), len(body.Daily.PrecipitationMax))
	daily := make([]Daily, 0, dailyCount)
	for i := 0; i < dailyCount; i++ {
		code := body.Daily.WeatherCode[i]
		daily = append(daily, Daily{
			Date:             body.Daily.Time[i],
			TemperatureMax:   body.Daily.TemperatureMax[i],
			TemperatureMin:   body.Daily.TemperatureMin[i],
			WeatherCode:      code,
			Weather:          WeatherDescription(code),
			PrecipitationMax: body.Daily.PrecipitationMax[i],
		})
	}

	hourlyCount := minNonNegative(hours, len(body.Hourly.Time), len(body.Hourly.WeatherCode), len(body.Hourly.Temperature), len(body.Hourly.Precipitation))
	hourly := make([]Hourly, 0, hourlyCount)
	for i := 0; i < hourlyCount; i++ {
		code := body.Hourly.WeatherCode[i]
		hourly = append(hourly, Hourly{
			Time:          body.Hourly.Time[i],
			Temperature:   body.Hourly.Temperature[i],
			WeatherCode:   code,
			Weather:       WeatherDescription(code),
			Precipitation: body.Hourly.Precipitation[i],
		})
	}

	return Forecast{
		Current: current,
		Daily:   daily,
		Hourly:  hourly,
	}
}

func minNonNegative(values ...int) int {
	minValue := math.MaxInt
	for _, value := range values {
		if value < minValue {
			minValue = value
		}
	}
	if minValue == math.MaxInt || minValue < 0 {
		return 0
	}
	return minValue
}

func WeatherDescription(code int) string {
	switch code {
	case 0:
		return "Clear"
	case 1:
		return "Mainly clear"
	case 2:
		return "Partly cloudy"
	case 3:
		return "Overcast"
	case 45:
		return "Fog"
	case 48:
		return "Rime fog"
	case 51:
		return "Light drizzle"
	case 53:
		return "Drizzle"
	case 55:
		return "Heavy drizzle"
	case 56:
		return "Light freezing drizzle"
	case 57:
		return "Freezing drizzle"
	case 61:
		return "Light rain"
	case 63:
		return "Rain"
	case 65:
		return "Heavy rain"
	case 66:
		return "Light freezing rain"
	case 67:
		return "Freezing rain"
	case 71:
		return "Light snow"
	case 73:
		return "Snow"
	case 75:
		return "Heavy snow"
	case 77:
		return "Snow grains"
	case 80:
		return "Light showers"
	case 81:
		return "Showers"
	case 82:
		return "Heavy showers"
	case 85:
		return "Snow showers"
	case 86:
		return "Heavy snow showers"
	case 95:
		return "Thunderstorm"
	case 96, 99:
		return "Thunderstorm with hail"
	default:
		return "Unknown"
	}
}

type forecastResponse struct {
	Current currentResponse `json:"current"`
	Daily   dailyResponse   `json:"daily"`
	Hourly  hourlyResponse  `json:"hourly"`
	Error   bool            `json:"error"`
	Reason  string          `json:"reason"`
}

type currentResponse struct {
	Time        string  `json:"time"`
	Temperature float64 `json:"temperature_2m"`
	WeatherCode int     `json:"weather_code"`
}

type dailyResponse struct {
	Time             []string  `json:"time"`
	WeatherCode      []int     `json:"weather_code"`
	TemperatureMax   []float64 `json:"temperature_2m_max"`
	TemperatureMin   []float64 `json:"temperature_2m_min"`
	PrecipitationMax []int     `json:"precipitation_probability_max"`
}

type hourlyResponse struct {
	Time          []string  `json:"time"`
	Temperature   []float64 `json:"temperature_2m"`
	WeatherCode   []int     `json:"weather_code"`
	Precipitation []int     `json:"precipitation_probability"`
}
