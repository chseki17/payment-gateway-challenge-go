package e2e

import (
	"context"
	"encoding/json"
	"net/http"
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestPayments_AuthorizationBehavior(t *testing.T) {
	t.Parallel()

	apiURL := os.Getenv("TEST_API_BASE_URL")
	if apiURL == "" {
		apiURL = "http://localhost:8090"
	}

	client := NewTestClient(apiURL)

	type response struct {
		ID     string `json:"id"`
		Status string `json:"status"`
	}

	tests := []struct {
		name           string
		cardNumber     string
		expectedStatus int
		expectedState  string
	}{
		{
			name:           "authorized when card ends with odd digit",
			cardNumber:     "4111111111111111", // ends with 1
			expectedStatus: http.StatusOK,
			expectedState:  "authorized",
		},
		{
			name:           "declined when card ends with even digit",
			cardNumber:     "4111111111111112", // ends with 2
			expectedStatus: http.StatusOK,
			expectedState:  "declined",
		},
		{
			name:           "service unavailable when card ends with zero",
			cardNumber:     "4111111111111110", // ends with 0
			expectedStatus: http.StatusInternalServerError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx, cancel := context.WithTimeout(
				context.Background(),
				10*time.Second,
			)
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
			require.Equal(t, tt.expectedStatus, resp.StatusCode)

			if resp.StatusCode == http.StatusOK {
				var r response
				require.NoError(t, json.Unmarshal(body, &r))
				require.Equal(t, tt.expectedState, r.Status)
				require.NotEmpty(t, r.ID)
			}
		})
	}
}
