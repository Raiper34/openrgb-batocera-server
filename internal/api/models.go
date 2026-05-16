package api

// RGBColor represents an RGB color
type RGBColor struct {
	R uint8 `json:"r"`
	G uint8 `json:"g"`
	B uint8 `json:"b"`
}

// RGBLed represents a single LED
type RGBLed struct {
	Name  string `json:"name"`
	Value uint32 `json:"value"`
}

// RGBZone represents a lighting zone
type RGBZone struct {
	Name         string `json:"name"`
	Type         int32  `json:"type"`
	LedsMin      uint32 `json:"leds_min"`
	LedsMax      uint32 `json:"leds_max"`
	LedsCount    uint32 `json:"leds_count"`
	MatrixHeight uint32 `json:"matrix_height"`
	MatrixWidth  uint32 `json:"matrix_width"`
}

// RGBMode represents a lighting mode
type RGBMode struct {
	Name          string     `json:"name"`
	Value         int32      `json:"value"`
	Flags         uint32     `json:"flags"`
	SpeedMin      uint32     `json:"speed_min"`
	SpeedMax      uint32     `json:"speed_max"`
	BrightnessMin uint32     `json:"brightness_min"`
	BrightnessMax uint32     `json:"brightness_max"`
	ColorsMin     uint32     `json:"colors_min"`
	ColorsMax     uint32     `json:"colors_max"`
	Speed         uint32     `json:"speed"`
	Brightness    uint32     `json:"brightness"`
	Direction     uint32     `json:"direction"`
	ColorMode     uint32     `json:"color_mode"`
	Colors        []RGBColor `json:"colors"`
}

// RGBDevice represents an RGB device
type RGBDevice struct {
	ID          int         `json:"id"`
	Type        int32       `json:"type"`
	Name        string      `json:"name"`
	Vendor      string      `json:"vendor"`
	Description string      `json:"description"`
	Version     string      `json:"version"`
	Serial      string      `json:"serial"`
	Location    string      `json:"location"`
	ActiveMode  int32       `json:"active_mode"`
	Modes       []RGBMode   `json:"modes"`
	Zones       []RGBZone   `json:"zones"`
	Leds        []RGBLed    `json:"leds"`
	Colors      []RGBColor  `json:"colors"`
}

// SetColorRequest is the request body for setting device colors
type SetColorRequest struct {
	Colors []RGBColor `json:"colors"`
}

// SetZoneColorRequest is the request body for setting zone colors
type SetZoneColorRequest struct {
	Colors []RGBColor `json:"colors"`
}

// SetModeRequest is the request body for changing mode
type SetModeRequest struct {
	ModeID int32 `json:"mode_id"`
}

// ConnectionStatus represents the current connection state
type ConnectionStatus struct {
	Connected  bool   `json:"connected"`
	Host       string `json:"host"`
	Port       int    `json:"port"`
	Error      string `json:"error,omitempty"`
	SavedHost  string `json:"saved_host,omitempty"`
	SavedPort  int    `json:"saved_port,omitempty"`
	// EnvHost is set when OPENRGB_HOST env var is configured; warns the UI
	// that after restart the env var will override any manual connection change.
	EnvHost string `json:"env_host,omitempty"`
	FromEnv bool   `json:"from_env,omitempty"`
}

// ConnectRequest is the request body for connecting to OpenRGB
type ConnectRequest struct {
	Host string `json:"host"`
	Port int    `json:"port"`
}

// SetAllColorsRequest sets all devices to one color
type SetAllColorsRequest struct {
	R uint8 `json:"r"`
	G uint8 `json:"g"`
	B uint8 `json:"b"`
}

// ErrorResponse is a generic error response
type ErrorResponse struct {
	Error string `json:"error"`
}
