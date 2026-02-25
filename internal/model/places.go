package model

// --- PRIM /marketplace/places response (custom format) ---

type PRIMPlacesResponse struct {
	Places []PRIMPlace `json:"places"`
}

type PRIMPlace struct {
	ID      string     `json:"id"`
	Name    string     `json:"name"`
	Type    string     `json:"type"` // "StopArea", "Address", "City"
	City    string     `json:"city"`
	ZipCode string     `json:"zipCode"`
	X       float64    `json:"x"`
	Y       float64    `json:"y"`
	Modes   []string   `json:"modes,omitempty"`
	Lines   []PRIMLine `json:"lines,omitempty"`
}

type PRIMLine struct {
	ID        string     `json:"id"`
	ShortName string     `json:"shortName"`
	Color     string     `json:"color"`
	TextColor string     `json:"textColor"`
	Mode      []PRIMMode `json:"mode"`
}

type PRIMMode struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

// --- Navitia /places response (standard Navitia format) ---

type NavitiaPlacesResponse struct {
	Places []NavitiaPlace `json:"places"`
}

type NavitiaPlace struct {
	ID           string    `json:"id"`
	Name         string    `json:"name"`
	EmbeddedType string    `json:"embedded_type"`
	Quality      int       `json:"quality"`
	StopArea     *StopArea `json:"stop_area,omitempty"`
	Address      *Address  `json:"address,omitempty"`
}

type Address struct {
	ID    string `json:"id"`
	Name  string `json:"name"`
	Coord Coord  `json:"coord"`
}

// --- Navitia places_nearby response ---

type PlacesNearbyResponse struct {
	PlacesNearby []PlaceNearby `json:"places_nearby"`
}

type PlaceNearby struct {
	ID           string     `json:"id"`
	Name         string     `json:"name"`
	EmbeddedType string     `json:"embedded_type"`
	Quality      int        `json:"quality"`
	Distance     string     `json:"distance"`
	StopArea     *StopArea  `json:"stop_area,omitempty"`
	StopPoint    *StopPoint `json:"stop_point,omitempty"`
}

// --- Shared Navitia types ---

type StopArea struct {
	ID    string `json:"id"`
	Name  string `json:"name"`
	Coord Coord  `json:"coord"`
	Lines []Line `json:"lines,omitempty"`
	Codes []Code `json:"codes,omitempty"`
}

type StopPoint struct {
	ID       string    `json:"id"`
	Name     string    `json:"name"`
	Coord    Coord     `json:"coord"`
	StopArea *StopArea `json:"stop_area,omitempty"`
	Lines    []Line    `json:"lines,omitempty"`
}

type Coord struct {
	Lon string `json:"lon"`
	Lat string `json:"lat"`
}

type Code struct {
	Type  string `json:"type"`
	Value string `json:"value"`
}
