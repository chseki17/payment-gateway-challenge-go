package api

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"

	"github.com/cko-recruitment/payment-gateway-challenge-go/internal/payments"
	"github.com/go-chi/chi/v5"
)

type PaymentsHandler struct {
	service *payments.Service
}

func NewPaymentsHandler(svc *payments.Service) *PaymentsHandler {
	return &PaymentsHandler{svc}
}

// GetPayment godoc
// @Summary Get payment by ID
// @Description Retrieves a payment by its unique identifier
// @Tags payments
// @Accept json
// @Produce json
// @Param id path string true "Payment ID"
// @Success 200 {object} payments.Payment
// @Failure 404 {object} api.ErrorResponseBody
// @Failure 500 {object} api.ErrorResponseBody
// @Router /api/v1/payments/{id} [get]
func (h *PaymentsHandler) GetHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		id := chi.URLParam(r, "id")

		payment, err := h.service.GetPayment(r.Context(), id)
		if err != nil {
			if errors.Is(err, payments.NotFoundPaymentErr) {
				ErrorResponse(w, http.StatusNotFound, err.Error())
			} else {
				ErrorResponse(w, http.StatusInternalServerError, err.Error())
			}
			return
		}

		OKResponse(w, payment)
	}
}

// CreatePayment godoc
// @Summary Create a payment
// @Description Creates a new payment and authorizes it with the bank
// @Description Amount must be expressed in minor units of the currency (e.g. 1099 = $10.99 USD).
// @Description The only currencies supported for now are USD, EUR, and BRL (ISO 4217).
// @Tags payments
// @Accept json
// @Produce json
// @Param request body payments.PaymentRequest true "Payment request"
// @Success 200 {object} payments.Payment
// @Failure 400 {object} api.ErrorResponseBody
// @Failure 500 {object} api.ErrorResponseBody
// @Router /api/v1/payments [post]
func (h *PaymentsHandler) PostHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		log := LoggingFromContext(r.Context())
		var paymentReq payments.PaymentRequest

		if err := json.NewDecoder(r.Body).Decode(&paymentReq); err != nil {
			ErrorResponse(w, http.StatusBadRequest, "invalid request body format")
			return
		}

		log.Info("Creating payment")
		payment, err := h.service.CreatePayment(r.Context(), paymentReq)
		if err != nil {
			log.Error(fmt.Sprintf("Creating payment: %s", err.Error()))
			var invalidPaymentRequestErr *payments.InvalidPaymentRequestErr
			if errors.As(err, &invalidPaymentRequestErr) {
				ErrorResponse(w, http.StatusBadRequest, invalidPaymentRequestErr.Message)
				return
			}
			ErrorResponse(w, http.StatusInternalServerError, err.Error())
			return
		}

		if payment.Status != payments.StatusAuthorized {
			log.Warn("payment status is not authorized", "payment_id", payment.ID, "payment_status", payment.Status.String())
		}

		OKResponse(w, payment)
	}
}
