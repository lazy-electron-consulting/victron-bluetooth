package main

import (
	"context"
	"fmt"
	"os"

	"github.com/lazy-electron-consulting/victron-bluetooth/internal/exporter"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"golang.org/x/exp/slog"
)

func cmd() *cobra.Command {
	var rootCmd = &cobra.Command{
		Use:   "exporter CONFIG_FILE",
		Short: "exporter is a Prometheus exporter for victron bluetooth-enabled devices",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return run(cmd.Context(), args[0])
		},
	}

	rootCmd.PersistentFlags().Bool("verbose", false, "print verbose logs")
	rootCmd.PersistentFlags().String("listen", ":8080", "address to listen on for the metrics server")

	if err := viper.BindPFlags(rootCmd.PersistentFlags()); err != nil {
		panic(fmt.Errorf("cannot bind flags to viper: %w", err))
	}

	return rootCmd
}

func run(ctx context.Context, path string) error {
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()
	cfg, err := readConfig(path)
	if err != nil {
		return fmt.Errorf("could not read config: %w", err)
	}
	opts := slog.HandlerOptions{}
	if cfg.Verbose {
		opts.Level = slog.LevelDebug
	}
	slog.SetDefault(slog.New(opts.NewTextHandler(os.Stderr)))
	exp := exporter.NewExporter(cfg.Config)
	return exp.Run(ctx)
}
