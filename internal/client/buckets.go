package client

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
)

type Bucket struct {
	ID            string   `json:"id"`
	GlobalAliases []string `json:"globalAliases"`
}

func (c *Client) GetBucket(ctx context.Context, id string) (*Bucket, error) {
	bucket := &Bucket{}
	_, err := c.do(
		ctx,
		http.MethodGet,
		fmt.Sprintf("/v2/GetBucketInfo?id=%s", url.QueryEscape(id)),
		nil,
		bucket,
	)
	if err != nil {
		return nil, fmt.Errorf("get bucket: %w", err)
	}
	return bucket, nil
}

func (c *Client) CreateBucket(ctx context.Context, name string) (*Bucket, error) {
	bucket := &Bucket{}
	_, err := c.do(ctx, http.MethodPost, "/v2/CreateBucket", map[string]any{
		"globalAlias": name,
	}, bucket)
	if err != nil {
		return nil, fmt.Errorf("create bucket: %w", err)
	}
	return bucket, nil
}

func (c *Client) DeleteBucket(ctx context.Context, id string) error {
	_, err := c.do(ctx, http.MethodPost, fmt.Sprintf("/v2/DeleteBucket?id=%s", id), nil, nil)
	if err != nil {
		return fmt.Errorf("delete bucket: %w", err)
	}
	return nil
}
