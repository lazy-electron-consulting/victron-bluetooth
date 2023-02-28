package devices

import (
	"fmt"

	"github.com/lazy-electron-consulting/victron-bluetooth/internal"
)

type BatteryMonitor struct {
	device[BatteryMonitorReading]
}

func NewBatteryMonitor(d Device, handler func(BatteryMonitorReading)) *BatteryMonitor {
	return &BatteryMonitor{
		device: device[BatteryMonitorReading]{
			Device:  d,
			reader:  ReadBatteryMonitor,
			handler: handler,
		},
	}
}

type BatteryMonitorReading struct {
	RemainingMinutes uint
	Voltage          float64
	Current          float64
}

func ReadBatteryMonitor(record []byte, data *BatteryMonitorReading) error {
	type layout struct {
		RemainingMinutes uint16
		Voltage          uint16
		Alarm            [2]byte
		Aux              int16
		Current          [3]byte
	}

	l, err := internal.Decode[layout](record)
	if err != nil {
		return fmt.Errorf("unable to decode: %w", err)
	}

	buf := internal.PrependZero(l.Current[:], 4)
	v, err := internal.Decode[int32](buf)
	if err != nil {
		return fmt.Errorf("unable to decode current: %w", err)
	}

	data.RemainingMinutes = uint(l.RemainingMinutes)
	data.Voltage = float64(l.Voltage) / 100
	data.Current = float64(v>>10) / 1000

	return nil
}
