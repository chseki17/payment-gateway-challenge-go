package repository_test

import (
	"context"
	"testing"

	"github.com/cko-recruitment/payment-gateway-challenge-go/internal/payments"
	"github.com/cko-recruitment/payment-gateway-challenge-go/internal/repository"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestPaymentsRepositoryInMemory_AddAndGet(t *testing.T) {
	repo := repository.NewPaymentsRepositoryInMemory()

	payment := &payments.Payment{
		Status: payments.StatusPending,
		Amount: 1000,
	}

	err := repo.AddPayment(context.Background(), payment)
	require.NoError(t, err, "AddPayment should not return an error")

	require.NotEmpty(t, payment.ID, "payment ID should be set")

	got, err := repo.GetPayment(context.Background(), payment.ID)
	require.NoError(t, err, "GetPayment should not return an error")
	require.NotNil(t, got, "expected payment, got nil")

	assert.Equal(t, payment.ID, got.ID)
	assert.Equal(t, payment.Status, got.Status)
	assert.Equal(t, payment.Amount, got.Amount)
}

func TestPaymentsRepositoryInMemory_ConcurrentAdd(t *testing.T) {
	repo := repository.NewPaymentsRepositoryInMemory()

	const workers = 50

	done := make(chan struct{}, workers)

	for range workers {
		go func() {
			payment := &payments.Payment{
				Status: payments.StatusPending,
				Amount: 1000,
			}

			err := repo.AddPayment(context.Background(), payment)
			assert.NoError(t, err)

			done <- struct{}{}
		}()
	}

	for range workers {
		<-done
	}
}

func TestPaymentsRepositoryInMemory_IdempotencyKey(t *testing.T) {
	repo := repository.NewPaymentsRepositoryInMemory()
	ctx := context.Background()

	idempotencyKey := "test-idempotency-key-123"

	// First request
	payment1 := &payments.Payment{
		IdempotencyKey: idempotencyKey,
		Status:         payments.StatusPending,
		Amount:         1000,
	}

	err := repo.AddPayment(ctx, payment1)
	require.NoError(t, err)
	require.NotEmpty(t, payment1.ID)

	firstID := payment1.ID

	// Second request with the same idempotency key
	payment2 := &payments.Payment{
		IdempotencyKey: idempotencyKey,
		Status:         payments.StatusPending,
		Amount:         1000,
	}

	err = repo.AddPayment(ctx, payment2)
	require.NoError(t, err)
	require.NotEmpty(t, payment2.ID)

	// Assert idempotency behavior
	assert.Equal(t, firstID, payment2.ID, "expected same payment ID for idempotent request")

	// Ensure the stored payment is the same
	stored, err := repo.GetPayment(ctx, firstID)
	require.NoError(t, err)
	require.NotNil(t, stored)

	assert.Equal(t, firstID, stored.ID)
	assert.Equal(t, payment1.Amount, stored.Amount)
	assert.Equal(t, payment1.Status, stored.Status)
}
