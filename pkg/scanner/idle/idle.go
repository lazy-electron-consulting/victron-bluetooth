package idle

import (
	"context"
	"errors"
	"sync"
	"time"
)

var ErrTimedOut = errors.New("timed out due to no activity")

// IdleTimer is a dead-mans switch to detect inactivity. Create with New.
type IdleTimer struct {
	timeout    time.Duration
	done       chan struct{}
	lastActive time.Time
	mu         sync.RWMutex // protects lastActive
}

// New creates an IdleTimer that will return errors if the there is no activity
// seen for the given duration.
func New(d time.Duration) *IdleTimer {
	return &IdleTimer{
		timeout: d,
		done:    make(chan struct{}),
	}
}

// SetActive marks this timer as still active.
func (t *IdleTimer) SetActive() {
	t.mu.Lock()
	t.lastActive = time.Now()
	t.mu.Unlock()
}

// Run runs until the context is cancelled or no-one calls SetActive for more
// than the timer's duration. Always returns an error, either the context's or
// ErrTimedOut.
func (t *IdleTimer) Run(ctx context.Context) error {
	ticker := time.NewTicker(t.timeout)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case now := <-ticker.C:
			tooOld := now.Add(-t.timeout)
			t.mu.RLock()
			isIdle := t.lastActive.Before(tooOld)
			t.mu.RUnlock()
			if isIdle {
				return ErrTimedOut
			}
		}
	}
}
