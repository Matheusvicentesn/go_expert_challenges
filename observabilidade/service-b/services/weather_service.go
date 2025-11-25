package services

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
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

func GetWeatherByCity(ctx context.Context, city, uf string) (*WeatherResult, error) {
	tracer := otel.Tracer("service-b")
	ctx, span := tracer.Start(ctx, "fetch-weather-data")
	defer span.End()

	span.SetAttributes(
		attribute.String("city", city),
		attribute.String("uf", uf),
	)

	apiKey := os.Getenv("WEATHER_API_KEY")
	if apiKey == "" {
		err := fmt.Errorf("WEATHER_API_KEY not configured")
		span.RecordError(err)
		span.SetStatus(codes.Error, "missing api key")
		return nil, err
	}

	query := url.QueryEscape(fmt.Sprintf("%s,%s,BR", city, uf))
	weatherURL := fmt.Sprintf(
		"https://api.openweathermap.org/data/2.5/weather?q=%s&appid=%s&units=metric",
		query,
		apiKey,
	)

	span.SetAttributes(attribute.String("weather.api.url", weatherURL))

	req, err := http.NewRequestWithContext(ctx, "GET", weatherURL, nil)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, "failed to create request")
		return nil, err
	}

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, "failed to call weather api")
		return nil, err
	}
	defer resp.Body.Close()

	span.SetAttributes(attribute.Int("http.status_code", resp.StatusCode))

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, "failed to read response")
		return nil, err
	}

	var data OpenWeatherResponse
	if err := json.Unmarshal(body, &data); err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, "failed to parse json")
		return nil, err
	}

	result := &WeatherResult{
		TempC:     data.Main.Temp,
		FeelsLike: data.Main.FeelsLike,
		Min:       data.Main.TempMin,
		Max:       data.Main.TempMax,
	}

	span.SetAttributes(
		attribute.Float64("temperature.celsius", result.TempC),
		attribute.Float64("temperature.feels_like", result.FeelsLike),
	)
	span.SetStatus(codes.Ok, "success")

	return result, nil
}
