package model

// LinesResponse is returned by /lines endpoint with embedded disruptions.
type LinesResponse struct {
	Lines       []Line       `json:"lines"`
	Disruptions []Disruption `json:"disruptions"`
	Pagination  Pagination   `json:"pagination"`
}

type Disruption struct {
	ID                 string           `json:"id"`
	DisruptionID       string           `json:"disruption_id"`
	Status             string           `json:"status"`
	ApplicationPeriods []Period         `json:"application_periods"`
	Severity           Severity         `json:"severity"`
	Messages           []Message        `json:"messages"`
	ImpactedObjects    []ImpactedObject `json:"impacted_objects,omitempty"`
	Cause              string           `json:"cause"`
	Category           string           `json:"category,omitempty"`
	Tags               []string         `json:"tags,omitempty"`
}

type Severity struct {
	Name     string `json:"name"`
	Effect   string `json:"effect"`
	Color    string `json:"color"`
	Priority int    `json:"priority"`
}

type Period struct {
	Begin string `json:"begin"`
	End   string `json:"end"`
}

type Message struct {
	Text    string  `json:"text"`
	Channel Channel `json:"channel"`
}

type Channel struct {
	ID          string   `json:"id"`
	Name        string   `json:"name"`
	ContentType string   `json:"content_type"`
	Types       []string `json:"types,omitempty"`
}

type ImpactedObject struct {
	PTObject      PTObject       `json:"pt_object"`
	ImpactedStops []ImpactedStop `json:"impacted_stops,omitempty"`
}

type PTObject struct {
	ID           string `json:"id"`
	Name         string `json:"name"`
	EmbeddedType string `json:"embedded_type"`
}

type ImpactedStop struct {
	StopPoint     StopPoint `json:"stop_point"`
	BaseDeparture string    `json:"base_departure_date_time"`
	Cause         string    `json:"cause"`
}

type Pagination struct {
	TotalResult  int `json:"total_result"`
	StartPage    int `json:"start_page"`
	ItemsPerPage int `json:"items_per_page"`
	ItemsOnPage  int `json:"items_on_page"`
}
