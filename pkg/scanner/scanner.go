package scanner

import (
	"context"
	"fmt"
	"time"

	"github.com/lazy-electron-consulting/victron-bluetooth/pkg/scanner/idle"
	"golang.org/x/exp/slog"
	"golang.org/x/sync/errgroup"
	"tinygo.org/x/bluetooth"
)

// Handler handlers BLE advertisements.
type Handler interface {
	Handle(advert Advertisement)
}

type HandlerFunc func(advert Advertisement)

func (f HandlerFunc) Handle(advert Advertisement) { f(advert) }

// FanoutHandler sends adverts to every provided handler.
func FanoutHandler(handlers ...Handler) Handler {
	return HandlerFunc(func(advert Advertisement) {
		for _, hh := range handlers {
			hh.Handle(advert)
		}
	})
}

// Run scans bluetooth devices for the Victron advertisement data, and calls the
// callback each time it sees one. Stops when the context is canceled.
func Run(ctx context.Context, h Handler) error {
	adapter := bluetooth.DefaultAdapter
	if err := adapter.Enable(); err != nil {
		return fmt.Errorf("adapter not enabled: %w", err)
	}

	timer := idle.New(5 * time.Minute)

	g, ctx := errgroup.WithContext(ctx)

	g.Go(func() error { return timer.Run(ctx) })

	g.Go(func() error {
		<-ctx.Done()
		return ctx.Err()
	})

	g.Go(func() error {
		return adapter.Scan(func(_ *bluetooth.Adapter, sr bluetooth.ScanResult) {
			timer.SetActive()
			addr := sr.Address.String()
			logger := slog.Default().With(slog.String("addr", addr), slog.String("name", sr.LocalName()))
			data, ok := sr.AdvertisementPayload.ManufacturerData()[0x02e1]
			logger.Debug("scanned", slog.Bool("hasVictronAdvert", ok))
			if ok {
				ad, err := readAdvertisement(addr, data)
				if err != nil {
					logger.Error("failed to decode advert, skipping", err)
				} else {
					h.Handle(ad)
				}
			}
		})
	})

	slog.Info("scanning devices")
	defer slog.Info("stopped scanning")
	defer func() {
		if err := adapter.StopScan(); err != nil {
			slog.Error("failed to stop scan while shutting down", err)
		}
	}()
	return g.Wait()
}
