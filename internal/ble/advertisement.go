package ble

import (
	"encoding/hex"
	"fmt"
	"time"
)

// Advertisement represents a single advertisement packet
type Advertisement struct {
	Timestamp        time.Time
	RSSI             int16
	RawData          []byte
	ManufacturerData []byte
	ServiceUUIDs     []string
	ServiceData      map[string][]byte
	LocalName        string
	TxPowerLevel     *int8
	Connectable      bool
}

// NewAdvertisement creates a new Advertisement with the current timestamp
func NewAdvertisement() Advertisement {
	return Advertisement{
		Timestamp:   time.Now(),
		ServiceData: make(map[string][]byte),
	}
}

// FormatRawData returns the raw data as a hex string
func (a *Advertisement) FormatRawData() string {
	if len(a.RawData) == 0 {
		return ""
	}
	return hex.EncodeToString(a.RawData)
}

// FormatManufacturerData returns the manufacturer data as a hex string
func (a *Advertisement) FormatManufacturerData() string {
	if len(a.ManufacturerData) == 0 {
		return ""
	}
	return hex.EncodeToString(a.ManufacturerData)
}

// String returns a formatted string representation of the advertisement
func (a *Advertisement) String() string {
	timeStr := a.Timestamp.Format("15:04:05.000")
	mfgData := a.FormatManufacturerData()
	if len(mfgData) > 40 {
		mfgData = mfgData[:40] + "..."
	}
	if mfgData == "" {
		mfgData = "-"
	}
	return fmt.Sprintf("%s RSSI:%d %s", timeStr, a.RSSI, mfgData)
}
