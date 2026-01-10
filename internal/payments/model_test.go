package payments_test

import (
	"testing"
	"time"

	"github.com/cko-recruitment/payment-gateway-challenge-go/internal/payments"
	"github.com/stretchr/testify/require"
)

func validRequest() payments.PaymentRequest {
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

func TestPaymentRequest_Validate(t *testing.T) {
	tests := []struct {
		name          string
		req           func() payments.PaymentRequest
		expectErr     bool
		expectedField string
	}{
		{
			name:      "valid request",
			req:       validRequest,
			expectErr: false,
		},
		{
			name: "invalid card number length",
			req: func() payments.PaymentRequest {
				r := validRequest()
				r.CardNumber = "123"
				return r
			},
			expectErr:     true,
			expectedField: "card_number",
		},
		{
			name: "non-numeric card number",
			req: func() payments.PaymentRequest {
				r := validRequest()
				r.CardNumber = "4111abcd11111111"
				return r
			},
			expectErr:     true,
			expectedField: "card_number",
		},
		{
			name: "invalid expiry month",
			req: func() payments.PaymentRequest {
				r := validRequest()
				r.ExpiryMonth = 13
				return r
			},
			expectErr:     true,
			expectedField: "expiry_month",
		},
		{
			name: "expiry date in the past",
			req: func() payments.PaymentRequest {
				r := validRequest()
				r.ExpiryYear = time.Now().Year() - 1
				return r
			},
			expectErr:     true,
			expectedField: "expiry_date",
		},
		{
			name: "unsupported currency",
			req: func() payments.PaymentRequest {
				r := validRequest()
				r.Currency = "GBP"
				return r
			},
			expectErr:     true,
			expectedField: "currency",
		},
		{
			name: "amount is zero",
			req: func() payments.PaymentRequest {
				r := validRequest()
				r.Amount = 0
				return r
			},
			expectErr:     true,
			expectedField: "amount",
		},
		{
			name: "invalid cvv",
			req: func() payments.PaymentRequest {
				r := validRequest()
				r.CVV = "12a"
				return r
			},
			expectErr:     true,
			expectedField: "cvv",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			err := tt.req().Validate()

			if !tt.expectErr {
				require.NoError(t, err)
				return
			}

			require.Error(t, err)

			var invalidErr *payments.InvalidPaymentRequestErr
			require.ErrorAs(t, err, &invalidErr)
			require.Equal(t, tt.expectedField, invalidErr.Field)
			require.NotEmpty(t, invalidErr.Message)
		})
	}
}
