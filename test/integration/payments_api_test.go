package e2e

import (
	"context"
	"encoding/json"
	"maps"
	"net/http"
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestPayments_AuthorizationAndRetrieval_Behavior(t *testing.T) {
	t.Parallel()

	apiURL := os.Getenv("TEST_API_BASE_URL")
	if apiURL == "" {
		apiURL = "http://localhost:8090"
	}

	client := NewTestClient(apiURL)

	type createResponse struct {
		ID     string `json:"id"`
		Status string `json:"status"`
	}

	type getResponse struct {
		ID     string `json:"id"`
		Status string `json:"status"`
	}

	tests := []struct {
		name                  string
		cardNumber            string
		expectedStatusCode    int
		expectedPaymentStatus string
	}{
		{
			name:                  "authorized when card ends with odd digit",
			cardNumber:            "4111111111111111",
			expectedStatusCode:    http.StatusOK,
			expectedPaymentStatus: "authorized",
		},
		{
			name:                  "declined when card ends with even digit",
			cardNumber:            "4111111111111112",
			expectedStatusCode:    http.StatusOK,
			expectedPaymentStatus: "declined",
		},
		{
			name:               "service unavailable when card ends with zero",
			cardNumber:         "4111111111111110",
			expectedStatusCode: http.StatusInternalServerError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
			defer cancel()

			req := map[string]any{
				"card_number":  tt.cardNumber,
				"expiry_month": 12,
				"expiry_year":  2050,
				"currency":     "USD",
				"amount":       1000,
				"cvv":          "123",
			}

			resp, body, err := client.Post(ctx, "/api/v1/payments", req)
			require.NoError(t, err)
			require.Equal(t, tt.expectedStatusCode, resp.StatusCode)

			// Only retrieve when creation succeeded
			if resp.StatusCode != http.StatusOK {
				return
			}

			var created createResponse
			require.NoError(t, json.Unmarshal(body, &created))
			require.NotEmpty(t, created.ID)
			require.Equal(t, tt.expectedPaymentStatus, created.Status)

			// --- Retrieve payment by ID ---
			var fetched getResponse
			getResp, err := client.Get(
				ctx,
				"/api/v1/payments/"+created.ID,
				&fetched,
			)
			require.NoError(t, err)
			require.Equal(t, http.StatusOK, getResp.StatusCode)

			require.Equal(t, created.ID, fetched.ID)
			require.Equal(t, created.Status, fetched.Status)
		})
	}
}

func TestPayments_BadRequest_ValidationErrors(t *testing.T) {
	t.Parallel()

	apiURL := os.Getenv("TEST_API_BASE_URL")
	if apiURL == "" {
		apiURL = "http://localhost:8090"
	}

	client := NewTestClient(apiURL)

	validRequest := map[string]any{
		"card_number":  "4111111111111111",
		"expiry_month": 12,
		"expiry_year":  2050,
		"currency":     "USD",
		"amount":       1000,
		"cvv":          "123",
	}

	tests := []struct {
		name   string
		mutate func(req map[string]any)
	}{
		{
			name: "missing card number",
			mutate: func(req map[string]any) {
				delete(req, "card_number")
			},
		},
		{
			name: "card number too short",
			mutate: func(req map[string]any) {
				req["card_number"] = "123"
			},
		},
		{
			name: "card number with non-numeric characters",
			mutate: func(req map[string]any) {
				req["card_number"] = "4111abcd11111111"
			},
		},
		{
			name: "expiry month out of range",
			mutate: func(req map[string]any) {
				req["expiry_month"] = 13
			},
		},
		{
			name: "expiry date in the past",
			mutate: func(req map[string]any) {
				req["expiry_year"] = time.Now().Year() - 1
			},
		},
		{
			name: "unsupported currency",
			mutate: func(req map[string]any) {
				req["currency"] = "JPY"
			},
		},
		{
			name: "currency not 3 characters",
			mutate: func(req map[string]any) {
				req["currency"] = "US"
			},
		},
		{
			name: "amount missing",
			mutate: func(req map[string]any) {
				delete(req, "amount")
			},
		},
		{
			name: "amount not integer",
			mutate: func(req map[string]any) {
				req["amount"] = "1000"
			},
		},
		{
			name: "cvv too short",
			mutate: func(req map[string]any) {
				req["cvv"] = "12"
			},
		},
		{
			name: "cvv contains non-numeric characters",
			mutate: func(req map[string]any) {
				req["cvv"] = "12a"
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx, cancel := context.WithTimeout(
				context.Background(),
				10*time.Second,
			)
			defer cancel()

			req := maps.Clone(validRequest)
			tt.mutate(req)

			resp, _, err := client.Post(ctx, "/api/v1/payments", req)
			require.NoError(t, err)
			require.Equal(t, http.StatusBadRequest, resp.StatusCode)
		})
	}
}
