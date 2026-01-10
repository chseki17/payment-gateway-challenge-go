package repository

import (
	"context"
	"sync"

	"github.com/cko-recruitment/payment-gateway-challenge-go/internal/payments"
	"github.com/google/uuid"
)

type PaymentID string

type PaymentsRepositoryInMemory struct {
	mu       sync.RWMutex
	payments map[PaymentID]*payments.Payment

	idempotency map[string]PaymentID
}

func NewPaymentsRepositoryInMemory() *PaymentsRepositoryInMemory {
	return &PaymentsRepositoryInMemory{
		payments:    map[PaymentID]*payments.Payment{},
		idempotency: map[string]PaymentID{},
	}
}

func (ps *PaymentsRepositoryInMemory) GetPayment(_ context.Context, id string) (*payments.Payment, error) {
	ps.mu.RLock()
	defer ps.mu.RUnlock()

	payment := ps.payments[PaymentID(id)]
	return payment, nil
}

func (ps *PaymentsRepositoryInMemory) AddPayment(_ context.Context, payment *payments.Payment) error {
	ps.mu.Lock()
	defer ps.mu.Unlock()

	existingID, requestAlreadyProcessed := ps.idempotency[payment.IdempotencyKey]
	if requestAlreadyProcessed {
		p := ps.payments[existingID]
		*payment = *p
		return nil
	}

	id, err := uuid.NewV7()
	if err != nil {
		return err
	}

	payment.ID = id.String()
	ps.payments[PaymentID(id.String())] = payment

	if payment.IdempotencyKey != "" {
		ps.idempotency[payment.IdempotencyKey] = PaymentID(payment.ID)
	}

	return nil
}
