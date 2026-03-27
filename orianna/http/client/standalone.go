// Copyright (c) 2026 anthanhphan <an.thanhphan.work@gmail.com>

package client

import (
	"context"
	"net/http"
	"sync"
)

// ============================================================================
// Standalone HTTP Request Functions (for simple use cases)
// ============================================================================

// defaultClients caches clients by baseURL to enable connection reuse
// across standalone function calls. Limited to maxCachedClients entries
// to prevent unbounded memory growth from dynamic URLs.
var (
	defaultClients   = make(map[string]*Client)
	defaultClientsMu sync.RWMutex
)

// maxCachedClients is the maximum number of cached clients.
// When exceeded, the oldest entry is evicted (simple eviction).
const maxCachedClients = 64

// getOrCreateClient returns a cached client for the given baseURL,
// creating one with default settings if none exists yet.
func getOrCreateClient(baseURL string) (*Client, error) {
	defaultClientsMu.RLock()
	if c, ok := defaultClients[baseURL]; ok {
		defaultClientsMu.RUnlock()
		return c, nil
	}
	defaultClientsMu.RUnlock()

	defaultClientsMu.Lock()
	defer defaultClientsMu.Unlock()
	// Double-check after acquiring write lock
	if c, ok := defaultClients[baseURL]; ok {
		return c, nil
	}

	// Evict one random entry if at capacity (map iteration order is random in Go)
	if len(defaultClients) >= maxCachedClients {
		for key, old := range defaultClients {
			if transport, ok := old.httpClient.Transport.(*http.Transport); ok {
				transport.CloseIdleConnections()
			}
			delete(defaultClients, key)
			break
		}
	}

	c, err := NewClient(WithBaseURL(baseURL))
	if err != nil {
		return nil, err
	}
	defaultClients[baseURL] = c
	return c, nil
}

// Get performs a simple GET request.
func Get(ctx context.Context, baseURL, path string, opts ...RequestOption) (*Response, error) {
	client, err := getOrCreateClient(baseURL)
	if err != nil {
		return nil, err
	}
	return client.Get(ctx, path, opts...)
}

// Post performs a simple POST request.
func Post(ctx context.Context, baseURL, path string, body any, opts ...RequestOption) (*Response, error) {
	client, err := getOrCreateClient(baseURL)
	if err != nil {
		return nil, err
	}
	return client.Post(ctx, path, body, opts...)
}

// Put performs a simple PUT request.
func Put(ctx context.Context, baseURL, path string, body any, opts ...RequestOption) (*Response, error) {
	client, err := getOrCreateClient(baseURL)
	if err != nil {
		return nil, err
	}
	return client.Put(ctx, path, body, opts...)
}

// Patch performs a simple PATCH request.
func Patch(ctx context.Context, baseURL, path string, body any, opts ...RequestOption) (*Response, error) {
	client, err := getOrCreateClient(baseURL)
	if err != nil {
		return nil, err
	}
	return client.Patch(ctx, path, body, opts...)
}

// Delete performs a simple DELETE request.
func Delete(ctx context.Context, baseURL, path string, opts ...RequestOption) (*Response, error) {
	client, err := getOrCreateClient(baseURL)
	if err != nil {
		return nil, err
	}
	return client.Delete(ctx, path, opts...)
}
