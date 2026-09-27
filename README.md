# ⚡ Redis Inspector

<div align="center">

[![Go Version](https://img.shields.io/badge/Go-1.21%2B-00ADD8?style=flat-square&logo=go)](https://golang.org)
[![License](https://img.shields.io/badge/license-MIT-blue.svg?style=flat-square)](LICENSE)
[![Platform](https://img.shields.io/badge/platform-windows%20%7C%20linux%20%7C%20macos-lightgrey?style=flat-square)](#-installation)
[![Tests](https://img.shields.io/badge/tests-passing-brightgreen?style=flat-square)](#-testing--quality)

**A high-performance, interactive Terminal User Interface (TUI) for real-time Redis memory profiling, namespace aggregation, and database administration.**

Built with Go, [Bubble Tea](https://github.com/charmbracelet/bubbletea), and [Lip Gloss](https://github.com/charmbracelet/lipgloss).

</div>

---

## 📖 What is Redis Inspector?

When running Redis in production, memory is your most expensive and scarce resource. Standard CLI tools like `redis-cli --bigkeys` or `redis-cli --memkeys` only sample a fraction of key types, and running `KEYS *` or blocking `DEL` commands can freeze an active Redis event loop.

**Redis Inspector** is a lightweight, non-blocking TUI diagnostic and management console. It allows Site Reliability Engineers (SREs), backend developers, and database administrators to:

1. **Spot Memory Hogs Instantly:** Automatically profile keys using pipelined `MEMORY USAGE` queries without degrading server latency.
2. **Understand Namespace Footprints:** Aggregate keys into hierarchical namespaces (e.g., `user:session:*`, `cache:product:*`) to see total allocated memory and count per prefix.
3. **Inspect & Edit Any Data Type:** Safely view raw values for strings, hashes, lists, sets, sorted sets, and streams (with JSON pretty-printing). Edit strings and hash payloads directly inline.
4. **Clean Up Safely at Scale:** Delete individual keys or bulk-purge entire namespaces using non-blocking asynchronous `UNLINK` commands in chunks of 500 keys.
5. **Switch Databases on the Fly:** Seamlessly switch between Redis databases (`DB 0` through `DB 15`) with full connection pool re-negotiation.

---

## 🖥️ Terminal Interface Preview

```text
  ⚡ REDIS INSPECTOR   [DB 0]   ● ONLINE   localhost:6379

 ╭──────────────────╮ ╭──────────────────╮ ╭──────────────────╮ ╭─────────────────────────╮
 │ USED MEMORY      │ │ PEAK MEMORY      │ │ FRAG RATIO       │ │ 30s MEMORY TREND        │
 │ 11.24 MB         │ │ 14.82 MB         │ │ 1.08             │ │ ▂▃▅▆▇█▇▆▄▃▂  ▂▃▄▅▆▇██   │
 │ 11,785,984 B     │ │ Alloc: jemalloc  │ │ ● Optimal        │ │ Live telemetry window   │
 ╰──────────────────╯ ╰──────────────────╯ ╰──────────────────╯ ╰─────────────────────────╯

 ╭────────────────────────────────────────────────────────────────────────────────────────╮
 │   1. Top Keys (48)    2. Namespaces (6)                                                │
 │                                                                                        │
 │  ╭──────────────────────────────────────────────╮                                      │
 │  │ 🔎 Filter: [ type to filter...             ] │                                      │
 │  ╰──────────────────────────────────────────────╯                                      │
 │                                                                                        │
 │    Top Memory Consumers   Page 1 of 5  (Keys: 48)                                      │
 │                                                                                        │
 │    #    KEY NAME                         TYPE       MEMORY     PROPORTION              │
 │   ─────────────────────────────────────────────────────────────────────────────        │
 │  👉  1. user:session:109482              [HASH]   248.50 KB    ████████░░              │
 │      2. cache:feed:homepage              [STR]     84.12 KB    ███░░░░░░░              │
 │      3. queue:notifications:dlq          [LIST]    32.80 KB    █░░░░░░░░░              │
 │      4. leaderboard:global:v2            [ZSET]    18.45 KB    █░░░░░░░░░              │
 │      5. telemetry:events:stream          [STRM]    12.20 KB    ░░░░░░░░░░              │
 │      6. tags:active:catalog              [SET]      4.10 KB    ░░░░░░░░░░              │
 ╰────────────────────────────────────────────────────────────────────────────────────────╯

   [Tab] Namespaces   [Enter] Inspect Key   [v] Raw Value   [b] Switch DB   [/] Search   [q] Quit
```

---

## ✨ Features

- **⚡ Real-Time Memory Analytics:**
  - Live memory metrics: `used_memory`, `peak_memory`, fragmentation ratio, and active memory allocator (jemalloc/libc).
  - Dynamic 30-second memory trend graph rendered via Unicode block sparklines (` ▂▃▄▅▆▇█`).
  - Real-time keyspace metrics across all 16 databases (`keys`, `expires`, `avg_ttl`).

- **📂 Namespace & Key Profiling:**
  - **Top Keys View:** Lists highest memory-consuming keys, their data type, and exact bytes.
  - **Namespaces View:** Automatically aggregates keys by `:` prefix, calculating total memory, key counts, and average bytes per key.
  - **Non-blocking Server SCAN:** Uses iterative `SCAN` cursors (100–500 keys per iteration) to prevent Redis latency spikes.

- **🔍 Interactive Search & Filter:**
  - Client-side live substring search as you type across keys and namespaces.
  - Server-side wildcard `SCAN` pattern matching (e.g. `user:*`) triggered by pressing `Enter`.

- **📄 Deep Key Inspection & Raw Value Viewers:**
  - Pipelined query for `TYPE`, `TTL`, `OBJECT ENCODING`, `MEMORY USAGE`, and element count (`STRLEN`, `HLEN`, `LLEN`, `SCARD`, `ZCARD`, `XLEN`).
  - Specialized viewers for **Strings** (with auto-detect JSON pretty-printing), **Hashes**, **Lists**, **Sets**, **Sorted Sets** (with scores), and **Streams** (with entry IDs and fields).

- **✏️ Safe Mutations & TTL Management:**
  - **Inline Value Editor:** Modify String values or update Hash fields directly via JSON. Keybinding is safely gated to prevent editing incompatible binary types.
  - **TTL & Expiration:** Set key TTL in seconds or remove expiration entirely using `PERSIST` (`-1`).
  - **Safe Asynchronous Deletion:** Single-key deletion uses `UNLINK` instead of blocking `DEL`.
  - **Bulk Namespace Purge:** Fast, iterative pattern deletion with confirmation modals and protection against accidental root key purging.

- **💾 JSON Report Export:**
  - Dump complete database memory diagnostics, top keys, and namespace breakdowns to a local timestamped JSON report (`redis_memory_report_<timestamp>.json`).

- **🔒 Production Security:**
  - Full support for Redis 6+ Access Control Lists (ACLs) via `--user` and `--auth`.
  - Encrypted TLS connections with optional certificate verification bypass (`--tls` and `--tls-skip-verify`).

---

## ⌨️ Keyboard Shortcuts Cheatsheet

### Global Navigation
| Key | Action |
|:---|:---|
| `Tab` / `1` / `2` | Toggle between **Top Keys** (1) and **Namespaces** (2) views |
| `↑` / `k` | Move selection up (wraps across pages) |
| `↓` / `j` | Move selection down (wraps across pages) |
| `n` / `PgDown` | Next page |
| `p` / `PgUp` | Previous page |
| `b` | Open Database Selector modal (`DB0` – `DB15`) |
| `/` | Focus filter / search input bar |
| `x` | Export memory statistics report to JSON |
| `s` | Seed mock sample keys for demonstration |
| `q` / `Ctrl+C` | Quit Redis Inspector |

### Key & Item Actions
| Key | Context | Action |
|:---|:---|:---|
| `Enter` | Top Keys List | Open **Key Inspector** modal with deep metadata |
| `Enter` | Namespaces List | Filter Top Keys by selected namespace prefix |
| `Enter` | Search Bar | Execute server-side `SCAN` with pattern |
| `v` | Key List / Details | Open **Raw Value Viewer** |
| `e` | Key Details | Open **Value Editor** (Strings & Hashes) |
| `t` | Key Details | Open **TTL Configuration** modal |
| `d` | Key List | Unlink single key (prompts confirmation) |
| `d` | Namespaces List | Bulk unlink all keys in namespace (prompts confirmation) |
| `Esc` | Any Modal / Search | Close modal, cancel search, or return to list |

---

## 🚀 Installation

### Option 1: Via `go install` (Recommended)
```bash
go install github.com/dheeraj080/redis-inspector/cmd/redis-inspector@latest
```
*Make sure `$GOPATH/bin` is in your system `PATH`.*

### Option 2: Pre-Compiled Binaries
Download the pre-built binary for your platform from the [GitHub Releases](https://github.com/dheeraj080/redis-inspector/releases) page:
- **Linux:** `amd64`, `arm64`
- **macOS (Darwin):** `amd64` (Intel), `arm64` (Apple Silicon)
- **Windows:** `amd64.exe`, `arm64.exe`

### Option 3: Build from Source
```bash
# Clone the repository
git clone https://github.com/dheeraj080/redis-inspector.git
cd redis-inspector

# Build the binary
go build -ldflags="-s -w" -o redis-inspector ./cmd/redis-inspector

# Run
./redis-inspector -addr="localhost:6379"
```

---

## 🛠️ Command-Line Flags

```text
Usage of redis-inspector:
  -addr string
        Redis server address (host:port) (default "localhost:6379")
  -auth string
        Redis password (-auth="secret")
  -password string
        Redis password (alias for -auth)
  -user string
        Redis ACL username (for Redis 6+)
  -db int
        Redis database index (0-15) (default 0)
  -tls
        Enable TLS connection
  -tls-skip-verify
        Skip TLS certificate verification (development only)
```

### Examples

**Local Redis (standard):**
```bash
redis-inspector
```

**Remote Redis with password authentication:**
```bash
redis-inspector -addr="redis.internal.net:6379" -auth="mypassword" -db=2
```

**AWS ElastiCache / Redis Cloud with ACLs and TLS:**
```bash
redis-inspector \
  -addr="clustercfg.my-cache.cache.amazonaws.com:6379" \
  -user="app_admin" \
  -auth="superSecretToken" \
  -tls
```

---

## 📐 Architecture & Internal Design

```text
redis-inspector/
├── cmd/
│   └── redis-inspector/
│       └── main.go           # CLI entrypoint, flag parsing, lifecycle management
├── internal/
│   ├── redis/
│   │   ├── client.go         # Domain layer: Redis RESP queries, pipelines, non-blocking UNLINK
│   │   └── client_test.go    # Unit tests for namespace extraction and INFO parser
│   ├── ui/
│   │   ├── model.go          # Bubble Tea state struct, constructor, clamp/pagination math
│   │   ├── model_test.go     # Unit tests for filtering, pagination, and option propagation
│   │   ├── update.go         # Event dispatch loop, modals, hotkey routing
│   │   ├── view.go           # Lip Gloss terminal rendering, sparklines, responsive layouts
│   │   ├── commands.go       # Asynchronous tea.Cmd factories with context timeouts
│   │   ├── cmds.go           # Value/TTL mutation commands
│   │   └── msg.go            # Bubble Tea message definitions
│   └── utils/
│       ├── format.go         # Byte formatter, string truncator, sparkline generator
│       └── format_test.go    # Unit tests for formatters and Unicode sparklines
├── .github/
│   └── workflows/
│       ├── build.yml         # CI workflow: dependency verification, go vet, test, build
│       └── release.yml       # Release pipeline: multi-arch cross-compilation & asset upload
├── build.ps1                 # PowerShell script for multi-platform local builds
├── go.mod                    # Module definition (github.com/dheeraj080/redis-inspector)
└── go.sum
```

### Key Architectural Safeguards
- **Non-Blocking Commands Only:** Redis Inspector never issues dangerous commands like `KEYS *`, `FLUSHALL`, or blocking `DEL`. It uses cursor-based `SCAN` loops and asynchronous background `UNLINK`.
- **Connection Isolation:** Switching databases (`b`) instantiates a clean `*redis.Client` pool with connection verification (`PING`) and safely closes obsolete connection pools to prevent socket leaks.
- **Context Timeouts:** All background commands are executed asynchronously via `tea.Cmd` bounded by timeouts (3s for reads, 5s for scans, 15s for bulk operations).

---

## 🧪 Testing & Quality

Run the test suite with coverage:
```bash
go test -v -cover ./...
```

Run race detector:
```bash
go test -race ./...
```

Run static analysis:
```bash
go vet ./...
```

---

## 🤝 Contributing

Contributions, bug reports, and feature requests are welcome!

1. Fork the repository.
2. Create a feature branch (`git checkout -b feature/my-enhancement`).
3. Commit your changes (`git commit -m 'Add support for Redis Cluster mode'`).
4. Ensure tests and linters pass (`go test ./... && go vet ./...`).
5. Push to the branch (`git push origin feature/my-enhancement`).
6. Open a Pull Request.

---

## 📄 License

This project is licensed under the MIT License — see the [LICENSE](LICENSE) file for details.