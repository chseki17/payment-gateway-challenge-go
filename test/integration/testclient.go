package e2e

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"net/http"
	"time"
)

type TestClient struct {
	BaseURL string
	Client  *http.Client
}

func NewTestClient(baseURL string) *TestClient {
	return &TestClient{
		BaseURL: baseURL,
		Client: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

func (c *TestClient) Post(
	ctx context.Context,
	path string,
	body any,
) (*http.Response, []byte, error) {

	b, err := json.Marshal(body)
	if err != nil {
		return nil, nil, err
	}

	req, err := http.NewRequestWithContext(
		ctx,
		http.MethodPost,
		c.BaseURL+path,
		bytes.NewReader(b),
	)
	if err != nil {
		return nil, nil, err
	}

	req.Header.Set("Content-Type", "application/json")

	resp, err := c.Client.Do(req)
	if err != nil {
		return nil, nil, err
	}
	defer resp.Body.Close()

	data, err := io.ReadAll(resp.Body)
	return resp, data, err
}

func (c *TestClient) Get(
	ctx context.Context,
	path string,
	out any,
) (*http.Response, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, c.BaseURL+path, nil)
	if err != nil {
		return nil, err
	}

	resp, err := c.Client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if err := json.NewDecoder(resp.Body).Decode(out); err != nil {
		return nil, err
	}

	return resp, nil
}
