package main

import (
	"context"
	"encoding/hex"
	"flag"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"golang.org/x/exp/slog"

	"github.com/lazy-electron-consulting/victron-bluetooth/pkg/devices"
	"github.com/lazy-electron-consulting/victron-bluetooth/pkg/scanner"
)

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
	}, func(bmr devices.BatteryMonitorReading) {
		slog.Info("read battery",
			slog.Float64("Current", bmr.Current),
			slog.Float64("Voltage", bmr.Voltage),
			slog.Int("RemainingMinutes", int(bmr.RemainingMinutes)))
	})

	if err := scanner.Run(ctx, b); err != nil {
		slog.Error("failed", err)
		os.Exit(1)
	}
}
