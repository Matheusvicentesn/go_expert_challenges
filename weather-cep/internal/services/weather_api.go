package services

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
)

type OpenWeatherResponse struct {
	Main struct {
		Temp      float64 `json:"temp"`
		FeelsLike float64 `json:"feels_like"`
		TempMin   float64 `json:"temp_min"`
		TempMax   float64 `json:"temp_max"`
		Pressure  float64 `json:"pressure"`
		Humidity  float64 `json:"humidity"`
	} `json:"main"`
}

type WeatherResult struct {
	TempC     float64
	FeelsLike float64
	Min       float64
	Max       float64
}

func GetWeatherByCity(city, uf string) (*WeatherResult, error) {
	apiKey := os.Getenv("WEATHER_API_KEY")

	query := url.QueryEscape(fmt.Sprintf("%s,%s,BR", city, uf))

	url := fmt.Sprintf(
		"https://api.openweathermap.org/data/2.5/weather?q=%s&appid=%s&units=metric",
		query,
		apiKey,
	)

	resp, err := http.Get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var data OpenWeatherResponse
	if err := json.Unmarshal(body, &data); err != nil {
		return nil, err
	}

	result := &WeatherResult{
		TempC:     data.Main.Temp,
		FeelsLike: data.Main.FeelsLike,
		Min:       data.Main.TempMin,
		Max:       data.Main.TempMax,
	}

	return result, nil
}
