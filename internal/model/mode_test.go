package model

import "testing"

func TestParseMode(t *testing.T) {
	tests := []struct {
		input   string
		want    string
		wantErr bool
	}{
		{"metro", "metro", false},
		{"rer", "rer", false},
		{"train", "train", false},
		{"tram", "tram", false},
		{"bus", "bus", false},
		{"all", "all", false},
		{"METRO", "metro", false},
		{"  RER  ", "rer", false},
		{"unknown", "", true},
		{"", "", true},
	}
	for _, tt := range tests {
		m, err := ParseMode(tt.input)
		if tt.wantErr {
			if err == nil {
				t.Errorf("ParseMode(%q) expected error, got %q", tt.input, m.Name)
			}
			continue
		}
		if err != nil {
			t.Errorf("ParseMode(%q) unexpected error: %v", tt.input, err)
			continue
		}
		if m.Name != tt.want {
			t.Errorf("ParseMode(%q) = %q, want %q", tt.input, m.Name, tt.want)
		}
	}
}

func TestIsAll(t *testing.T) {
	all, _ := ParseMode("all")
	if !all.IsAll() {
		t.Error("expected IsAll() = true for 'all' mode")
	}
	metro, _ := ParseMode("metro")
	if metro.IsAll() {
		t.Error("expected IsAll() = false for 'metro' mode")
	}
}

func TestLineLabel(t *testing.T) {
	tests := []struct {
		code, mode, want string
	}{
		{"1", "Metro", "M1"},
		{"14", "Métro", "M14"},
		{"14", "métro", "M14"},
		{"14", "METRO", "M14"},
		{"A", "RER", "RER A"},
		{"A", "rer", "RER A"},
		{"3a", "Tramway", "T3a"},
		{"3a", "tramway", "T3a"},
		{"27", "Bus", "27"},
		{"N15", "", "N15"},
	}
	for _, tt := range tests {
		got := LineLabel(tt.code, tt.mode)
		if got != tt.want {
			t.Errorf("LineLabel(%q, %q) = %q, want %q", tt.code, tt.mode, got, tt.want)
		}
	}
}

func TestParseModeFilters(t *testing.T) {
	// Ensure known modes have non-empty filters
	for _, name := range ModeNames {
		m, err := ParseMode(name)
		if err != nil {
			t.Fatalf("ParseMode(%q) error: %v", name, err)
		}
		if m.Filter == "" {
			t.Errorf("mode %q has empty filter", name)
		}
	}
	// "all" should have empty filter
	all, _ := ParseMode("all")
	if all.Filter != "" {
		t.Errorf("expected empty filter for 'all', got %q", all.Filter)
	}
}
