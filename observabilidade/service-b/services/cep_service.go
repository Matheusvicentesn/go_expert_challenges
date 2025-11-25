package services

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
)

var ErrCEPNotFound = errors.New("can not find zipcode")

type ViaCEPResponse struct {
	CEP         string `json:"cep"`
	Logradouro  string `json:"logradouro"`
	Complemento string `json:"complemento"`
	Bairro      string `json:"bairro"`
	Localidade  string `json:"localidade"`
	UF          string `json:"uf"`
	IBGE        string `json:"ibge"`
	GIA         string `json:"gia"`
	DDD         string `json:"ddd"`
	SIAFI       string `json:"siafi"`
	Erro        bool   `json:"erro,omitempty"`
}

type Location struct {
	City string
	UF   string
}

func GetLocationByCEP(ctx context.Context, cep string) (*Location, error) {
	tracer := otel.Tracer("service-b")
	ctx, span := tracer.Start(ctx, "fetch-cep-data")
	defer span.End()

	span.SetAttributes(attribute.String("cep", cep))

	url := fmt.Sprintf("https://viacep.com.br/ws/%s/json/", cep)
	span.SetAttributes(attribute.String("viacep.url", url))

	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, "failed to create request")
		return nil, err
	}

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, "failed to call viacep")
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

	log.Printf("ViaCEP response for CEP %s: status=%d, body=%s", cep, resp.StatusCode, string(body))

	var data ViaCEPResponse
	if err := json.Unmarshal(body, &data); err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, "failed to parse json")
		return nil, err
	}

	if data.Erro {
		log.Printf("CEP %s not found (erro field = true)", cep)
		span.SetStatus(codes.Error, "cep not found")
		return nil, ErrCEPNotFound
	}

	if data.Localidade == "" {
		log.Printf("CEP %s not found (empty localidade)", cep)
		span.SetStatus(codes.Error, "cep not found")
		return nil, ErrCEPNotFound
	}

	location := &Location{
		City: data.Localidade,
		UF:   data.UF,
	}

	span.SetAttributes(
		attribute.String("location.city", location.City),
		attribute.String("location.uf", location.UF),
	)
	span.SetStatus(codes.Ok, "success")

	log.Printf("CEP %s found: %s, %s", cep, location.City, location.UF)

	return location, nil
}
