// Package dagger wraps Dagger SDK initialization. The actual SDK calls live
// in package-specific files (e.g., repo.Clone, run.Execute) — this package
// is just connection lifecycle.
package dagger

import (
	"context"
	"fmt"
	"io"
	"os"
	"strings"
	"sync"

	dag "dagger.io/dagger"
)

// Client wraps a singleton dagger.Client. The Dagger SDK opens a long-lived
// session per Client; we keep one per API process.
//
// IMPORTANT: dag.Connect ties the SDK session's lifetime to the context
// passed in. If we connected with a per-request ctx, the session would die
// the moment that request finished and every subsequent run would fail with
// "connection refused" against the local session port. The session ctx must
// outlive any individual HTTP request — usually the main process's context.
type Client struct {
	sessionCtx context.Context

	mu     sync.Mutex
	client *dag.Client
	stale  bool
}

// NewClient creates a Client whose Dagger session is bound to sessionCtx.
// When sessionCtx is cancelled (e.g. on SIGTERM), the dagger session
// terminates cleanly.
func NewClient(sessionCtx context.Context) *Client {
	return &Client{sessionCtx: sessionCtx}
}

// Get lazily initializes and returns the underlying dagger client.
// Safe for concurrent use. The reqCtx parameter is honored only for the
// connect call's deadline — the resulting session itself is bound to the
// long-lived sessionCtx supplied at construction time.
func (c *Client) Get(reqCtx context.Context) (*dag.Client, error) {
	c.mu.Lock()
	defer c.mu.Unlock()
	if c.client != nil && !c.stale {
		// Sanity: if the session ctx has been cancelled (process is shutting
		// down), reject before the caller hits a confusing dial error.
		if err := c.sessionCtx.Err(); err != nil {
			return nil, fmt.Errorf("dagger session terminated: %w", err)
		}
		return c.client, nil
	}
	// Stale or first call — (re)connect.
	if c.client != nil {
		_ = c.client.Close()
		c.client = nil
	}
	c.stale = false
	if c.sessionCtx == nil {
		c.sessionCtx = context.Background()
	}
	client, err := dag.Connect(c.sessionCtx, dag.WithLogOutput(io.Discard))
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

// MarkStale flags the current session as dead so the next Get() call
// transparently reconnects. Safe to call even if the session is healthy.
func (c *Client) MarkStale() {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.stale = true
}

// IsSessionError returns true if err looks like a dead / evicted Dagger
// engine session. These are not retryable on the same *dag.Client —
// callers should MarkStale + re-call Get.
func IsSessionError(err error) bool {
	if err == nil {
		return false
	}
	msg := err.Error()
	return strings.Contains(msg, "no active session") ||
		strings.Contains(msg, "connect: connection refused") ||
		(strings.Contains(msg, "DeadlineExceeded") && strings.Contains(msg, "session"))
}

// VerboseLog enables stderr logging for SDK debugging.
// Set WORKEND_DAGGER_VERBOSE=true to use.
func VerboseLog() bool {
	return os.Getenv("WORKEND_DAGGER_VERBOSE") == "true"
}
