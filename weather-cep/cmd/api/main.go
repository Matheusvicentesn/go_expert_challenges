package main

import (
	"log"
	"net/http"

	"weather-cep/internal/handlers"

	"github.com/go-chi/chi/v5"
	"github.com/joho/godotenv"
)

func init() {
	godotenv.Load()
}

func main() {
	r := chi.NewRouter()

	r.Get("/weather/{cep}", handlers.GetWeatherByCEP)

	log.Println("API rodando na porta 8080")
	http.ListenAndServe(":8080", r)
}
