package client

import (
	"context"
	"fmt"
	"net/http"
)

type Permission struct {
	AccessKeyID string
	BucketID    string
	Owner       bool
	Read        bool
	Write       bool
}

func (c *Client) GetPermissions(
	ctx context.Context,
	keyID string,
	bucketID string,
) (*Permission, error) {
	bucket, err := c.GetBucket(ctx, bucketID)
	if err != nil {
		return nil, fmt.Errorf("get bucket for permissions: %w", err)
	}

	for _, perm := range bucket.Keys {
		if perm.AccessKeyID == keyID {
			return &Permission{
				AccessKeyID: keyID,
				BucketID:    bucketID,
				Owner:       perm.Permissions.Owner,
				Read:        perm.Permissions.Read,
				Write:       perm.Permissions.Write,
			}, nil
		}
	}

	return nil, fmt.Errorf("could not find permission for key/bucket")
}

type CreatePermissionsBlock struct {
	Owner bool `json:"owner"`
	Read  bool `json:"read"`
	Write bool `json:"write"`
}

type CreatePermissionRequest struct {
	AccessKeyID string                 `json:"accessKeyId"`
	BucketID    string                 `json:"bucketId"`
	Permissions CreatePermissionsBlock `json:"permissions"`
}

func (c *Client) CreatePermission(
	ctx context.Context,
	req CreatePermissionRequest,
) (*Permission, error) {
	err := c.do(ctx, http.MethodPost, "/v2/AllowBucketKey", req, nil)
	if err != nil {
		return nil, fmt.Errorf("grant bucket permissions: %w", err)
	}
	return c.GetPermissions(ctx, req.AccessKeyID, req.BucketID)
}

func (c *Client) UpdatePermission(
	ctx context.Context,
	req CreatePermissionRequest,
) (*Permission, error) {
	if _, err := c.CreatePermission(ctx, req); err != nil {
		return nil, err
	}

	inverse := &CreatePermissionRequest{
		AccessKeyID: req.AccessKeyID,
		BucketID:    req.BucketID,
		Permissions: CreatePermissionsBlock{
			Owner: !req.Permissions.Owner,
			Read:  !req.Permissions.Read,
			Write: !req.Permissions.Write,
		},
	}

	err := c.do(ctx, http.MethodPost, "/v2/DenyBucketKey", inverse, nil)
	if err != nil {
		return nil, fmt.Errorf("remove bucket permissions: %w", err)
	}
	return c.GetPermissions(ctx, req.AccessKeyID, req.BucketID)
}
