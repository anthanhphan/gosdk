// Copyright (c) 2026 anthanhphan <an.thanhphan.work@gmail.com>

package client

import (
	"context"
	"fmt"

	"github.com/anthanhphan/gosdk/jcodec"
)

// Decode unmarshals the response body into the target type T using jcodec.
// Use this to avoid manual jcodec.Unmarshal calls after Client.Do/Get/Post/etc.
//
// Example:
//
//	resp, err := c.Get(ctx, "/users/123")
//	user, err := client.Decode[User](resp)
func Decode[T any](resp *Response) (T, error) {
	var result T
	if resp == nil {
		return result, fmt.Errorf("decode: nil response")
	}
	if err := jcodec.Unmarshal(resp.Body, &result); err != nil {
		return result, fmt.Errorf("decode response body: %w", err)
	}
	return result, nil
}

// GetJSON performs a GET request and decodes the response body into type T.
//
// Example:
//
//	user, err := client.GetJSON[User](c, ctx, "/users/123")
func GetJSON[T any](c *Client, ctx context.Context, path string, opts ...RequestOption) (T, error) {
	resp, err := c.Get(ctx, path, opts...)
	if err != nil {
		var zero T
		return zero, err
	}
	return Decode[T](resp)
}

// PostJSON performs a POST request and decodes the response body into type T.
func PostJSON[T any](c *Client, ctx context.Context, path string, body any, opts ...RequestOption) (T, error) {
	resp, err := c.Post(ctx, path, body, opts...)
	if err != nil {
		var zero T
		return zero, err
	}
	return Decode[T](resp)
}

// PutJSON performs a PUT request and decodes the response body into type T.
func PutJSON[T any](c *Client, ctx context.Context, path string, body any, opts ...RequestOption) (T, error) {
	resp, err := c.Put(ctx, path, body, opts...)
	if err != nil {
		var zero T
		return zero, err
	}
	return Decode[T](resp)
}

// PatchJSON performs a PATCH request and decodes the response body into type T.
func PatchJSON[T any](c *Client, ctx context.Context, path string, body any, opts ...RequestOption) (T, error) {
	resp, err := c.Patch(ctx, path, body, opts...)
	if err != nil {
		var zero T
		return zero, err
	}
	return Decode[T](resp)
}

// DeleteJSON performs a DELETE request and decodes the response body into type T.
func DeleteJSON[T any](c *Client, ctx context.Context, path string, opts ...RequestOption) (T, error) {
	resp, err := c.Delete(ctx, path, opts...)
	if err != nil {
		var zero T
		return zero, err
	}
	return Decode[T](resp)
}
