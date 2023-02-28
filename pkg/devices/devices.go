package devices

import (
	"fmt"

	"github.com/lazy-electron-consulting/victron-bluetooth/pkg/scanner"
	"golang.org/x/exp/slog"
)

type Device struct {
	Addr string
	Key  []byte
}

type device[T any] struct {
	Device
	reader  func(record []byte, data *T) error
	handler func(T)
}

func (d *device[T]) Handle(ad scanner.Advertisement) {
	logger := slog.Default().With(
		slog.String("Mode", fmt.Sprintf("%x", ad.Mode)),
		slog.String("Model", fmt.Sprintf("%x", ad.Model)),
		slog.String("addr", ad.Addr),
	)
	logger.Debug("handling ad")
	if ad.Addr != d.Addr {
		logger.Debug("ignoring uninteresting addr")
		return
	}
	data, err := ad.Decrypt(d.Key)
	if err != nil {
		logger.Error("unable to decrypt", err)
		return
	}
	var rec T
	if err := d.reader(data, &rec); err != nil {
		logger.Error("unable to read record", err)
		return
	}
	d.handler(rec)
}
