package main

import (
	"fmt"
	"os"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/buckleypaul/blescan/internal/ble"
	"github.com/buckleypaul/blescan/internal/ui"
)

var version = "dev"

func main() {
	// Check for version flag
	if len(os.Args) > 1 && (os.Args[1] == "--version" || os.Args[1] == "-v") {
		fmt.Printf("blescan version %s\n", version)
		os.Exit(0)
	}

	// Create scanner
	scanner := ble.NewScanner()

	// Start scanning
	if err := scanner.Start(); err != nil {
		fmt.Fprintf(os.Stderr, "Error starting BLE scanner: %v\n", err)
		fmt.Fprintln(os.Stderr, "")
		fmt.Fprintln(os.Stderr, "Troubleshooting tips:")
		fmt.Fprintln(os.Stderr, "  - macOS: Ensure Bluetooth is enabled and terminal has Bluetooth permission")
		fmt.Fprintln(os.Stderr, "  - Linux: Ensure bluez is installed and you have proper permissions")
		fmt.Fprintln(os.Stderr, "           Try running with sudo or adding your user to the bluetooth group")
		os.Exit(1)
	}
	defer scanner.Stop()

	// Create and run TUI
	model := ui.NewModel(scanner)
	p := tea.NewProgram(model, tea.WithAltScreen())

	if _, err := p.Run(); err != nil {
		fmt.Fprintf(os.Stderr, "Error running application: %v\n", err)
		os.Exit(1)
	}
}
