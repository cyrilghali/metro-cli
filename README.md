<div align="center">

<br>

# 🚇 metro-cli

_C'est dans combien le prochain ?_

Real-time Ile-de-France departures and disruptions in your terminal.

<br>

[![CI](https://img.shields.io/github/actions/workflow/status/cyrilghali/metro-cli/ci.yml?style=flat-square&label=tests)](https://github.com/cyrilghali/metro-cli/actions/workflows/ci.yml)
[![Release](https://img.shields.io/github/v/release/cyrilghali/metro-cli?style=flat-square&color=0055A4)](https://github.com/cyrilghali/metro-cli/releases/latest)
[![License](https://img.shields.io/badge/license-MIT-0055A4?style=flat-square)](LICENSE)
[![Go](https://img.shields.io/badge/Go-1.23-00ADD8?style=flat-square&logo=go&logoColor=white)](https://go.dev)

</div>

<br>

```
$ metro departures chatelet

  Châtelet (Paris)

  Line   Direction                  Next departures
  ----   ---------                  ---------------
  M1     La Défense                 2 min, 5 min, 9 min
  M1     Château de Vincennes       now, 4 min, 8 min
  M4     Porte de Clignancourt      3 min, 7 min
  M14    Olympiades                 2 min, 5 min

$ metro departures chatelet --mode rer

  Châtelet les Halles (Paris)

  Line    Direction                  Next departures
  ----    ---------                  ---------------
  RER A   Marne-la-Vallée Chessy     2 min, 14 min
  RER B   Aéroport CDG 2 TGV        4 min, 19 min
  RER D   Creil                      7 min

$ metro disruptions --mode rer

  Line    Status       Info
  ----    ------       ----
  RER A   OK
  RER B   OK
  RER C   Modified     Issy : gare non desservie
  RER D   Delays       Plan de transport adapté
  RER E   OK
```

---

<br>

## Install

**Quick install (Linux / macOS):**

```bash
curl -sSfL https://raw.githubusercontent.com/cyrilghali/metro-cli/master/install.sh | sh
```

This downloads the latest release binary and installs it to `/usr/local/bin` (may prompt for `sudo`).
Run it again at any time to update to the latest version.

**From source:**

```bash
go install github.com/cyrilghali/metro-cli@latest
```

Or clone and build:

```bash
git clone https://github.com/cyrilghali/metro-cli.git
cd metro-cli
go build -o metro .
```

> `go install` produces a binary named `metro-cli`. To get `metro`, use `go build -o metro .` or alias it.

<br>

## Setup

Get a **free** API token at [prim.iledefrance-mobilites.fr](https://prim.iledefrance-mobilites.fr/), then:

```bash
metro config --token YOUR_TOKEN
```

Set a default station so you can just run `metro departures`:

```bash
metro config --default-station chatelet
```

<details>
<summary>Other token methods</summary>

<br>

```bash
# Environment variable
export PRIM_TOKEN=your_token

# .env file in current directory
echo 'token=your_token' > .env
```

Token lookup order: `PRIM_TOKEN` env → `~/.metro.toml` → `.env`

</details>

<br>

## Usage

### `metro departures` — next trains

```bash
metro departures chatelet              # search by station name (all modes)
metro d chatelet                       # short alias
metro dep "gare de lyon"              # quotes for multi-word names
metro d "73 rue rivoli"                # search by address (finds nearby stops)
metro d --here                         # auto-detect location via browser
metro d                                # uses your default station
metro d chatelet -m metro              # metro only
metro d chatelet -m rer                # RER only
```

When multiple stations match, an interactive picker lets you choose:

```
Multiple results found:
  1. Châtelet (Stop [M1, M4, M7, M11, M14]) - Paris
  2. Châtelet les Halles (Stop [RER A, RER B, RER D]) - Paris
  3. Château d'Eau (Stop [M4]) - Paris

Pick a number:
```

<br>

### `metro disruptions` — line status

```bash
metro disruptions                      # all lines (default)
metro dis                              # short alias
metro status                           # another alias
metro dis -m metro                     # metro lines only
metro dis -m rer                       # RER lines only
metro dis --line A                     # filter by line
```

Status is color-coded in your terminal:

| Color | Meaning |
|:------|:--------|
| 🟢 Green | Normal service |
| 🟡 Yellow | Delays / reduced / modified service |
| 🔴 Red | Service interrupted |

<br>

### `--mode` — transport modes

Both `departures` and `disruptions` accept a `--mode` / `-m` flag:

| Mode | What | Lines |
|:-----|:-----|:------|
| `all` | Everything (default) | All modes |
| `metro` | Metro | M1-M14, M3B, M7B |
| `rer` | RER | A, B, C, D, E |
| `train` | Transilien | H, J, K, L, N, P, R, U |
| `tram` | Tramway | T1-T13 |
| `bus` | Bus | All IDF bus lines |

<br>

### `metro config` — settings

```bash
metro config                           # view current config
metro config --token YOUR_TOKEN        # save API token
metro config --default-station nation  # save default station
```

Config is stored in `~/.metro.toml`.

<br>

## The `--here` flag

The `--here` flag finds stops near your **current location**:

1. Starts a temporary local HTTP server
2. Opens your browser
3. Browser asks for geolocation permission
4. Coordinates are sent back to the CLI
5. Nearby stops are found within 500m

Works on **macOS**, **Linux**, and **Windows**.

<br>

## How it works

| Feature | How |
|:--------|:----|
| **Station search** | PRIM places API with mode filtering + interactive picker |
| **Address search** | Navitia geocoding → nearby stops within 500m |
| **Geolocation** | Temporary localhost server + browser `navigator.geolocation` |
| **Departures** | Navitia v2 real-time API, filtered by transport mode |
| **Disruptions** | Navitia lines endpoint with embedded disruption data |

All data comes from the [PRIM Ile-de-France Mobilites](https://prim.iledefrance-mobilites.fr/) API gateway.

<br>

## License

[MIT](LICENSE)
