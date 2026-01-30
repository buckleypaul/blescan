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
	enabledColumns []string
	columnDefs     map[string]*ColumnDefinition
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

	// Build column lookup map
	columnDefs := BuildColumnLookup()

	// Default enabled columns
	enabledColumns := DefaultEnabledColumns()

	// Find index of "rssi" column for default sorting
	rssiIdx := -1
	for i, colID := range enabledColumns {
		if colID == "rssi" {
			rssiIdx = i
			break
		}
	}
	if rssiIdx == -1 {
		rssiIdx = 0 // Fallback if rssi isn't enabled
	}

	m := DeviceListModel{
		devices:        make([]ble.Device, 0),
		filtered:       make([]ble.Device, 0),
		table:          t,
		filter:         NewFilterModel(),
		sortColumn:     rssiIdx,
		sortAscending:  false,
		selectedColumn: rssiIdx,
		columnWidths:   make([]int, len(enabledColumns)),
		enabledColumns: enabledColumns,
		columnDefs:     columnDefs,
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
		wasInColumnMode := m.filter.Mode == FilterModeColumns
		m.filter, cmd = m.filter.Update(msg)

		// If we just exited column mode, apply the changes
		if wasInColumnMode && m.filter.Mode == FilterModeNone && m.filter.tempEnabledColumns != nil {
			m.ApplyColumnConfiguration()
		}

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
			if m.selectedColumn < len(m.enabledColumns)-1 {
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
		case "tab":
			// Start column configuration
			m.filter.tempEnabledColumns = append([]string(nil), m.enabledColumns...)
			return m, m.filter.SetMode(FilterModeColumns)
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

		// Calculate total percentage
		totalPct := 0
		for _, colID := range m.enabledColumns {
			totalPct += m.columnDefs[colID].WidthPct
		}

		// Distribute proportionally
		m.columnWidths = make([]int, len(m.enabledColumns))
		for i, colID := range m.enabledColumns {
			def := m.columnDefs[colID]
			pctWidth := availableWidth * def.WidthPct / totalPct

			// Enforce minimum width
			if pctWidth < def.MinWidth {
				pctWidth = def.MinWidth
			}

			m.columnWidths[i] = pctWidth
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

	// Filter bar or column selector
	filterBarStyle := lipgloss.NewStyle().
		Foreground(styles.SecondaryColor).
		Background(lipgloss.Color("236")).
		Padding(0, 2).
		Width(m.width)

	var filterContent string
	if m.filter.Mode == FilterModeColumns {
		// Show column selector
		filterContent = ""
	} else if m.filter.Mode != FilterModeNone {
		filterContent = m.filter.View()
	} else if m.filter.IsFiltering() {
		filterContent = m.filter.FilterSummary()
	} else {
		filterContent = "Use ←/→ to select column, 's' to sort"
	}
	b.WriteString(filterBarStyle.Render(filterContent))
	b.WriteString("\n")

	// Show column selector modal if in column mode
	if m.filter.Mode == FilterModeColumns {
		b.WriteString(m.renderColumnSelector())
	} else {
		// Table with dynamic header
		tableStyle := lipgloss.NewStyle().
			BorderStyle(lipgloss.RoundedBorder()).
			BorderForeground(styles.MutedColor).
			Width(m.width - 2)

		b.WriteString(tableStyle.Render(m.table.View()))
		b.WriteString("\n")
	}

	// Help bar
	helpStyle := lipgloss.NewStyle().
		Foreground(styles.MutedColor).
		Background(lipgloss.Color("235")).
		Padding(0, 2).
		Width(m.width)

	help := "↑/↓ Row • ←/→ Column • s Sort • Enter View • / Name • r RSSI • Tab Columns • c Clear • q Quit"
	b.WriteString(helpStyle.Render(help))

	return b.String()
}

func (m *DeviceListModel) updateColumns() {
	columns := make([]table.Column, len(m.enabledColumns))

	for i, colID := range m.enabledColumns {
		def := m.columnDefs[colID]
		displayTitle := def.Title

		// Add selection marker
		if i == m.selectedColumn {
			displayTitle = ">" + displayTitle + "<"
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
		row := make(table.Row, len(m.enabledColumns))

		for j, colID := range m.enabledColumns {
			def := m.columnDefs[colID]

			// Extract data using formatter
			value := def.Formatter(&device)

			// Truncate to fit column width
			maxLen := m.columnWidths[j] - 2
			if maxLen > 3 && len(value) > maxLen {
				value = value[:maxLen-3] + "..."
			}

			row[j] = value
		}

		rows[i] = row
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
	if m.filter.Config.MinRSSI != nil && d.RSSIAverage < float64(*m.filter.Config.MinRSSI) {
		return false
	}
	return true
}

func (m DeviceListModel) compareDevices(a, b ble.Device) bool {
	cmp := m.compareByColumn(a, b, m.sortColumn)

	// Check if sorting by rssi column
	sortByRSSI := false
	if m.sortColumn >= 0 && m.sortColumn < len(m.enabledColumns) {
		sortByRSSI = m.enabledColumns[m.sortColumn] == "rssi"
	}

	if cmp == 0 && !sortByRSSI {
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
	if col < 0 || col >= len(m.enabledColumns) {
		return 0
	}

	colID := m.enabledColumns[col]
	def := m.columnDefs[colID]

	// For numeric columns, provide better sorting
	switch colID {
	case "rssi":
		return compareFloat(b.RSSIAverage, a.RSSIAverage) // Higher RSSI first
	case "count":
		return compareInt(int(b.AdvCount), int(a.AdvCount)) // Higher count first
	case "interval":
		return compareInt(int(a.AdvInterval), int(b.AdvInterval)) // Lower interval first
	case "flags":
		aFlags := uint8(0)
		bFlags := uint8(0)
		if a.Flags != nil {
			aFlags = *a.Flags
		}
		if b.Flags != nil {
			bFlags = *b.Flags
		}
		return compareInt(int(aFlags), int(bFlags))
	case "name":
		return strings.Compare(strings.ToLower(a.GetDisplayName()), strings.ToLower(b.GetDisplayName()))
	case "service_uuids":
		return compareInt(len(b.ServiceUUIDs), len(a.ServiceUUIDs)) // More UUIDs first
	case "service_data":
		return compareInt(len(b.ServiceData), len(a.ServiceData)) // More data first
	case "appearance":
		aApp := uint16(0)
		bApp := uint16(0)
		if a.Appearance != nil {
			aApp = *a.Appearance
		}
		if b.Appearance != nil {
			bApp = *b.Appearance
		}
		return compareInt(int(aApp), int(bApp))
	case "other_ad":
		return compareInt(len(b.ADTypes), len(a.ADTypes)) // More AD types first
	case "company":
		aCompany := ""
		bCompany := ""
		if a.ManufacturerID != nil {
			aCompany = ble.GetManufacturerName(*a.ManufacturerID)
		}
		if b.ManufacturerID != nil {
			bCompany = ble.GetManufacturerName(*b.ManufacturerID)
		}
		return strings.Compare(strings.ToLower(aCompany), strings.ToLower(bCompany))
	default:
		// Generic string comparison
		aVal := def.Formatter(&a)
		bVal := def.Formatter(&b)
		return strings.Compare(aVal, bVal)
	}
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

// renderColumnSelector renders the column configuration modal
func (m DeviceListModel) renderColumnSelector() string {
	var b strings.Builder

	selectorStyle := lipgloss.NewStyle().
		BorderStyle(lipgloss.RoundedBorder()).
		BorderForeground(styles.PrimaryColor).
		Padding(1, 2).
		Width(m.width - 4)

	title := lipgloss.NewStyle().
		Bold(true).
		Foreground(styles.PrimaryColor).
		Render("Column Configuration")

	b.WriteString(title + "\n\n")
	b.WriteString("Use ↑/↓ or j/k to navigate, Space to toggle, Enter to apply, Esc to cancel\n\n")

	// Group columns by category
	currentCategory := ColumnCategory(-1)
	for i, def := range ColumnRegistry {
		// Add category header if this is a new category
		if def.Category != currentCategory {
			if currentCategory != -1 {
				b.WriteString("\n") // Blank line between categories
			}
			currentCategory = def.Category

			categoryTitle := ""
			switch def.Category {
			case CategoryAdvertisement:
				categoryTitle = "Advertisement Data:"
			case CategoryMetadata:
				categoryTitle = "Metadata:"
			}

			categoryStyle := lipgloss.NewStyle().
				Bold(true).
				Foreground(styles.SecondaryColor).
				Render(categoryTitle)
			b.WriteString(categoryStyle + "\n")
		}

		enabled := m.filter.isColumnEnabled(def.ID)

		checkbox := "[ ]"
		if enabled {
			checkbox = "[✓]"
		}

		available := ""
		if !def.Available {
			available = " (not available)"
		}

		cursor := "  "
		if i == m.filter.columnSelectorIdx {
			cursor = "> "
		}

		line := fmt.Sprintf("%s%s %s%s", cursor, checkbox, def.Title, available)

		if i == m.filter.columnSelectorIdx {
			line = lipgloss.NewStyle().
				Foreground(styles.PrimaryColor).
				Bold(true).
				Render(line)
		}

		b.WriteString(line + "\n")
	}

	return selectorStyle.Render(b.String())
}

// ApplyColumnConfiguration applies the temporary column configuration
func (m *DeviceListModel) ApplyColumnConfiguration() {
	if m.filter.tempEnabledColumns != nil && len(m.filter.tempEnabledColumns) > 0 {
		m.enabledColumns = append([]string(nil), m.filter.tempEnabledColumns...)
		m.columnWidths = make([]int, len(m.enabledColumns))

		// Reset selected column if out of bounds
		if m.selectedColumn >= len(m.enabledColumns) {
			m.selectedColumn = len(m.enabledColumns) - 1
		}
		if m.selectedColumn < 0 {
			m.selectedColumn = 0
		}

		// Reset sort column if out of bounds
		if m.sortColumn >= len(m.enabledColumns) {
			m.sortColumn = 0
		}

		// Clear rows before updating columns to avoid index out of bounds
		m.table.SetRows([]table.Row{})

		m.updateTableSize()
		m.applyFilterAndSort()
	}
	m.filter.tempEnabledColumns = nil
}
