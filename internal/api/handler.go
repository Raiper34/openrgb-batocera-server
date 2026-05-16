package api

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"strings"

	sdk "github.com/csutorasa/go-openrgb-sdk"

	"github.com/raiper34/openrgb-batocera-server/internal/openrgb"
	"github.com/raiper34/openrgb-batocera-server/internal/state"
)

// Handler handles HTTP API requests
type Handler struct {
	manager *openrgb.Manager
	state   *state.State
	envHost string // non-empty when OPENRGB_HOST env var is set
}

func NewHandler(manager *openrgb.Manager, s *state.State, envHost string) *Handler {
	return &Handler{manager: manager, state: s, envHost: envHost}
}

func (h *Handler) RegisterRoutes(mux *http.ServeMux) {
	mux.HandleFunc("GET /api/status", h.getStatus)
	mux.HandleFunc("POST /api/connect", h.connect)
	mux.HandleFunc("POST /api/disconnect", h.disconnect)
	mux.HandleFunc("GET /api/devices", h.getDevices)
	mux.HandleFunc("GET /api/devices/{id}", h.getDevice)
	mux.HandleFunc("POST /api/devices/{id}/colors", h.setDeviceColors)
	mux.HandleFunc("POST /api/devices/{id}/zones/{zone_id}/colors", h.setZoneColors)
	mux.HandleFunc("POST /api/devices/{id}/mode", h.setMode)
	mux.HandleFunc("POST /api/all-colors", h.setAllColors)
}

func (h *Handler) getStatus(w http.ResponseWriter, r *http.Request) {
	host, port, connected := h.manager.GetStatus()
	resp := ConnectionStatus{
		Connected: connected,
		Host:      host,
		Port:      port,
	}
	if saved := h.state.GetConnection(); saved != nil {
		resp.SavedHost = saved.Host
		resp.SavedPort = saved.Port
	}
	if h.envHost != "" {
		resp.EnvHost = h.envHost
		resp.FromEnv = true
	}
	respondJSON(w, http.StatusOK, resp)
}

func (h *Handler) connect(w http.ResponseWriter, r *http.Request) {
	var req ConnectRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	if req.Host == "" {
		req.Host = "localhost"
	}
	if req.Port == 0 {
		req.Port = 6742
	}

	if err := h.manager.Connect(req.Host, req.Port); err != nil {
		respondJSON(w, http.StatusOK, ConnectionStatus{
			Connected: false,
			Host:      req.Host,
			Port:      req.Port,
			Error:     err.Error(),
		})
		return
	}

	h.state.SetConnection(req.Host, req.Port)

	respondJSON(w, http.StatusOK, ConnectionStatus{
		Connected: true,
		Host:      req.Host,
		Port:      req.Port,
	})
}

func (h *Handler) disconnect(w http.ResponseWriter, r *http.Request) {
	h.manager.Disconnect()
	h.state.ClearConnection()
	respondJSON(w, http.StatusOK, ConnectionStatus{
		Connected: false,
	})
}

func (h *Handler) getDevices(w http.ResponseWriter, r *http.Request) {
	if !h.manager.IsConnected() {
		respondError(w, http.StatusServiceUnavailable, "not connected to OpenRGB")
		return
	}

	sdkDevices, err := h.manager.GetAllDevices()
	if err != nil {
		respondError(w, http.StatusInternalServerError, err.Error())
		return
	}

	devices := make([]RGBDevice, len(sdkDevices))
	for i, d := range sdkDevices {
		devices[i] = convertDevice(i, d)
	}
	respondJSON(w, http.StatusOK, devices)
}

func (h *Handler) getDevice(w http.ResponseWriter, r *http.Request) {
	id, err := parsePathInt(r, "id")
	if err != nil {
		respondError(w, http.StatusBadRequest, "invalid device ID")
		return
	}

	d, err := h.manager.GetDevice(uint32(id))
	if err != nil {
		respondError(w, http.StatusInternalServerError, err.Error())
		return
	}

	respondJSON(w, http.StatusOK, convertDevice(id, d))
}

func (h *Handler) setDeviceColors(w http.ResponseWriter, r *http.Request) {
	id, err := parsePathInt(r, "id")
	if err != nil {
		respondError(w, http.StatusBadRequest, "invalid device ID")
		return
	}

	var req SetColorRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	colors := convertColors(req.Colors)
	if err := h.manager.SetDeviceColors(uint32(id), colors); err != nil {
		respondError(w, http.StatusInternalServerError, err.Error())
		return
	}

	// Persist: fetch device name then save colors
	if d, err := h.manager.GetDevice(uint32(id)); err == nil {
		h.state.SetDeviceColors(id, trimNull(d.Name), colorsToHex(req.Colors))
	}

	w.WriteHeader(http.StatusNoContent)
}

func (h *Handler) setZoneColors(w http.ResponseWriter, r *http.Request) {
	id, err := parsePathInt(r, "id")
	if err != nil {
		respondError(w, http.StatusBadRequest, "invalid device ID")
		return
	}

	zoneId, err := parsePathInt(r, "zone_id")
	if err != nil {
		respondError(w, http.StatusBadRequest, "invalid zone ID")
		return
	}

	var req SetZoneColorRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	colors := convertColors(req.Colors)
	if err := h.manager.SetZoneColors(uint32(id), uint32(zoneId), colors); err != nil {
		respondError(w, http.StatusInternalServerError, err.Error())
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func (h *Handler) setMode(w http.ResponseWriter, r *http.Request) {
	id, err := parsePathInt(r, "id")
	if err != nil {
		respondError(w, http.StatusBadRequest, "invalid device ID")
		return
	}

	var req SetModeRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	device, err := h.manager.GetDevice(uint32(id))
	if err != nil {
		respondError(w, http.StatusInternalServerError, err.Error())
		return
	}

	if err := h.manager.SetMode(uint32(id), req.ModeID, device); err != nil {
		respondError(w, http.StatusInternalServerError, err.Error())
		return
	}

	h.state.SetDeviceMode(id, trimNull(device.Name), int(req.ModeID))

	w.WriteHeader(http.StatusNoContent)
}

func (h *Handler) setAllColors(w http.ResponseWriter, r *http.Request) {
	var req SetAllColorsRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	count, err := h.manager.GetDeviceCount()
	if err != nil {
		respondError(w, http.StatusInternalServerError, err.Error())
		return
	}

	color := sdk.Color{R: req.R, G: req.G, B: req.B}
	hexColor := fmt.Sprintf("%02X%02X%02X", req.R, req.G, req.B)

	var bulk []state.DeviceWithColors
	for i := uint32(0); i < count; i++ {
		device, err := h.manager.GetDevice(i)
		if err != nil {
			continue
		}

		colors := make([]sdk.Color, len(device.Leds))
		hexColors := make([]string, len(device.Leds))
		for j := range colors {
			colors[j] = color
			hexColors[j] = hexColor
		}

		_ = h.manager.SetDeviceColors(i, colors)
		bulk = append(bulk, state.DeviceWithColors{
			Index:  int(i),
			Name:   trimNull(device.Name),
			Colors: hexColors,
		})
	}

	if len(bulk) > 0 {
		h.state.SetAllDeviceColors(bulk)
	}

	w.WriteHeader(http.StatusNoContent)
}

// Helpers

func respondJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(v)
}

func respondError(w http.ResponseWriter, status int, msg string) {
	respondJSON(w, status, ErrorResponse{Error: msg})
}

func parsePathInt(r *http.Request, key string) (int, error) {
	return strconv.Atoi(r.PathValue(key))
}

func convertColors(colors []RGBColor) []sdk.Color {
	result := make([]sdk.Color, len(colors))
	for i, c := range colors {
		result[i] = sdk.Color{R: c.R, G: c.G, B: c.B}
	}
	return result
}

func colorsToHex(colors []RGBColor) []string {
	result := make([]string, len(colors))
	for i, c := range colors {
		result[i] = fmt.Sprintf("%02X%02X%02X", c.R, c.G, c.B)
	}
	return result
}

func trimNull(s string) string {
	return strings.TrimRight(s, "\x00")
}

func convertDevice(id int, d *sdk.ControllerData) RGBDevice {
	modes := make([]RGBMode, len(d.Modes))
	for i, m := range d.Modes {
		modeColors := make([]RGBColor, len(m.ModeColors))
		for j, mc := range m.ModeColors {
			modeColors[j] = RGBColor{R: mc.R, G: mc.G, B: mc.B}
		}
		modes[i] = RGBMode{
			Name:          trimNull(m.ModeName),
			Value:         m.ModeValue,
			Flags:         m.ModeFlags,
			SpeedMin:      m.ModeSpeedMin,
			SpeedMax:      m.ModeSpeedMax,
			BrightnessMin: m.ModeBrightnessMin,
			BrightnessMax: m.ModeBrightnessMax,
			ColorsMin:     m.ModeColorsMin,
			ColorsMax:     m.ModeColorsMax,
			Speed:         m.ModeSpeed,
			Brightness:    m.ModeBrightness,
			Direction:     m.ModeDirection,
			ColorMode:     m.ModeColorMode,
			Colors:        modeColors,
		}
	}

	zones := make([]RGBZone, len(d.Zones))
	for i, z := range d.Zones {
		matrixH := uint32(0)
		matrixW := uint32(0)
		if len(z.ZoneMatrixData) > 0 {
			matrixH = uint32(len(z.ZoneMatrixData))
			if len(z.ZoneMatrixData[0]) > 0 {
				matrixW = uint32(len(z.ZoneMatrixData[0]))
			}
		}
		zones[i] = RGBZone{
			Name:         trimNull(z.ZoneName),
			Type:         z.ZoneType,
			LedsMin:      z.ZoneLedsMin,
			LedsMax:      z.ZoneLedsMax,
			LedsCount:    z.ZoneLedsCount,
			MatrixHeight: matrixH,
			MatrixWidth:  matrixW,
		}
	}

	leds := make([]RGBLed, len(d.Leds))
	for i, l := range d.Leds {
		leds[i] = RGBLed{
			Name:  trimNull(l.LedName),
			Value: l.LedValue,
		}
	}

	colors := make([]RGBColor, len(d.Colors))
	for i, c := range d.Colors {
		colors[i] = RGBColor{R: c.R, G: c.G, B: c.B}
	}

	return RGBDevice{
		ID:          id,
		Type:        d.Type,
		Name:        trimNull(d.Name),
		Vendor:      trimNull(d.Vendor),
		Description: trimNull(d.Description),
		Version:     trimNull(d.Version),
		Serial:      trimNull(d.Serial),
		Location:    trimNull(d.Location),
		ActiveMode:  d.ActiveMode,
		Modes:       modes,
		Zones:       zones,
		Leds:        leds,
		Colors:      colors,
	}
}
