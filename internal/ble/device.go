package ble

import (
	"fmt"
	"strings"
	"sync"
	"time"
)

// Device represents a discovered BLE device
type Device struct {
	Address          string
	Name             string
	RSSIHistory      []int16
	RSSICurrent      int16
	RSSIAverage      float64
	Advertisements   []Advertisement
	FirstSeen        time.Time
	LastSeen         time.Time
	AdvInterval      time.Duration
	AdvCount         int
	ManufacturerID   *uint16
	ManufacturerData []byte
	ServiceUUIDs     []string
	ServiceData      map[string][]byte
	TxPowerLevel     *int8
	Connectable      bool

	mu sync.RWMutex
}

const (
	maxRSSIHistory    = 20
	maxAdvertisements = 100
)

// NewDevice creates a new Device with the given address
func NewDevice(address string) *Device {
	now := time.Now()
	return &Device{
		Address:      address,
		RSSIHistory:  make([]int16, 0, maxRSSIHistory),
		FirstSeen:    now,
		LastSeen:     now,
		ServiceData:  make(map[string][]byte),
		ServiceUUIDs: make([]string, 0),
	}
}

// Update updates the device with new advertisement data
func (d *Device) Update(adv Advertisement) {
	d.mu.Lock()
	defer d.mu.Unlock()

	d.LastSeen = adv.Timestamp
	d.AdvCount++

	// Update name if provided
	if adv.LocalName != "" {
		d.Name = adv.LocalName
	}

	// Update RSSI
	d.RSSICurrent = adv.RSSI
	d.RSSIHistory = append(d.RSSIHistory, adv.RSSI)
	if len(d.RSSIHistory) > maxRSSIHistory {
		d.RSSIHistory = d.RSSIHistory[1:]
	}
	d.RSSIAverage = d.calculateRSSIAverage()

	// Update manufacturer data
	if len(adv.ManufacturerData) >= 2 {
		companyID := uint16(adv.ManufacturerData[0]) | uint16(adv.ManufacturerData[1])<<8
		d.ManufacturerID = &companyID
		d.ManufacturerData = adv.ManufacturerData
	}

	// Update service UUIDs
	if len(adv.ServiceUUIDs) > 0 {
		d.ServiceUUIDs = adv.ServiceUUIDs
	}

	// Update service data
	for k, v := range adv.ServiceData {
		d.ServiceData[k] = v
	}

	// Update TX power
	if adv.TxPowerLevel != nil {
		d.TxPowerLevel = adv.TxPowerLevel
	}

	// Update connectable flag
	d.Connectable = adv.Connectable

	// Store advertisement
	d.Advertisements = append(d.Advertisements, adv)
	if len(d.Advertisements) > maxAdvertisements {
		d.Advertisements = d.Advertisements[1:]
	}

	// Calculate advertisement interval
	d.calculateAdvInterval()
}

func (d *Device) calculateRSSIAverage() float64 {
	if len(d.RSSIHistory) == 0 {
		return 0
	}
	var sum int64
	for _, rssi := range d.RSSIHistory {
		sum += int64(rssi)
	}
	return float64(sum) / float64(len(d.RSSIHistory))
}

func (d *Device) calculateAdvInterval() {
	// Need at least 5 advertisements for a meaningful interval calculation
	if len(d.Advertisements) < 5 {
		d.AdvInterval = 0
		return
	}

	// Calculate intervals between consecutive advertisements
	intervals := make([]time.Duration, 0, len(d.Advertisements)-1)
	for i := 1; i < len(d.Advertisements); i++ {
		interval := d.Advertisements[i].Timestamp.Sub(d.Advertisements[i-1].Timestamp)
		// Filter out unreasonable intervals:
		// - BLE minimum advertisement interval is 20ms (we use 10ms to be safe)
		// - Maximum reasonable interval is 10 seconds
		if interval >= 10*time.Millisecond && interval < 10*time.Second {
			intervals = append(intervals, interval)
		}
	}

	// Need at least 3 valid intervals
	if len(intervals) < 3 {
		d.AdvInterval = 0
		return
	}

	// Use median interval
	d.AdvInterval = medianDuration(intervals)
}

func medianDuration(durations []time.Duration) time.Duration {
	n := len(durations)
	if n == 0 {
		return 0
	}

	// Simple selection for median - could use quickselect for better performance
	sorted := make([]time.Duration, n)
	copy(sorted, durations)
	for i := 0; i < n; i++ {
		for j := i + 1; j < n; j++ {
			if sorted[j] < sorted[i] {
				sorted[i], sorted[j] = sorted[j], sorted[i]
			}
		}
	}

	if n%2 == 0 {
		return (sorted[n/2-1] + sorted[n/2]) / 2
	}
	return sorted[n/2]
}

// GetDisplayName returns the device name or address if no name is set
func (d *Device) GetDisplayName() string {
	d.mu.RLock()
	defer d.mu.RUnlock()

	if d.Name != "" {
		return d.Name
	}
	return d.Address
}

// Copy returns a thread-safe copy of the device for UI rendering
func (d *Device) Copy() Device {
	d.mu.RLock()
	defer d.mu.RUnlock()

	copy := Device{
		Address:          d.Address,
		Name:             d.Name,
		RSSICurrent:      d.RSSICurrent,
		RSSIAverage:      d.RSSIAverage,
		FirstSeen:        d.FirstSeen,
		LastSeen:         d.LastSeen,
		AdvInterval:      d.AdvInterval,
		AdvCount:         d.AdvCount,
		ManufacturerData: append([]byte(nil), d.ManufacturerData...),
		ServiceUUIDs:     append([]string(nil), d.ServiceUUIDs...),
		TxPowerLevel:     d.TxPowerLevel,
		Connectable:      d.Connectable,
	}

	if d.ManufacturerID != nil {
		id := *d.ManufacturerID
		copy.ManufacturerID = &id
	}

	copy.RSSIHistory = append([]int16(nil), d.RSSIHistory...)
	copy.Advertisements = append([]Advertisement(nil), d.Advertisements...)

	copy.ServiceData = make(map[string][]byte)
	for k, v := range d.ServiceData {
		copy.ServiceData[k] = append([]byte(nil), v...)
	}

	return copy
}

// ADType represents an advertisement data type with its value
type ADType struct {
	Name  string
	Value string
}

// GetADTypes returns all AD types seen from this device with their current values
func (d *Device) GetADTypes() []ADType {
	d.mu.RLock()
	defer d.mu.RUnlock()

	var types []ADType

	if d.Name != "" {
		types = append(types, ADType{Name: "Local Name", Value: d.Name})
	}

	if d.ManufacturerID != nil {
		company := GetManufacturerName(*d.ManufacturerID)
		dataHex := ""
		if len(d.ManufacturerData) > 2 {
			dataHex = fmt.Sprintf(" [%x]", d.ManufacturerData[2:])
		}
		types = append(types, ADType{Name: "Manufacturer Data", Value: company + dataHex})
	}

	if len(d.ServiceUUIDs) > 0 {
		types = append(types, ADType{Name: "Service UUIDs", Value: strings.Join(d.ServiceUUIDs, ", ")})
	}

	if len(d.ServiceData) > 0 {
		var parts []string
		for uuid, data := range d.ServiceData {
			shortUUID := uuid
			if len(uuid) > 8 {
				shortUUID = uuid[:8]
			}
			parts = append(parts, fmt.Sprintf("%s:[%x]", shortUUID, data))
		}
		types = append(types, ADType{Name: "Service Data", Value: strings.Join(parts, ", ")})
	}

	if d.TxPowerLevel != nil {
		types = append(types, ADType{Name: "TX Power", Value: fmt.Sprintf("%d dBm", *d.TxPowerLevel)})
	}

	return types
}

// FormatADTypesSummary returns a short summary of AD types for the list view
func (d *Device) FormatADTypesSummary(maxLen int) string {
	types := d.GetADTypes()
	if len(types) == 0 {
		return "-"
	}

	var parts []string
	for _, t := range types {
		// Abbreviate the name for the summary
		shortName := t.Name
		switch t.Name {
		case "Local Name":
			shortName = "Name"
		case "Manufacturer Data":
			shortName = "Mfg"
		case "Service UUIDs":
			shortName = "Svc"
		case "Service Data":
			shortName = "SvcData"
		case "TX Power":
			shortName = "TxPwr"
		}

		// Truncate value if too long
		val := t.Value
		if len(val) > 20 {
			val = val[:17] + "..."
		}
		parts = append(parts, shortName+":"+val)
	}

	result := strings.Join(parts, ", ")
	if maxLen > 0 && len(result) > maxLen {
		if maxLen > 3 {
			result = result[:maxLen-3] + "..."
		} else {
			result = result[:maxLen]
		}
	}
	return result
}
