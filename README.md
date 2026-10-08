# aloh TUI

> A terminal-based voice/video/SNS client built with Go and Bubble Tea.

**aloh TUI** is a feature-rich terminal user interface (TUI) application for real-time voice chat, video streaming (webcam & screen sharing), instant messaging, and social networking. It uses SSH-based authentication for identity management and ICE/QUIC for peer-to-peer media streaming.

---

## Table of Contents

- [Features](#features)
- [Architecture](#architecture)
- [Technology Stack](#technology-stack)
- [Prerequisites](#prerequisites)
  - [Linux](#linux)
  - [Windows](#windows)
- [Installation](#installation)
- [Quick Start](#quick-start)
- [Usage](#usage)
  - [Authentication](#authentication)
  - [Navigation & Keybindings](#navigation--keybindings)
  - [Tabs Overview](#tabs-overview)
- [Project Structure](#project-structure)
- [Building](#building)
- [Dependencies](#dependencies)
- [Logging](#logging)
- [Data Storage](#data-storage)
- [Contributing](#contributing)
- [License](#license)
- [Acknowledgments](#acknowledgments)

---

## Features

| Feature | Description |
| :--- | :--- |
| **Voice Chat** | Real-time Opus-encoded voice chat with RNNoise + SpeexDSP noise suppression, echo cancellation, and per-user volume control |
| **Webcam** | Live webcam streaming in the terminal using H.264/OpenH264 encoding via custom fork `pion/mediadevices` |
| **Screen Sharing** | Full desktop capture shared directly to other peers |
| **Instant Messaging** | Text chat with image paste support (clipboard integration) |
| **Friend System** | Add friends, send/accept/deny requests, and block/unblock users |
| **Presence** | Real-time online/offline status tracking via SSH events |
| **Notifications** | Audio, desktop, and in-app notification system |
| **Theme & Appearance** | Customizable theme colors, emoji tags, and date/time formats |
| **Account Management** | Change nickname, password, tagline, and profile color |
| **Device Management** | Switch microphones, headphones, and webcams at runtime |
| **Statistics** | Track connection time, message count, best friend, and more |
| **Cross-Platform** | Native binaries for Linux and Windows with Docker cross-compilation |

---

## Architecture

```text
┌─────────────────────────────────────────────────────────────────────┐
│                         cmd/app/main.go                             │
│                    (entry point, launches app)                      │
├─────────────────────────────────────────────────────────────────────┤
│                     internal/app/                                   │
│   • app.go            — main app runner (file setup, logger, TUI)   │
│   • app_linux.go      — Linux: UDP buffer tuning, error redirect    │
│   • app_windows.go    — Windows: error redirect, env vars           │
├─────────────────────────────────────────────────────────────────────┤
│                     internal/tui/                                   │
│   • model.go          — Bubble Tea Model (state, initialization)     │
│   • view.go           — Lipgloss-based terminal rendering            │
│   • update.go         — Event loop, key/mouse handling               │
│   • inputs.go         — Input focus/unfocus management               │
│   • utils.go          — Utility functions (truncation, validation)   │
│   • setupers.go       — TUI setup (lists, colors, animations)          │
│   • commands/         — Bubble Tea commands (I/O, async operations)   │
│   • components/       — Styles, states, windows, titles, lists        │
│   • helpers/          — Auth helper (SSH + engine initialization)      │
├─────────────────────────────────────────────────────────────────────┤
│                     internal/entities/                              │
│   • users/            — User data model (friends, mutes, stats)       │
│   • setups/           — User-specific audio/per-connection setup      │
├─────────────────────────────────────────────────────────────────────┤
│                     internal/media/                                 │
│   • audio/            — Audio engine (malgo, Opus, RNNoise, SpeexDSP) │
│   • video/            — Video engine (pion/mediadevices, OpenH264)    │
│   • devices.go        — Shared device type alias                      │
├─────────────────────────────────────────────────────────────────────┤
│                     internal/networking/                            │
│   • networking.go     — Networking interface & config               │
│   • events.go         — Event constants & helpers                    │
│   • config.yaml       — Embedded STUN/TURN/timeout config           │
├─────────────────────────────────────────────────────────────────────┤
│                     internal/sshclient/                             │
│   • client.go         — SSH client (auth, channel management)         │
│   • keys.go           — SSH key generation (Ed25519)                  │
│   • event.go          — SSH event processing & casting                │
│   • request.go        — SSH custom request handlers                   │
│   • errs.go           — SSH error helpers                             │
├─────────────────────────────────────────────────────────────────────┤
│                     internal/utils/                                 │
│   • utils.go          — File path setup, color utilities              │
├─────────────────────────────────────────────────────────────────────┤
│                     pkg/                                            │
│   • logger/           — Structured logger (slog-based)                │
│   • errs/             — Application error types                       │
└─────────────────────────────────────────────────────────────────────┘
```

---

## Technology Stack

| Layer | Technology |
| :--- | :--- |
| **TUI Framework** | [Bubble Tea](https://github.com/charmbracelet/bubbletea) v1.3.10 |
| **Rendering** | [Lipgloss](https://github.com/charmbracelet/lipgloss) v1.1.0 + [Bubblezone](https://github.com/lrstanley/bubblezone) |
| **Audio** | [malgo](https://github.com/gen2brain/malgo) for device I/O, Opus for encoding, RNNoise & SpeexDSP for noise suppression |
| **Video** | [pion/mediadevices](https://github.com/pion/mediadevices) for capture, OpenH264 for H.264 encoding/decoding |
| **Networking** | Custom `aloh-networking` (ICE/QUIC-based P2P), `aloh-signalling` |
| **Authentication** | SSH-based auth via `aloh-ssh`, Ed25519 key pairs |
| **Terminal Images** | [go-termimg](https://github.com/blacktop/go-termimg) for rendering webcam/screen frames |
| **Clipboard** | [golang.design/x/clipboard](https://github.com/golang-design/clipboard) for image paste |
| **Password Validation** | `go-password-validator` |

---

## Prerequisites

### Linux

- **Go 1.26+** (or Go 1.27+ as specified in `go.mod`)
- C compiler (GCC) for CGO
- ALSA / PulseAudio development libraries
- `libogg`, `libopus`, `libopusfile`, `libspeexdsp`, `libvpx`
- `libX11`, `libxcb`, `libXau`, `libXdmcp` (for desktop capture)
- OpenH264 shared library (for video encoding)
- RNNoise shared library (for neural noise suppression)
- OpenCV (for video frame rendering utilities)
- Terminal with true color and image protocol support (e.g., Kitty, WezTerm)

#### Install Dependencies (Debian/Ubuntu):

```bash
sudo apt-get update
sudo apt-get install -y \
  build-essential pkg-config git autoconf automake libtool \
  libx11-dev libxcb1-dev libxau-dev libxdmcp-dev \
  libasound2-dev libpulse-dev \
  libogg-dev libopus-dev libopusfile-dev libspeexdsp-dev libvpx-dev \
  cmake libwebp-dev
```

#### Build RNNoise:

```bash
git clone https://github.com/xiph/rnnoise.git
cd rnnoise
./autogen.sh
./configure --prefix=/usr/local --enable-shared --disable-static
make -j$(nproc) && sudo make install
sudo ldconfig
```

### Windows

- Go 1.26+
- MSYS2 with MinGW-w64 (POSIX threads)
- OpenH264 DLL
- Terminal with image protocol support (Windows Terminal recommended)

---

## Installation

### Pre-built Releases

Download pre-built binaries from the Releases page:

```bash
# Linux
wget https://github.com/kiryuhakipyatok/aloh-tui/releases/latest/download/aloh-linux-amd64-s.tar.gz
tar -xzf aloh-linux-amd64-s.tar.gz
cd dist-linux && ./aloh-linux
```

---

## Quick Start

```bash
./aloh
# 1. Press ENTER or TAB to start
# 2. Register with a unique nickname and strong password (min 55 bits entropy)
# 3. Or log in with existing credentials
# 4. After login: connect to friends, chat, share webcam/screen
```

---

## Usage

### Authentication

The application uses SSH-based authentication with Ed25519 key pairs stored in `keys/id_ed25519`.

- **Registration**: Create a nickname + password (validated for $\ge 55$ bits of entropy).
- **Login**: Enter your nickname + password.
- **Auto-login**: Existing credentials in `userdata.json` are automatically detected on startup.

### Navigation & Keybindings

| Key | Action |
| :--- | :--- |
| <kbd>ALT</kbd> + <kbd>Q</kbd> / <kbd>ALT</kbd> + <kbd>Й</kbd> | Quit application |
| <kbd>TAB</kbd> / <kbd>→</kbd> | Next tab |
| <kbd>SHIFT</kbd> + <kbd>TAB</kbd> / <kbd>←</kbd> | Previous tab |
| <kbd>ALT</kbd> + <kbd>→</kbd> | Switch to right pane |
| <kbd>ALT</kbd> + <kbd>←</kbd> | Switch to left pane |
| <kbd>↑</kbd> / <kbd>↓</kbd> | Move cursor |
| <kbd>ESC</kbd> | Back / close error dialog |
| <kbd>ALT</kbd> + <kbd>H</kbd> | Open help dialog |

- **7 tabs when logged in:** `Friends`, `Chat`, `Voice`, `Webcam`, `Screen`, `Profile`, `Settings`
- **2 tabs when logged out:** `Registration`, `Login`

### Tabs Overview

#### Friends Tab
- Enter nickname $\to$ press <kbd>ENTER</kbd> to initiate connection.
- Send, accept, or deny incoming friend requests.
- Block and unblock users.

#### Voice Chat
- <kbd>ALT</kbd> + <kbd>V</kbd> — Toggle microphone mute
- <kbd>ALT</kbd> + <kbd>B</kbd> — Toggle deafen (mute all audio output)
- <kbd>ALT</kbd> + <kbd>Z</kbd> — Mute selected user
- <kbd>ALT</kbd> + <kbd>W</kbd> — Toggle personal soft denoise
- <kbd>ALT</kbd> + <kbd>R</kbd> — Toggle personal hard denoise
- <kbd>ALT</kbd> + <kbd>↑</kbd> / <kbd>ALT</kbd> + <kbd>↓</kbd> — Adjust selected user volume

**Status Indicators:**
- 🔊 Speaking
- 🔇 Muted
- 🙊 Microphone muted
- 🙊🙉 Fully deafened

#### Webcam & Screen
- Press <kbd>ENTER</kbd> on the webcam/screen tab to toggle your stream.
- Click user video frames to open or close individual windows.
- Video streams use H.264 encoding and are rendered as Unicode / terminal graphics.

#### Profile Tab
- View nickname, registration date, and connection statistics.
- Manage friend requests (<kbd>ENTER</kbd> to accept, <kbd>ALT</kbd> + <kbd>X</kbd> to deny).
- Appearance settings (theme color, profile tags).

#### Settings Tab
Split into left (general app settings) and right (contextual settings) panes:
- **Audio:** AEC (acoustic echo cancellation), hard/soft denoise, equalizer.
- **Notifications:** In-app, desktop, and audio alerts.
- **Devices:** Microphones, headphones, and webcams.
- **Binds:** Fully customizable keyboard shortcuts.
- **Account:** Nickname, profile color, password, and tagline.

---

## Building

### From Source (Linux)

```bash
# Simple build
go build -tags nomicrophone -o aloh ./cmd/app

# Makefile build targets:
# make bs  — static cross-compile (Linux + Windows)
# make bd  — dynamic cross-compile
# make b   — full build with OpenCV
# make lb  — local Linux build
```

### CI/CD
GitHub Actions workflow (`.github/workflows/build.yml`) triggered via `workflow_dispatch` with version input — builds native binaries for Ubuntu and Windows with all native dependencies included.

---

## Dependencies

- **Go (Key Packages):** `Bubble Tea`, `Lipgloss`, `Bubblezone`, `malgo`, `pion/mediadevices`, `openh264`, `Opus`, `RNNoise`, `SpeexDSP`, `go-termimg`, `clipboard`, `password-validator`, `slog`.
- **C Libraries:** `libogg`, `libopus`, `libopusfile`, `libspeexdsp`, `rnnoise`, `libX11`, `OpenCV`, `OpenH264`, `ALSA` / `PulseAudio`.

---

## Logging

- **Levels:** `local` (Debug plain text), `dev` (Debug JSON), `prod` (Info JSON).
- **Files:** `app-logs/log.log`, `networking-logs/log.log`, `errors.log`.
- **Profiling:** Built-in HTTP server on `localhost:6060` (Linux only) for `pprof`.

---

## Data Storage

Application data is persisted to `userdata.json`:
- User identity (UUID + nickname)
- Friends list with per-user denoise and audio settings
- Friend requests and blocked user list
- Hardware device preferences (microphone, headphones, webcam)
- Audio configuration (AEC, denoise levels, mutes, filters)
- Appearance options (theme color, emoji tags)
- Notification preferences
- Custom key bindings
- Account details (tagline, custom color)
- Statistics (session time, total messages sent, top contact)

---

## Contributing

```bash
git clone https://github.com/kiryuhakipyatok/aloh-tui.git
cd aloh-tui
go mod download
go run ./cmd/app
```

> **Code style:** Follow `gofmt`, `goimports`, interface-driven design, and thread safety via `sync.Mutex` and `sync/atomic`.

---

## License

This project is licensed under the **MIT License** — see the [LICENSE](LICENSE) file for details.

---

## Acknowledgments

- [Charm](https://charm.sh/) — Bubble Tea, Lipgloss, Bubblezone, keygen
- [Pion](https://pion.ly/) — mediadevices and WebRTC foundation
- [gen2brain](https://github.com/gen2brain) — malgo
- [Xiph.Org](https://www.xiph.org/) — Opus, SpeexDSP, RNNoise
- [Cisco](https://www.openh264.org/) — OpenH264
- [blacktop](https://github.com/blacktop) — go-termimg