package openrgb

import (
	"fmt"
	"sync"

	sdk "github.com/csutorasa/go-openrgb-sdk"
)

// Manager manages connection to OpenRGB server
type Manager struct {
	mu     sync.Mutex
	client *sdk.Client
	host   string
	port   int
}

func NewManager() *Manager {
	return &Manager{}
}

func (m *Manager) Connect(host string, port int) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if m.client != nil {
		_ = m.client.Close()
		m.client = nil
	}

	client, err := sdk.NewClientHostPort(host, port)
	if err != nil {
		return fmt.Errorf("failed to connect to OpenRGB at %s:%d: %w", host, port, err)
	}

	if err := client.RequestProtocolVersion(); err != nil {
		_ = client.Close()
		return fmt.Errorf("protocol version negotiation failed: %w", err)
	}

	if err := client.Initialize("OpenRGB Batocera Server"); err != nil {
		_ = client.Close()
		return fmt.Errorf("client initialization failed: %w", err)
	}

	m.client = client
	m.host = host
	m.port = port
	return nil
}

func (m *Manager) Disconnect() {
	m.mu.Lock()
	defer m.mu.Unlock()

	if m.client != nil {
		_ = m.client.Close()
		m.client = nil
	}
	m.host = ""
	m.port = 0
}

func (m *Manager) IsConnected() bool {
	m.mu.Lock()
	defer m.mu.Unlock()
	return m.client != nil
}

func (m *Manager) GetStatus() (host string, port int, connected bool) {
	m.mu.Lock()
	defer m.mu.Unlock()
	return m.host, m.port, m.client != nil
}

func (m *Manager) GetDeviceCount() (uint32, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	if m.client == nil {
		return 0, fmt.Errorf("not connected")
	}

	resp, err := m.client.RequestControllerCount()
	if err != nil {
		return 0, err
	}
	return resp.Count, nil
}

func (m *Manager) GetDevice(id uint32) (*sdk.ControllerData, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	if m.client == nil {
		return nil, fmt.Errorf("not connected")
	}

	resp, err := m.client.RequestControllerData(id)
	if err != nil {
		return nil, fmt.Errorf("failed to get device %d: %w", id, err)
	}
	return resp.Controller, nil
}

func (m *Manager) GetAllDevices() ([]*sdk.ControllerData, error) {
	count, err := m.GetDeviceCount()
	if err != nil {
		return nil, err
	}

	devices := make([]*sdk.ControllerData, 0, count)
	for i := uint32(0); i < count; i++ {
		device, err := m.GetDevice(i)
		if err != nil {
			return nil, fmt.Errorf("failed to get device %d: %w", i, err)
		}
		devices = append(devices, device)
	}
	return devices, nil
}

func (m *Manager) SetDeviceColors(deviceId uint32, colors []sdk.Color) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if m.client == nil {
		return fmt.Errorf("not connected")
	}

	if err := m.client.RGBControllerSetCustomMode(deviceId); err != nil {
		return fmt.Errorf("failed to set custom mode: %w", err)
	}

	return m.client.RGBControllerUpdateLeds(deviceId, &sdk.RGBControllerUpdateLedsRequest{
		LedColor: colors,
	})
}

func (m *Manager) SetZoneColors(deviceId uint32, zoneId uint32, colors []sdk.Color) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if m.client == nil {
		return fmt.Errorf("not connected")
	}

	if err := m.client.RGBControllerSetCustomMode(deviceId); err != nil {
		return fmt.Errorf("failed to set custom mode: %w", err)
	}

	return m.client.RGBControllerUpdateZoneLeds(deviceId, &sdk.RGBControllerUpdateZoneLedsRequest{
		ZoneIdx:  zoneId,
		LedColor: colors,
	})
}

func (m *Manager) SetMode(deviceId uint32, modeId int32, device *sdk.ControllerData) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if m.client == nil {
		return fmt.Errorf("not connected")
	}

	if int(modeId) >= len(device.Modes) {
		return fmt.Errorf("invalid mode ID %d", modeId)
	}

	return m.client.RGBControllerUpdateMode(deviceId, &sdk.RGBControllerUpdateModeRequest{
		ModeIdx: modeId,
		Mode:    device.Modes[modeId],
	})
}
