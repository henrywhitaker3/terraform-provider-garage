// Package client
//
// An API client for garage whilst the go sdk is broken
package client

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

type Client struct {
	url   string
	token string
}

func New(url, token string) *Client {
	return &Client{
		url:   url,
		token: token,
	}
}

func (c *Client) do(
	ctx context.Context,
	method, path string,
	body any,
	output any,
) (*http.Response, error) {
	var reader io.Reader
	if body != nil {
		by, err := json.Marshal(body)
		if err != nil {
			return nil, fmt.Errorf("marshal request body: %w", err)
		}
		reader = bytes.NewBuffer(by)
	}
	req, err := http.NewRequestWithContext(
		ctx,
		method,
		fmt.Sprintf("%s%s", c.url, path),
		reader,
	)
	req.Header.Set("Content-Type", "application/json")
	if err != nil {
		return nil, fmt.Errorf("create request: %w", err)
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("do request: %w", err)
	}

	if resp.StatusCode > 299 {
		return resp, nil
	}

	if output != nil {
		out, err := io.ReadAll(resp.Body)
		if err != nil {
			return resp, fmt.Errorf("read response body: %w", err)
		}
		if err := json.Unmarshal(out, output); err != nil {
			return resp, fmt.Errorf("unmarhsal body: %w", err)
		}
	}

	return resp, nil
}
