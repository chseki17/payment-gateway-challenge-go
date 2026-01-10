package simulator_test

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/cko-recruitment/payment-gateway-challenge-go/internal/banks/simulator"
)

func TestClient_Authorize_Success(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		require.Equal(t, http.MethodPost, r.Method)
		require.Equal(t, "/payments", r.URL.Path)

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)

		_ = json.NewEncoder(w).Encode(simulator.AuthorizationResponse{
			Authorized:        true,
			AuthorizationCode: "auth_123",
		})
	}))
	t.Cleanup(server.Close)

	client := simulator.NewClient(server.URL, server.Client())

	resp, err := client.Authorize(context.Background(), simulator.AuthorizationRequest{
		CardNumber: "4111111111111111",
		ExpiryDate: "12/2029",
		Currency:   "USD",
		Amount:     1000,
		CVV:        "123",
	})

	require.NoError(t, err)
	require.NotNil(t, resp)

	assert.True(t, resp.Authorized)
	assert.Equal(t, "auth_123", resp.AuthorizationCode)
}

func TestClient_Authorize_Rejected(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusBadRequest)
		_ = json.NewEncoder(w).Encode(map[string]string{
			"error_message": "card declined",
		})
	}))
	t.Cleanup(server.Close)

	client := simulator.NewClient(server.URL, server.Client())

	_, err := client.Authorize(context.Background(), simulator.AuthorizationRequest{})
	require.Error(t, err)

	assert.ErrorIs(t, err, simulator.ErrAuthorizationRejected)
	assert.Contains(t, err.Error(), "card declined")
}

func TestClient_Authorize_Unavailable(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusServiceUnavailable)
	}))
	t.Cleanup(server.Close)

	client := simulator.NewClient(server.URL, server.Client())

	_, err := client.Authorize(context.Background(), simulator.AuthorizationRequest{})
	require.Error(t, err)

	assert.ErrorIs(t, err, simulator.ErrAuthorizationUnavailable)
}

func TestClient_Authorize_UnexpectedStatus(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusTeapot)
	}))
	t.Cleanup(server.Close)

	client := simulator.NewClient(server.URL, server.Client())

	_, err := client.Authorize(context.Background(), simulator.AuthorizationRequest{})
	require.Error(t, err)

	assert.ErrorIs(t, err, simulator.ErrAuthorizationUnexpected)
}
