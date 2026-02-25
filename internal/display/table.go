package display

import (
	"fmt"
	"math"
	"os"
	"strings"
	"text/tabwriter"
	"time"

	"github.com/cyrilghali/metro-cli/internal/model"
)

const (
	reset  = "\033[0m"
	bold   = "\033[1m"
	dim    = "\033[2m"
	red    = "\033[31m"
	green  = "\033[32m"
	yellow = "\033[33m"
	cyan   = "\033[36m"
)

// ParseNavitiaTime parses "20260225T143000" into time.Time.
func ParseNavitiaTime(s string) (time.Time, error) {
	return time.ParseInLocation("20060102T150405", s, time.Now().Location())
}

// FormatMinutesUntil returns "2 min", "now", "~2h30", etc.
func FormatMinutesUntil(t time.Time) string {
	diff := time.Until(t)
	mins := int(math.Round(diff.Minutes()))
	if mins <= 0 {
		return green + "now" + reset
	}
	if mins == 1 {
		return cyan + "1 min" + reset
	}
	if mins >= 90 {
		h := mins / 60
		m := mins % 60
		return fmt.Sprintf("%s~%dh%02d%s", dim, h, m, reset)
	}
	return fmt.Sprintf("%s%d min%s", cyan, mins, reset)
}

// truncate safely truncates a string to maxRunes, appending "..." if needed.
// Uses rune-level slicing so multi-byte UTF-8 characters are not split.
func truncate(s string, maxRunes int) string {
	runes := []rune(s)
	if len(runes) <= maxRunes {
		return s
	}
	return string(runes[:maxRunes-3]) + "..."
}

// lineLabel returns a formatted label like "M1", "RER A", "T3a".
func lineLabel(code, commercialMode string) string {
	label := model.LineLabel(code, commercialMode)
	return fmt.Sprintf("%s%s%s", bold, label, reset)
}

// Departures prints next departures grouped by line+direction, followed
// by any active disruptions affecting the displayed lines.
func Departures(deps []model.Departure, disruptions []model.Disruption, showMode bool) {
	if len(deps) == 0 {
		fmt.Printf("  %s(no upcoming departures)%s\n", dim, reset)
		return
	}

	type key struct {
		lineCode       string
		commercialMode string
		direction      string
	}
	type entry struct {
		times []string
	}
	groups := make(map[key]*entry)
	var order []key

	for _, d := range deps {
		k := key{
			lineCode:       d.DisplayInformations.Code,
			commercialMode: d.DisplayInformations.CommercialMode,
			direction:      d.DisplayInformations.Direction,
		}
		if _, ok := groups[k]; !ok {
			groups[k] = &entry{}
			order = append(order, k)
		}
		if len(groups[k].times) >= 3 {
			continue
		}
		t, err := ParseNavitiaTime(d.StopDateTime.DepartureDateTime)
		if err != nil {
			continue
		}
		groups[k].times = append(groups[k].times, FormatMinutesUntil(t))
	}

	w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
	fmt.Fprintf(w, "  %sLine\tDirection\tNext departures%s\n", bold, reset)
	fmt.Fprintf(w, "  %s----\t---------\t---------------%s\n", dim, reset)

	for _, k := range order {
		e := groups[k]
		label := lineLabel(k.lineCode, k.commercialMode)

		dir := truncate(k.direction, 30)

		timesStr := strings.Join(e.times, ", ")
		fmt.Fprintf(w, "  %s\t%s\t%s\n", label, dir, timesStr)
	}
	w.Flush()

	// Show active disruptions affecting the displayed lines
	showDepartureDisruptions(deps, disruptions)
}

// showDepartureDisruptions prints active disruptions for lines present in the departures.
func showDepartureDisruptions(deps []model.Departure, disruptions []model.Disruption) {
	if len(disruptions) == 0 {
		return
	}

	// Collect line IDs seen in departures
	lineIDs := make(map[string]bool)
	for _, d := range deps {
		if d.Route.Line != nil {
			lineIDs[d.Route.Line.ID] = true
		}
	}

	// Find active disruptions impacting those lines (deduplicate by disruption ID)
	type match struct {
		label      string
		disruption *model.Disruption
	}
	seen := make(map[string]bool)
	var matches []match
	for i, d := range disruptions {
		if d.Status != "active" {
			continue
		}
		for _, io := range d.ImpactedObjects {
			if !lineIDs[io.PTObject.ID] {
				continue
			}
			if seen[d.ID] {
				continue
			}
			seen[d.ID] = true
			// Build label from the impacted line name
			label := io.PTObject.Name
			// Try to find a better label from departures
			for _, dep := range deps {
				if dep.Route.Line != nil && dep.Route.Line.ID == io.PTObject.ID {
					label = model.LineLabel(dep.DisplayInformations.Code, dep.DisplayInformations.CommercialMode)
					break
				}
			}
			matches = append(matches, match{label: label, disruption: &disruptions[i]})
		}
	}

	if len(matches) == 0 {
		return
	}

	fmt.Println()
	for _, m := range matches {
		severity := formatSeverity(m.disruption.Severity)
		msg := truncate(extractMessage(*m.disruption), 80)
		fmt.Printf("  %s!%s %s%s%s  %s  %s\n", yellow, reset, bold, m.label, reset, severity, msg)
	}
}

// DisruptionsSummary prints disruption status for lines of a given mode.
func DisruptionsSummary(resp *model.LinesResponse, filterLine string, mode model.TransportMode) {
	if resp == nil || len(resp.Lines) == 0 {
		fmt.Printf("%sNo lines found.%s\n", dim, reset)
		return
	}

	// Index disruptions by ID
	dmap := make(map[string]*model.Disruption)
	for i := range resp.Disruptions {
		d := &resp.Disruptions[i]
		dmap[d.ID] = d
	}

	// Build map: line ID -> active disruptions
	lineDisruptions := make(map[string][]*model.Disruption)
	for _, d := range resp.Disruptions {
		if d.Status != "active" {
			continue
		}
		for _, io := range d.ImpactedObjects {
			lineDisruptions[io.PTObject.ID] = append(lineDisruptions[io.PTObject.ID], dmap[d.ID])
		}
	}

	w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
	fmt.Fprintf(w, "%sLine\tStatus\tInfo%s\n", bold, reset)
	fmt.Fprintf(w, "%s----\t------\t----%s\n", dim, reset)

	for _, line := range resp.Lines {
		code := line.Code
		lineLabel := mode.Prefix + code

		if filterLine != "" && !matchesLineFilter(code, lineLabel, filterLine) {
			continue
		}

		label := fmt.Sprintf("%s%-6s%s", bold, lineLabel, reset)
		disruptions := lineDisruptions[line.ID]

		if len(disruptions) == 0 {
			fmt.Fprintf(w, "%s\t%sOK%s\t\n", label, green, reset)
		} else {
			for i, d := range disruptions {
				prefix := label
				if i > 0 {
					prefix = "      "
				}
				status := formatSeverity(d.Severity)
				msg := truncate(extractMessage(*d), 70)
				fmt.Fprintf(w, "%s\t%s\t%s\n", prefix, status, msg)
			}
		}
	}
	w.Flush()
}

func matchesLineFilter(code, lineLabel, filter string) bool {
	f := strings.ToUpper(strings.TrimSpace(filter))
	return strings.EqualFold(code, f) ||
		strings.EqualFold(lineLabel, f) ||
		strings.EqualFold("M"+code, f) ||
		strings.EqualFold("RER "+code, f) ||
		strings.EqualFold("RER"+code, f) ||
		strings.EqualFold("T"+code, f)
}

func formatSeverity(s model.Severity) string {
	switch s.Effect {
	case "NO_SERVICE":
		return red + "Interrupted" + reset
	case "REDUCED_SERVICE":
		return yellow + "Reduced" + reset
	case "SIGNIFICANT_DELAYS":
		return yellow + "Delays" + reset
	case "MODIFIED_SERVICE":
		return yellow + "Modified" + reset
	case "ADDITIONAL_SERVICE":
		return green + "Extra" + reset
	case "UNKNOWN_EFFECT":
		return dim + "Info" + reset
	default:
		if s.Name != "" {
			return yellow + s.Name + reset
		}
		return dim + s.Effect + reset
	}
}

func extractMessage(d model.Disruption) string {
	for _, m := range d.Messages {
		if m.Channel.ContentType == "text/plain" {
			return m.Text
		}
	}
	if len(d.Messages) > 0 {
		text := d.Messages[0].Text
		text = stripHTMLTags(text)
		return text
	}
	if d.Cause != "" {
		return d.Cause
	}
	return ""
}

func stripHTMLTags(s string) string {
	var result strings.Builder
	inTag := false
	for _, r := range s {
		if r == '<' {
			inTag = true
			continue
		}
		if r == '>' {
			inTag = false
			continue
		}
		if !inTag {
			result.WriteRune(r)
		}
	}
	return result.String()
}
