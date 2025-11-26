package tests

import (
	"net/http/httptest"
	"testing"

	"weather-cep/internal/handlers"

	"github.com/go-chi/chi/v5"
)

func TestInvalidCEP(t *testing.T) {
	req := httptest.NewRequest("GET", "/weather/123", nil)
	w := httptest.NewRecorder()

	r := chi.NewRouter()
	r.Get("/weather/{cep}", handlers.GetWeatherByCEP)

	r.ServeHTTP(w, req)

	if w.Code != 422 {
		t.Errorf("Esperado 422, recebido %d", w.Code)
	}
}
