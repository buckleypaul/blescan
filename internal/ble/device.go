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
	Flags            *uint8
	Appearance       *uint16
	ADTypes          []uint8 // All AD type codes seen

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

	// Update flags
	if adv.Flags != nil {
		d.Flags = adv.Flags
	}

	// Update appearance
	if adv.Appearance != nil {
		d.Appearance = adv.Appearance
	}

	// Update AD types - merge with existing
	if len(adv.ADTypes) > 0 {
		// Build a set of all unique AD types seen
		typeSet := make(map[uint8]bool)
		for _, t := range d.ADTypes {
			typeSet[t] = true
		}
		for _, t := range adv.ADTypes {
			typeSet[t] = true
		}
		// Convert back to slice
		d.ADTypes = make([]uint8, 0, len(typeSet))
		for t := range typeSet {
			d.ADTypes = append(d.ADTypes, t)
		}
	}

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

	if d.Flags != nil {
		flags := *d.Flags
		copy.Flags = &flags
	}

	if d.Appearance != nil {
		appearance := *d.Appearance
		copy.Appearance = &appearance
	}

	copy.ADTypes = append([]uint8(nil), d.ADTypes...)

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

// FormatFlags returns a formatted string of the flags field
func (d *Device) FormatFlags() string {
	d.mu.RLock()
	defer d.mu.RUnlock()

	if d.Flags == nil {
		return "-"
	}

	flags := *d.Flags
	var parts []string

	if flags&0x01 != 0 {
		parts = append(parts, "LE-Ltd")
	}
	if flags&0x02 != 0 {
		parts = append(parts, "LE-Gen")
	}
	if flags&0x04 != 0 {
		parts = append(parts, "BR/EDR")
	}
	if flags&0x08 != 0 {
		parts = append(parts, "LE-BR/EDR-Ctrl")
	}
	if flags&0x10 != 0 {
		parts = append(parts, "LE-BR/EDR-Host")
	}

	if len(parts) == 0 {
		return fmt.Sprintf("0x%02x", flags)
	}
	return strings.Join(parts, ",")
}

// FormatServiceUUIDs returns a formatted string of service UUIDs
func (d *Device) FormatServiceUUIDs() string {
	d.mu.RLock()
	defer d.mu.RUnlock()

	if len(d.ServiceUUIDs) == 0 {
		return "-"
	}

	// Shorten UUIDs for display
	var shortened []string
	for _, uuid := range d.ServiceUUIDs {
		// If it's a full UUID, show just the characteristic part
		if len(uuid) > 8 {
			shortened = append(shortened, uuid[:8])
		} else {
			shortened = append(shortened, uuid)
		}
	}

	result := strings.Join(shortened, ",")
	if len(result) > 30 {
		return result[:27] + "..."
	}
	return result
}

// FormatServiceData returns a formatted string of service data
func (d *Device) FormatServiceData() string {
	d.mu.RLock()
	defer d.mu.RUnlock()

	if len(d.ServiceData) == 0 {
		return "-"
	}

	var parts []string
	for uuid, data := range d.ServiceData {
		shortUUID := uuid
		if len(uuid) > 8 {
			shortUUID = uuid[:8]
		}
		// Show first few bytes of data
		dataStr := fmt.Sprintf("%x", data)
		if len(dataStr) > 8 {
			dataStr = dataStr[:8] + "..."
		}
		parts = append(parts, fmt.Sprintf("%s:%s", shortUUID, dataStr))
	}

	result := strings.Join(parts, ",")
	if len(result) > 30 {
		return result[:27] + "..."
	}
	return result
}

// FormatAppearance returns a formatted string of the appearance value
func (d *Device) FormatAppearance() string {
	d.mu.RLock()
	defer d.mu.RUnlock()

	if d.Appearance == nil {
		return "-"
	}

	appearance := *d.Appearance

	// BLE Appearance values (common ones)
	// Category is upper 10 bits, subcategory is lower 6 bits
	category := appearance >> 6
	switch category {
	case 0:
		return "Unknown"
	case 1:
		return "Phone"
	case 2:
		return "Computer"
	case 3:
		return "Watch"
	case 4:
		return "Clock"
	case 5:
		return "Display"
	case 6:
		return "Remote"
	case 7:
		return "Eye-glasses"
	case 8:
		return "Tag"
	case 9:
		return "Keyring"
	case 10:
		return "Media Player"
	case 11:
		return "Barcode Scanner"
	case 12:
		return "Thermometer"
	case 13:
		return "Heart Rate"
	case 14:
		return "Blood Pressure"
	case 15:
		return "HID"
	case 16:
		return "Glucose"
	case 17:
		return "Running/Walking"
	case 18:
		return "Cycling"
	case 49:
		return "Pulse Oximeter"
	case 50:
		return "Weight Scale"
	case 51:
		return "Personal Mobility"
	case 52:
		return "Continuous Glucose"
	case 53:
		return "Insulin Pump"
	case 54:
		return "Medication Delivery"
	case 81:
		return "Outdoor Sports"
	default:
		return fmt.Sprintf("0x%04x", appearance)
	}
}

// FormatOtherADTypes returns a list of AD types not shown in other columns
func (d *Device) FormatOtherADTypes() string {
	d.mu.RLock()
	defer d.mu.RUnlock()

	if len(d.ADTypes) == 0 {
		return "-"
	}

	// AD types that have dedicated columns
	knownTypes := map[uint8]bool{
		0x01: true, // Flags
		0x02: true, // 16-bit Service UUIDs (incomplete)
		0x03: true, // 16-bit Service UUIDs (complete)
		0x06: true, // 128-bit Service UUIDs (incomplete)
		0x07: true, // 128-bit Service UUIDs (complete)
		0x08: true, // Shortened Local Name
		0x09: true, // Complete Local Name
		0x16: true, // Service Data - 16-bit UUID
		0x19: true, // Appearance
		0x20: true, // Service Data - 32-bit UUID
		0x21: true, // Service Data - 128-bit UUID
		0xFF: true, // Manufacturer Specific Data
	}

	var other []string
	for _, adType := range d.ADTypes {
		if !knownTypes[adType] {
			other = append(other, fmt.Sprintf("0x%02X", adType))
		}
	}

	if len(other) == 0 {
		return "-"
	}

	result := strings.Join(other, ",")
	if len(result) > 30 {
		return result[:27] + "..."
	}
	return result
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

// FormatRawData returns the raw advertisement data as hex string
func (d *Device) FormatRawData() string {
	d.mu.RLock()
	defer d.mu.RUnlock()

	if len(d.Advertisements) == 0 {
		return "-"
	}

	// Get the most recent advertisement
	latest := d.Advertisements[len(d.Advertisements)-1]
	if len(latest.RawData) == 0 {
		return "-"
	}

	// Format as hex, show first 32 bytes
	hexStr := fmt.Sprintf("%x", latest.RawData)
	if len(hexStr) > 64 {
		hexStr = hexStr[:64] + "..."
	}
	return hexStr
}

// FormatUnknownADTypes returns AD types that aren't shown in specific columns
func (d *Device) FormatUnknownADTypes() string {
	d.mu.RLock()
	defer d.mu.RUnlock()

	if len(d.ADTypes) == 0 {
		return "-"
	}

	// AD types that have dedicated columns
	knownTypes := map[uint8]bool{
		0x01: true, // Flags
		0x02: true, // 16-bit Service UUIDs (incomplete)
		0x03: true, // 16-bit Service UUIDs (complete)
		0x06: true, // 128-bit Service UUIDs (incomplete)
		0x07: true, // 128-bit Service UUIDs (complete)
		0x08: true, // Shortened Local Name
		0x09: true, // Complete Local Name
		0x0A: true, // TX Power
		0x0D: true, // Class of Device
		0x14: true, // 16-bit Service Solicitation UUIDs
		0x15: true, // 128-bit Service Solicitation UUIDs
		0x16: true, // Service Data - 16-bit UUID
		0x19: true, // Appearance
		0x1A: true, // Advertising Interval
		0x1B: true, // LE Bluetooth Device Address
		0x1C: true, // LE Role
		0x1F: true, // 32-bit Service Solicitation UUIDs
		0x20: true, // Service Data - 32-bit UUID
		0x21: true, // Service Data - 128-bit UUID
		0x24: true, // URI
		0xFF: true, // Manufacturer Specific Data
	}

	var unknown []string
	for _, adType := range d.ADTypes {
		if !knownTypes[adType] {
			unknown = append(unknown, fmt.Sprintf("0x%02X", adType))
		}
	}

	if len(unknown) == 0 {
		return "-"
	}

	result := strings.Join(unknown, ",")
	if len(result) > 30 {
		return result[:27] + "..."
	}
	return result
}
