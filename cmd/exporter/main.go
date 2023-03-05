package main

import (
	"context"
	"os"
	"os/signal"
	"syscall"

	"golang.org/x/exp/slog"
)

func main() {
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGHUP, syscall.SIGABRT, syscall.SIGINT)
	defer stop()

	if err := cmd().ExecuteContext(ctx); err != nil {
		slog.Error("failed", err)
		os.Exit(1)
	}
}
