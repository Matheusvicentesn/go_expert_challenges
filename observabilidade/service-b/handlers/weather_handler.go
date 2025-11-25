package handlers

import (
	"encoding/json"
	"log"
	"net/http"
	"regexp"
	"strings"

	"service-b/services"
	"service-b/utils"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/propagation"
)

type WeatherResponse struct {
	City  string  `json:"city"`
	TempC float64 `json:"temp_C"`
	TempF float64 `json:"temp_F"`
	TempK float64 `json:"temp_K"`
}

type ErrorResponse struct {
	Message string `json:"message"`
}

var cepRegex = regexp.MustCompile(`^\d{8}$`)

func HandleWeather(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	ctx = otel.GetTextMapPropagator().Extract(ctx, propagation.HeaderCarrier(r.Header))

	tracer := otel.Tracer("service-b")
	ctx, span := tracer.Start(ctx, "handle-weather-request")
	defer span.End()

	if r.Method != http.MethodGet {
		span.SetStatus(codes.Error, "method not allowed")
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	path := strings.TrimPrefix(r.URL.Path, "/weather/")
	cep := strings.TrimSpace(path)
	span.SetAttributes(attribute.String("cep", cep))

	if !cepRegex.MatchString(cep) {
		span.SetStatus(codes.Error, "invalid zipcode format")
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusUnprocessableEntity)
		json.NewEncoder(w).Encode(ErrorResponse{Message: "invalid zipcode"})
		return
	}

	location, err := services.GetLocationByCEP(ctx, cep)
	if err != nil {
		log.Printf("CEP %s not found or error: %v", cep, err)
		span.SetStatus(codes.Error, "can not find zipcode")
		span.SetAttributes(attribute.String("error.message", err.Error()))
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusNotFound)
		json.NewEncoder(w).Encode(ErrorResponse{Message: "can not find zipcode"})
		return
	}

	span.SetAttributes(
		attribute.String("location.city", location.City),
		attribute.String("location.uf", location.UF),
	)

	weather, err := services.GetWeatherByCity(ctx, location.City, location.UF)
	if err != nil {
		log.Printf("Error getting weather for city %s: %v", location.City, err)
		span.RecordError(err)
		span.SetStatus(codes.Error, "failed to get weather")
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(ErrorResponse{Message: "internal server error"})
		return
	}

	tempF := utils.CToF(weather.TempC)
	tempK := utils.CToK(weather.TempC)

	response := WeatherResponse{
		City:  location.City,
		TempC: weather.TempC,
		TempF: tempF,
		TempK: tempK,
	}

	span.SetAttributes(
		attribute.Float64("temperature.celsius", weather.TempC),
		attribute.Float64("temperature.fahrenheit", tempF),
		attribute.Float64("temperature.kelvin", tempK),
	)
	span.SetStatus(codes.Ok, "success")

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(response)
}
