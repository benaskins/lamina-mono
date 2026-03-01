package tool

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestOpenMeteoClient_ParsesWeatherResponse(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("/v1/search", func(w http.ResponseWriter, r *http.Request) {
		resp := geocodeResponse{
			Results: []geoResult{
				{Name: "Melbourne", Country: "Australia", Admin1: "Victoria", Latitude: -37.8142, Longitude: 144.9632},
			},
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(resp)
	})
	mux.HandleFunc("/v1/forecast", func(w http.ResponseWriter, r *http.Request) {
		resp := weatherAPIResponse{}
		resp.CurrentWeather.Temperature = 22.5
		resp.CurrentWeather.WindSpeed = 15.0
		resp.CurrentWeather.WindDirection = 180
		resp.CurrentWeather.WeatherCode = 0
		resp.CurrentWeather.IsDay = 1
		resp.Hourly.RelativeHumidity = []int{65}
		resp.Hourly.ApparentTemp = []float64{21.0}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(resp)
	})

	srv := httptest.NewServer(mux)
	defer srv.Close()

	client := &OpenMeteoClient{
		httpClient: srv.Client(),
		geocodeURL: srv.URL + "/v1/search",
		weatherURL: srv.URL + "/v1/forecast",
	}

	result, err := client.GetWeather(context.Background(), "Melbourne")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.Temperature != 22.5 {
		t.Errorf("got temperature %f, want 22.5", result.Temperature)
	}
	if result.FeelsLike != 21.0 {
		t.Errorf("got feels_like %f, want 21.0", result.FeelsLike)
	}
	if result.Humidity != 65 {
		t.Errorf("got humidity %d, want 65", result.Humidity)
	}
	if result.WindSpeed != 15.0 {
		t.Errorf("got wind_speed %f, want 15.0", result.WindSpeed)
	}
	if result.Description != "Clear sky" {
		t.Errorf("got description %q, want %q", result.Description, "Clear sky")
	}
	if !result.IsDay {
		t.Error("expected IsDay to be true")
	}
}

func TestOpenMeteoClient_GeocodingAndWeatherFlow(t *testing.T) {
	var receivedGeoName string
	var receivedLat string

	mux := http.NewServeMux()
	mux.HandleFunc("/v1/search", func(w http.ResponseWriter, r *http.Request) {
		receivedGeoName = r.URL.Query().Get("name")
		resp := geocodeResponse{
			Results: []geoResult{
				{Name: "Gosford", Country: "Australia", Admin1: "New South Wales", Latitude: -33.4265, Longitude: 151.3418},
				{Name: "Gosford", Country: "United Kingdom", Admin1: "England", Latitude: 55.0, Longitude: -1.6},
			},
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(resp)
	})
	mux.HandleFunc("/v1/forecast", func(w http.ResponseWriter, r *http.Request) {
		receivedLat = r.URL.Query().Get("latitude")
		resp := weatherAPIResponse{}
		resp.CurrentWeather.Temperature = 28.0
		resp.CurrentWeather.WeatherCode = 2
		resp.CurrentWeather.IsDay = 1
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(resp)
	})

	srv := httptest.NewServer(mux)
	defer srv.Close()

	client := &OpenMeteoClient{
		httpClient: srv.Client(),
		geocodeURL: srv.URL + "/v1/search",
		weatherURL: srv.URL + "/v1/forecast",
	}

	result, err := client.GetWeather(context.Background(), "Gosford, Australia")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if receivedGeoName != "Gosford" {
		t.Errorf("geocode received name %q, want %q", receivedGeoName, "Gosford")
	}

	// Should have picked the Australian result (-33.4265)
	if !strings.HasPrefix(receivedLat, "-33.4265") {
		t.Errorf("weather API received lat %q, expected Australian coordinates", receivedLat)
	}

	if !strings.Contains(result.Location, "Australia") {
		t.Errorf("got location %q, expected it to contain 'Australia'", result.Location)
	}
	if result.Description != "Partly cloudy" {
		t.Errorf("got description %q, want %q", result.Description, "Partly cloudy")
	}
}

func TestOpenMeteoClient_LocationNotFound(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("/v1/search", func(w http.ResponseWriter, r *http.Request) {
		// Return empty results
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(geocodeResponse{})
	})

	srv := httptest.NewServer(mux)
	defer srv.Close()

	client := &OpenMeteoClient{
		httpClient: srv.Client(),
		geocodeURL: srv.URL + "/v1/search",
		weatherURL: srv.URL + "/v1/forecast",
	}

	_, err := client.GetWeather(context.Background(), "Xyzzyville")
	if err == nil {
		t.Fatal("expected error for unknown location")
	}
	if !strings.Contains(err.Error(), "location not found") {
		t.Errorf("got error %q, expected 'location not found'", err.Error())
	}
}

func TestOpenMeteoClient_GeocodingError(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("/v1/search", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	})

	srv := httptest.NewServer(mux)
	defer srv.Close()

	client := &OpenMeteoClient{
		httpClient: srv.Client(),
		geocodeURL: srv.URL + "/v1/search",
		weatherURL: srv.URL + "/v1/forecast",
	}

	_, err := client.GetWeather(context.Background(), "Melbourne")
	if err == nil {
		t.Fatal("expected error for geocode failure")
	}
}

func TestOpenMeteoClient_WeatherAPIError(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("/v1/search", func(w http.ResponseWriter, r *http.Request) {
		resp := geocodeResponse{
			Results: []geoResult{
				{Name: "Melbourne", Country: "Australia", Latitude: -37.8142, Longitude: 144.9632},
			},
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(resp)
	})
	mux.HandleFunc("/v1/forecast", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusServiceUnavailable)
	})

	srv := httptest.NewServer(mux)
	defer srv.Close()

	client := &OpenMeteoClient{
		httpClient: srv.Client(),
		geocodeURL: srv.URL + "/v1/search",
		weatherURL: srv.URL + "/v1/forecast",
	}

	_, err := client.GetWeather(context.Background(), "Melbourne")
	if err == nil {
		t.Fatal("expected error for weather API failure")
	}
}

func TestOpenMeteoClient_ImplementsWeatherProvider(t *testing.T) {
	var _ WeatherProvider = (*OpenMeteoClient)(nil)
}

func TestWeatherCodeToDescription(t *testing.T) {
	tests := []struct {
		code int
		want string
	}{
		{0, "Clear sky"},
		{1, "Mainly clear"},
		{2, "Partly cloudy"},
		{3, "Overcast"},
		{45, "Foggy"},
		{51, "Drizzle"},
		{61, "Rain"},
		{71, "Snow"},
		{95, "Thunderstorm"},
		{99, "Thunderstorm with hail"},
		{999, "Unknown"},
	}
	for _, tt := range tests {
		got := weatherCodeToDescription(tt.code)
		if got != tt.want {
			t.Errorf("weatherCodeToDescription(%d) = %q, want %q", tt.code, got, tt.want)
		}
	}
}

func TestParseLocation(t *testing.T) {
	tests := []struct {
		input      string
		wantCity   string
		wantQuals  int
	}{
		{"Melbourne", "Melbourne", 0},
		{"Gosford, Australia", "Gosford", 1},
		{"Portland, Oregon, USA", "Portland", 2},
	}
	for _, tt := range tests {
		city, quals := parseLocation(tt.input)
		if city != tt.wantCity {
			t.Errorf("parseLocation(%q) city = %q, want %q", tt.input, city, tt.wantCity)
		}
		if len(quals) != tt.wantQuals {
			t.Errorf("parseLocation(%q) qualifiers = %d, want %d", tt.input, len(quals), tt.wantQuals)
		}
	}
}
