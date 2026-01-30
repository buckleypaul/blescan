package ble

import (
	"sync"
	"time"

	"tinygo.org/x/bluetooth"
)

// Scanner handles BLE device scanning
type Scanner struct {
	adapter *bluetooth.Adapter
	devices map[string]*Device
	mu      sync.RWMutex

	// Channel for notifying UI of updates
	Updates chan struct{}

	scanning    bool
	stopChan    chan struct{}
	cleanupTicker *time.Ticker
}

const (
	deviceTimeout    = 30 * time.Second
	cleanupInterval  = 5 * time.Second
)

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

	// Start BLE scanning
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

	// Start cleanup goroutine
	s.cleanupTicker = time.NewTicker(cleanupInterval)
	go s.cleanupStaleDevices()

	return nil
}

// Stop stops the BLE scanning
func (s *Scanner) Stop() {
	if !s.scanning {
		return
	}
	s.scanning = false
	close(s.stopChan)
	if s.cleanupTicker != nil {
		s.cleanupTicker.Stop()
	}
	_ = s.adapter.StopScan()
}

// cleanupStaleDevices runs periodically to remove devices not seen recently
func (s *Scanner) cleanupStaleDevices() {
	for {
		select {
		case <-s.stopChan:
			return
		case <-s.cleanupTicker.C:
			now := time.Now()
			s.mu.Lock()
			var removed bool
			for address, device := range s.devices {
				device.mu.RLock()
				lastSeen := device.LastSeen
				device.mu.RUnlock()

				if now.Sub(lastSeen) > deviceTimeout {
					delete(s.devices, address)
					removed = true
				}
			}
			s.mu.Unlock()

			// Notify UI if any devices were removed
			if removed {
				select {
				case s.Updates <- struct{}{}:
				default:
					// Channel full, skip notification
				}
			}
		}
	}
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

	// Infer AD types present from what we can detect
	var adTypes []uint8
	if adv.LocalName != "" {
		adTypes = append(adTypes, 0x09) // Complete Local Name (we can't distinguish from shortened)
	}
	if len(adv.ManufacturerData) > 0 {
		adTypes = append(adTypes, 0xFF) // Manufacturer Specific Data
	}
	if len(adv.ServiceUUIDs) > 0 {
		// Could be 0x02, 0x03, 0x06, or 0x07 depending on UUID length and completeness
		// For now, assume complete 16-bit service UUIDs
		adTypes = append(adTypes, 0x03)
	}
	if len(adv.ServiceData) > 0 {
		// Could be 0x16, 0x20, or 0x21 depending on UUID length
		// For now, assume 16-bit UUID service data
		adTypes = append(adTypes, 0x16)
	}
	adv.ADTypes = adTypes

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
