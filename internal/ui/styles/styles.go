package styles

import (
	"fmt"

	"github.com/charmbracelet/lipgloss"
)

var (
	// Colors
	PrimaryColor   = lipgloss.Color("39")  // Cyan
	SecondaryColor = lipgloss.Color("147") // Light purple
	AccentColor    = lipgloss.Color("208") // Orange
	MutedColor     = lipgloss.Color("241") // Gray
	ErrorColor     = lipgloss.Color("196") // Red
	SuccessColor   = lipgloss.Color("46")  // Green

	// Signal strength colors
	SignalExcellent = lipgloss.Color("46")  // Bright green
	SignalGood      = lipgloss.Color("226") // Yellow
	SignalFair      = lipgloss.Color("208") // Orange
	SignalWeak      = lipgloss.Color("196") // Red

	// Base styles
	BaseStyle = lipgloss.NewStyle()

	// Title styles
	TitleStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(PrimaryColor).
			MarginBottom(1)

	// Header styles
	HeaderStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(SecondaryColor).
			BorderStyle(lipgloss.NormalBorder()).
			BorderBottom(true).
			BorderForeground(MutedColor)

	// Table styles
	TableHeaderStyle = lipgloss.NewStyle().
				Bold(true).
				Foreground(SecondaryColor).
				Padding(0, 1)

	TableCellStyle = lipgloss.NewStyle().
			Padding(0, 1)

	TableSelectedStyle = lipgloss.NewStyle().
				Background(lipgloss.Color("236")).
				Foreground(lipgloss.Color("255")).
				Padding(0, 1)

	// Box styles
	BoxStyle = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(MutedColor).
			Padding(0, 1)

	DetailBoxStyle = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(PrimaryColor).
			Padding(1, 2)

	// Status bar
	StatusBarStyle = lipgloss.NewStyle().
			Foreground(MutedColor).
			MarginTop(1)

	// Help style
	HelpStyle = lipgloss.NewStyle().
			Foreground(MutedColor)

	// Filter input style
	FilterStyle = lipgloss.NewStyle().
			Border(lipgloss.NormalBorder()).
			BorderForeground(PrimaryColor).
			Padding(0, 1)

	FilterLabelStyle = lipgloss.NewStyle().
				Foreground(PrimaryColor).
				Bold(true)

	// Detail view styles
	LabelStyle = lipgloss.NewStyle().
			Foreground(MutedColor)

	ValueStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("255"))

	HighlightValueStyle = lipgloss.NewStyle().
				Foreground(PrimaryColor).
				Bold(true)

	// Raw data stream style
	RawDataStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("245"))
)

// GetRSSIStyle returns a style based on signal strength
func GetRSSIStyle(rssi int16) lipgloss.Style {
	var color lipgloss.Color
	switch {
	case rssi >= -50:
		color = SignalExcellent
	case rssi >= -70:
		color = SignalGood
	case rssi >= -85:
		color = SignalFair
	default:
		color = SignalWeak
	}
	return lipgloss.NewStyle().Foreground(color)
}

// FormatRSSI returns a styled RSSI string
func FormatRSSI(rssi int16) string {
	return GetRSSIStyle(rssi).Render(fmt.Sprintf("%d", rssi))
}
