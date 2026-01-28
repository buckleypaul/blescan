package views

import (
	"fmt"
	"sort"
	"strings"

	"github.com/charmbracelet/bubbles/table"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/buckleypaul/blescan/internal/ble"
	"github.com/buckleypaul/blescan/internal/ui/styles"
)

// Column indices
const (
	ColName = iota
	ColCompany
	ColADTypes
	ColRSSI
	ColAvg
	ColCount
	ColInterval
	NumColumns
)

// DeviceListModel represents the device list view
type DeviceListModel struct {
	devices        []ble.Device
	filtered       []ble.Device
	table          table.Model
	width          int
	height         int
	filter         FilterModel
	sortColumn     int
	sortAscending  bool
	selectedColumn int
	columnWidths   []int
}

// NewDeviceListModel creates a new device list model
func NewDeviceListModel() DeviceListModel {
	t := table.New(
		table.WithColumns([]table.Column{}),
		table.WithRows([]table.Row{}),
		table.WithFocused(true),
		table.WithHeight(10),
	)

	s := table.DefaultStyles()
	s.Header = s.Header.
		BorderStyle(lipgloss.NormalBorder()).
		BorderForeground(styles.MutedColor).
		BorderBottom(true).
		Bold(true).
		Foreground(styles.PrimaryColor)
	s.Selected = s.Selected.
		Foreground(lipgloss.Color("229")).
		Background(lipgloss.Color("57")).
		Bold(true)
	s.Cell = s.Cell.Padding(0, 1)
	t.SetStyles(s)

	m := DeviceListModel{
		devices:        make([]ble.Device, 0),
		filtered:       make([]ble.Device, 0),
		table:          t,
		filter:         NewFilterModel(),
		sortColumn:     ColAvg,
		sortAscending:  false,
		selectedColumn: ColAvg,
		columnWidths:   []int{20, 18, 20, 8, 8, 8, 10},
	}
	m.updateColumns()
	return m
}

// Init initializes the device list model
func (m DeviceListModel) Init() tea.Cmd {
	return nil
}

// Update handles device list updates
func (m DeviceListModel) Update(msg tea.Msg) (DeviceListModel, tea.Cmd) {
	var cmd tea.Cmd

	// If filter is active, route to filter
	if m.filter.Mode != FilterModeNone {
		m.filter, cmd = m.filter.Update(msg)
		m.applyFilterAndSort()
		return m, cmd
	}

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "left", "h":
			if m.selectedColumn > 0 {
				m.selectedColumn--
				m.updateColumns()
			}
		case "right", "l":
			if m.selectedColumn < NumColumns-1 {
				m.selectedColumn++
				m.updateColumns()
			}
		case "s":
			// Sort by selected column
			if m.sortColumn == m.selectedColumn {
				m.sortAscending = !m.sortAscending
			} else {
				m.sortColumn = m.selectedColumn
				m.sortAscending = false
			}
			m.updateColumns()
			m.applyFilterAndSort()
		case "/", "n":
			return m, m.filter.SetMode(FilterModeName)
		case "r":
			return m, m.filter.SetMode(FilterModeRSSI)
		case "c":
			m.filter.ClearFilters()
			m.applyFilterAndSort()
		default:
			m.table, cmd = m.table.Update(msg)
		}
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		m.updateTableSize()
	default:
		m.table, cmd = m.table.Update(msg)
	}

	return m, cmd
}

func (m *DeviceListModel) updateTableSize() {
	// Reserve space for title, filter, and help
	tableHeight := m.height - 7
	if tableHeight < 5 {
		tableHeight = 5
	}
	m.table.SetHeight(tableHeight)

	// Distribute width across columns
	if m.width > 20 {
		availableWidth := m.width - 8 // margins and borders
		// Proportional column widths
		m.columnWidths = []int{
			availableWidth * 16 / 100, // Name
			availableWidth * 16 / 100, // Company
			availableWidth * 20 / 100, // Services
			availableWidth * 10 / 100, // RSSI
			availableWidth * 10 / 100, // Avg
			availableWidth * 12 / 100, // Count
			availableWidth * 12 / 100, // Interval
		}

		// Minimum widths
		if m.columnWidths[0] < 10 {
			m.columnWidths[0] = 10
		}
		if m.columnWidths[1] < 10 {
			m.columnWidths[1] = 10
		}
		if m.columnWidths[2] < 12 {
			m.columnWidths[2] = 12
		}

		m.updateColumns()
	}
}

// View renders the device list
func (m DeviceListModel) View() string {
	var b strings.Builder

	// Title bar
	titleStyle := lipgloss.NewStyle().
		Bold(true).
		Foreground(styles.PrimaryColor).
		Background(lipgloss.Color("235")).
		Padding(0, 2).
		Width(m.width)

	title := "BLE Device Scanner"
	deviceCount := fmt.Sprintf("%d devices", len(m.filtered))
	if len(m.filtered) != len(m.devices) {
		deviceCount = fmt.Sprintf("%d/%d devices", len(m.filtered), len(m.devices))
	}
	titleContent := title + strings.Repeat(" ", max(0, m.width-len(title)-len(deviceCount)-6)) + deviceCount
	b.WriteString(titleStyle.Render(titleContent))
	b.WriteString("\n")

	// Filter bar
	filterBarStyle := lipgloss.NewStyle().
		Foreground(styles.SecondaryColor).
		Background(lipgloss.Color("236")).
		Padding(0, 2).
		Width(m.width)

	var filterContent string
	if m.filter.Mode != FilterModeNone {
		filterContent = m.filter.View()
	} else if m.filter.IsFiltering() {
		filterContent = m.filter.FilterSummary()
	} else {
		filterContent = "Use ←/→ to select column, 's' to sort"
	}
	b.WriteString(filterBarStyle.Render(filterContent))
	b.WriteString("\n")

	// Table with dynamic header
	tableStyle := lipgloss.NewStyle().
		BorderStyle(lipgloss.RoundedBorder()).
		BorderForeground(styles.MutedColor).
		Width(m.width - 2)

	b.WriteString(tableStyle.Render(m.table.View()))
	b.WriteString("\n")

	// Help bar
	helpStyle := lipgloss.NewStyle().
		Foreground(styles.MutedColor).
		Background(lipgloss.Color("235")).
		Padding(0, 2).
		Width(m.width)

	help := "↑/↓ Row • ←/→ Column • s Sort • Enter View • / Filter • c Clear • q Quit"
	b.WriteString(helpStyle.Render(help))

	return b.String()
}

func (m *DeviceListModel) updateColumns() {
	baseTitles := []string{"Name", "Company", "AD Types", "RSSI", "Avg", "Count", "Interval"}

	columns := make([]table.Column, len(baseTitles))
	for i, title := range baseTitles {
		displayTitle := title

		// Add selection marker with visible unicode brackets
		if i == m.selectedColumn {
			displayTitle = ">" + title + "<"
		}

		// Add sort indicator
		if i == m.sortColumn {
			if m.sortAscending {
				displayTitle += " ↑"
			} else {
				displayTitle += " ↓"
			}
		}

		columns[i] = table.Column{
			Title: displayTitle,
			Width: m.columnWidths[i],
		}
	}
	m.table.SetColumns(columns)
}

func (m *DeviceListModel) applyFilterAndSort() {
	// Filter
	m.filtered = make([]ble.Device, 0, len(m.devices))
	for _, d := range m.devices {
		if m.matchesFilter(d) {
			m.filtered = append(m.filtered, d)
		}
	}

	// Sort
	sort.Slice(m.filtered, func(i, j int) bool {
		return m.compareDevices(m.filtered[i], m.filtered[j])
	})

	// Update table rows
	m.updateTableRows()
}

func (m *DeviceListModel) updateTableRows() {
	rows := make([]table.Row, len(m.filtered))
	for i, device := range m.filtered {
		name := device.Name
		if name == "" {
			name = "(unnamed)"
		}
		// Truncate long names
		maxNameLen := m.columnWidths[ColName] - 2
		if maxNameLen > 0 && len(name) > maxNameLen {
			name = name[:maxNameLen-3] + "..."
		}

		// Company/Manufacturer
		company := "-"
		if device.ManufacturerID != nil {
			company = ble.GetManufacturerName(*device.ManufacturerID)
			// Truncate long company names
			maxCompanyLen := m.columnWidths[ColCompany] - 2
			if maxCompanyLen > 0 && len(company) > maxCompanyLen {
				company = company[:maxCompanyLen-3] + "..."
			}
		}

		// AD Types summary
		adTypes := device.FormatADTypesSummary(m.columnWidths[ColADTypes] - 2)

		avgStr := fmt.Sprintf("%.1f", device.RSSIAverage)
		countStr := fmt.Sprintf("%d", device.AdvCount)

		intervalStr := "-"
		if device.AdvInterval > 0 {
			intervalStr = fmt.Sprintf("%dms", device.AdvInterval.Milliseconds())
		}

		rows[i] = table.Row{
			name,
			company,
			adTypes,
			fmt.Sprintf("%d", device.RSSICurrent),
			avgStr,
			countStr,
			intervalStr,
		}
	}
	m.table.SetRows(rows)
}

func (m DeviceListModel) matchesFilter(d ble.Device) bool {
	if m.filter.Config.NameContains != "" {
		name := strings.ToLower(d.GetDisplayName())
		if !strings.Contains(name, strings.ToLower(m.filter.Config.NameContains)) {
			return false
		}
	}
	if m.filter.Config.MinRSSI != nil && d.RSSICurrent < *m.filter.Config.MinRSSI {
		return false
	}
	return true
}

func (m DeviceListModel) compareDevices(a, b ble.Device) bool {
	cmp := m.compareByColumn(a, b, m.sortColumn)
	if cmp == 0 && m.sortColumn != ColAvg {
		// Secondary sort by average RSSI (higher first) to reduce jumping
		cmp = compareFloat(b.RSSIAverage, a.RSSIAverage)
	}
	if m.sortAscending {
		return cmp > 0
	}
	return cmp < 0
}

// compareByColumn returns -1 if a < b, 0 if a == b, 1 if a > b for the given column
func (m DeviceListModel) compareByColumn(a, b ble.Device, col int) int {
	switch col {
	case ColName:
		return strings.Compare(strings.ToLower(a.GetDisplayName()), strings.ToLower(b.GetDisplayName()))
	case ColCompany:
		aCompany := ""
		bCompany := ""
		if a.ManufacturerID != nil {
			aCompany = ble.GetManufacturerName(*a.ManufacturerID)
		}
		if b.ManufacturerID != nil {
			bCompany = ble.GetManufacturerName(*b.ManufacturerID)
		}
		return strings.Compare(strings.ToLower(aCompany), strings.ToLower(bCompany))
	case ColADTypes:
		return compareInt(len(b.GetADTypes()), len(a.GetADTypes())) // More AD types first
	case ColRSSI:
		return compareInt(int(b.RSSICurrent), int(a.RSSICurrent)) // Higher RSSI first
	case ColAvg:
		return compareFloat(b.RSSIAverage, a.RSSIAverage) // Higher avg first
	case ColCount:
		return compareInt(int(b.AdvCount), int(a.AdvCount)) // Higher count first
	case ColInterval:
		return compareInt(int(a.AdvInterval), int(b.AdvInterval)) // Lower interval first
	}
	return 0
}

func compareInt(a, b int) int {
	if a < b {
		return -1
	}
	if a > b {
		return 1
	}
	return 0
}

func compareFloat(a, b float64) int {
	if a < b {
		return -1
	}
	if a > b {
		return 1
	}
	return 0
}

// SetDevices updates the device list
func (m *DeviceListModel) SetDevices(devices []ble.Device) {
	m.devices = devices
	m.applyFilterAndSort()
}

// SelectedDevice returns the currently selected device
func (m DeviceListModel) SelectedDevice() (ble.Device, bool) {
	idx := m.table.Cursor()
	if idx >= 0 && idx < len(m.filtered) {
		return m.filtered[idx], true
	}
	return ble.Device{}, false
}

// IsFilterActive returns true if filter input is focused
func (m DeviceListModel) IsFilterActive() bool {
	return m.filter.Mode != FilterModeNone
}

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}
