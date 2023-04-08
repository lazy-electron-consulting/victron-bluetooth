package exporter

import (
	"context"
	"fmt"
	"sync"

	_ "net/http/pprof"

	"github.com/lazy-electron-consulting/victron-bluetooth/internal/exporter/prom"
	"github.com/lazy-electron-consulting/victron-bluetooth/pkg/devices"
	"github.com/lazy-electron-consulting/victron-bluetooth/pkg/scanner"
	"golang.org/x/exp/slog"
	"golang.org/x/sync/errgroup"
)

type Device struct {
	devices.Device `mapstructure:",squash"`
	Type           string
}

type Config struct {
	Listen  string
	Devices []Device
}

type Exporter struct {
	config   Config
	devices  map[string]Device
	mu       sync.Mutex // protects registry
	registry map[string]scanner.Handler
}

func NewExporter(cfg Config) *Exporter {
	byAddr := make(map[string]Device, len(cfg.Devices))
	for _, d := range cfg.Devices {
		byAddr[d.Addr] = d
	}

	return &Exporter{
		config:   cfg,
		devices:  byAddr,
		registry: make(map[string]scanner.Handler),
	}
}

// Run runs the exporter until the context is cancelled
func (e *Exporter) Run(ctx context.Context) error {
	g, ctx := errgroup.WithContext(ctx)
	g.Go(func() error {
		return prom.Run(ctx, e.config.Listen)
	})
	g.Go(func() error {
		return scanner.Run(ctx, e)
	})
	slog.Info("running exporter", slog.Int("count", len(e.devices)))
	defer slog.Info("exporter stopped")

	return g.Wait()
}

func (e *Exporter) Handle(ad scanner.Advertisement) {
	logger := slog.Default().With(
		slog.String("Mode", fmt.Sprintf("%x", ad.Mode)),
		slog.String("Model", fmt.Sprintf("%x", ad.Model)),
		slog.String("addr", ad.Addr),
	)
	prom.ObserveAdvertisement(ad)
	e.mu.Lock()
	defer e.mu.Unlock()

	if h, ok := e.registry[ad.Addr]; ok {
		if h != nil {
			h.Handle(ad)
		}
	} else if d, ok := e.devices[ad.Addr]; ok {
		logger.Info("registering device", slog.String("type", d.Type))
		h, ok := detect(ad, d)
		if ok {
			e.registry[ad.Addr] = h
			h.Handle(ad)
		} else {
			logger.Warn("cannot detect device, ignoring")
			e.registry[ad.Addr] = nil
		}
	} else {
		logger.Warn("ignoring unconfigured device")
		e.registry[ad.Addr] = nil
	}
}

func detect(ad scanner.Advertisement, d Device) (scanner.Handler, bool) {
	// TODO: validate the types
	switch {
	case d.Type == "battery monitor", ad.Mode == 0x02:
		return devices.NewBatteryMonitor(d.Device, prom.ObserveBatteryMonitor), true
	case d.Type == "dcdc charger", ad.Mode == 0x04:
		return devices.NewDCDCCharger(d.Device, prom.ObserveDCDCCharger), true
	default:
		return nil, false
	}
}
