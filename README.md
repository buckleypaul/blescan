# blescan

A terminal UI application for scanning and analyzing BLE (Bluetooth Low Energy) device advertisements.

## Features

- Real-time BLE device scanning
- Device list with RSSI, advertisement count, and interval
- Detailed device view with manufacturer lookup
- Raw advertisement data stream
- Filter by device name or minimum RSSI
- Sortable device list
- Color-coded signal strength indicators

## Installation

### Homebrew (macOS/Linux)

```bash
brew tap buckleypaul/tap
brew install blescan
```

### From Source

```bash
go install github.com/buckleypaul/blescan/cmd/blescan@latest
```

### Using Make

```bash
git clone https://github.com/buckleypaul/blescan.git
cd blescan
make build
```

## Usage

```bash
blescan
```

### Keyboard Shortcuts

#### Device List View

| Key | Action |
|-----|--------|
| `Up/k` | Previous device |
| `Down/j` | Next device |
| `Enter` | View device details |
| `/` or `n` | Filter by name |
| `r` | Filter by minimum RSSI |
| `c` | Clear filters |
| `s` | Cycle sort column |
| `q` | Quit |

#### Device Detail View

| Key | Action |
|-----|--------|
| `Up/k` | Scroll up |
| `Down/j` | Scroll down |
| `Esc` | Back to list |
| `q` | Quit |

## Platform Requirements

### macOS

- Bluetooth must be enabled
- Terminal app needs Bluetooth permission (System Preferences > Privacy & Security > Bluetooth)

### Linux

- `bluez` package must be installed
- May require root access or `CAP_NET_ADMIN` capability
- Alternatively, add your user to the `bluetooth` group:
  ```bash
  sudo usermod -a -G bluetooth $USER
  ```

## License

MIT License - see [LICENSE](LICENSE) file.
