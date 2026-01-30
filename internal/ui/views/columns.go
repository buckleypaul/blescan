package views

import (
	"fmt"

	"github.com/buckleypaul/blescan/internal/ble"
)

// ColumnCategory represents the category of a column
type ColumnCategory int

const (
	CategoryAdvertisement ColumnCategory = iota // Data from BLE advertisement packets
	CategoryMetadata                            // Computed/derived metadata
)

// ColumnDefinition describes a single column in the device list
type ColumnDefinition struct {
	ID           string         // "flags", "name", "service_uuids", etc.
	Title        string         // Display name: "Flags", "Name", "Svc UUIDs"
	ShortTitle   string         // Abbreviated: "Flg", "Name", "Svc"
	Category     ColumnCategory // Category for grouping in UI
	MinWidth     int            // Minimum column width
	DefaultWidth int            // Default width
	WidthPct     int            // Percentage for proportional sizing
	ADTypes      []uint8        // Associated AD type codes
	Formatter    func(*ble.Device) string // Data extraction function
	Available    bool // Whether this AD type is available from library
}

// ColumnRegistry defines all possible columns
var ColumnRegistry = []ColumnDefinition{
	{
		ID:           "flags",
		Title:        "Flags",
		ShortTitle:   "Flg",
		Category:     CategoryAdvertisement,
		MinWidth:     10,
		DefaultWidth: 10,
		WidthPct:     7,
		ADTypes:      []uint8{0x01},
		Formatter: func(d *ble.Device) string {
			return d.FormatFlags()
		},
		Available: true,
	},
	{
		ID:           "name",
		Title:        "Name",
		ShortTitle:   "Name",
		Category:     CategoryAdvertisement,
		MinWidth:     10,
		DefaultWidth: 20,
		WidthPct:     13,
		ADTypes:      []uint8{0x08, 0x09},
		Formatter: func(d *ble.Device) string {
			if d.Name == "" {
				return "-"
			}
			return d.Name
		},
		Available: true,
	},
	{
		ID:           "service_uuids",
		Title:        "Svc UUIDs",
		ShortTitle:   "Svc",
		Category:     CategoryAdvertisement,
		MinWidth:     12,
		DefaultWidth: 12,
		WidthPct:     11,
		ADTypes:      []uint8{0x02, 0x03, 0x06, 0x07},
		Formatter: func(d *ble.Device) string {
			return d.FormatServiceUUIDs()
		},
		Available: true,
	},
	{
		ID:           "service_data",
		Title:        "Svc Data",
		ShortTitle:   "SvcD",
		Category:     CategoryAdvertisement,
		MinWidth:     12,
		DefaultWidth: 12,
		WidthPct:     11,
		ADTypes:      []uint8{0x16, 0x20, 0x21},
		Formatter: func(d *ble.Device) string {
			return d.FormatServiceData()
		},
		Available: true,
	},
	{
		ID:           "appearance",
		Title:        "Appearance",
		ShortTitle:   "App",
		Category:     CategoryAdvertisement,
		MinWidth:     12,
		DefaultWidth: 12,
		WidthPct:     9,
		ADTypes:      []uint8{0x19},
		Formatter: func(d *ble.Device) string {
			return d.FormatAppearance()
		},
		Available: true,
	},
	{
		ID:           "class_of_device",
		Title:        "Class of Device",
		ShortTitle:   "CoD",
		Category:     CategoryAdvertisement,
		MinWidth:     10,
		DefaultWidth: 12,
		WidthPct:     8,
		ADTypes:      []uint8{0x0D},
		Formatter: func(d *ble.Device) string {
			return "-"
		},
		Available: false,
	},
	{
		ID:           "service_solicitation",
		Title:        "Svc Solicitation",
		ShortTitle:   "SvcSol",
		Category:     CategoryAdvertisement,
		MinWidth:     12,
		DefaultWidth: 14,
		WidthPct:     9,
		ADTypes:      []uint8{0x14, 0x15, 0x1F},
		Formatter: func(d *ble.Device) string {
			return "-"
		},
		Available: false,
	},
	{
		ID:           "adv_interval",
		Title:        "Adv Interval",
		ShortTitle:   "AdvInt",
		Category:     CategoryAdvertisement,
		MinWidth:     10,
		DefaultWidth: 12,
		WidthPct:     8,
		ADTypes:      []uint8{0x1A},
		Formatter: func(d *ble.Device) string {
			return "-"
		},
		Available: false,
	},
	{
		ID:           "le_address",
		Title:        "LE Address",
		ShortTitle:   "LEAddr",
		Category:     CategoryAdvertisement,
		MinWidth:     12,
		DefaultWidth: 14,
		WidthPct:     9,
		ADTypes:      []uint8{0x1B},
		Formatter: func(d *ble.Device) string {
			return "-"
		},
		Available: false,
	},
	{
		ID:           "le_role",
		Title:        "LE Role",
		ShortTitle:   "Role",
		Category:     CategoryAdvertisement,
		MinWidth:     8,
		DefaultWidth: 10,
		WidthPct:     7,
		ADTypes:      []uint8{0x1C},
		Formatter: func(d *ble.Device) string {
			return "-"
		},
		Available: false,
	},
	{
		ID:           "uri",
		Title:        "URI",
		ShortTitle:   "URI",
		Category:     CategoryAdvertisement,
		MinWidth:     10,
		DefaultWidth: 15,
		WidthPct:     10,
		ADTypes:      []uint8{0x24},
		Formatter: func(d *ble.Device) string {
			return "-"
		},
		Available: false,
	},
	{
		ID:           "unknown_ad",
		Title:        "Unknown AD",
		ShortTitle:   "Unk",
		Category:     CategoryAdvertisement,
		MinWidth:     10,
		DefaultWidth: 12,
		WidthPct:     8,
		ADTypes:      []uint8{}, // Catch-all for truly unknown types
		Formatter: func(d *ble.Device) string {
			return d.FormatUnknownADTypes()
		},
		Available: true,
	},
	{
		ID:           "raw_data",
		Title:        "Raw Data",
		ShortTitle:   "Raw",
		Category:     CategoryAdvertisement,
		MinWidth:     20,
		DefaultWidth: 40,
		WidthPct:     15,
		ADTypes:      []uint8{}, // All raw advertisement data
		Formatter: func(d *ble.Device) string {
			return d.FormatRawData()
		},
		Available: false, // TinyGo doesn't expose raw advertisement bytes
	},
	{
		ID:           "company",
		Title:        "Company",
		ShortTitle:   "Mfg",
		Category:     CategoryAdvertisement,
		MinWidth:     10,
		DefaultWidth: 18,
		WidthPct:     13,
		ADTypes:      []uint8{0xFF},
		Formatter: func(d *ble.Device) string {
			if d.ManufacturerID != nil {
				return ble.GetManufacturerName(*d.ManufacturerID)
			}
			return "-"
		},
		Available: true,
	},
	{
		ID:           "tx_power",
		Title:        "TX Power",
		ShortTitle:   "TxPwr",
		Category:     CategoryAdvertisement,
		MinWidth:     8,
		DefaultWidth: 10,
		WidthPct:     7,
		ADTypes:      []uint8{0x0A},
		Formatter: func(d *ble.Device) string {
			if d.TxPowerLevel != nil {
				return fmt.Sprintf("%d dBm", *d.TxPowerLevel)
			}
			return "-"
		},
		Available: true,
	},
	{
		ID:           "rssi",
		Title:        "RSSI",
		ShortTitle:   "RSSI",
		Category:     CategoryMetadata,
		MinWidth:     8,
		DefaultWidth: 10,
		WidthPct:     9,
		ADTypes:      []uint8{},
		Formatter: func(d *ble.Device) string {
			return fmt.Sprintf("%.1f", d.RSSIAverage)
		},
		Available: true,
	},
	{
		ID:           "count",
		Title:        "Count",
		ShortTitle:   "Cnt",
		Category:     CategoryMetadata,
		MinWidth:     6,
		DefaultWidth: 8,
		WidthPct:     8,
		ADTypes:      []uint8{},
		Formatter: func(d *ble.Device) string {
			return fmt.Sprintf("%d", d.AdvCount)
		},
		Available: true,
	},
	{
		ID:           "interval",
		Title:        "Interval",
		ShortTitle:   "Int",
		Category:     CategoryMetadata,
		MinWidth:     8,
		DefaultWidth: 10,
		WidthPct:     9,
		ADTypes:      []uint8{},
		Formatter: func(d *ble.Device) string {
			if d.AdvInterval > 0 {
				return fmt.Sprintf("%dms", d.AdvInterval.Milliseconds())
			}
			return "-"
		},
		Available: true,
	},
}

// DefaultEnabledColumns returns the default set of enabled column IDs
func DefaultEnabledColumns() []string {
	return []string{
		"flags",
		"name",
		"service_uuids",
		"service_data",
		"appearance",
		"unknown_ad",
		"company",
		"rssi",
		"count",
		"interval",
	}
}

// BuildColumnLookup creates a map for quick column definition lookup by ID
func BuildColumnLookup() map[string]*ColumnDefinition {
	lookup := make(map[string]*ColumnDefinition)
	for i := range ColumnRegistry {
		lookup[ColumnRegistry[i].ID] = &ColumnRegistry[i]
	}
	return lookup
}
