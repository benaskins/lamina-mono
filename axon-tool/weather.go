package tool

import "context"

// WeatherProvider abstracts weather lookup functionality.
type WeatherProvider interface {
	GetWeather(ctx context.Context, location string) (*WeatherResult, error)
}

// WeatherResult holds current weather data for a location.
type WeatherResult struct {
	Location    string
	Description string
	Temperature float64
	FeelsLike   float64
	Humidity    int
	WindSpeed   float64
	IsDay       bool
}
