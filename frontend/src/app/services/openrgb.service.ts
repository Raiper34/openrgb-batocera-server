import { Injectable } from '@angular/core';
import { HttpClient } from '@angular/common/http';
import { Observable } from 'rxjs';
import {
  RGBDevice,
  SetColorRequest,
  SetZoneColorRequest,
  SetModeRequest,
  ConnectionStatus,
  ConnectRequest
} from '../models/openrgb.models';

@Injectable({
  providedIn: 'root'
})
export class OpenRgbService {
  private readonly baseUrl = '/api';

  constructor(private http: HttpClient) {}

  getStatus(): Observable<ConnectionStatus> {
    return this.http.get<ConnectionStatus>(`${this.baseUrl}/status`);
  }

  connect(request: ConnectRequest): Observable<ConnectionStatus> {
    return this.http.post<ConnectionStatus>(`${this.baseUrl}/connect`, request);
  }

  disconnect(): Observable<ConnectionStatus> {
    return this.http.post<ConnectionStatus>(`${this.baseUrl}/disconnect`, {});
  }

  getDevices(): Observable<RGBDevice[]> {
    return this.http.get<RGBDevice[]>(`${this.baseUrl}/devices`);
  }

  getDevice(id: number): Observable<RGBDevice> {
    return this.http.get<RGBDevice>(`${this.baseUrl}/devices/${id}`);
  }

  setDeviceColors(request: SetColorRequest): Observable<void> {
    return this.http.post<void>(`${this.baseUrl}/devices/${request.device_id}/colors`, request);
  }

  setZoneColors(request: SetZoneColorRequest): Observable<void> {
    return this.http.post<void>(
      `${this.baseUrl}/devices/${request.device_id}/zones/${request.zone_id}/colors`,
      request
    );
  }

  setMode(request: SetModeRequest): Observable<void> {
    return this.http.post<void>(
      `${this.baseUrl}/devices/${request.device_id}/mode`,
      request
    );
  }

  setAllColors(color: { r: number; g: number; b: number }): Observable<void> {
    return this.http.post<void>(`${this.baseUrl}/all-colors`, color);
  }
}
