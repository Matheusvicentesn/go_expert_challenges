package main

import (
	"context"
	"encoding/json"
	"log"
	"net/http"
	"os"
	"time"
)

type ServerResponse struct {
	Bid string `json:"bid"`
}

func main() {
	ctx, cancel := context.WithTimeout(context.Background(), 300*time.Millisecond)
	defer cancel()

	req, _ := http.NewRequestWithContext(ctx, http.MethodGet, "http://localhost:8080/cotacao", nil)

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		log.Println("Erro ao chamar servidor:", err)
		return
	}
	defer resp.Body.Close()

	var data ServerResponse
	if err := json.NewDecoder(resp.Body).Decode(&data); err != nil {
		log.Println("Erro ao decodificar resposta:", err)
		return
	}

	file, err := os.Create("cotacao.txt")
	if err != nil {
		log.Println("Erro ao criar arquivo:", err)
		return
	}
	defer file.Close()

	file.WriteString("Dólar: " + data.Bid)
	log.Println("Cotação salva em cotacao.txt:", data.Bid)
}
