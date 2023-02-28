package devices_test

import (
	"encoding/hex"
	"testing"

	"github.com/lazy-electron-consulting/victron-bluetooth/pkg/devices"
	"github.com/stretchr/testify/require"
)

func TestReadBatteryMonitor(t *testing.T) {
	rec, _ := hex.DecodeString("4038190500000000abfaff010080fe84")

	var bm devices.BatteryMonitorReading

	require.NoError(t, devices.ReadBatteryMonitor(rec, &bm))

	require.Equal(t, 13.05, bm.Voltage)
	require.Equal(t, -.342, bm.Current)
}
