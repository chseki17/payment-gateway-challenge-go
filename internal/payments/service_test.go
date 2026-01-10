package payments_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/cko-recruitment/payment-gateway-challenge-go/internal/banks/simulator"
	"github.com/cko-recruitment/payment-gateway-challenge-go/internal/payments"
	"github.com/stretchr/testify/require"
)

type mockPaymentsRepository struct {
	addFn func(ctx context.Context, payment *payments.Payment) error
	getFn func(ctx context.Context, id string) (*payments.Payment, error)
}

func (m *mockPaymentsRepository) AddPayment(
	ctx context.Context,
	payment *payments.Payment,
) error {
	return m.addFn(ctx, payment)
}

func (m *mockPaymentsRepository) GetPayment(
	ctx context.Context,
	id string,
) (*payments.Payment, error) {
	return m.getFn(ctx, id)
}

type mockBankingSimulator struct {
	authorizeFn func(
		ctx context.Context,
		req simulator.AuthorizationRequest,
	) (*simulator.AuthorizationResponse, error)
}

func (m *mockBankingSimulator) Authorize(
	ctx context.Context,
	req simulator.AuthorizationRequest,
) (*simulator.AuthorizationResponse, error) {
	return m.authorizeFn(ctx, req)
}

func validPaymentRequest() payments.PaymentRequest {
	now := time.Now()

	return payments.PaymentRequest{
		CardNumber:  "4111111111111111",
		ExpiryMonth: int(now.Month()) + 1,
		ExpiryYear:  now.Year(),
		Currency:    "USD",
		Amount:      1000,
		CVV:         "123",
	}
}

func TestService_CreatePayment_ValidationError(t *testing.T) {
	repo := &mockPaymentsRepository{}
	bank := &mockBankingSimulator{}

	service := payments.NewService(repo, bank)

	req := validPaymentRequest()
	req.CardNumber = "123"

	payment, err := service.CreatePayment(context.Background(), req)

	require.Nil(t, payment)
	require.Error(t, err)

	var invalidErr *payments.InvalidPaymentRequestErr
	require.ErrorAs(t, err, &invalidErr)
}

func TestService_CreatePayment_BankError(t *testing.T) {
	repo := &mockPaymentsRepository{}
	bank := &mockBankingSimulator{
		authorizeFn: func(ctx context.Context, req simulator.AuthorizationRequest) (*simulator.AuthorizationResponse, error) {
			return &simulator.AuthorizationResponse{}, errors.New("bank unavailable")
		},
	}

	service := payments.NewService(repo, bank)

	payment, err := service.CreatePayment(context.Background(), validPaymentRequest())

	require.Nil(t, payment)
	require.Error(t, err)
}

func TestService_CreatePayment_Authorized(t *testing.T) {
	repo := &mockPaymentsRepository{
		addFn: func(ctx context.Context, p *payments.Payment) error {
			require.Equal(t, payments.StatusAuthorized, p.Status)
			require.Equal(t, "1111", p.CardNumberLastFour)
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

	service := payments.NewService(repo, bank)

	paymentReq := validPaymentRequest()
	payment, err := service.CreatePayment(context.Background(), paymentReq)

	require.NoError(t, err)
	require.NotNil(t, payment)

	// Assert full domain result
	require.Equal(t, payments.StatusAuthorized, payment.Status)
	require.Equal(t, "1111", payment.CardNumberLastFour)
	require.Equal(t, paymentReq.ExpiryMonth, payment.ExpiryMonth)
	require.Equal(t, paymentReq.ExpiryYear, payment.ExpiryYear)
	require.Equal(t, paymentReq.Currency, payment.Currency)
	require.Equal(t, paymentReq.Amount, payment.Amount)
}

func TestService_CreatePayment_Declined(t *testing.T) {
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

	service := payments.NewService(repo, bank)

	payment, err := service.CreatePayment(context.Background(), validPaymentRequest())

	require.NoError(t, err)
	require.Equal(t, payments.StatusDeclined, payment.Status)
}

func TestService_CreatePayment_RepositoryError(t *testing.T) {
	repo := &mockPaymentsRepository{
		addFn: func(ctx context.Context, p *payments.Payment) error {
			return errors.New("db error")
		},
	}

	bank := &mockBankingSimulator{
		authorizeFn: func(ctx context.Context, req simulator.AuthorizationRequest) (*simulator.AuthorizationResponse, error) {
			return &simulator.AuthorizationResponse{Authorized: true}, nil
		},
	}

	service := payments.NewService(repo, bank)

	payment, err := service.CreatePayment(context.Background(), validPaymentRequest())

	require.Nil(t, payment)
	require.Error(t, err)
}

func TestService_GetPayment(t *testing.T) {
	expected := &payments.Payment{ID: "123"}

	repo := &mockPaymentsRepository{
		getFn: func(ctx context.Context, id string) (*payments.Payment, error) {
			require.Equal(t, "123", id)
			return expected, nil
		},
	}

	service := payments.NewService(repo, nil)

	payment, err := service.GetPayment(context.Background(), "123")

	require.NoError(t, err)
	require.Equal(t, expected, payment)
}
