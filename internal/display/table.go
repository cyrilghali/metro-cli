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

// Departures prints next departures grouped by line+direction for a station.
func Departures(stationName string, deps []model.Departure) {
	if len(deps) == 0 {
		fmt.Printf("  %s(no upcoming departures)%s\n", dim, reset)
		return
	}

	type key struct {
		lineCode  string
		direction string
	}
	type entry struct {
		times []string
	}
	groups := make(map[key]*entry)
	var order []key

	for _, d := range deps {
		k := key{
			lineCode:  d.DisplayInformations.Code,
			direction: d.DisplayInformations.Direction,
		}
		if _, ok := groups[k]; !ok {
			groups[k] = &entry{}
			order = append(order, k)
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
		lineLabel := fmt.Sprintf("%sM%s%s", bold, k.lineCode, reset)

		dir := k.direction
		if len(dir) > 30 {
			dir = dir[:27] + "..."
		}

		timesStr := strings.Join(e.times, ", ")
		fmt.Fprintf(w, "  %s\t%s\t%s\n", lineLabel, dir, timesStr)
	}
	w.Flush()
}

// DisruptionsSummary prints disruption status for each metro line.
func DisruptionsSummary(resp *model.LinesResponse, filterLine string) {
	if resp == nil || len(resp.Lines) == 0 {
		fmt.Printf("%sNo metro lines found.%s\n", dim, reset)
		return
	}

	// Index disruptions by ID
	dmap := make(map[string]*model.Disruption)
	for i := range resp.Disruptions {
		d := &resp.Disruptions[i]
		dmap[d.ID] = d
	}

	// Each line has links to disruptions via impacted_objects
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
		if filterLine != "" && !strings.EqualFold(code, filterLine) && !strings.EqualFold("M"+code, filterLine) {
			continue
		}

		lineLabel := fmt.Sprintf("%sM%-3s%s", bold, code, reset)
		disruptions := lineDisruptions[line.ID]

		if len(disruptions) == 0 {
			fmt.Fprintf(w, "%s\t%sOK%s\t\n", lineLabel, green, reset)
		} else {
			for i, d := range disruptions {
				prefix := lineLabel
				if i > 0 {
					prefix = "    "
				}
				status := formatSeverity(d.Severity)
				msg := extractMessage(*d)
				if len(msg) > 70 {
					msg = msg[:67] + "..."
				}
				fmt.Fprintf(w, "%s\t%s\t%s\n", prefix, status, msg)
			}
		}
	}
	w.Flush()
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
		// Strip HTML tags for web messages
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
