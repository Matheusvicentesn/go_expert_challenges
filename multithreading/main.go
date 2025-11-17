package main

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"time"
)

type BrasilAPIResponse struct {
	Cep          string `json:"cep"`
	State        string `json:"state"`
	City         string `json:"city"`
	Neighborhood string `json:"neighborhood"`
	Street       string `json:"street"`
}

type ViaCEPResponse struct {
	Cep        string `json:"cep"`
	Logradouro string `json:"logradouro"`
	Bairro     string `json:"bairro"`
	Localidade string `json:"localidade"`
	Uf         string `json:"uf"`
}

func fetchFromBrasilAPI(ctx context.Context, cep string) (string, error) {
	url := "https://brasilapi.com.br/api/cep/v1/" + cep
	req, _ := http.NewRequestWithContext(ctx, "GET", url, nil)

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	var data BrasilAPIResponse
	if err := json.NewDecoder(resp.Body).Decode(&data); err != nil {
		return "", err
	}

	return fmt.Sprintf(
		"[BrasilAPI]\nCEP: %s\nEstado: %s\nCidade: %s\nBairro: %s\nRua: %s\n",
		data.Cep, data.State, data.City, data.Neighborhood, data.Street,
	), nil
}

func fetchFromViaCEP(ctx context.Context, cep string) (string, error) {
	url := "http://viacep.com.br/ws/" + cep + "/json/"
	req, _ := http.NewRequestWithContext(ctx, "GET", url, nil)

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	var data ViaCEPResponse
	if err := json.NewDecoder(resp.Body).Decode(&data); err != nil {
		return "", err
	}

	return fmt.Sprintf(
		"[ViaCEP]\nCEP: %s\nEstado: %s\nCidade: %s\nBairro: %s\nRua: %s\n",
		data.Cep, data.Uf, data.Localidade, data.Bairro, data.Logradouro,
	), nil
}

func main() {
	if len(os.Args) < 2 {
		fmt.Println("Uso: go run main.go <CEP>")
		return
	}

	cep := os.Args[1]

	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
	defer cancel()

	result := make(chan string)
	errChan := make(chan error)

	go func() {
		resp, err := fetchFromBrasilAPI(ctx, cep)
		if err != nil {
			errChan <- err
			return
		}

		result <- resp
	}()

	go func() {
		resp, err := fetchFromViaCEP(ctx, cep)
		if err != nil {
			errChan <- err
			return
		}

		result <- resp
	}()

	select {
	case r := <-result:
		fmt.Println("Resultado mais rápido:")
		fmt.Println(r)
	case err := <-errChan:
		fmt.Println("Erro:", err)
	case <-ctx.Done():
		fmt.Println("Timeout de 1 segundo!")
	}
}
