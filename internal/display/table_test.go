package display

import (
	"strings"
	"testing"
	"time"

	"github.com/cyrilghali/metro-cli/internal/model"
)

func TestParseNavitiaTime(t *testing.T) {
	got, err := ParseNavitiaTime("20260225T143000")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got.Hour() != 14 || got.Minute() != 30 || got.Second() != 0 {
		t.Errorf("got %v, want 14:30:00", got)
	}
	if got.Year() != 2026 || got.Month() != 2 || got.Day() != 25 {
		t.Errorf("got %v, want 2026-02-25", got)
	}
}

func TestParseNavitiaTimeInvalid(t *testing.T) {
	_, err := ParseNavitiaTime("not-a-date")
	if err == nil {
		t.Error("expected error for invalid input")
	}
}

func TestFormatMinutesUntil(t *testing.T) {
	// "now" for past times
	past := time.Now().Add(-1 * time.Minute)
	got := FormatMinutesUntil(past)
	if !strings.Contains(got, "now") {
		t.Errorf("expected 'now' for past time, got %q", got)
	}

	// "1 min"
	oneMin := time.Now().Add(1*time.Minute + 10*time.Second)
	got = FormatMinutesUntil(oneMin)
	if !strings.Contains(got, "1 min") {
		t.Errorf("expected '1 min', got %q", got)
	}

	// "N min" for normal range
	fiveMin := time.Now().Add(5*time.Minute + 10*time.Second)
	got = FormatMinutesUntil(fiveMin)
	if !strings.Contains(got, "5 min") {
		t.Errorf("expected '5 min', got %q", got)
	}

	// "~Xh" format for >= 90 min
	twoHours := time.Now().Add(2*time.Hour + 30*time.Minute)
	got = FormatMinutesUntil(twoHours)
	if !strings.Contains(got, "~2h30") {
		t.Errorf("expected '~2h30', got %q", got)
	}
}

func TestStripHTMLTags(t *testing.T) {
	tests := []struct {
		input, want string
	}{
		{"hello", "hello"},
		{"<p>hello</p>", "hello"},
		{"<b>bold</b> and <i>italic</i>", "bold and italic"},
		{"no tags here", "no tags here"},
		{"<div><span>nested</span></div>", "nested"},
		{"", ""},
	}
	for _, tt := range tests {
		got := stripHTMLTags(tt.input)
		if got != tt.want {
			t.Errorf("stripHTMLTags(%q) = %q, want %q", tt.input, got, tt.want)
		}
	}
}

func TestExtractMessage(t *testing.T) {
	// Prefers text/plain
	d := model.Disruption{
		Messages: []model.Message{
			{Text: "<p>html version</p>", Channel: model.Channel{ContentType: "text/html"}},
			{Text: "plain version", Channel: model.Channel{ContentType: "text/plain"}},
		},
	}
	got := extractMessage(d)
	if got != "plain version" {
		t.Errorf("expected 'plain version', got %q", got)
	}

	// Falls back to first message with HTML stripped
	d2 := model.Disruption{
		Messages: []model.Message{
			{Text: "<p>only html</p>", Channel: model.Channel{ContentType: "text/html"}},
		},
	}
	got = extractMessage(d2)
	if got != "only html" {
		t.Errorf("expected 'only html', got %q", got)
	}

	// Falls back to cause
	d3 := model.Disruption{Cause: "travaux"}
	got = extractMessage(d3)
	if got != "travaux" {
		t.Errorf("expected 'travaux', got %q", got)
	}

	// Empty
	got = extractMessage(model.Disruption{})
	if got != "" {
		t.Errorf("expected empty, got %q", got)
	}
}

func TestFormatSeverity(t *testing.T) {
	tests := []struct {
		effect string
		want   string
	}{
		{"NO_SERVICE", "Interrupted"},
		{"REDUCED_SERVICE", "Reduced"},
		{"SIGNIFICANT_DELAYS", "Delays"},
		{"MODIFIED_SERVICE", "Modified"},
		{"ADDITIONAL_SERVICE", "Extra"},
		{"UNKNOWN_EFFECT", "Info"},
	}
	for _, tt := range tests {
		got := formatSeverity(model.Severity{Effect: tt.effect})
		if !strings.Contains(got, tt.want) {
			t.Errorf("formatSeverity(%q) = %q, want it to contain %q", tt.effect, got, tt.want)
		}
	}

	// Unknown effect with name
	got := formatSeverity(model.Severity{Effect: "CUSTOM", Name: "Custom"})
	if !strings.Contains(got, "Custom") {
		t.Errorf("expected 'Custom' in output, got %q", got)
	}

	// Unknown effect without name
	got = formatSeverity(model.Severity{Effect: "SOMETHING"})
	if !strings.Contains(got, "SOMETHING") {
		t.Errorf("expected 'SOMETHING' in output, got %q", got)
	}
}

func TestMatchesLineFilter(t *testing.T) {
	tests := []struct {
		code, label, filter string
		want                bool
	}{
		{"14", "M14", "M14", true},
		{"14", "M14", "m14", true},
		{"14", "M14", "14", true},
		{"A", "RER A", "RER A", true},
		{"A", "RER A", "RERA", true},
		{"3a", "T3a", "T3a", true},
		{"14", "M14", "M7", false},
		{"14", "M14", "A", false},
	}
	for _, tt := range tests {
		got := matchesLineFilter(tt.code, tt.label, tt.filter)
		if got != tt.want {
			t.Errorf("matchesLineFilter(%q, %q, %q) = %v, want %v",
				tt.code, tt.label, tt.filter, got, tt.want)
		}
	}
}

func TestTruncate(t *testing.T) {
	tests := []struct {
		input string
		max   int
		want  string
	}{
		{"short", 10, "short"},
		{"exactly ten", 11, "exactly ten"},
		{"this is a long string", 10, "this is..."},
		// French accented characters — must not split mid-rune
		{"Château de Vincennes direction", 15, "Château de V..."},
		{"Châtelet", 30, "Châtelet"},
		{"", 5, ""},
	}
	for _, tt := range tests {
		got := truncate(tt.input, tt.max)
		if got != tt.want {
			t.Errorf("truncate(%q, %d) = %q, want %q", tt.input, tt.max, got, tt.want)
		}
	}
}
