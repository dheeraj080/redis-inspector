# ⚡ Redis Inspector

**Redis Inspector** is a fast, feature-rich, and interactive Terminal User Interface (TUI) built in Go (using Charm's Bubble Tea framework) designed for SREs, developers, and DevOps engineers to monitor, troubleshoot, and manage Redis instances directly from the terminal.

![Go Version](https://img.shields.io/badge/Go-1.21%2B-blue.svg)
![License](https://img.shields.io/badge/license-MIT-green.svg)
![Platform](https://img.shields.io/badge/platform-windows%20%7C%20linux%20%7C%20macos-lightgrey)

---

## ✨ Key Features

* **📊 Real-Time Memory Analytics:** Live sparkline memory trends, peak memory tracking, fragmentation ratios, and database statistics across all 16 DBs (`0-15`).
* **📂 Namespace & Key Breakdown:** Aggregate keys by namespace prefixes (e.g., `user:session:*`) or inspect individual top-memory keys sorted by byte size.
* **✏️ Interactive Key Content Editor:** View raw values or inline-edit strings, hashes, lists, and sets directly inside Redis without leaving the TUI.
* **🔍 Server-Side `SCAN MATCH` Search:** Scan millions of keys in massive production environments using server-side pattern matching.
* **⏱️ TTL Management & Cleanup:** Easily update key expirations, persist keys, or safely unlink single keys and bulk namespace patterns.
* **📦 Cross-Platform Distribution:** Pre-configured build script for seamless compilation across Windows, Linux, and macOS (`amd64` / `arm64`).

---

## ⌨️ Keyboard Shortcuts

| Key | Action |
| :--- | :--- |
| `1` / `2` / `Tab` | Switch between **Top Keys** and **Namespaces** views |
| `/` | Open search / pattern filter bar |
| `Enter` | Inspect selected key details (or scan server pattern when searching) |
| `v` | View raw content of the selected key |
| `e` | Open interactive value editor modal |
| `t` | Modify TTL / expiration of a key |
| `d` | Unlink / delete selected key or bulk namespace |
| `b` | Select and switch Redis database (`0-15`) |
| `s` | Seed mock test data |
| `x` | Export memory usage report to JSON |
| `q` / `Ctrl+C` | Quit application |

---

## 🚀 Getting Started

### Prerequisites
* Go 1.21 or higher installed on your machine.
* A running Redis instance (local or remote).

### Installation & Running Locally

1. **Clone the repository:**
   ```bash
   git clone [https://github.com/your-username/redis-inspector.git](https://github.com/your-username/redis-inspector.git)
   cd redis-inspector