# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

BlueGo is a Bluetooth Low Energy (BLE) management system specifically designed for K18S smart bracelets. The project provides both a web interface and command-line tools for discovering, connecting to, and interacting with K18S fitness trackers over Bluetooth. The system enables real-time heart rate monitoring, battery status checking, and sending notifications to the bracelets.

## Development Commands

### Building and Running
```bash
# Build the project
go build -o bluego

# Run the main application (starts HTTP server on port 8000)
go run main.go

# Run with specific adapter (default is hci0)
go run main.go --adapterID=hci0
```

### Testing
```bash
# Run all tests
go test ./...

# Run tests for specific package
go test ./s18

# Run tests with verbose output
go test -v ./s18
```

### Dependencies
```bash
# Download and install dependencies
go mod download

# Tidy up dependencies
go mod tidy
```

## Architecture

### Package Structure

**`discovery/`** - Bluetooth device discovery and adapter management
- `Run()`: Continuous discovery that filters for K18S bracelets (names starting with "K18S")
- `RunWithin()`: Time-limited discovery (returns all devices within specified duration)
- Uses `go-bluetooth` library to interact with BlueZ D-Bus API
- Default adapter: `hci0`

**`s18/`** - K18S bracelet protocol implementation
- Core type: `Bracelet` struct managing device connection state
- Communication via GATT characteristics:
  - RX channel (0000fff1...): Receives responses via notifications
  - TX channel (0000fff2...): Sends commands
- Protocol uses 0x68/0x16 frame delimiters with packet reassembly for fragmented responses
- Key operations: `GetBattery()`, `GetVersion()`, `StartTracing()`, `StopTracing()`, `Tracing()`, `Notification()`, `Reset()`
- Automatic reconnection handling via `connQueue` channel

**`http/`** - Web server and real-time communication
- Gin framework for HTTP/REST endpoints
- Socket.IO for websocket-based real-time communication
- Global state maps:
  - `ConnectedMap`: Active bracelet connections by name
  - `DiscoveredMap`: Discovered devices by D-Bus path
- Key endpoints:
  - `GET /scan`: Discover devices for 10 seconds, returns list
  - `GET /get_base_data`: Returns connected bracelets status
  - Socket.IO events: `open` (connect to bracelet), `start` (begin monitoring), `stop` (end monitoring)
- Server runs on port 8000

**`hid/`** - USB HID device enumeration and reading (separate from BLE functionality)

**`cmd/`** - Cobra-based CLI commands
- Root command: `go-bluetooth`
- Subcommands: `discovery` (scan for devices)
- Configuration via Viper (supports environment variables and config files)

### Data Flow

1. **Device Discovery**: `discovery.Run()` or `RunWithin()` scans for BLE devices via BlueZ adapter
2. **Connection**: Socket.IO `open` event triggers `s18.RBracelet()` to connect and initialize GATT characteristics
3. **Command/Response**: Commands sent via TX characteristic, responses received asynchronously on RX channel
4. **Real-time Monitoring**: `start` event begins goroutine that polls heart rate data every second and broadcasts via Socket.IO
5. **Packet Handling**: RX listener reassembles fragmented BLE packets based on frame markers before parsing

### Important Implementation Details

- **Bluetooth Adapter**: Code assumes `hci0` adapter; must be powered and available
- **Packet Reassembly**: S18 protocol may split responses across multiple BLE notifications; buffering logic handles 0x68/0x16 frame boundaries
- **Goroutine Management**: Each bracelet connection spawns multiple goroutines (reconnection handler, RX listener, trace loop); cleanup requires channel signals
- **Socket.IO Channels**: Default namespace for control, `/test` namespace for broadcasting trace data to separate clients
- **Chinese Comments**: Some log messages and comments are in Chinese; code logic is standard Go

### Protocol Notes (s18 package)

Command IDs are single-byte:
- `0x03`: Get battery level
- `0x07`: Get firmware version
- `0x06`: Trace commands (0x00=query, 0x01=start continuous, 0x02=stop)
- `0x08`: Send text notification to bracelet
- `0x01`: Send call notification
- `0x11`: Reset/reboot bracelet

Responses contain structured data parsed in `response.go` (e.g., `HeartBeatResponse` with heart rate, step count, distance, calories).

## Environment Requirements

- Linux system with BlueZ stack installed
- Bluetooth adapter (hci0 or specified)
- Go 1.13+ (as specified in go.mod)
- Root or CAP_NET_ADMIN permissions may be required for BLE operations
