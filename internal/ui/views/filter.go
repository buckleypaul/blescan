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
)

// FilterModel handles filter input
type FilterModel struct {
	Mode      FilterMode
	Config    stats.FilterConfig
	textInput textinput.Model
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
