package model

type DeparturesResponse struct {
	Departures  []Departure  `json:"departures"`
	Disruptions []Disruption `json:"disruptions,omitempty"`
}

type Departure struct {
	DisplayInformations DisplayInfo `json:"display_informations"`
	StopPoint           StopPoint   `json:"stop_point"`
	StopDateTime        StopDateTime `json:"stop_date_time"`
	Route               Route       `json:"route"`
}

type DisplayInfo struct {
	Direction      string `json:"direction"`
	Code           string `json:"code"`
	Network        string `json:"network"`
	Color          string `json:"color"`
	TextColor      string `json:"text_color"`
	CommercialMode string `json:"commercial_mode"`
	Label          string `json:"label"`
	Name           string `json:"name"`
}

type StopDateTime struct {
	DepartureDateTime string `json:"departure_date_time"`
	ArrivalDateTime   string `json:"arrival_date_time"`
	BaseDateTime      string `json:"base_departure_date_time"`
	DataFreshness     string `json:"data_freshness"`
}

type Route struct {
	ID        string    `json:"id"`
	Name      string    `json:"name"`
	Direction Direction `json:"direction"`
	Line      *Line     `json:"line,omitempty"`
}

type Direction struct {
	ID           string    `json:"id"`
	Name         string    `json:"name"`
	EmbeddedType string    `json:"embedded_type"`
	StopArea     *StopArea `json:"stop_area,omitempty"`
}

type Line struct {
	ID             string `json:"id"`
	Name           string `json:"name"`
	Code           string `json:"code"`
	Color          string `json:"color"`
	TextColor      string `json:"text_color"`
	CommercialMode *Mode  `json:"commercial_mode,omitempty"`
	PhysicalModes  []Mode `json:"physical_modes,omitempty"`
}

type Mode struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}
