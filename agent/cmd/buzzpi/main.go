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
	"bufio"
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"os/signal"
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
  devices           List paired devices
  device            Get device information or statistics
  pair              Pair with a device
  unpair            Unpair from a device
  terminal          Open a terminal session with a device
  session           Manage sessions
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
	case "devices":
		cmdDevices(ctx, args[1:])
	case "device":
		cmdDevice(ctx, args[1:])
	case "pair":
		cmdPair(ctx, args[1:])
	case "unpair":
		cmdUnpair(ctx, args[1:])
	case "terminal":
		cmdTerminal(ctx, args[1:])
	case "session":
		cmdSession(ctx, args[1:])
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

// ── devices ─────────────────────────────────────────────────────────────────

func cmdDevices(ctx context.Context, args []string) {
	state, err := loadClientState()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}

	devices := state.listDevices()
	if len(devices) == 0 {
		fmt.Println("No paired devices.")
		fmt.Println()
		fmt.Println("Pair with a device:")
		fmt.Println("  buzzpi discover          # find devices on your network")
		fmt.Println("  buzzpi pair <device_id>  # pair with a device")
		return
	}

	fmt.Println()
	fmt.Printf("%-24s %-24s %-16s %-6s %s\n", "DEVICE ID", "NAME", "ADDRESS", "PORT", "PAIRED")
	fmt.Println("──────────────────────────────────────────────────────────────────────────────────────────────")
	for _, d := range devices {
		pairedAt, _ := time.Parse(time.RFC3339, d.PairedAt)
		fmt.Printf("%-24s %-24s %-16s %-6d %s\n",
			d.DeviceID, d.DeviceName, d.Address, d.Port, pairedAt.Format("2006-01-02"))
	}
	fmt.Printf("\n%d paired device(s)\n", len(devices))
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

// ── terminal ────────────────────────────────────────────────────────────────

func cmdTerminal(ctx context.Context, args []string) {
	if len(args) < 1 {
		fmt.Fprintln(os.Stderr, "Usage: buzzpi terminal <device_id>")
		os.Exit(1)
	}

	deviceID := args[0]

	devInfo := discoverDevice(ctx, deviceID)
	if devInfo == nil {
		os.Exit(1)
	}

	state, err := loadClientState()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}

	dev, ok := state.getDevice(deviceID)
	if !ok {
		fmt.Fprintf(os.Stderr, "Error: device %q is not paired\n", deviceID)
		fmt.Fprintln(os.Stderr, "Run 'buzzpi pair <device_id>' first")
		os.Exit(1)
	}

	fmt.Printf("Connecting to %s (%s)...\n", dev.DeviceName, deviceID)

	c := bpp.NewClient(deviceID, devInfo.Addr.String(), devInfo.Port)
	if err := c.Connect(ctx); err != nil {
		fmt.Fprintf(os.Stderr, "Error: connection failed: %v\n", err)
		os.Exit(1)
	}
	defer c.Close()

	fmt.Print("Authenticating... ")
	accept, err := c.Handshake(ctx, dev.SessionToken)
	if err != nil {
		if authErr, ok := err.(*bpp.HandshakeAuthError); ok {
			fmt.Printf("\nSession expired for device %q.\n", deviceID)
			fmt.Printf("Please re-pair: buzzpi pair %s (PIN: %s)\n", deviceID, authErr.PIN)
		} else {
			fmt.Printf("\nHandshake failed: %v\n", err)
		}
		os.Exit(1)
	}
	fmt.Println("OK")

	sesCtx, cancel := context.WithCancel(ctx)
	defer cancel()

	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, os.Interrupt)

	fmt.Print("Opening terminal session... ")
	openResp, err := c.Call(sesCtx, "terminal.open", map[string]string{})
	if err != nil {
		fmt.Fprintf(os.Stderr, "\nError: terminal.open failed: %v\n", err)
		os.Exit(1)
	}

	var openResult struct {
		SessionID string `json:"session_id"`
	}
	if err := json.Unmarshal(openResp.Result, &openResult); err != nil {
		fmt.Fprintf(os.Stderr, "\nError: parse terminal.open response: %v\n", err)
		os.Exit(1)
	}
	sessionID := openResult.SessionID
	fmt.Printf("OK (session: %s)\n", sessionID)
	fmt.Println()
	fmt.Printf("Terminal connected to %s. Type commands, Ctrl+C to exit.\n", accept.DeviceName)
	fmt.Println()

	stdinCh := make(chan string, 10)
	go func() {
		scanner := bufio.NewScanner(os.Stdin)
		for scanner.Scan() {
			select {
			case stdinCh <- scanner.Text():
			case <-sesCtx.Done():
				return
			}
		}
	}()

	done := make(chan struct{})
	go func() {
		defer close(done)
		for {
			select {
			case <-sigCh:
				fmt.Println("\nClosing terminal...")
				cancel()
				return
			case <-sesCtx.Done():
				return
			case line := <-stdinCh:
				env, err := c.Call(sesCtx, "terminal.input", map[string]interface{}{
					"session_id": sessionID,
					"data":       []byte(line + "\n"),
				})
				if err != nil {
					select {
					case <-sesCtx.Done():
						return
					default:
						fmt.Fprintf(os.Stderr, "\nError: %v\n", err)
					}
					return
				}

				var result struct {
					Output []byte `json:"output"`
				}
				if env.Result != nil {
					if err := json.Unmarshal(env.Result, &result); err != nil {
						fmt.Fprintf(os.Stderr, "\nParse error: %v\n", err)
						continue
					}
				}

				if len(result.Output) > 0 {
					os.Stdout.Write(result.Output)
				}
			}
		}
	}()

	<-done

	// Clean up the terminal session
	fmt.Print("Closing terminal session... ")
	closeCtx, closeCancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer closeCancel()
	if _, err := c.Call(closeCtx, "terminal.close", map[string]string{
		"session_id": sessionID,
	}); err != nil {
		fmt.Fprintf(os.Stderr, "\nWarning: terminal.close: %v\n", err)
	} else {
		fmt.Println("OK")
	}
}

// ── pair / unpair ───────────────────────────────────────────────────────────

func cmdPair(ctx context.Context, args []string) {
	if len(args) < 1 {
		fmt.Fprintln(os.Stderr, "Usage: buzzpi pair <device_id>")
		os.Exit(1)
	}

	deviceID := args[0]

	devInfo := discoverDevice(ctx, deviceID)
	if devInfo == nil {
		os.Exit(1)
	}

	state, err := loadClientState()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("Initiating pairing with:\n")
	fmt.Printf("  Device ID:    %s\n", devInfo.DeviceID)
	fmt.Printf("  Name:         %s\n", devInfo.FriendlyName)
	fmt.Printf("  Address:      %s\n", devInfo.Addr.String())
	fmt.Printf("  Port:         %d\n", devInfo.Port)
	fmt.Println()

	c := bpp.NewClient(deviceID, devInfo.Addr.String(), devInfo.Port)
	if err := c.Connect(ctx); err != nil {
		fmt.Fprintf(os.Stderr, "Error: connection failed: %v\n", err)
		os.Exit(1)
	}
	defer c.Close()

	initParams := map[string]string{"device_id": deviceID}
	initResp, err := c.Call(ctx, "pair.initiate", initParams)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: initiate failed: %v\n", err)
		os.Exit(1)
	}

	var initResult struct {
		PIN       string `json:"pin"`
		SessionID string `json:"session_id"`
		DeviceID  string `json:"device_id"`
		ExpiresAt int64  `json:"expires_at"`
	}
	if err := json.Unmarshal(initResp.Result, &initResult); err != nil {
		fmt.Fprintf(os.Stderr, "Error: parse initiate response: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("Pairing PIN: %s\n", initResult.PIN)
	fmt.Printf("This PIN expires in 2 minutes.\n")
	fmt.Println()

	fmt.Print("Press Enter to confirm pairing with this PIN...")
	fmt.Scanln()

	verifyParams := map[string]interface{}{
		"session_id":        initResult.SessionID,
		"pin":               initResult.PIN,
		"client_public_key": "",
		"client_name":       "buzzpi-cli",
	}
	verifyResp, err := c.Call(ctx, "pair.verify", verifyParams)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: verify failed: %v\n", err)
		os.Exit(1)
	}

	var verifyResult struct {
		SessionToken string `json:"session_token"`
		ExpiresAt    int64  `json:"expires_at"`
		DeviceID     string `json:"device_id"`
	}
	if err := json.Unmarshal(verifyResp.Result, &verifyResult); err != nil {
		fmt.Fprintf(os.Stderr, "Error: parse verify response: %v\n", err)
		os.Exit(1)
	}

	state.addDevice(pairedDevice{
		DeviceID:     initResult.DeviceID,
		DeviceName:   devInfo.FriendlyName,
		Address:      devInfo.Addr.String(),
		Port:         devInfo.Port,
		SessionToken: verifyResult.SessionToken,
		PairedAt:     time.Now().Format(time.RFC3339),
	})
	if err := state.save(); err != nil {
		fmt.Fprintf(os.Stderr, "Warning: failed to save pairing state: %v\n", err)
	}

	fmt.Println("Pairing successful!")
	fmt.Printf("  Session token: %s\n", verifyResult.SessionToken)
	fmt.Printf("  Expires at:    %s\n", time.Unix(verifyResult.ExpiresAt, 0).Format(time.RFC3339))
}

func cmdUnpair(ctx context.Context, args []string) {
	if len(args) < 1 {
		fmt.Fprintln(os.Stderr, "Usage: buzzpi unpair <device_id>")
		os.Exit(1)
	}

	deviceID := args[0]

	state, err := loadClientState()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}

	dev, ok := state.getDevice(deviceID)
	if !ok {
		fmt.Fprintf(os.Stderr, "Error: device %q not found in local state\n", deviceID)
		fmt.Fprintln(os.Stderr, "Run 'buzzpi pair <device_id>' first")
		os.Exit(1)
	}

	c := bpp.NewClient(deviceID, dev.Address, dev.Port)
	if err := c.Connect(ctx); err != nil {
		fmt.Fprintf(os.Stderr, "Warning: connection failed: %v\n", err)
		fmt.Println("Removing pairing from local state only.")
	} else {
		unpairParams := map[string]string{
			"session_token": dev.SessionToken,
			"device_id":     deviceID,
		}
		c.Call(ctx, "pair.unpair", unpairParams)
		c.Close()
	}

	state.removeDevice(deviceID)
	if err := state.save(); err != nil {
		fmt.Fprintf(os.Stderr, "Warning: failed to save state: %v\n", err)
	}

	fmt.Printf("Unpaired from %s (%s)\n", dev.DeviceName, deviceID)
}

func discoverDevice(ctx context.Context, deviceID string) *bpp.DiscoveredDevice {
	browser := bpp.NewBrowser()
	devices, err := browser.Discover(ctx, 3*time.Second)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: discovery failed: %v\n", err)
		return nil
	}

	for _, d := range devices {
		if d.DeviceID == deviceID {
			return &d
		}
	}

	fmt.Fprintf(os.Stderr, "Error: device %q not found on local network\n", deviceID)
	fmt.Fprintln(os.Stderr, "Run 'buzzpi discover' to find available devices")
	return nil
}

// ── session ─────────────────────────────────────────────────────────────────

func cmdSession(ctx context.Context, args []string) {
	if len(args) < 2 {
		fmt.Fprintln(os.Stderr, "Usage: buzzpi session list <device_id>")
		os.Exit(1)
	}

	switch args[0] {
	case "list":
		cmdSessionList(ctx, args[1])
	default:
		fmt.Fprintf(os.Stderr, "Error: unknown subcommand: %s\n", args[0])
		fmt.Fprintln(os.Stderr, "Usage: buzzpi session list <device_id>")
		os.Exit(1)
	}
}

func cmdSessionList(ctx context.Context, deviceID string) {
	state, err := loadClientState()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}

	dev, ok := state.getDevice(deviceID)
	if !ok {
		fmt.Fprintf(os.Stderr, "Error: device %q is not paired\n", deviceID)
		fmt.Fprintln(os.Stderr, "Run 'buzzpi pair <device_id>' first")
		os.Exit(1)
	}

	c := bpp.NewClient(deviceID, dev.Address, dev.Port)
	if err := c.Connect(ctx); err != nil {
		fmt.Fprintf(os.Stderr, "Error: connection failed: %v\n", err)
		os.Exit(1)
	}
	defer c.Close()

	fmt.Print("Authenticating... ")
	if _, err := c.Handshake(ctx, dev.SessionToken); err != nil {
		fmt.Printf("\nAuthentication failed: %v\n", err)
		fmt.Println("Session may have expired. Re-pair with: buzzpi pair " + deviceID)
		os.Exit(1)
	}
	fmt.Println("OK")

	fmt.Print("Querying sessions... ")
	resp, err := c.Call(ctx, "pair.status", map[string]string{
		"device_id": deviceID,
	})
	if err != nil {
		fmt.Printf("\nError: %v\n", err)
		os.Exit(1)
	}

	var status struct {
		Sessions []struct {
			ClientID  string `json:"client_id"`
			ExpiresAt int64  `json:"expires_at"`
			CreatedAt int64  `json:"created_at"`
		} `json:"sessions"`
	}
	if err := json.Unmarshal(resp.Result, &status); err != nil {
		fmt.Printf("\nError: %v\n", err)
		os.Exit(1)
	}
	fmt.Println("OK")
	fmt.Println()

	if len(status.Sessions) == 0 {
		fmt.Println("No active sessions.")
		return
	}

	fmt.Printf("%-24s %-20s %-20s\n", "CLIENT", "CREATED", "EXPIRES")
	fmt.Println("──────────────────────────────────────────────────────────────")
	for _, s := range status.Sessions {
		created := time.Unix(s.CreatedAt, 0).Format("2006-01-02 15:04")
		expires := time.Unix(s.ExpiresAt, 0).Format("2006-01-02 15:04")
		fmt.Printf("%-24s %-20s %-20s\n", s.ClientID, created, expires)
	}
	fmt.Printf("\n%d active session(s)\n", len(status.Sessions))
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
		case "devices":
			fmt.Print(`Usage: buzzpi devices

List all paired devices stored locally.

Examples:
  buzzpi devices
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
		case "terminal":
			fmt.Print(`Usage: buzzpi terminal <device_id>

Open a terminal session with a paired device.

The device must already be paired via 'buzzpi pair <device_id>'.
Type commands and press Enter to execute. Ctrl+C to exit.

Examples:
  buzzpi terminal dev_a1b2c3d4
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
