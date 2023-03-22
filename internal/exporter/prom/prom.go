package prom

import (
	"context"
	"errors"
	"fmt"
	"net/http"

	"github.com/lazy-electron-consulting/victron-bluetooth/pkg/devices"
	"github.com/lazy-electron-consulting/victron-bluetooth/pkg/scanner"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"golang.org/x/exp/slices"
	"golang.org/x/exp/slog"
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

	dcInputVolts = promauto.NewGauge(prometheus.GaugeOpts{
		Subsystem: "dcdc_charger",
		Name:      "input_volts",
	})
	dcOutputVolts = promauto.NewGauge(prometheus.GaugeOpts{
		Subsystem: "dcdc_charger",
		Name:      "output_volts",
	})
	dcMode = promauto.NewGaugeVec(prometheus.GaugeOpts{
		Subsystem: "dcdc_charger",
		Name:      "mode",
	}, []string{"state"})
	dcOffReason = promauto.NewGaugeVec(prometheus.GaugeOpts{
		Subsystem: "dcdc_charger",
		Name:      "off_reason",
	}, []string{"state"})
	dcError = promauto.NewGaugeVec(prometheus.GaugeOpts{
		Subsystem: "dcdc_charger",
		Name:      "error",
	}, []string{"state"})

	btAdverts = promauto.NewCounterVec(prometheus.CounterOpts{
		Subsystem: "bluetooth",
		Name:      "advertisements_total",
	}, []string{"addr", "mode", "model"})
)

// ObserveAdvertisement updates metrics for newly seen advertisements.
func ObserveAdvertisement(ad scanner.Advertisement) {
	btAdverts.WithLabelValues(
		ad.Addr,
		fmt.Sprintf("%x", ad.Mode),
		fmt.Sprintf("%x", ad.Model),
	).Inc()
}

// ObserveBatteryMonitor updates metrics for the reading.
func ObserveBatteryMonitor(r devices.BatteryMonitorReading) {
	volts.Set(r.Voltage)
	amps.Set(r.Current)
}

// ObserveDCDCCharger updates metrics for the reading.
func ObserveDCDCCharger(r devices.DCDCChargerReading) {
	dcInputVolts.Set(r.InputVoltage)
	dcOutputVolts.Set(r.OutputVoltage)

	for _, mode := range devices.DCDCOperationMode() {
		var a float64
		if mode == r.Mode {
			a = 1
		}
		dcMode.WithLabelValues(mode).Set(a)
	}

	for _, mode := range devices.DCDCOffReasons() {
		var a float64
		if slices.Contains(r.OffReasons, mode) {
			a = 1
		}
		dcOffReason.WithLabelValues(mode).Set(a)
	}

	for _, mode := range devices.DCDCErrors() {
		var a float64
		if mode == r.Error {
			a = 1
		}
		dcError.WithLabelValues(mode).Set(a)
	}
}

// Run runs the metrics server on the given addr. Exits when the context is canceled.
func Run(ctx context.Context, addr string) error {
	http.Handle("/metrics", promhttp.Handler())
	srv := &http.Server{Addr: addr}
	go func() {
		<-ctx.Done()
		srv.Close()
	}()
	defer slog.Info("http server stopped")
	slog.Info("http server started", slog.String("addr", addr))

	if err := srv.ListenAndServe(); !errors.Is(err, http.ErrServerClosed) {
		return err
	}
	return nil
}
