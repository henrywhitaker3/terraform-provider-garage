package client

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
)

type AccessKey struct {
	Name            string  `json:"name"`
	AccessKeyID     string  `json:"accessKeyId"`
	SecretAccessKey *string `json:"secretAccessKey"`
	Expiration      *string `json:"expiration"`
}

func (c *Client) GetAccessKey(ctx context.Context, id string) (*AccessKey, error) {
	key := &AccessKey{}
	_, err := c.do(
		ctx,
		http.MethodGet,
		fmt.Sprintf("/v2/GetKeyInfo?id=%s", url.QueryEscape(id)),
		nil,
		key,
	)
	if err != nil {
		return nil, fmt.Errorf("get key: %w", err)
	}
	return key, nil
}

type CreateKeyRequest struct {
	Name         string `json:"name"`
	Expiration   string `json:"expiration,omitempty"`
	NeverExpires bool   `json:"neverExpires,omitempty"`
}

func (c *Client) CreateAccessKey(ctx context.Context, req CreateKeyRequest) (*AccessKey, error) {
	key := &AccessKey{}
	_, err := c.do(ctx, http.MethodPost, "/v2/CreateKey", req, key)
	if err != nil {
		return nil, fmt.Errorf("create key: %w", err)
	}
	return key, nil
}

func (c *Client) UpdateAccessKey(
	ctx context.Context,
	id string,
	req CreateKeyRequest,
) (*AccessKey, error) {
	key := &AccessKey{}
	_, err := c.do(
		ctx,
		http.MethodPost,
		fmt.Sprintf("/v2/UpdateKey?id=%s", url.QueryEscape(id)),
		req,
		key,
	)
	if err != nil {
		return nil, fmt.Errorf("update key: %w", err)
	}
	return key, nil
}

func (c *Client) DeleteAccessKey(ctx context.Context, id string) error {
	_, err := c.do(
		ctx,
		http.MethodPost,
		fmt.Sprintf("/v2/DeleteKey?id=%s", url.QueryEscape(id)),
		nil,
		nil,
	)
	if err != nil {
		return fmt.Errorf("delete key: %w", err)
	}
	return nil
}
