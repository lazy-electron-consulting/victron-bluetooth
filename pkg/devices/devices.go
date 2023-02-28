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

func (d *device[T]) Handle(addr string, ad scanner.Advertisement) {
	slog.Debug("handling ad",
		slog.String("Mode", fmt.Sprintf("%x", ad.Mode)),
		slog.String("Model", fmt.Sprintf("%x", ad.Model)),
	)
	if addr != d.Addr {
		slog.Debug("ignoring uninteresting addr", slog.String("addr", addr))
		return
	}
	data, err := ad.Decrypt(d.Key)
	if err != nil {
		slog.Error("unable to decrypt", err, slog.String("addr", d.Addr))
		return
	}
	var rec T
	if err := d.reader(data, &rec); err != nil {
		slog.Error("unable to read record", err)
		return
	}
	d.handler(rec)
}
