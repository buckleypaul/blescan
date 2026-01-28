package ui

import (
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/paulbuckley/blescan/internal/ble"
	"github.com/paulbuckley/blescan/internal/ui/views"
)

// ViewState represents the current view
type ViewState int

const (
	ViewDeviceList ViewState = iota
	ViewDeviceDetail
)

// Model is the main application model
type Model struct {
	scanner      *ble.Scanner
	viewState    ViewState
	deviceList   views.DeviceListModel
	deviceDetail views.DeviceDetailModel
	width        int
	height       int
	err          error
}

// tickMsg is sent periodically to refresh the UI
type tickMsg time.Time

// scanUpdateMsg is sent when new scan data is available
type scanUpdateMsg struct{}

// errMsg is sent when an error occurs
type errMsg struct{ err error }

// NewModel creates a new application model
func NewModel(scanner *ble.Scanner) Model {
	return Model{
		scanner:    scanner,
		viewState:  ViewDeviceList,
		deviceList: views.NewDeviceListModel(),
	}
}

// Init initializes the application
func (m Model) Init() tea.Cmd {
	return tea.Batch(
		m.tickCmd(),
		m.waitForScanUpdate(),
	)
}

func (m Model) tickCmd() tea.Cmd {
	return tea.Tick(500*time.Millisecond, func(t time.Time) tea.Msg {
		return tickMsg(t)
	})
}

func (m Model) waitForScanUpdate() tea.Cmd {
	return func() tea.Msg {
		<-m.scanner.Updates
		return scanUpdateMsg{}
	}
}

// Update handles messages
func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		// Global key handling
		switch msg.String() {
		case "ctrl+c", "q":
			// Don't quit if filter is active
			if m.viewState == ViewDeviceList && m.deviceList.IsFilterActive() {
				break
			}
			m.scanner.Stop()
			return m, tea.Quit
		case "esc":
			if m.viewState == ViewDeviceDetail {
				m.viewState = ViewDeviceList
				return m, nil
			}
		case "enter":
			if m.viewState == ViewDeviceList && !m.deviceList.IsFilterActive() {
				if device, ok := m.deviceList.SelectedDevice(); ok {
					m.deviceDetail = views.NewDeviceDetailModel(device)
					m.viewState = ViewDeviceDetail
					// Initialize detail view with current window size
					m.deviceDetail, _ = m.deviceDetail.Update(tea.WindowSizeMsg{
						Width:  m.width,
						Height: m.height,
					})
					return m, nil
				}
			}
		}

	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		// Pass size to current view
		switch m.viewState {
		case ViewDeviceList:
			m.deviceList, _ = m.deviceList.Update(msg)
		case ViewDeviceDetail:
			m.deviceDetail, _ = m.deviceDetail.Update(msg)
		}
		return m, nil

	case tickMsg:
		// Refresh device list periodically
		m.refreshDevices()
		return m, m.tickCmd()

	case scanUpdateMsg:
		// New scan data available
		m.refreshDevices()
		return m, m.waitForScanUpdate()

	case errMsg:
		m.err = msg.err
		return m, nil
	}

	// Route to current view
	var cmd tea.Cmd
	switch m.viewState {
	case ViewDeviceList:
		m.deviceList, cmd = m.deviceList.Update(msg)
	case ViewDeviceDetail:
		m.deviceDetail, cmd = m.deviceDetail.Update(msg)
		// Also update the device data
		if device, ok := m.scanner.GetDevice(m.deviceDetail.Device.Address); ok {
			m.deviceDetail.UpdateDevice(device)
		}
	}

	return m, cmd
}

func (m *Model) refreshDevices() {
	devices := m.scanner.GetDevices()
	m.deviceList.SetDevices(devices)

	// Update detail view if open
	if m.viewState == ViewDeviceDetail {
		if device, ok := m.scanner.GetDevice(m.deviceDetail.Device.Address); ok {
			m.deviceDetail.UpdateDevice(device)
		}
	}
}

// View renders the application
func (m Model) View() string {
	if m.err != nil {
		return "Error: " + m.err.Error() + "\n\nPress q to quit."
	}

	switch m.viewState {
	case ViewDeviceList:
		return m.deviceList.View()
	case ViewDeviceDetail:
		return m.deviceDetail.View()
	}

	return ""
}
