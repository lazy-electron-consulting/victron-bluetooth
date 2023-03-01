package main

import (
	"context"
	"encoding/hex"
	"errors"
	"flag"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"golang.org/x/exp/slog"
	"golang.org/x/sync/errgroup"

	"github.com/lazy-electron-consulting/victron-bluetooth/cmd/exporter/prom"
	"github.com/lazy-electron-consulting/victron-bluetooth/pkg/devices"
	"github.com/lazy-electron-consulting/victron-bluetooth/pkg/scanner"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

var addr = flag.String("addr", ":8000", "addr to run the HTTP server for metrics")

func main() {
	flag.Usage = func() {
		fmt.Printf("Usage: %s [OPTIONS] DEVICE_ID SECRET\n", os.Args[0])
		flag.PrintDefaults()
	}

	flag.Parse()
	if flag.NArg() != 2 {
		flag.Usage()
		os.Exit(1)
	}

	slog.SetDefault(slog.New(slog.HandlerOptions{Level: slog.LevelDebug}.NewTextHandler(os.Stderr)))
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGHUP, syscall.SIGABRT, syscall.SIGINT)
	defer stop()

	key, err := hex.DecodeString(flag.Arg(1))
	if err != nil {
		fmt.Printf("SECRET must be a valid hex string: %v\n", err)
		os.Exit(1)
	}

	b := devices.NewBatteryMonitor(devices.Device{
		Addr: flag.Arg(0),
		Key:  key,
	}, prom.ObserveBatteryMonitor)

	g, ctx := errgroup.WithContext(ctx)

	g.Go(func() error {
		return runHttp(ctx, *addr)
	})
	g.Go(func() error {
		return scanner.Run(ctx, b)
	})

	if err := g.Wait(); err != nil && !errors.Is(err, context.Canceled) {
		slog.Error("failed", err)
		os.Exit(1)
	}
}

func runHttp(ctx context.Context, addr string) error {
	http.Handle("/metrics", promhttp.Handler())
	srv := &http.Server{Addr: addr}
	go func() {
		<-ctx.Done()
		srv.Close()
	}()
	defer slog.Info("http server stopped")
	slog.Info("http server started", slog.String("addr", addr))

	if err := srv.ListenAndServe(); !errors.Is(err, http.ErrServerClosed) {
		return err
	}
	return nil
}
