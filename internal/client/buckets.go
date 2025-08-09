package client

import (
	"context"
	"fmt"
	"net/http"
)

type Bucket struct {
	ID            string   `json:"id"`
	GlobalAliases []string `json:"globalAliases"`
}

func (c *Client) GetBucket(ctx context.Context, id string) (*Bucket, error) {
	bucket := &Bucket{}
	resp, err := c.do(ctx, http.MethodGet, "/v2/GetBucketInfo", nil, bucket)
	if err != nil {
		return nil, fmt.Errorf("get bucket: %w", err)
	}
	if resp.StatusCode > 299 {
		return nil, fmt.Errorf("got status code %d", resp.StatusCode)
	}
	return bucket, nil
}

func (c *Client) CreateBucket(ctx context.Context, name string) (*Bucket, error) {
	bucket := &Bucket{}
	resp, err := c.do(ctx, http.MethodPost, "/v2/CreateBucket", map[string]any{
		"globalAlias": name,
	}, bucket)
	if err != nil {
		return nil, fmt.Errorf("create bucket: %w", err)
	}
	if resp.StatusCode > 299 {
		return nil, fmt.Errorf("got status code %d", resp.StatusCode)
	}
	return bucket, nil
}

func (c *Client) DeleteBucket(ctx context.Context, id string) error {
	resp, err := c.do(ctx, http.MethodPost, fmt.Sprintf("/v2/DeleteBucket/%s", id), nil, nil)
	if err != nil {
		return fmt.Errorf("delete bucket: %w", err)
	}
	if resp.StatusCode > 299 {
		return fmt.Errorf("got status code %d", resp.StatusCode)
	}
	return nil
}
