export interface RGBColor {
  r: number;
  g: number;
  b: number;
}

export interface RGBLed {
  name: string;
  value: RGBColor;
}

export interface RGBZone {
  name: string;
  type: number;
  leds_min: number;
  leds_max: number;
  leds_count: number;
  matrix_height: number;
  matrix_width: number;
}

export interface RGBMode {
  name: string;
  value: number;
  flags: number;
  speed_min: number;
  speed_max: number;
  brightness_min: number;
  brightness_max: number;
  colors_min: number;
  colors_max: number;
  speed: number;
  brightness: number;
  direction: number;
  color_mode: number;
  colors: RGBColor[];
}

export interface RGBDevice {
  id: number;
  type: number;
  name: string;
  vendor: string;
  description: string;
  version: string;
  serial: string;
  location: string;
  active_mode: number;
  modes: RGBMode[];
  zones: RGBZone[];
  leds: RGBLed[];
  colors: RGBColor[];
}

export interface SetColorRequest {
  device_id: number;
  colors: RGBColor[];
}

export interface SetZoneColorRequest {
  device_id: number;
  zone_id: number;
  colors: RGBColor[];
}

export interface SetModeRequest {
  device_id: number;
  mode_id: number;
}

export interface ConnectionStatus {
  connected: boolean;
  host: string;
  port: number;
  error?: string;
  saved_host?: string;
  saved_port?: number;
  /** Set when OPENRGB_HOST env var is configured on the server. */
  env_host?: string;
  /** True when the server was started with OPENRGB_HOST env var. */
  from_env?: boolean;
}

export interface ConnectRequest {
  host: string;
  port: number;
}
