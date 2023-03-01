package prom

import (
	"github.com/lazy-electron-consulting/victron-bluetooth/pkg/devices"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

var (
	volts = promauto.NewGauge(prometheus.GaugeOpts{
		Subsystem: "smart_shunt",
		Name:      "battery_volts",
		Help:      "Main battery voltage",
	})
	amps = promauto.NewGauge(prometheus.GaugeOpts{
		Subsystem: "smart_shunt",
		Name:      "battery_amps",
		Help:      "Main battery current",
	})
)

// ObserveBatteryMonitor increments metrics for the reading.
func ObserveBatteryMonitor(bmr devices.BatteryMonitorReading) {
	volts.Set(bmr.Voltage)
	amps.Set(bmr.Current)
}
