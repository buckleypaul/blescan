package views

import (
	"strconv"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/buckleypaul/blescan/internal/stats"
	"github.com/buckleypaul/blescan/internal/ui/styles"
)

// FilterMode represents the current filter input mode
type FilterMode int

const (
	FilterModeNone FilterMode = iota
	FilterModeName
	FilterModeRSSI
	FilterModeColumns
)

// FilterModel handles filter input
type FilterModel struct {
	Mode               FilterMode
	Config             stats.FilterConfig
	textInput          textinput.Model
	columnSelectorIdx  int
	tempEnabledColumns []string // Temporary storage during column selection
}

// NewFilterModel creates a new filter model
func NewFilterModel() FilterModel {
	ti := textinput.New()
	ti.Placeholder = ""
	ti.CharLimit = 50
	ti.Width = 30

	return FilterModel{
		Mode:      FilterModeNone,
		textInput: ti,
	}
}

// Init initializes the filter model
func (m FilterModel) Init() tea.Cmd {
	return nil
}

// Update handles filter input updates
func (m FilterModel) Update(msg tea.Msg) (FilterModel, tea.Cmd) {
	if m.Mode == FilterModeNone {
		return m, nil
	}

	// Handle column selection mode separately
	if m.Mode == FilterModeColumns {
		switch msg := msg.(type) {
		case tea.KeyMsg:
			switch msg.String() {
			case "up", "k":
				m.columnSelectorIdx--
				if m.columnSelectorIdx < 0 {
					m.columnSelectorIdx = len(ColumnRegistry) - 1
				}
			case "down", "j":
				m.columnSelectorIdx++
				if m.columnSelectorIdx >= len(ColumnRegistry) {
					m.columnSelectorIdx = 0
				}
			case " ":
				// Toggle selected column
				selectedColID := ColumnRegistry[m.columnSelectorIdx].ID
				m.toggleColumn(selectedColID)
			case "enter":
				// Apply changes and exit
				m.Mode = FilterModeNone
				return m, nil
			case "esc":
				// Cancel changes and exit
				m.tempEnabledColumns = nil
				m.Mode = FilterModeNone
				return m, nil
			}
		}
		return m, nil
	}

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "enter":
			m.applyFilter()
			m.Mode = FilterModeNone
			m.textInput.Blur()
			return m, nil
		case "esc":
			m.Mode = FilterModeNone
			m.textInput.Blur()
			return m, nil
		}
	}

	var cmd tea.Cmd
	m.textInput, cmd = m.textInput.Update(msg)
	return m, cmd
}

// View renders the filter input
func (m FilterModel) View() string {
	if m.Mode == FilterModeNone {
		return ""
	}

	if m.Mode == FilterModeColumns {
		return "" // Column selector has its own view in devicelist.go
	}

	var label string
	switch m.Mode {
	case FilterModeName:
		label = "Filter by name: "
	case FilterModeRSSI:
		label = "Min RSSI (dBm): "
	}

	return styles.FilterLabelStyle.Render(label) + m.textInput.View()
}

// SetMode sets the filter mode and prepares the input
func (m *FilterModel) SetMode(mode FilterMode) tea.Cmd {
	m.Mode = mode

	switch mode {
	case FilterModeName:
		m.textInput.Placeholder = "device name..."
		m.textInput.SetValue(m.Config.NameContains)
	case FilterModeRSSI:
		m.textInput.Placeholder = "-70"
		if m.Config.MinRSSI != nil {
			m.textInput.SetValue(strconv.Itoa(int(*m.Config.MinRSSI)))
		} else {
			m.textInput.SetValue("")
		}
	}

	m.textInput.Focus()
	return textinput.Blink
}

func (m *FilterModel) applyFilter() {
	value := m.textInput.Value()

	switch m.Mode {
	case FilterModeName:
		m.Config.NameContains = value
	case FilterModeRSSI:
		if value == "" {
			m.Config.MinRSSI = nil
		} else {
			if rssi, err := strconv.Atoi(value); err == nil {
				r := int16(rssi)
				m.Config.MinRSSI = &r
			}
		}
	}
}

// ClearFilters clears all filter criteria
func (m *FilterModel) ClearFilters() {
	m.Config = stats.FilterConfig{}
	m.textInput.SetValue("")
}

// IsFiltering returns true if any filter is active
func (m FilterModel) IsFiltering() bool {
	return m.Config.NameContains != "" || m.Config.MinRSSI != nil
}

// FilterSummary returns a string describing active filters
func (m FilterModel) FilterSummary() string {
	if !m.IsFiltering() {
		return ""
	}

	var parts []string
	if m.Config.NameContains != "" {
		parts = append(parts, "name:"+m.Config.NameContains)
	}
	if m.Config.MinRSSI != nil {
		parts = append(parts, "rssi>="+strconv.Itoa(int(*m.Config.MinRSSI)))
	}

	result := "Filters: "
	for i, p := range parts {
		if i > 0 {
			result += ", "
		}
		result += p
	}
	return result
}

// toggleColumn toggles a column in the temporary enabled columns list
func (m *FilterModel) toggleColumn(colID string) {
	// Find and remove if present
	for i, id := range m.tempEnabledColumns {
		if id == colID {
			// Remove from slice
			m.tempEnabledColumns = append(m.tempEnabledColumns[:i], m.tempEnabledColumns[i+1:]...)
			return
		}
	}
	// Not found, add it
	m.tempEnabledColumns = append(m.tempEnabledColumns, colID)
}

// isColumnEnabled checks if a column is enabled in the temp list
func (m FilterModel) isColumnEnabled(colID string) bool {
	for _, id := range m.tempEnabledColumns {
		if id == colID {
			return true
		}
	}
	return false
}
