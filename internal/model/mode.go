package model

import (
	"fmt"
	"strings"
)

// TransportMode represents a supported public transport mode.
type TransportMode struct {
	Name     string // user-facing name ("metro", "rer", ...)
	Filter   string // Navitia physical_mode filter value
	Prefix   string // line label prefix ("M", "RER ", "T", ...)
	MaxLines int    // expected count for API pagination
}

var Modes = map[string]TransportMode{
	"metro": {
		Name:     "metro",
		Filter:   "physical_mode.id=physical_mode:Metro",
		Prefix:   "M",
		MaxLines: 20,
	},
	"rer": {
		Name:     "rer",
		Filter:   "physical_mode.id=physical_mode:RapidTransit",
		Prefix:   "RER ",
		MaxLines: 10,
	},
	"train": {
		Name:     "train",
		Filter:   "physical_mode.id=physical_mode:LocalTrain",
		Prefix:   "",
		MaxLines: 30,
	},
	"tram": {
		Name:     "tram",
		Filter:   "physical_mode.id=physical_mode:Tramway",
		Prefix:   "T",
		MaxLines: 20,
	},
	"bus": {
		Name:     "bus",
		Filter:   "physical_mode.id=physical_mode:Bus",
		Prefix:   "",
		MaxLines: 100,
	},
}

// ModeNames returns sorted mode names for help text.
var ModeNames = []string{"metro", "rer", "train", "tram", "bus"}

// AllFilter returns a combined OR filter for all supported modes.
// Navitia doesn't support OR filters directly, so "all" means no filter.
const AllFilter = ""

// ParseMode validates a mode string and returns the filter(s) to use.
// Returns ("", true) for "all" meaning no filter should be applied.
func ParseMode(s string) (TransportMode, error) {
	s = strings.ToLower(strings.TrimSpace(s))
	if s == "all" {
		return TransportMode{Name: "all"}, nil
	}
	m, ok := Modes[s]
	if !ok {
		return TransportMode{}, fmt.Errorf("unknown mode %q (valid: %s, all)", s, strings.Join(ModeNames, ", "))
	}
	return m, nil
}

// IsAll returns true if this is the "all modes" wildcard.
func (m TransportMode) IsAll() bool {
	return m.Name == "all"
}

// LineLabel returns a formatted line label like "M1", "RER A", "T3a".
func LineLabel(code string, commercialMode string) string {
	switch strings.ToLower(commercialMode) {
	case "metro", "métro":
		return "M" + code
	case "rer":
		return "RER " + code
	case "tramway":
		return "T" + code
	default:
		return code
	}
}
