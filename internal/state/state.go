package state

import (
	"encoding/json"
	"log"
	"os"
	"sync"
)

// DeviceState holds the persisted state for a single device.
type DeviceState struct {
	Index  int      `json:"index"`
	Name   string   `json:"name"`
	ModeID int      `json:"mode_id"`
	Colors []string `json:"colors"` // hex strings, e.g. "FF0000"
}

// State is the full persisted application state.
type State struct {
	mu   sync.Mutex
	path string

	Connection *ConnectionState `json:"connection,omitempty"`
	Devices    []DeviceState    `json:"devices,omitempty"`
}

// ConnectionState holds the last-used OpenRGB connection info.
type ConnectionState struct {
	Host string `json:"host"`
	Port int    `json:"port"`
}

// New creates a State backed by the given file path.
// If the file exists it is loaded immediately.
func New(path string) *State {
	s := &State{path: path}
	if err := s.load(); err != nil && !os.IsNotExist(err) {
		log.Printf("state: could not load %s: %v", path, err)
	}
	return s
}

// GetConnection returns the saved connection or nil if none.
func (s *State) GetConnection() *ConnectionState {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.Connection == nil {
		return nil
	}
	c := *s.Connection
	return &c
}

// SetConnection saves the connection info and persists it.
func (s *State) SetConnection(host string, port int) {
	s.mu.Lock()
	s.Connection = &ConnectionState{Host: host, Port: port}
	s.mu.Unlock()
	s.save()
}

// ClearConnection removes the saved connection.
func (s *State) ClearConnection() {
	s.mu.Lock()
	s.Connection = nil
	s.mu.Unlock()
	s.save()
}

// GetDevices returns a copy of the saved device states.
func (s *State) GetDevices() []DeviceState {
	s.mu.Lock()
	defer s.mu.Unlock()
	out := make([]DeviceState, len(s.Devices))
	copy(out, s.Devices)
	return out
}

// SetDeviceColors updates the colors for a device (matched by index) and persists.
func (s *State) SetDeviceColors(index int, name string, colors []string) {
	s.mu.Lock()
	s.ensureDevice(index, name)
	s.Devices[s.findDevice(index)].Colors = colors
	s.mu.Unlock()
	s.save()
}

// SetDeviceMode updates the mode for a device (matched by index) and persists.
func (s *State) SetDeviceMode(index int, name string, modeID int) {
	s.mu.Lock()
	s.ensureDevice(index, name)
	s.Devices[s.findDevice(index)].ModeID = modeID
	s.mu.Unlock()
	s.save()
}

// SetAllDeviceColors sets the same color on all devices and persists.
func (s *State) SetAllDeviceColors(devices []DeviceWithColors) {
	s.mu.Lock()
	for _, d := range devices {
		s.ensureDevice(d.Index, d.Name)
		s.Devices[s.findDevice(d.Index)].Colors = d.Colors
	}
	s.mu.Unlock()
	s.save()
}

// DeviceWithColors is a helper for bulk updates.
type DeviceWithColors struct {
	Index  int
	Name   string
	Colors []string
}

// --- internal helpers (must be called with mu held) ---

func (s *State) findDevice(index int) int {
	for i, d := range s.Devices {
		if d.Index == index {
			return i
		}
	}
	return -1
}

func (s *State) ensureDevice(index int, name string) {
	if s.findDevice(index) == -1 {
		s.Devices = append(s.Devices, DeviceState{Index: index, Name: name})
	}
}

// --- persistence ---

func (s *State) load() error {
	data, err := os.ReadFile(s.path)
	if err != nil {
		return err
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	return json.Unmarshal(data, s)
}

func (s *State) save() {
	s.mu.Lock()
	data, err := json.MarshalIndent(s, "", "  ")
	s.mu.Unlock()
	if err != nil {
		log.Printf("state: marshal error: %v", err)
		return
	}
	if err := os.WriteFile(s.path, data, 0644); err != nil {
		log.Printf("state: write error: %v", err)
	}
}
