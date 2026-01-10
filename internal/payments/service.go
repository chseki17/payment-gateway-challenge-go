package payments

import (
	"context"
	"errors"
	"fmt"

	"github.com/cko-recruitment/payment-gateway-challenge-go/internal/banks/simulator"
)

type Service struct {
	repo PaymentsRepository
	bank simulator.BankingSimulator
}

func NewService(repo PaymentsRepository, bank simulator.BankingSimulator) *Service {
	return &Service{repo, bank}
}

func (s *Service) CreatePayment(ctx context.Context, paymentReq PaymentRequest) (*Payment, error) {
	if err := paymentReq.Validate(); err != nil {
		return nil, fmt.Errorf("payment validation: %w", err)
	}

	res, err := s.bank.Authorize(ctx, simulator.AuthorizationRequest{
		CardNumber: paymentReq.CardNumber,
		ExpiryDate: fmt.Sprintf("%02d/%d", paymentReq.ExpiryMonth, paymentReq.ExpiryYear),
		Currency:   paymentReq.Currency,
		Amount:     paymentReq.Amount,
		CVV:        paymentReq.CVV,
	})

	paymentStatus := StatusAuthorized

	if err != nil {
		switch {
		case errors.Is(err, simulator.ErrAuthorizationRejected):
			paymentStatus = StatusRejected

		case errors.Is(err, simulator.ErrAuthorizationUnavailable):
			return nil, err // retry higher up

		default:
			return nil, fmt.Errorf("authorize payment: %w", err)
		}
	} else if !res.Authorized {
		paymentStatus = StatusDeclined
	}

	payment := &Payment{
		Status:             paymentStatus,
		CardNumberLastFour: paymentReq.CardNumber[len(paymentReq.CardNumber)-4:],
		ExpiryMonth:        paymentReq.ExpiryMonth,
		ExpiryYear:         paymentReq.ExpiryYear,
		Currency:           paymentReq.Currency,
		Amount:             paymentReq.Amount,
	}

	err = s.repo.AddPayment(ctx, payment)
	if err != nil {
		return nil, fmt.Errorf("persist payment: %w", err)
	}

	return payment, nil
}

func (s *Service) GetPayment(ctx context.Context, id string) (*Payment, error) {
	p, err := s.repo.GetPayment(ctx, id)
	if err != nil {
		return nil, err
	}

	if p == nil {
		return nil, NotFoundPaymentErr
	}

	return p, nil
}
