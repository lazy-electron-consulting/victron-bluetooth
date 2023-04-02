package idle_test

import (
	"context"
	"testing"
	"time"

	"github.com/lazy-electron-consulting/victron-bluetooth/pkg/scanner/idle"
	"github.com/stretchr/testify/require"
)

func TestTimer_Idle(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	d := 10 * time.Millisecond
	timer := idle.New(d)

	go func() {
		<-time.After(2 * d)
		cancel()
	}()

	err := timer.Run(ctx)
	require.ErrorIs(t, err, idle.ErrTimedOut)
}

func TestTimer_Active(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())

	d := 100 * time.Millisecond
	timer := idle.New(d)

	go func() {
		defer cancel()
		timer.SetActive()
		<-time.After(50 * time.Millisecond)
		timer.SetActive()
		<-time.After(50 * time.Millisecond)
		timer.SetActive()
		<-time.After(50 * time.Millisecond)
	}()

	err := timer.Run(ctx)
	require.ErrorIs(t, err, context.Canceled)
}
