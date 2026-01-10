package api_test

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/cko-recruitment/payment-gateway-challenge-go/internal/api"
	"github.com/cko-recruitment/payment-gateway-challenge-go/internal/banks/simulator"
	"github.com/cko-recruitment/payment-gateway-challenge-go/internal/payments"
	"github.com/go-chi/chi/v5"
	"github.com/stretchr/testify/require"
)

type mockPaymentsRepository struct {
	getFn func(ctx context.Context, id string) (*payments.Payment, error)
	addFn func(ctx context.Context, payment *payments.Payment) error
}

func (m *mockPaymentsRepository) GetPayment(ctx context.Context, id string) (*payments.Payment, error) {
	return m.getFn(ctx, id)
}

func (m *mockPaymentsRepository) AddPayment(ctx context.Context, payment *payments.Payment) error {
	return m.addFn(ctx, payment)
}

type mockBankingSimulator struct {
	authorizeFn func(ctx context.Context, req simulator.AuthorizationRequest) (*simulator.AuthorizationResponse, error)
}

func (m *mockBankingSimulator) Authorize(
	ctx context.Context,
	req simulator.AuthorizationRequest,
) (*simulator.AuthorizationResponse, error) {
	return m.authorizeFn(ctx, req)
}

func TestPaymentsHandler_PostHandler_Authorized(t *testing.T) {
	repo := &mockPaymentsRepository{
		addFn: func(ctx context.Context, p *payments.Payment) error {
			require.Equal(t, payments.StatusAuthorized, p.Status)
			return nil
		},
	}

	bank := &mockBankingSimulator{
		authorizeFn: func(ctx context.Context, req simulator.AuthorizationRequest) (*simulator.AuthorizationResponse, error) {
			return &simulator.AuthorizationResponse{
				Authorized:        true,
				AuthorizationCode: "AUTH123",
			}, nil
		},
	}

	svc := payments.NewService(repo, bank)
	handler := api.NewPaymentsHandler(svc)

	body := payments.PaymentRequest{
		CardNumber:  "4111111111111111",
		ExpiryMonth: 4,
		ExpiryYear:  2026,
		Currency:    "USD",
		Amount:      1000,
		CVV:         "123",
	}

	payload, err := json.Marshal(body)
	require.NoError(t, err)

	req := httptest.NewRequest(http.MethodPost, "/payments", bytes.NewReader(payload))
	rec := httptest.NewRecorder()

	handler.PostHandler().ServeHTTP(rec, req)

	require.Equal(t, http.StatusOK, rec.Code)
}

func TestPaymentsHandler_PostHandler_Declined(t *testing.T) {
	repo := &mockPaymentsRepository{
		addFn: func(ctx context.Context, p *payments.Payment) error {
			require.Equal(t, payments.StatusDeclined, p.Status)
			return nil
		},
	}

	bank := &mockBankingSimulator{
		authorizeFn: func(ctx context.Context, req simulator.AuthorizationRequest) (*simulator.AuthorizationResponse, error) {
			return &simulator.AuthorizationResponse{Authorized: false}, nil
		},
	}

	svc := payments.NewService(repo, bank)
	handler := api.NewPaymentsHandler(svc)

	req := httptest.NewRequest(
		http.MethodPost,
		"/payments",
		bytes.NewBufferString(`{
			"card_number":"4111111111111111",
			"expiry_month":4,
			"expiry_year":2026,
			"currency":"USD",
			"amount":1000,
			"cvv":"123"
		}`),
	)

	rec := httptest.NewRecorder()
	handler.PostHandler().ServeHTTP(rec, req)

	require.Equal(t, http.StatusOK, rec.Code)
}

func TestPaymentsHandler_PostHandler_InvalidJSON(t *testing.T) {
	svc := payments.NewService(nil, nil)

	handler := api.NewPaymentsHandler(svc)
	req := httptest.NewRequest(
		http.MethodPost,
		"/payments",
		bytes.NewBufferString(`{invalid json`),
	)
	rec := httptest.NewRecorder()

	handler.PostHandler().ServeHTTP(rec, req)

	require.Equal(t, http.StatusBadRequest, rec.Code)
}

func TestPaymentsHandler_PostHandler_BankError(t *testing.T) {
	repo := &mockPaymentsRepository{}
	bank := &mockBankingSimulator{
		authorizeFn: func(ctx context.Context, req simulator.AuthorizationRequest) (*simulator.AuthorizationResponse, error) {
			return &simulator.AuthorizationResponse{}, errors.New("bank down")
		},
	}

	svc := payments.NewService(repo, bank)
	handler := api.NewPaymentsHandler(svc)

	req := httptest.NewRequest(
		http.MethodPost,
		"/payments",
		bytes.NewBufferString(`{
			"card_number":"4111111111111111",
			"expiry_month":4,
			"expiry_year":2026,
			"currency":"USD",
			"amount":1000,
			"cvv":"123"
		}`),
	)

	rec := httptest.NewRecorder()
	handler.PostHandler().ServeHTTP(rec, req)

	require.Equal(t, http.StatusInternalServerError, rec.Code)
}

func TestPaymentsHandler_GetHandler_Found(t *testing.T) {
	repo := &mockPaymentsRepository{
		getFn: func(ctx context.Context, id string) (*payments.Payment, error) {
			return &payments.Payment{ID: id}, nil
		},
	}

	svc := payments.NewService(repo, nil)
	handler := api.NewPaymentsHandler(svc)

	req := httptest.NewRequest(http.MethodGet, "/payments/123", nil)
	rec := httptest.NewRecorder()

	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("id", "123")
	req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

	handler.GetHandler().ServeHTTP(rec, req)

	require.Equal(t, http.StatusOK, rec.Code)
}
