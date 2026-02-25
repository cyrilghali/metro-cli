# metro

A CLI tool to check **next metro departures** and **traffic disruptions** on the Paris metro network.

Uses the [PRIM Ile-de-France Mobilites](https://prim.iledefrance-mobilites.fr/) API.

## Install

```bash
go install github.com/cyrilghali/metro-cli@latest
```

Or build from source:

```bash
git clone https://github.com/cyrilghali/metro-cli.git
cd metro-cli
go build -o metro .
```

> **Note:** `go install` produces a binary named `metro-cli`. To get `metro`, use `go build -o metro .` or create an alias.

## Setup

Get a free API token at [prim.iledefrance-mobilites.fr](https://prim.iledefrance-mobilites.fr/), then:

```bash
metro config --token YOUR_TOKEN
```

Optionally set a default station:

```bash
metro config --default-station chatelet
```

The token can also be provided via the `PRIM_TOKEN` environment variable or a `.env` file with `token=...`.

## Usage

### Departures

```bash
# Search by station name
metro departures chatelet
metro departures "gare de lyon"

# Search by address (finds nearest metro stops)
metro departures "73 rue rivoli"

# Auto-detect location via browser geolocation
metro departures --here

# Use your default station
metro departures
```

Output shows upcoming departures grouped by line and direction:

```
Chatelet (Paris)
  Line  Direction                  Next departures
  ----  ---------                  ---------------
  M1    La Defense                 2 min, 5 min, 9 min
  M1    Chateau de Vincennes       now, 4 min, 8 min
  M4    Porte de Clignancourt      3 min, 7 min
  M4    Mairie de Montrouge        1 min, 6 min
```

### Disruptions

```bash
# All metro lines
metro disruptions

# Filter by line
metro disruptions --line M14
metro disruptions --line 1
```

Shows the status of all 16 Paris metro lines (M1-M14, M3B, M7B) with color-coded severity.

### Config

```bash
# View current config
metro config

# Set API token
metro config --token YOUR_TOKEN

# Set default station
metro config --default-station "nation"
```

Config is stored in `~/.metro.toml`.

## How it works

- **Station search** uses the PRIM places API with metro filtering and an interactive picker when there are multiple matches
- **Address search** geocodes the address via the Navitia API, then finds nearby metro stops within 500m
- **`--here` flag** opens a temporary local HTTP server, launches your browser, uses the Geolocation JS API, and sends the coordinates back. Works on macOS, Linux, and Windows.
- **Departures** are fetched via the Navitia v2 real-time API, filtered to metro only
- **Disruptions** are pulled from the Navitia lines endpoint with embedded disruption data

## License

MIT
