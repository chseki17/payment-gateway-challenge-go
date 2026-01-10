package api

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
)

// ErrJSONResponseSerialization is returned when the response payload
// cannot be serialized to JSON or written to the client.
var ErrJSONResponseSerialization = errors.New("json response serialization error")

// OKResponse writes a 200 OK JSON response.
//
// It sets the Content-Type header to application/json and serializes
// the given data as JSON.
//
// If serialization or writing fails, an error is returned.
func OKResponse(w http.ResponseWriter, data any) error {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)

	if err := json.NewEncoder(w).Encode(data); err != nil {
		return fmt.Errorf("encode ok response: %w", ErrJSONResponseSerialization)
	}

	return nil
}

type ErrorResponseBody struct {
	Error string `json:"error"`
}

// ErrorResponse writes an error response with the given HTTP status code.
func ErrorResponse(w http.ResponseWriter, status int, message string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)

	json.NewEncoder(w).Encode(ErrorResponseBody{Error: message})
}
