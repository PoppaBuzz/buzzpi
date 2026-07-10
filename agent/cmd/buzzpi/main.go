// Command buzzpi is the BuzzPi CLI client.
//
// The CLI discovers devices on the LAN, queries device info/stats,
// and manages pairings. It is the primary development tool and
// secondary client after the Android app.
//
// Usage:
//
//	buzzpi discover             # scan for devices on LAN
//	buzzpi device info <id>     # get device information
//	buzzpi device stats <id>    # get device statistics
//	buzzpi pair <id>            # pair with a device
//	buzzpi session list <id>    # list active sessions
package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"time"

	"github.com/buzzpi/agent/internal/version"
	"github.com/buzzpi/agent/pkg/bpp"
)

func main() {
	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, `BuzzPi CLI — Manage your BuzzPi devices.

Usage:
  buzzpi <command> [arguments]

Commands:
  discover          Scan for BuzzPi devices on the local network
  device            Get device information or statistics
  pair              Pair with a device
  unpair            Unpair from a device
  session           Manage sessions
  plugin            Manage plugins
  version           Show version information
  help              Show help for a command

Global Flags:
  --version         Show version and exit
  --help            Show help

Use "buzzpi help <command>" for more information about a command.
`)
	}

	showVersion := flag.Bool("version", false, "show version and exit")
	flag.Parse()

	if *showVersion {
		fmt.Println(version.Info())
		os.Exit(0)
	}

	args := flag.Args()
	if len(args) == 0 {
		flag.Usage()
		os.Exit(1)
	}

	ctx := context.Background()

	switch args[0] {
	case "discover":
		cmdDiscover(ctx, args[1:])
	case "device":
		cmdDevice(ctx, args[1:])
	case "pair":
		cmdPair(ctx, args[1:])
	case "unpair":
		cmdUnpair(ctx, args[1:])
	case "session":
		cmdSession(ctx, args[1:])
	case "plugin":
		cmdPlugin(ctx, args[1:])
	case "version":
		fmt.Println(version.Info())
	case "help":
		cmdHelp(args[1:])
	default:
		fmt.Fprintf(os.Stderr, "Error: unknown command: %s\n", args[0])
		flag.Usage()
		os.Exit(1)
	}
}

// ── discover ────────────────────────────────────────────────────────────────

func cmdDiscover(ctx context.Context, args []string) {
	timeout := 5 * time.Second
	if len(args) > 0 {
		if d, err := time.ParseDuration(args[0]); err == nil {
			timeout = d
		}
	}

	fmt.Printf("Scanning for BuzzPi devices (timeout: %s)...\n", timeout)

	browser := bpp.NewBrowser()
	devices, err := browser.Discover(ctx, timeout)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: discovery failed: %v\n", err)
		os.Exit(1)
	}

	if len(devices) == 0 {
		fmt.Println("No BuzzPi devices found on the local network.")
		return
	}

	fmt.Println()
	fmt.Printf("%-24s %-24s %-16s %-6s %s\n", "DEVICE ID", "NAME", "ADDRESS", "PORT", "PLATFORM")
	fmt.Println("──────────────────────────────────────────────────────────────────────────────────────────────")
	for _, d := range devices {
		fmt.Printf("%-24s %-24s %-16s %-6d %s\n",
			d.DeviceID, d.FriendlyName, d.Addr, d.Port, d.Platform)
	}
	fmt.Printf("\nFound %d device(s)\n", len(devices))
}

// ── device ──────────────────────────────────────────────────────────────────

func cmdDevice(ctx context.Context, args []string) {
	if len(args) < 1 {
		fmt.Fprintln(os.Stderr, "Usage: buzzpi device info|stats <device_id>")
		os.Exit(1)
	}

	switch args[0] {
	case "info":
		printDeviceInfo(ctx, args[1:])
	case "stats":
		printDeviceStats(ctx, args[1:])
	default:
		fmt.Fprintf(os.Stderr, "Error: unknown subcommand: %s\n", args[0])
		fmt.Fprintln(os.Stderr, "Usage: buzzpi device info|stats <device_id>")
		os.Exit(1)
	}
}

func printDeviceInfo(ctx context.Context, args []string) {
	if len(args) < 1 {
		fmt.Fprintln(os.Stderr, "Usage: buzzpi device info <device_id>")
		os.Exit(1)
	}

	deviceID := args[0]

	// Try to discover first.
	browser := bpp.NewBrowser()
	devices, err := browser.Discover(ctx, 3*time.Second)
	if err == nil {
		for _, d := range devices {
			if d.DeviceID == deviceID {
				fmt.Printf("Device ID:       %s\n", d.DeviceID)
				fmt.Printf("Name:            %s\n", d.FriendlyName)
				fmt.Printf("Address:         %s\n", d.Addr)
				fmt.Printf("Port:            %d\n", d.Port)
				fmt.Printf("Platform:        %s\n", d.Platform)
				fmt.Printf("Version:         %s\n", d.Version)
				if len(d.Capabilities) > 0 {
					fmt.Printf("Capabilities:    %v\n", d.Capabilities)
				}
				return
			}
		}
	}

	fmt.Fprintf(os.Stderr, "Error: device %q not found on local network\n", deviceID)
	fmt.Fprintln(os.Stderr, "Run 'buzzpi discover' to find available devices")
	os.Exit(1)
}

func printDeviceStats(ctx context.Context, args []string) {
	if len(args) < 1 {
		fmt.Fprintln(os.Stderr, "Usage: buzzpi device stats <device_id>")
		os.Exit(1)
	}

	deviceID := args[0]

	browser := bpp.NewBrowser()
	devices, err := browser.Discover(ctx, 3*time.Second)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: discovery failed: %v\n", err)
		os.Exit(1)
	}

	for _, d := range devices {
		if d.DeviceID == deviceID {
			fmt.Printf("Device ID:       %s\n", d.DeviceID)
			fmt.Printf("Name:            %s\n", d.FriendlyName)
			fmt.Printf("Address:         %s\n", d.Addr)
			fmt.Printf("Port:            %d\n", d.Port)
			fmt.Printf("Platform:        %s\n", d.Platform)
			fmt.Printf("Version:         %s\n", d.Version)
			fmt.Println()
			fmt.Println("Connect to the device via WebSocket to retrieve live CPU/memory/disk stats.")
			fmt.Println("  buzzpi pair " + deviceID)
			return
		}
	}

	fmt.Fprintf(os.Stderr, "Error: device %q not found on local network\n", deviceID)
	fmt.Fprintln(os.Stderr, "Run 'buzzpi discover' to find available devices")
	os.Exit(1)
}

// ── pair / unpair ───────────────────────────────────────────────────────────

func cmdPair(ctx context.Context, args []string) {
	if len(args) < 1 {
		fmt.Fprintln(os.Stderr, "Usage: buzzpi pair <device_id>")
		os.Exit(1)
	}

	deviceID := args[0]

	browser := bpp.NewBrowser()
	devices, err := browser.Discover(ctx, 3*time.Second)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: discovery failed: %v\n", err)
		os.Exit(1)
	}

	for _, d := range devices {
		if d.DeviceID == deviceID {
			fmt.Printf("Initiating pairing with:\n")
			fmt.Printf("  Device ID:    %s\n", d.DeviceID)
			fmt.Printf("  Name:         %s\n", d.FriendlyName)
			fmt.Printf("  Address:      %s\n", d.Addr)
			fmt.Printf("  Port:         %d\n", d.Port)
			fmt.Printf("  Platform:     %s\n", d.Platform)
			fmt.Println()
			fmt.Printf("Connect to %s:%d via WebSocket and complete the BPP handshake.\n", d.Addr, d.Port)
			fmt.Println("The protocol supports PIN-based authentication (see bpp.Handshake).")
			return
		}
	}

	fmt.Fprintf(os.Stderr, "Error: device %q not found on local network\n", deviceID)
	fmt.Fprintln(os.Stderr, "Run 'buzzpi discover' to find available devices")
	os.Exit(1)
}

func cmdUnpair(ctx context.Context, args []string) {
	if len(args) < 1 {
		fmt.Fprintln(os.Stderr, "Usage: buzzpi unpair <device_id>")
		os.Exit(1)
	}

	deviceID := args[0]

	browser := bpp.NewBrowser()
	devices, err := browser.Discover(ctx, 3*time.Second)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: discovery failed: %v\n", err)
		os.Exit(1)
	}

	for _, d := range devices {
		if d.DeviceID == deviceID {
			fmt.Printf("Unpairing from %s (%s)...\n", d.FriendlyName, deviceID)
			fmt.Println("Send unpair request to the device to remove the pairing.")
			return
		}
	}

	fmt.Fprintf(os.Stderr, "Error: device %q not found on local network\n", deviceID)
	fmt.Fprintln(os.Stderr, "Run 'buzzpi discover' to find available devices")
	os.Exit(1)
}

// ── session ─────────────────────────────────────────────────────────────────

func cmdSession(ctx context.Context, args []string) {
	if len(args) < 1 {
		fmt.Fprintln(os.Stderr, "Usage: buzzpi session list|revoke <device_id>")
		os.Exit(1)
	}
	switch args[0] {
	case "list":
		fmt.Println("Active sessions (not yet implemented)")
	case "revoke":
		if len(args) < 2 {
			fmt.Fprintln(os.Stderr, "Usage: buzzpi session revoke <device_id>")
			os.Exit(1)
		}
		fmt.Printf("Revoking sessions for: %s (not yet implemented)\n", args[1])
	default:
		fmt.Fprintf(os.Stderr, "Error: unknown subcommand: %s\n", args[0])
		os.Exit(1)
	}
}

// ── plugin ──────────────────────────────────────────────────────────────────

func cmdPlugin(ctx context.Context, args []string) {
	if len(args) < 1 {
		fmt.Fprintln(os.Stderr, "Usage: buzzpi plugin list|install|uninstall")
		os.Exit(1)
	}
	switch args[0] {
	case "list":
		fmt.Println("Installed plugins (not yet implemented)")
	case "install":
		if len(args) < 2 {
			fmt.Fprintln(os.Stderr, "Usage: buzzpi plugin install <plugin_id>")
			os.Exit(1)
		}
		fmt.Printf("Installing plugin: %s (not yet implemented)\n", args[1])
	case "uninstall":
		if len(args) < 2 {
			fmt.Fprintln(os.Stderr, "Usage: buzzpi plugin uninstall <plugin_id>")
			os.Exit(1)
		}
		fmt.Printf("Uninstalling plugin: %s (not yet implemented)\n", args[1])
	default:
		fmt.Fprintf(os.Stderr, "Error: unknown subcommand: %s\n", args[0])
		os.Exit(1)
	}
}

// ── help ────────────────────────────────────────────────────────────────────

func cmdHelp(args []string) {
	if len(args) > 0 {
		switch args[0] {
		case "discover":
			fmt.Print(`Usage: buzzpi discover [timeout]

Scan the local network for BuzzPi devices using mDNS.

Arguments:
  timeout   Scan duration (e.g. "5s", "10s"). Default: 5s

Examples:
  buzzpi discover
  buzzpi discover 10s
`)
		case "device":
			fmt.Print(`Usage: buzzpi device info|stats <device_id>

Get device information or live statistics.

Subcommands:
  info <id>    Show device details (version, platform, capabilities)
  stats <id>   Show live device statistics (CPU, memory, etc.)

Examples:
  buzzpi device info dev_a1b2c3d4
`)
		case "pair":
			fmt.Print(`Usage: buzzpi pair <device_id>

Pair with a device. Initiates the pairing handshake.

Examples:
  buzzpi pair dev_a1b2c3d4
`)
		case "unpair":
			fmt.Print(`Usage: buzzpi unpair <device_id>

Remove a pairing with a device.

Examples:
  buzzpi unpair dev_a1b2c3d4
`)
		default:
			fmt.Fprintf(os.Stderr, "No help available for %q\n", args[0])
		}
		return
	}
	flag.Usage()
}
