# Netmon-mcp (Go + eBPF + Model Context Protocol)

This project implements a **Model Context Protocol (MCP)** server written in Go, exposing an MCP tool capable of monitoring **incoming network traffic using eBPF**.  
The server currently works **only on Linux distributions** with kernel support for XDP/eBPF and must be run **with sudo privileges**.

The project structure and behavior follow the MCP specification and use the official `mark3labs/mcp-go` SDK for MCP server development.

## Features

### 1. MCP Server in Go (stdio transport)
The server uses `ServeStdio` to communicate with MCP clients (currently available for Opencode).  
It exposes a single tool:

| Tool name | Description |
|----------|-------------|
| `network-monitor` | Monitors packet rate on a given network interface using eBPF |

### 2. eBPF Network Monitoring
The project uses:

- `cilium/ebpf` Go library
- XDP program compiled from C using `bpf2go`
- A kernel-attached eBPF program that counts incoming packets

The tool returns **packets per second** (averaged over 5 seconds).

## Project Structure

```
netmon-mcp/
│
├── main.go                   # MCP server entry point
├── go.mod / go.sum           # Go dependencies
│
├── netmon/
    ├── include                 # header files
    ├── netmon.c                # eBPF XDP program (C)
    ├── netmon_bpfel.go         # Auto-generated BPF bindings
    ├── netmon_bpfel.o          # Compiled ELF object
    └── netmon.go               # Go wrapper for eBPF loading & traffic monitoring

```

## Requirements

### Operating System
✔ Linux distribution  
✔ Kernel with eBPF + XDP support  
✔ `clang` and `llc` (LLVM) for building eBPF code  
✔ `sudo` privileges (required to load eBPF programs and running Opencode)

### Go Requirements
- Go 1.22+
- Cilium/ebpf library
- `bpf2go` (included with Go toolchain)

Install dependencies:

```bash
sudo apt update
sudo apt install clang llvm
```

## Building the eBPF Program

The Go file includes:

```go
//go:generate go tool bpf2go -tags linux -go-package ebpf -output-dir ebpf/ netmon ebpf/netmonitor.c
```

To regenerate the BPF bindings:

```bash
go generate ./...
```

## Running the MCP Server

### 1. Build
```bash
go build -o mcp-test
```

### 2. Run (Linux only)
Because XDP requires privileged access:

```bash
sudo ./mcp-test
```

You will see logs such as:

```
Listening on stdio...
Counting incoming packets on wlp1s0...
```

To call the tool manually (if using an MCP client):

```json
{
  "name": "network-monitor",
  "arguments": {
    "network_operation": "incoming"
  }
}
```

## How the Project Works (Current State)

### 1. MCP Server
The server initializes using:

```go
server.NewMCPServer(...)
server.ServeStdio(s)
```

This means it uses **stdio transport**, which Claude Desktop supports.

### 2. Exposed Tool: `network-monitor`

Registered with:

```go
mcp.NewTool("network-monitor", ...)
```

The tool expects:

```json
{
  "network_operation": "incoming" | "outgoing"
}
```

Currently only `"incoming"` is implemented.

### 3. eBPF Logic
When the tool is executed:

1. The Go program loads the auto-generated eBPF objects
2. XDP program attaches to the interface (hardcoded: wlp1s0)
3. For 5 seconds, packet counts are read from the BPF map
4. The average packets/second is returned

This is done in:

```go
ebpf.IncomingPacketsPerSecond()
```

If MEMLOCK ulimit is too low, fix:

```bash
sudo sysctl -w kernel.unprivileged_bpf_disabled=0
ulimit -l unlimited
```

Or add to `/etc/security/limits.conf`:

```
* hard memlock unlimited
* soft memlock unlimited
```

## Claude MCP Client Usage

Example config:

```json
{
  "mcpServers": {
    "mcp-test": {
      "command": "wsl.exe",
      "args": [
        "bash",
        "-lc",
        "cd /home/youruser/dev/mcp-test && ./mcp-test"
      ]
    }
  }
}
```

Opencode automatically connects via **stdio**.

## Current Limitations

- Linux only (eBPF cannot run natively on Windows)
- Requires sudo to attach XDP

## Future Work

- Replace stdio + sudo with:
    - MCP gateway (non-root)
    - Privileged daemon (root)
- Add HTTP transport
- Add outgoing traffic monitoring
- Interface auto-detection
- Docker support
- Remote monitoring via SSH or HTTP tunnel

## License

MIT License