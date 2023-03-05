package devices

import (
	"fmt"

	"github.com/lazy-electron-consulting/victron-bluetooth/internal"
)

type DCDCCharger struct {
	device[DCDCChargerReading]
}

func NewDCDCCharger(d Device, handler func(DCDCChargerReading)) *DCDCCharger {
	return &DCDCCharger{
		device: device[DCDCChargerReading]{
			Device:  d,
			reader:  ReadDCDCCharger,
			handler: handler,
		},
	}
}

type DCDCChargerReading struct {
	InputVoltage  float64
	OutputVoltage float64
	Mode, Error   string
	OffReasons    []string
}

// from ve direct protocol
func DCDCOperationMode() map[uint8]string {
	return map[uint8]string{
		0:   "Off",
		1:   "Low power",
		2:   "Fault",
		3:   "Bulk",
		4:   "Absorption",
		5:   "Float",
		6:   "Storage",
		7:   "Equalize (manual)",
		9:   "Inverting",
		11:  "Power supply",
		245: "Starting-up",
		246: "Repeated absorption",
		247: "Auto equalize / Recondition",
		248: "BatterySafe",
		252: "External Control",
	}
}

func DCDCOffReasons() map[uint32]string {
	return map[uint32]string{
		0x00000001: "No input power",
		0x00000002: "Switched off (power switch)",
		0x00000004: "Switched off (device mode register)",
		0x00000008: "Remote input",
		0x00000010: "Protection active",
		0x00000020: "Paygo",
		0x00000040: "BMS",
		0x00000080: "Engine shutdown detection",
		0x00000100: "Analysing input voltage",
	}
}

func DCDCErrors() map[uint8]string {
	return map[uint8]string{
		0:   "No error",
		2:   "Battery voltage too high",
		17:  "Charger temperature too high",
		18:  "Charger over current",
		19:  "Charger current reversed",
		20:  "Bulk time limit exceeded",
		21:  "Current sensor issue (sensor bias/sensor broken)",
		26:  "Terminals overheated",
		28:  "Converter issue (dual converter models only)",
		33:  "Input voltage too high (solar panel)",
		34:  "Input current too high (solar panel)",
		38:  "Input shutdown (due to excessive battery voltage)",
		39:  "Input shutdown (due to current flow during off mode)",
		65:  "Lost communication with one of devices",
		66:  "Synchronised charging device configuration issue",
		67:  "BMS connection lost",
		68:  "Network misconfigured",
		116: "Factory calibration data lost",
		117: "Invalid/incompatible firmware",
		119: "User settings invalid",
	}
}

func ReadDCDCCharger(record []byte, data *DCDCChargerReading) error {
	type layout struct {
		State       uint8
		Error       uint8
		InputVolts  uint16
		OutputVolts int16
		OffReason   uint32
	}

	l, err := internal.Decode[layout](record)
	if err != nil {
		return fmt.Errorf("unable to decode: %w", err)
	}

	data.InputVoltage = float64(l.InputVolts) / 100
	if l.OutputVolts != 0x7FFF {
		data.OutputVoltage = float64(l.OutputVolts) / 100
	}

	if m, ok := DCDCOperationMode()[l.State]; ok {
		data.Mode = m
	} else {
		data.Mode = "unknown"
	}

	for k, v := range DCDCOffReasons() {
		if k&l.OffReason == k {
			data.OffReasons = append(data.OffReasons, v)
		}
	}

	if m, ok := DCDCErrors()[l.Error]; ok {
		data.Error = m
	} else {
		data.Error = "unknown"
	}

	return nil
}
