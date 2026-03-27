// Copyright (c) 2026 anthanhphan <an.thanhphan.work@gmail.com>

package core

import "google.golang.org/grpc/metadata"

// GRPCMetadataCarrier adapts gRPC metadata.MD to a text map carrier interface
// for trace context propagation. Used by both client and server interceptors.
type GRPCMetadataCarrier struct {
	MD metadata.MD
}

// Get returns the first value for the given key.
func (c *GRPCMetadataCarrier) Get(key string) string {
	vals := c.MD.Get(key)
	if len(vals) == 0 {
		return ""
	}
	return vals[0]
}

// Set sets the value for the given key.
func (c *GRPCMetadataCarrier) Set(key, value string) {
	c.MD.Set(key, value)
}

// Keys returns all keys in the metadata.
func (c *GRPCMetadataCarrier) Keys() []string {
	keys := make([]string, 0, len(c.MD))
	for k := range c.MD {
		keys = append(keys, k)
	}
	return keys
}
