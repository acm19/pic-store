# Contributing to Pics

## Development Setup

### Prerequisites

#### Required
- **Go 1.24+**: [Install Go](https://go.dev/doc/install)
- **Make**: Build automation
- **curl, unzip, tar**: For downloading external binaries

#### For UI Development (optional)
- **Node.js 18+**: For Svelte frontend development
- **System dependencies** (Linux only):
  ```bash
  # Ubuntu/Debian
  sudo apt-get install libgtk-3-dev libwebkit2gtk-4.1-dev

  # Fedora/RHEL
  sudo dnf install gtk3-devel webkit2gtk4.1-devel

  # Arch
  sudo pacman -S gtk3 webkit2gtk-4.1
  ```

### Project Structure

This is a multi-module Go monorepo:

```
.
├── internal/pics/          # Shared business logic
├── apps/
│   ├── cli/               # CLI application (separate go.mod)
│   └── ui/                # Wails desktop UI (separate go.mod)
├── go.mod                 # Root module (shared dependencies)
└── Makefile              # Build automation
```

### Building

#### CLI Only
```bash
make build-cli
# Output: ./pics
```

#### UI Only
```bash
make build-ui
# Output: apps/ui/build/bin/pics-ui
```

#### Both
```bash
make build-all
```

### Development Workflow

#### Running the CLI
```bash
make run ARGS="parse /source /target"
```

#### Running the UI in development mode
```bash
make dev-ui
# Opens UI with hot reload
```

### Dependencies

#### Go Modules
```bash
# Tidy all modules (root, CLI, and UI)
make tidy
```

#### External Binaries (for UI)
The UI embeds exiftool and jpegoptim binaries for all platforms. These are downloaded automatically when building the UI, but you can download them manually:

```bash
make download-binaries
```

**Note**: Binaries are only downloaded if they don't exist, so subsequent builds are fast.

### Testing

```bash
make test
# Runs tests in all modules
```

### Common Issues

#### "go: command not found"
After installing Go, open a new terminal session or run:
```bash
source ~/.bashrc  # or ~/.profile
```

#### "Package webkit2gtk-4.0 was not found"
Install the correct version for your system:
```bash
# Ubuntu/Debian (newer versions use 4.1)
sudo apt-get install libwebkit2gtk-4.1-dev
```

#### "make: wails: command not found"
The Makefile uses `go run` to execute Wails, so no global installation is needed. Just ensure Go is in your PATH.

### Code Style

- Use `gofmt` for Go code formatting
- Run tests before submitting PRs
- Keep commits focused and atomic

### Makefile Targets

- `make build` - Build CLI (default)
- `make build-cli` - Build CLI only
- `make build-ui` - Build UI only
- `make build-all` - Build both CLI and UI
- `make dev-ui` - Run UI in development mode
- `make test` - Run all tests
- `make tidy` - Tidy all Go modules
- `make clean` - Remove build artifacts
- `make download-binaries` - Download external binaries for UI embedding

### Release Process

See [RELEASING.md](RELEASING.md) for release procedures.
