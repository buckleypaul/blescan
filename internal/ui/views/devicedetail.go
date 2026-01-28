package views

import (
	"fmt"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/buckleypaul/blescan/internal/ble"
	"github.com/buckleypaul/blescan/internal/stats"
	"github.com/buckleypaul/blescan/internal/ui/styles"
)

// DeviceDetailModel represents the device detail view
type DeviceDetailModel struct {
	Device   ble.Device
	viewport viewport.Model
	width    int
	height   int
	ready    bool
}

// NewDeviceDetailModel creates a new device detail model
func NewDeviceDetailModel(device ble.Device) DeviceDetailModel {
	return DeviceDetailModel{
		Device: device,
	}
}

// Init initializes the device detail model
func (m DeviceDetailModel) Init() tea.Cmd {
	return nil
}

// Update handles device detail updates
func (m DeviceDetailModel) Update(msg tea.Msg) (DeviceDetailModel, tea.Cmd) {
	var cmd tea.Cmd

	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height

		headerHeight := 4
		footerHeight := 2
		verticalMargins := headerHeight + footerHeight

		if !m.ready {
			m.viewport = viewport.New(msg.Width-4, msg.Height-verticalMargins)
			m.viewport.YPosition = headerHeight
			m.ready = true
		} else {
			m.viewport.Width = msg.Width - 4
			m.viewport.Height = msg.Height - verticalMargins
		}
		m.viewport.SetContent(m.renderContent())
	default:
		m.viewport, cmd = m.viewport.Update(msg)
	}

	return m, cmd
}

// View renders the device detail view
func (m DeviceDetailModel) View() string {
	if !m.ready {
		return "\n  Initializing..."
	}

	var b strings.Builder

	// Title bar
	titleStyle := lipgloss.NewStyle().
		Bold(true).
		Foreground(styles.PrimaryColor).
		Background(lipgloss.Color("235")).
		Padding(0, 2).
		Width(m.width)

	title := m.Device.GetDisplayName()
	if len(title) > m.width-20 {
		title = title[:m.width-23] + "..."
	}
	b.WriteString(titleStyle.Render(title))
	b.WriteString("\n")

	// Address bar
	addrStyle := lipgloss.NewStyle().
		Foreground(styles.MutedColor).
		Background(lipgloss.Color("236")).
		Padding(0, 2).
		Width(m.width)
	b.WriteString(addrStyle.Render(m.Device.Address))
	b.WriteString("\n")

	// Content viewport
	viewportStyle := lipgloss.NewStyle().
		BorderStyle(lipgloss.RoundedBorder()).
		BorderForeground(styles.MutedColor).
		Padding(0, 1).
		Width(m.width - 2)

	b.WriteString(viewportStyle.Render(m.viewport.View()))
	b.WriteString("\n")

	// Help bar
	helpStyle := lipgloss.NewStyle().
		Foreground(styles.MutedColor).
		Background(lipgloss.Color("235")).
		Padding(0, 2).
		Width(m.width)

	scrollPercent := fmt.Sprintf("%3.f%%", m.viewport.ScrollPercent()*100)
	help := "↑/↓ Scroll • Esc Back • q Quit"
	helpContent := help + strings.Repeat(" ", max(0, m.width-len(help)-len(scrollPercent)-6)) + scrollPercent
	b.WriteString(helpStyle.Render(helpContent))

	return b.String()
}

func (m DeviceDetailModel) renderContent() string {
	var sections []string

	// Signal section
	sections = append(sections, m.renderSignalSection())

	// Statistics section
	sections = append(sections, m.renderStatsSection())

	// AD Types section
	adTypes := m.Device.GetADTypes()
	if len(adTypes) > 0 {
		sections = append(sections, m.renderADTypesSection(adTypes))
	}

	// Advertisements section
	sections = append(sections, m.renderAdvertisementsSection())

	return strings.Join(sections, "\n\n")
}

func (m DeviceDetailModel) renderSignalSection() string {
	sectionStyle := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(styles.PrimaryColor).
		Padding(0, 2).
		Width(m.width - 8)

	headerStyle := lipgloss.NewStyle().Bold(true).Foreground(styles.PrimaryColor)
	labelStyle := lipgloss.NewStyle().Foreground(styles.MutedColor).Width(16)
	valueStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("255"))

	deviceStats := stats.CalculateDeviceStats(m.Device)

	var content strings.Builder
	content.WriteString(headerStyle.Render("Signal"))
	content.WriteString("\n\n")

	// RSSI with color
	rssiColor := styles.GetRSSIStyle(m.Device.RSSICurrent)
	content.WriteString(labelStyle.Render("Current RSSI:"))
	content.WriteString(rssiColor.Render(fmt.Sprintf("%d dBm", m.Device.RSSICurrent)))
	content.WriteString("\n")

	content.WriteString(labelStyle.Render("Average RSSI:"))
	content.WriteString(valueStyle.Render(fmt.Sprintf("%.1f dBm", m.Device.RSSIAverage)))
	content.WriteString("\n")

	content.WriteString(labelStyle.Render("Signal Quality:"))
	qualityStyle := styles.GetRSSIStyle(m.Device.RSSICurrent)
	content.WriteString(qualityStyle.Render(stats.SignalStrengthLabel(deviceStats.SignalStrength)))
	content.WriteString("\n")

	if m.Device.TxPowerLevel != nil {
		content.WriteString(labelStyle.Render("TX Power:"))
		content.WriteString(valueStyle.Render(fmt.Sprintf("%d dBm", *m.Device.TxPowerLevel)))
		content.WriteString("\n")
	}

	connectStr := "No"
	if m.Device.Connectable {
		connectStr = "Yes"
	}
	content.WriteString(labelStyle.Render("Connectable:"))
	content.WriteString(valueStyle.Render(connectStr))

	return sectionStyle.Render(content.String())
}

func (m DeviceDetailModel) renderStatsSection() string {
	sectionStyle := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(styles.SecondaryColor).
		Padding(0, 2).
		Width(m.width - 8)

	headerStyle := lipgloss.NewStyle().Bold(true).Foreground(styles.SecondaryColor)
	labelStyle := lipgloss.NewStyle().Foreground(styles.MutedColor).Width(16)
	valueStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("255"))

	deviceStats := stats.CalculateDeviceStats(m.Device)

	var content strings.Builder
	content.WriteString(headerStyle.Render("Statistics"))
	content.WriteString("\n\n")

	content.WriteString(labelStyle.Render("Advertisements:"))
	content.WriteString(valueStyle.Render(fmt.Sprintf("%d", m.Device.AdvCount)))
	content.WriteString("\n")

	content.WriteString(labelStyle.Render("Interval:"))
	intervalStr := "-"
	if m.Device.AdvInterval > 0 {
		intervalStr = fmt.Sprintf("%dms", m.Device.AdvInterval.Milliseconds())
	}
	content.WriteString(valueStyle.Render(intervalStr))
	content.WriteString("\n")

	if deviceStats.AdvertisementsPerSecond > 0 {
		content.WriteString(labelStyle.Render("Rate:"))
		content.WriteString(valueStyle.Render(fmt.Sprintf("%.1f/sec", deviceStats.AdvertisementsPerSecond)))
		content.WriteString("\n")
	}

	content.WriteString(labelStyle.Render("First Seen:"))
	content.WriteString(valueStyle.Render(m.Device.FirstSeen.Format("15:04:05")))
	content.WriteString("\n")

	content.WriteString(labelStyle.Render("Last Seen:"))
	content.WriteString(valueStyle.Render(m.Device.LastSeen.Format("15:04:05")))

	return sectionStyle.Render(content.String())
}

func (m DeviceDetailModel) renderADTypesSection(adTypes []ble.ADType) string {
	sectionStyle := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(styles.AccentColor).
		Padding(0, 2).
		Width(m.width - 8)

	headerStyle := lipgloss.NewStyle().Bold(true).Foreground(styles.AccentColor)
	labelStyle := lipgloss.NewStyle().Foreground(styles.MutedColor).Width(20)
	valueStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("255"))

	var content strings.Builder
	content.WriteString(headerStyle.Render("Advertisement Data Types"))
	content.WriteString("\n\n")

	maxValueWidth := m.width - 35
	if maxValueWidth < 20 {
		maxValueWidth = 20
	}

	for _, adType := range adTypes {
		content.WriteString(labelStyle.Render(adType.Name + ":"))

		// Handle multi-line values (like long hex strings)
		value := adType.Value
		if len(value) > maxValueWidth {
			// Wrap long values
			for len(value) > 0 {
				lineLen := maxValueWidth
				if lineLen > len(value) {
					lineLen = len(value)
				}
				content.WriteString(valueStyle.Render(value[:lineLen]))
				value = value[lineLen:]
				if len(value) > 0 {
					content.WriteString("\n")
					content.WriteString(strings.Repeat(" ", 20)) // Indent continuation
				}
			}
		} else {
			content.WriteString(valueStyle.Render(value))
		}
		content.WriteString("\n")
	}

	return sectionStyle.Render(strings.TrimRight(content.String(), "\n"))
}

func (m DeviceDetailModel) renderAdvertisementsSection() string {
	sectionStyle := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(styles.MutedColor).
		Padding(0, 2).
		Width(m.width - 8)

	headerStyle := lipgloss.NewStyle().Bold(true).Foreground(styles.MutedColor)
	timeStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("245"))
	dataStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("240"))

	var content strings.Builder
	content.WriteString(headerStyle.Render(fmt.Sprintf("Recent Advertisements (%d total)", len(m.Device.Advertisements))))
	content.WriteString("\n\n")

	// Show last 20 advertisements
	ads := m.Device.Advertisements
	start := 0
	if len(ads) > 20 {
		start = len(ads) - 20
	}

	for i := len(ads) - 1; i >= start; i-- {
		adv := ads[i]
		timeStr := adv.Timestamp.Format("15:04:05.000")
		rssiStyle := styles.GetRSSIStyle(adv.RSSI)

		content.WriteString(timeStyle.Render(timeStr))
		content.WriteString("  ")
		content.WriteString(rssiStyle.Render(fmt.Sprintf("%4d", adv.RSSI)))
		content.WriteString(" dBm  ")

		dataHex := formatAdvPayload(adv, m.width-40)
		content.WriteString(dataStyle.Render(dataHex))
		content.WriteString("\n")
	}

	return sectionStyle.Render(strings.TrimRight(content.String(), "\n"))
}

func formatInterval(d time.Duration) string {
	if d == 0 {
		return "-"
	}
	return fmt.Sprintf("%dms", d.Milliseconds())
}

// formatAdvPayload returns a hex string of the advertisement payload
// Prefers manufacturer data, falls back to service data
func formatAdvPayload(adv ble.Advertisement, maxLen int) string {
	if maxLen < 20 {
		maxLen = 20
	}

	var dataHex string
	var prefix string

	if len(adv.ManufacturerData) > 0 {
		dataHex = fmt.Sprintf("%x", adv.ManufacturerData)
	} else if len(adv.ServiceData) > 0 {
		// Show first service data entry
		for uuid, data := range adv.ServiceData {
			if len(data) > 0 {
				// Show shortened UUID prefix
				shortUUID := uuid
				if len(uuid) > 8 {
					shortUUID = uuid[:8]
				}
				prefix = shortUUID + ":"
				dataHex = fmt.Sprintf("%x", data)
				break
			}
		}
	}

	if dataHex == "" {
		return "-"
	}

	full := prefix + dataHex
	if len(full) > maxLen {
		full = full[:maxLen] + "..."
	}
	return full
}

// UpdateDevice updates the device being displayed
func (m *DeviceDetailModel) UpdateDevice(device ble.Device) {
	m.Device = device
	if m.ready {
		m.viewport.SetContent(m.renderContent())
	}
}
