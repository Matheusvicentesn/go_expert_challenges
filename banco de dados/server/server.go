package main

import (
	"context"
	"database/sql"
	"encoding/json"
	"log"
	"net/http"
	"time"

	_ "github.com/mattn/go-sqlite3"
)

type APIResponse struct {
	USDBRL struct {
		Bid string `json:"bid"`
	} `json:"USDBRL"`
}

type ServerResponse struct {
	Bid string `json:"bid"`
}

func main() {
	db, err := sql.Open("sqlite3", "./cotacao.db")
	if err != nil {
		log.Fatal(err)
	}

	_, err = db.Exec(`CREATE TABLE IF NOT EXISTS cotacoes (id INTEGER PRIMARY KEY, bid TEXT, data TIMESTAMP)`)
	if err != nil {
		log.Fatal(err)
	}

	http.HandleFunc("/cotacao", func(w http.ResponseWriter, r *http.Request) {
		ctxAPI, cancelAPI := context.WithTimeout(context.Background(), 200*time.Millisecond)
		defer cancelAPI()

		req, _ := http.NewRequestWithContext(ctxAPI, http.MethodGet,
			"https://economia.awesomeapi.com.br/json/last/USD-BRL", nil)

		resp, err := http.DefaultClient.Do(req)
		if err != nil {
			log.Println("Erro ao chamar API externa:", err)
			http.Error(w, err.Error(), http.StatusGatewayTimeout)
			return
		}
		defer resp.Body.Close()

		var apiResp APIResponse
		if err := json.NewDecoder(resp.Body).Decode(&apiResp); err != nil {
			log.Println("Erro ao decodificar JSON externo:", err)
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		bid := apiResp.USDBRL.Bid

		ctxDB, cancelDB := context.WithTimeout(context.Background(), 10*time.Millisecond)
		defer cancelDB()

		_, err = db.ExecContext(ctxDB, "INSERT INTO cotacoes (bid, data) VALUES (?, ?)", bid, time.Now())
		if err != nil {
			log.Println("Erro ao inserir no banco:", err)
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(ServerResponse{Bid: bid})
	})

	log.Println("Servidor rodando na porta 8080...")
	http.ListenAndServe(":8080", nil)
}
