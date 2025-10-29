# SkyCLI

Terminal interface and command-line tools for Bluesky and SkyPanel server management.

## Features

- View status
- Basic CLI structure

See [ROADMAP](../ROADMAP.md) for planned features.

## Installation

### From Source

```sh
cd cli
task build
./tmp/skycli --help
```

## Usage

```sh
# View status
skycli status

# See all commands
skycli --help
```

See [ROADMAP](../ROADMAP.md) for planned commands and features.

## Development

### Prerequisites

- Go 1.24+

### Setup

```sh
cd cli

# Install dependencies
go mod download

# Run development version
go run main.go --help

# Build for testing
task build

# Run tests
task test
```

### Project Structure

```sh
cli/
├── cmd/                    # Command implementations
├── internal/               # Internal packages
│   ├── config/             # Configuration
│   └── client/             # API client
├── main.go                 # Entry point
├── Taskfile.yaml           # Task runner config
└── README.md               # This file
```

## References

- [CLI ROADMAP](../ROADMAP.md) - Planned features
