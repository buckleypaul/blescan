package stats

import (
	"time"

	"github.com/paulbuckley/blescan/internal/ble"
)

// FilterConfig defines filtering criteria for devices
type FilterConfig struct {
	NameContains string // Case-insensitive substring match
	MinRSSI      *int16 // Only show devices with RSSI >= this
}

// MatchesFilter checks if a device matches the filter criteria
func MatchesFilter(d *ble.Device, f FilterConfig) bool {
	if f.NameContains != "" {
		name := d.GetDisplayName()
		if !containsIgnoreCase(name, f.NameContains) {
			return false
		}
	}
	if f.MinRSSI != nil && d.RSSICurrent < *f.MinRSSI {
		return false
	}
	return true
}

func containsIgnoreCase(s, substr string) bool {
	sLower := toLower(s)
	substrLower := toLower(substr)
	return contains(sLower, substrLower)
}

func toLower(s string) string {
	b := make([]byte, len(s))
	for i := 0; i < len(s); i++ {
		c := s[i]
		if c >= 'A' && c <= 'Z' {
			c += 'a' - 'A'
		}
		b[i] = c
	}
	return string(b)
}

func contains(s, substr string) bool {
	if len(substr) > len(s) {
		return false
	}
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}

// SortField defines the field to sort devices by
type SortField int

const (
	SortByName SortField = iota
	SortByRSSI
	SortByAdvCount
	SortByLastSeen
)

// DeviceStats holds calculated statistics for a device
type DeviceStats struct {
	AdvertisementsPerSecond float64
	TimeSinceLastSeen       time.Duration
	SignalStrength          SignalStrength
}

// SignalStrength categorizes RSSI values
type SignalStrength int

const (
	SignalExcellent SignalStrength = iota // >= -50 dBm
	SignalGood                            // -50 to -70 dBm
	SignalFair                            // -70 to -85 dBm
	SignalWeak                            // < -85 dBm
)

// GetSignalStrength returns the signal strength category for an RSSI value
func GetSignalStrength(rssi int16) SignalStrength {
	switch {
	case rssi >= -50:
		return SignalExcellent
	case rssi >= -70:
		return SignalGood
	case rssi >= -85:
		return SignalFair
	default:
		return SignalWeak
	}
}

// SignalStrengthLabel returns a human-readable label for signal strength
func SignalStrengthLabel(s SignalStrength) string {
	switch s {
	case SignalExcellent:
		return "Excellent"
	case SignalGood:
		return "Good"
	case SignalFair:
		return "Fair"
	case SignalWeak:
		return "Weak"
	default:
		return "Unknown"
	}
}

// CalculateDeviceStats calculates statistics for a device
func CalculateDeviceStats(d ble.Device) DeviceStats {
	stats := DeviceStats{
		TimeSinceLastSeen: time.Since(d.LastSeen),
		SignalStrength:    GetSignalStrength(d.RSSICurrent),
	}

	// Calculate advertisements per second over last 10 seconds
	if len(d.Advertisements) > 1 {
		cutoff := time.Now().Add(-10 * time.Second)
		count := 0
		for _, adv := range d.Advertisements {
			if adv.Timestamp.After(cutoff) {
				count++
			}
		}
		stats.AdvertisementsPerSecond = float64(count) / 10.0
	}

	return stats
}
