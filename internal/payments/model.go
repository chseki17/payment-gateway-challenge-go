package payments

import (
	"context"
	"encoding/json"
	"errors"
	"time"
	"unicode"
)

type PaymentsRepository interface {
	GetPayment(ctx context.Context, id string) (*Payment, error)
	AddPayment(ctx context.Context, payment *Payment) error
}

var (
	NotFoundPaymentErr = errors.New("payment not found")
)

type InvalidPaymentRequestErr struct {
	Field   string
	Message string
}

func (e *InvalidPaymentRequestErr) Error() string {
	return e.Message
}

type PaymentStatus uint8

const (
	StatusPending PaymentStatus = iota
	StatusAuthorized
	StatusDeclined
	StatusRejected
)

func (s PaymentStatus) String() string {
	switch s {
	case StatusPending:
		return "pending"
	case StatusAuthorized:
		return "authorized"
	case StatusDeclined:
		return "declined"
	case StatusRejected:
		return "rejected"
	default:
		return "unknown"
	}
}

func (s PaymentStatus) MarshalJSON() ([]byte, error) {
	return json.Marshal(s.String())
}

type Payment struct {
	ID     string        `json:"id" example:"019ba901-48a1-7138-824e-d0e65a8dc38a"`                                             // Unique identifier of the payment.
	Status PaymentStatus `json:"status" swaggertype:"string" example:"authorized" enums:"authorized,declined,rejected,pending"` // Current status of the payment.
	// TODO: StatusDescription  string one possiblity to distinguich between errors better
	// StatusErrorCode int
	// 1 -> represents the card dont have enough money
	// 2 -> could be validation
	// and so on...
	CardNumberLastFour string `json:"card_number_last_four" example:"8877"`       // Last four digits of the card number used in the payment.
	ExpiryMonth        int    `json:"expiry_month" example:"12"`                  // Expiration month (1–12).
	ExpiryYear         int    `json:"expiry_year" example:"2050"`                 // Expiration year (four digits).
	Currency           string `json:"currency" example:"USD" enums:"USD,EUR,BRL"` // Currency code in ISO 4217 format (e.g. USD, EUR, BRL).
	Amount             int64  `json:"amount" example:"1000"`                      // Amount expressed in minor units of the given currency. Example: $10.99 USD → 1099
}

type PaymentRequest struct {
	CardNumber  string `json:"card_number" example:"2222405343248877"`     // Card number containing between 14 and 19 digits.
	ExpiryMonth int    `json:"expiry_month" example:"12"`                  // Expiration month (1–12).
	ExpiryYear  int    `json:"expiry_year" example:"2050"`                 // Expiration year (four digits).
	Currency    string `json:"currency" example:"USD" enums:"USD,EUR,BRL"` // Currency code in ISO 4217 format (e.g. USD, EUR, BRL).
	Amount      int64  `json:"amount" example:"1000"`                      // Amount expressed in minor units of the given currency. Example: $10.99 USD → 1099
	CVV         string `json:"cvv" example:"123"`                          // Card verification value (3 or 4 digits).
}

func (req PaymentRequest) Validate() error {
	if len(req.CardNumber) < 14 ||
		len(req.CardNumber) > 19 ||
		!isDigitsOnly(req.CardNumber) {
		return &InvalidPaymentRequestErr{
			Field:   "card_number",
			Message: "card number must contain between 14 and 19 numeric digits",
		}
	}

	if req.ExpiryMonth < 1 || req.ExpiryMonth > 12 {
		return &InvalidPaymentRequestErr{
			Field:   "expiry_month",
			Message: "expiry month must be between 1 and 12",
		}
	}

	// TODO: ensure that we test this before ship to production
	// tip: unit test to enforce different timezones and ensure that UTC is covering everything...
	now := time.Now().UTC()
	if req.ExpiryYear < now.Year() ||
		(req.ExpiryYear == now.Year() && req.ExpiryMonth < int(now.Month())) {
		return &InvalidPaymentRequestErr{
			Field:   "expiry_date",
			Message: "expiry date must be in the future",
		}
	}

	switch req.Currency {
	case "USD", "EUR", "BRL":
	default:
		return &InvalidPaymentRequestErr{
			Field:   "currency",
			Message: "currency must be one of: USD, EUR, BRL",
		}
	}

	if req.Amount <= 0 {
		return &InvalidPaymentRequestErr{
			Field:   "amount",
			Message: "amount must be greater than zero",
		}
	}

	if len(req.CVV) < 3 ||
		len(req.CVV) > 4 ||
		!isDigitsOnly(req.CVV) {
		return &InvalidPaymentRequestErr{
			Field:   "cvv",
			Message: "cvv must contain 3 or 4 numeric digits",
		}
	}

	return nil
}

func isDigitsOnly(s string) bool {
	if s == "" {
		return false
	}

	for _, r := range s {
		if !unicode.IsDigit(r) {
			return false
		}
	}
	return true
}
