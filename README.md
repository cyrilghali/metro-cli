<div align="center">

<br>

# 🚇 metro-cli

**Prochain passage.**

_Real-time Paris metro departures and disruptions in your terminal._

<br>

[![License](https://img.shields.io/badge/license-MIT-0055A4?style=flat-square)](LICENSE)
[![Go](https://img.shields.io/badge/Go-1.23-00ADD8?style=flat-square&logo=go&logoColor=white)](https://go.dev)

</div>

<br>

```
$ metro departures chatelet

  Châtelet (Paris)

  Line  Direction                  Next departures
  ----  ---------                  ---------------
  M1    La Défense                 2 min, 5 min, 9 min
  M1    Château de Vincennes       now, 4 min, 8 min
  M4    Porte de Clignancourt      3 min, 7 min
  M4    Mairie de Montrouge        1 min, 6 min
  M7    La Courneuve               4 min, 11 min
  M11   Mairie des Lilas           now, 8 min
  M14   Olympiades                 2 min, 5 min

$ metro disruptions

  Line  Status       Info
  ----  ------       ----
  M1    OK
  M2    OK
  M3    Delays       Ralentissement entre Villiers et Opéra
  M4    OK
  ...
  M13   Interrupted  Service interrompu entre Montparnasse et ...
  M14   OK
```

<br>

> **metro** — check when the next one arrives, without leaving your terminal.

---

<br>

## 📦 Install

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

> **Note:** `go install` produces a binary named `metro-cli`. To get `metro`, use `go build -o metro .` or alias it.

<br>

## 🔑 Setup

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

## 🚉 Usage

### `metro departures` — next trains

```bash
metro departures chatelet              # 🔍 search by station name
metro departures "gare de lyon"        # 🔍 quotes for multi-word names
metro departures "73 rue rivoli"       # 📍 search by address (finds nearby stops)
metro departures --here                # 🌐 auto-detect location via browser
metro departures                       # ⭐ uses your default station
```

When multiple stations match, an interactive picker lets you choose:

```
Multiple results found:
  1. Châtelet (Stop [M1, M4, M7, M11, M14]) - Paris
  2. Châtelet - Les Halles (Stop [RER A, RER B, RER D]) - Paris
  3. Château d'Eau (Stop [M4]) - Paris

Pick a number:
```

<br>

### `metro disruptions` — line status

```bash
metro disruptions                      # 📊 all 16 metro lines
metro disruptions --line M14           # 🔎 filter by line
metro disruptions --line 1             # 🔎 also works without the M prefix
```

Status is color-coded in your terminal:

| Color | Meaning |
|:------|:--------|
| 🟢 Green | Normal service |
| 🟡 Yellow | Delays / reduced / modified service |
| 🔴 Red | Service interrupted |

<br>

### `metro config` — settings

```bash
metro config                           # 👀 view current config
metro config --token YOUR_TOKEN        # 🔑 save API token
metro config --default-station nation   # ⭐ save default station
```

Config is stored in `~/.metro.toml`.

<br>

## 📍 The `--here` flag

The `--here` flag finds metro stops near your **current location**:

1. Starts a temporary local HTTP server
2. Opens your browser
3. Browser asks for geolocation permission
4. Coordinates are sent back to the CLI
5. Nearby metro stops are found within 500m

Works on **macOS**, **Linux**, and **Windows**.

<br>

## ⚙️ How it works

| Feature | How |
|:--------|:----|
| **Station search** | PRIM places API with metro filtering + interactive picker |
| **Address search** | Navitia geocoding → nearby metro stops within 500m |
| **Geolocation** | Temporary localhost server + browser `navigator.geolocation` |
| **Departures** | Navitia v2 real-time API, filtered to `physical_mode:Metro` |
| **Disruptions** | Navitia lines endpoint with embedded disruption data |

All data comes from the [PRIM Île-de-France Mobilités](https://prim.iledefrance-mobilites.fr/) API gateway.

<br>

## 🗺️ Covered lines

All **16 Paris metro lines**: M1 · M2 · M3 · M3B · M4 · M5 · M6 · M7 · M7B · M8 · M9 · M10 · M11 · M12 · M13 · M14

<br>

## 📄 License

[MIT](LICENSE)

<br>

---

<p align="center">
  <em>C'est dans combien le prochain ?</em>
</p>
