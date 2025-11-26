package handlers

import (
	"encoding/json"
	"fmt"
	"net/http"
	"regexp"
	"strings"

	"weather-cep/internal/services"
	"weather-cep/internal/utils"

	"github.com/go-chi/chi/v5"
)

func GetWeatherByCEP(w http.ResponseWriter, r *http.Request) {
	cep := chi.URLParam(r, "cep")

	pattern := regexp.MustCompile(`^\d{5}-?\d{3}$`)
	if !pattern.MatchString(cep) {
		http.Error(w, `invalid zipcode`, http.StatusUnprocessableEntity)
		return
	}

	cep = strings.ReplaceAll(cep, "-", "")

	viaCepData, err := services.GetLocationByCEP(cep)
	if err != nil {
		if err == services.ErrCEPNotFound {
			http.Error(w, `can not find zipcode`, http.StatusNotFound)
			return
		}

		http.Error(w, `internal error`, http.StatusInternalServerError)
		return
	}

	weather, err := services.GetWeatherByCity(viaCepData.Localidade, viaCepData.UF)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	response := map[string]interface{}{
		"location": fmt.Sprintf("%s, %s - %s",
			viaCepData.Logradouro,
			viaCepData.Localidade,
			viaCepData.UF,
		),
		"temperatures": map[string]string{
			"temp_C":            fmt.Sprintf("%.1f °C", weather.TempC),
			"temp_F":            fmt.Sprintf("%.1f °F", utils.CToF(weather.TempC)),
			"temp_K":            fmt.Sprintf("%.1f K", utils.CToK(weather.TempC)),
			"temp_C_feels_like": fmt.Sprintf("%.1f °C", weather.FeelsLike),
			"temp_C_min":        fmt.Sprintf("%.1f °C", weather.Min),
			"temp_C_max":        fmt.Sprintf("%.1f °C", weather.Max),
		},
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}
