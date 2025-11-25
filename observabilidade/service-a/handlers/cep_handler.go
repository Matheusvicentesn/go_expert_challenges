package handlers

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"regexp"
	"strings"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/propagation"
)

type CEPRequest struct {
	CEP string `json:"cep"`
}

type ErrorResponse struct {
	Message string `json:"message"`
}

var cepRegex = regexp.MustCompile(`^\d{5}-?\d{3}$`)

func HandleCEP(w http.ResponseWriter, r *http.Request) {
	tracer := otel.Tracer("service-a")
	ctx, span := tracer.Start(r.Context(), "handle-cep-request")
	defer span.End()

	if r.Method != http.MethodPost {
		span.SetStatus(codes.Error, "method not allowed")
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req CEPRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, "invalid json")
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(ErrorResponse{Message: "invalid request body"})
		return
	}

	req.CEP = strings.ReplaceAll(req.CEP, "-", "")

	span.SetAttributes(attribute.String("cep", req.CEP))

	if err := validateCEP(ctx, req.CEP); err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, "invalid zipcode")
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusUnprocessableEntity)
		json.NewEncoder(w).Encode(ErrorResponse{Message: "invalid zipcode"})
		return
	}

	response, statusCode, err := forwardToServiceB(ctx, req.CEP)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, "failed to forward to service-b")
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(ErrorResponse{Message: "internal server error"})
		return
	}

	span.SetStatus(codes.Ok, "success")
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	w.Write(response)
}

func validateCEP(ctx context.Context, cep string) error {
	tracer := otel.Tracer("service-a")
	_, span := tracer.Start(ctx, "validate-cep")
	defer span.End()

	span.SetAttributes(
		attribute.String("cep", cep),
		attribute.Int("cep.length", len(cep)),
	)

	if !cepRegex.MatchString(cep) {
		span.SetStatus(codes.Error, "invalid format")
		return fmt.Errorf("CEP must contain exactly 8 digits")
	}

	span.SetStatus(codes.Ok, "valid")
	return nil
}

func forwardToServiceB(ctx context.Context, cep string) ([]byte, int, error) {
	tracer := otel.Tracer("service-a")
	ctx, span := tracer.Start(ctx, "forward-to-service-b")
	defer span.End()

	serviceBURL := os.Getenv("SERVICE_B_URL")
	if serviceBURL == "" {
		serviceBURL = "http://localhost:8081"
	}

	url := fmt.Sprintf("%s/weather/%s", serviceBURL, cep)
	span.SetAttributes(attribute.String("service.b.url", url))

	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, "failed to create request")
		return nil, 0, err
	}

	otel.GetTextMapPropagator().Inject(ctx, propagation.HeaderCarrier(req.Header))

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, "failed to call service-b")
		return nil, 0, err
	}
	defer resp.Body.Close()

	span.SetAttributes(attribute.Int("http.status_code", resp.StatusCode))

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, "failed to read response")
		return nil, 0, err
	}

	if resp.StatusCode >= 200 && resp.StatusCode < 300 {
		span.SetStatus(codes.Ok, "success")
	} else {
		span.SetStatus(codes.Error, fmt.Sprintf("service-b returned %d", resp.StatusCode))
	}

	return body, resp.StatusCode, nil
}
