package ble

import (
	"sync"

	"tinygo.org/x/bluetooth"
)

// Scanner handles BLE device scanning
type Scanner struct {
	adapter *bluetooth.Adapter
	devices map[string]*Device
	mu      sync.RWMutex

	// Channel for notifying UI of updates
	Updates chan struct{}

	scanning bool
	stopChan chan struct{}
}

// NewScanner creates a new BLE scanner
func NewScanner() *Scanner {
	return &Scanner{
		adapter:  bluetooth.DefaultAdapter,
		devices:  make(map[string]*Device),
		Updates:  make(chan struct{}, 100),
		stopChan: make(chan struct{}),
	}
}

// Start begins scanning for BLE devices
func (s *Scanner) Start() error {
	if err := s.adapter.Enable(); err != nil {
		return err
	}

	s.scanning = true

	go func() {
		_ = s.adapter.Scan(func(adapter *bluetooth.Adapter, result bluetooth.ScanResult) {
			select {
			case <-s.stopChan:
				return
			default:
			}

			s.handleAdvertisement(result)
		})
	}()

	return nil
}

// Stop stops the BLE scanning
func (s *Scanner) Stop() {
	if !s.scanning {
		return
	}
	s.scanning = false
	close(s.stopChan)
	_ = s.adapter.StopScan()
}

func (s *Scanner) handleAdvertisement(result bluetooth.ScanResult) {
	address := result.Address.String()

	adv := NewAdvertisement()
	adv.RSSI = result.RSSI
	adv.LocalName = result.LocalName()

	// Extract manufacturer data
	mfgData := result.ManufacturerData()
	if len(mfgData) > 0 {
		// ManufacturerData returns a slice of manufacturer data entries
		// Each entry has CompanyID and Data
		for _, entry := range mfgData {
			// Combine company ID and data
			data := make([]byte, 2+len(entry.Data))
			data[0] = byte(entry.CompanyID & 0xFF)
			data[1] = byte(entry.CompanyID >> 8)
			copy(data[2:], entry.Data)
			adv.ManufacturerData = data
			break // Use first entry
		}
	}

	// Extract service data (which also tells us about service UUIDs)
	serviceData := result.ServiceData()
	for _, sd := range serviceData {
		adv.ServiceData[sd.UUID.String()] = sd.Data
		adv.ServiceUUIDs = append(adv.ServiceUUIDs, sd.UUID.String())
	}

	// Check if device appears connectable based on available info
	// Devices with names or service data are often connectable
	adv.Connectable = result.LocalName() != "" || len(serviceData) > 0

	s.mu.Lock()
	device, exists := s.devices[address]
	if !exists {
		device = NewDevice(address)
		s.devices[address] = device
	}
	device.Update(adv)
	s.mu.Unlock()

	// Notify UI of update
	select {
	case s.Updates <- struct{}{}:
	default:
		// Channel full, skip notification
	}
}

// GetDevices returns a copy of all discovered devices
func (s *Scanner) GetDevices() []Device {
	s.mu.RLock()
	defer s.mu.RUnlock()

	devices := make([]Device, 0, len(s.devices))
	for _, d := range s.devices {
		devices = append(devices, d.Copy())
	}
	return devices
}

// GetDevice returns a copy of a specific device by address
func (s *Scanner) GetDevice(address string) (Device, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	if d, exists := s.devices[address]; exists {
		return d.Copy(), true
	}
	return Device{}, false
}

// DeviceCount returns the number of discovered devices
func (s *Scanner) DeviceCount() int {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return len(s.devices)
}

// Clear removes all discovered devices
func (s *Scanner) Clear() {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.devices = make(map[string]*Device)
}
