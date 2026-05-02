// Package dagger wraps Dagger SDK initialization. The actual SDK calls live
// in package-specific files (e.g., repo.Clone, run.Execute) — this package
// is just connection lifecycle.
package dagger

import (
	"context"
	"fmt"
	"io"
	"os"
	"sync"

	dag "dagger.io/dagger"
)

// Client wraps a singleton dagger.Client. The Dagger SDK opens a long-lived
// session per Client; we keep one per API process.
type Client struct {
	mu     sync.Mutex
	client *dag.Client
}

func NewClient() *Client {
	return &Client{}
}

// Get lazily initializes and returns the underlying dagger client.
// Safe for concurrent use.
func (c *Client) Get(ctx context.Context) (*dag.Client, error) {
	c.mu.Lock()
	defer c.mu.Unlock()
	if c.client != nil {
		return c.client, nil
	}
	client, err := dag.Connect(ctx, dag.WithLogOutput(io.Discard))
	if err != nil {
		return nil, fmt.Errorf("dagger connect: %w", err)
	}
	c.client = client
	return c.client, nil
}

// Close releases the dagger session.
func (c *Client) Close() error {
	c.mu.Lock()
	defer c.mu.Unlock()
	if c.client == nil {
		return nil
	}
	err := c.client.Close()
	c.client = nil
	return err
}

// VerboseLog enables stderr logging for SDK debugging.
// Set WORKEND_DAGGER_VERBOSE=true to use.
func VerboseLog() bool {
	return os.Getenv("WORKEND_DAGGER_VERBOSE") == "true"
}
