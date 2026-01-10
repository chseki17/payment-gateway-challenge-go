package simulator

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

type BankingSimulator interface {
	Authorize(ctx context.Context, req AuthorizationRequest) (*AuthorizationResponse, error)
}

type AuthorizationRequest struct {
	CardNumber string `json:"card_number"`
	ExpiryDate string `json:"expiry_date"` // "MM/YYYY" format
	Currency   string `json:"currency"`
	Amount     int64  `json:"amount"` // amount in minor units (e.g. cents)
	CVV        string `json:"cvv"`
}

type AuthorizationResponse struct {
	Authorized        bool   `json:"authorized"`
	AuthorizationCode string `json:"authorization_code"`
}

type Client struct {
	baseURL    string
	httpClient *http.Client
}

type errorResponse struct {
	ErrorMessage string `json:"error_message"`
}

func NewClient(baseURL string, httpClient *http.Client) *Client {
	if httpClient == nil {
		httpClient = &http.Client{
			Timeout: 10 * time.Second,
		}
	}

	return &Client{
		baseURL:    baseURL,
		httpClient: httpClient,
	}
}

func (c *Client) Authorize(ctx context.Context, req AuthorizationRequest) (*AuthorizationResponse, error) {
	resp := &AuthorizationResponse{}
	url := fmt.Sprintf("%s/payments", c.baseURL)

	body, err := json.Marshal(req)
	if err != nil {
		return nil, fmt.Errorf(
			"%w: marshal authorization request: %v",
			ErrAuthorizationInternal,
			err,
		)
	}

	httpReq, err := http.NewRequestWithContext(
		ctx,
		http.MethodPost,
		url,
		bytes.NewReader(body),
	)
	if err != nil {
		return nil, fmt.Errorf(
			"%w: create http request: %v",
			ErrAuthorizationInternal,
			err,
		)
	}

	httpReq.Header.Set("Content-Type", "application/json")

	httpResp, err := c.httpClient.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf(
			"%w: perform authorization request: %v",
			ErrAuthorizationInternal,
			err,
		)
	}
	defer httpResp.Body.Close()

	switch {

	case httpResp.StatusCode >= 200 && httpResp.StatusCode <= 299:
		if err := json.NewDecoder(httpResp.Body).Decode(resp); err != nil {
			return nil, fmt.Errorf(
				"%w: decode authorization response: %v",
				ErrAuthorizationInternal,
				err,
			)
		}
		return resp, nil

	case httpResp.StatusCode == http.StatusBadRequest:
		// Business rejection (invalid data, card rejected, etc.)
		var errBody errorResponse
		if err := json.NewDecoder(httpResp.Body).Decode(&errBody); err != nil {
			return nil, fmt.Errorf(
				"%w: parse authorization rejection response: %v",
				ErrAuthorizationInternal,
				err,
			)
		}

		return nil, fmt.Errorf(
			"%w: %s",
			ErrAuthorizationRejected,
			errBody.ErrorMessage,
		)

	case httpResp.StatusCode == http.StatusServiceUnavailable:
		return nil, ErrAuthorizationUnavailable

	default:
		return nil, fmt.Errorf(
			"%w: status code %d",
			ErrAuthorizationUnexpected,
			httpResp.StatusCode,
		)
	}
}
