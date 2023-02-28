package scanner

import (
	"context"
	"fmt"

	"golang.org/x/exp/slog"
	"golang.org/x/sync/errgroup"
	"tinygo.org/x/bluetooth"
)

type Handler interface {
	Handle(addr string, advert Advertisement)
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
			data, ok := sr.AdvertisementPayload.ManufacturerData()[0x02e1]
			slog.Debug("scanned",
				slog.String("addr", sr.Address.String()),
				slog.String("name", sr.LocalName()),
				slog.Bool("hasVictronAdvert", ok),
			)
			if ok {
				ad, err := readAdvertisement(data)
				if err != nil {
					slog.Error("failed to decode advert, skipping", err, slog.String("addr", sr.Address.String()))
				} else {
					h.Handle(sr.Address.String(), ad)
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
