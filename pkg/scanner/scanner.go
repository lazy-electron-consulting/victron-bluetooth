package scanner

import (
	"context"
	"fmt"

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
	var g errgroup.Group
	g.Go(func() error {
		return adapter.Scan(func(_ *bluetooth.Adapter, sr bluetooth.ScanResult) {
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
	g.Go(func() error {
		<-ctx.Done()
		return adapter.StopScan()
	})
	slog.Info("scanning devices")
	defer slog.Info("stopped scanning")
	return g.Wait()
}
