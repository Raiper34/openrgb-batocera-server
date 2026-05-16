import { Component, OnInit, ChangeDetectorRef, DestroyRef, inject } from '@angular/core';
import { FormsModule } from '@angular/forms';
import { ButtonModule } from 'primeng/button';
import { InputTextModule } from 'primeng/inputtext';
import { InputNumberModule } from 'primeng/inputnumber';
import { ToastModule } from 'primeng/toast';
import { MessageService } from 'primeng/api';
import { ColorPickerModule } from 'primeng/colorpicker';
import { DividerModule } from 'primeng/divider';
import { DialogModule } from 'primeng/dialog';
import { MessageModule } from 'primeng/message';
import { CardModule } from 'primeng/card';
import { takeUntilDestroyed } from '@angular/core/rxjs-interop';

import { OpenRgbService } from '../../services/openrgb.service';
import { DeviceListComponent } from '../../components/device-list/device-list';
import { DeviceDetailComponent } from '../../components/device-detail/device-detail';
import { RGBDevice, RGBColor, ConnectionStatus } from '../../models/openrgb.models';
import { hexToRgb, rgbToHex } from '../../utils/color.utils';

@Component({
  selector: 'app-home',
  standalone: true,
  imports: [
    FormsModule,
    ButtonModule,
    InputTextModule,
    InputNumberModule,
    ToastModule,
    ColorPickerModule,
    DividerModule,
    DialogModule,
    MessageModule,
    CardModule,
    DeviceListComponent,
    DeviceDetailComponent,
  ],
  providers: [MessageService],
  templateUrl: './home.html',
  styleUrl: './home.scss'
})
export class HomeComponent implements OnInit {
  status: ConnectionStatus = { connected: false, host: 'localhost', port: 6742 };
  devices: RGBDevice[] = [];
  selectedDevice: RGBDevice | null = null;

  connectHost: string = 'localhost';
  connectPort: number = 6742;

  globalColor: string = '#ff0000';
  isLoading: boolean = false;
  isAutoConnecting: boolean = false;

  /** Controls the connection settings dialog. */
  showConnectionDialog: boolean = false;
  /** Editable copies inside the dialog — only applied on "Connect". */
  dialogHost: string = 'localhost';
  dialogPort: number = 6742;

  /** Mobile navigation: which panel is visible on small screens. */
  mobilePanel: 'list' | 'detail' = 'list';

  private readonly destroyRef = inject(DestroyRef);

  constructor(
    private openRgbService: OpenRgbService,
    private messageService: MessageService,
    private cdr: ChangeDetectorRef
  ) {}

  ngOnInit(): void {
    this.checkStatusAndAutoConnect();
  }

  openConnectionDialog(): void {
    this.dialogHost = this.connectHost;
    this.dialogPort = this.connectPort;
    this.showConnectionDialog = true;
    this.cdr.detectChanges();
  }

  /** On page load: check backend status.
   *  If already connected → load devices.
   *  If not connected → try auto-connecting using saved connection from state.json
   *  (with URL param ?host=...&port=... taking priority). */
  private checkStatusAndAutoConnect(): void {
    this.openRgbService.getStatus()
      .pipe(takeUntilDestroyed(this.destroyRef))
      .subscribe({
        next: (status) => {
          this.status = status;
          this.cdr.detectChanges();
          if (status.connected) {
            this.connectHost = status.host ?? this.connectHost;
            this.connectPort = status.port ?? this.connectPort;
            this.loadDevices();
            return;
          }

          // Resolve host/port to use for auto-connect:
          // 1. URL params  2. saved_host/saved_port from state.json  3. defaults
          const params = new URLSearchParams(window.location.search);
          const urlHost = params.get('host');
          const urlPort = params.get('port');

          if (urlHost) {
            this.connectHost = urlHost;
            this.connectPort = urlPort ? parseInt(urlPort, 10) : 6742;
          } else if (status.saved_host) {
            this.connectHost = status.saved_host;
            this.connectPort = status.saved_port ?? 6742;
          }

          this.autoConnect();
        },
        error: () => {
          this.messageService.add({ severity: 'error', summary: 'Error', detail: 'Failed to reach server' });
        }
      });
  }

  /** Silently connect using the resolved host/port. */
  private autoConnect(): void {
    this.isAutoConnecting = true;
    this.openRgbService.connect({ host: this.connectHost, port: this.connectPort })
      .pipe(takeUntilDestroyed(this.destroyRef))
      .subscribe({
        next: (status) => {
          this.isAutoConnecting = false;
          this.status = status;
          this.cdr.detectChanges();
          if (status.connected) {
            this.loadDevices();
          }
        },
        error: () => {
          this.isAutoConnecting = false;
          this.cdr.detectChanges();
        }
      });
  }

  loadDevices(): void {
    this.isLoading = true;
    this.openRgbService.getDevices()
      .pipe(takeUntilDestroyed(this.destroyRef))
      .subscribe({
        next: (devices) => {
          this.devices = devices;
          this.isLoading = false;
          if (this.selectedDevice) {
            const updated = devices.find(d => d.id === this.selectedDevice!.id);
            if (updated) this.selectedDevice = updated;
          }
          this.updateGlobalColorFromDevices(devices);
          this.cdr.detectChanges();
        },
        error: () => {
          this.isLoading = false;
          this.messageService.add({ severity: 'error', summary: 'Error', detail: 'Failed to load devices' });
          this.cdr.detectChanges();
        }
      });
  }

  private updateGlobalColorFromDevices(devices: RGBDevice[]): void {
    const freq = new Map<string, number>();
    for (const device of devices) {
      for (const color of device.colors) {
        const hex = rgbToHex(color);
        freq.set(hex, (freq.get(hex) ?? 0) + 1);
      }
    }
    if (freq.size === 0) return;
    let best = '#ff0000';
    let bestCount = 0;
    for (const [hex, count] of freq) {
      if (count > bestCount) { best = hex; bestCount = count; }
    }
    this.globalColor = best;
  }

  /** Connect using dialog values; called from the dialog "Connect" button. */
  connectFromDialog(): void {
    this.connectHost = this.dialogHost;
    this.connectPort = this.dialogPort;
    this.showConnectionDialog = false;
    this.connect();
  }

  connect(): void {
    this.isLoading = true;
    this.openRgbService.connect({ host: this.connectHost, port: this.connectPort })
      .pipe(takeUntilDestroyed(this.destroyRef))
      .subscribe({
        next: (status) => {
          this.status = status;
          this.isLoading = false;
          this.cdr.detectChanges();
          if (status.connected) {
            this.messageService.add({ severity: 'success', summary: 'Connected', detail: `Connected to ${status.host}:${status.port}` });
            this.loadDevices();
          } else {
            this.messageService.add({ severity: 'error', summary: 'Connection failed', detail: status.error ?? 'Unknown error' });
          }
        },
        error: () => {
          this.isLoading = false;
          this.cdr.detectChanges();
          this.messageService.add({ severity: 'error', summary: 'Error', detail: 'Connection failed' });
        }
      });
  }

  disconnect(): void {
    this.showConnectionDialog = false;
    this.openRgbService.disconnect()
      .pipe(takeUntilDestroyed(this.destroyRef))
      .subscribe({
        next: (status) => {
          this.status = status;
          this.devices = [];
          this.selectedDevice = null;
          this.cdr.detectChanges();
          this.messageService.add({ severity: 'info', summary: 'Disconnected', detail: 'Disconnected from OpenRGB' });
        },
        error: () => {
          this.messageService.add({ severity: 'error', summary: 'Error', detail: 'Failed to disconnect' });
        }
      });
  }

  onDeviceSelected(device: RGBDevice): void {
    this.selectedDevice = device;
    this.mobilePanel = 'detail';
    this.cdr.detectChanges();
  }

  goBackToList(): void {
    this.mobilePanel = 'list';
    this.cdr.detectChanges();
  }

  onSetDeviceColors(colors: RGBColor[]): void {
    if (!this.selectedDevice) return;
    this.openRgbService.setDeviceColors({ device_id: this.selectedDevice.id, colors })
      .pipe(takeUntilDestroyed(this.destroyRef))
      .subscribe({
        next: () => {
          this.messageService.add({ severity: 'success', summary: 'Applied', detail: 'Colors updated' });
          this.loadDevices();
        },
        error: () => this.messageService.add({ severity: 'error', summary: 'Error', detail: 'Failed to set colors' })
      });
  }

  onSetZoneColors(event: { zoneId: number; colors: RGBColor[] }): void {
    if (!this.selectedDevice) return;
    this.openRgbService.setZoneColors({
      device_id: this.selectedDevice.id,
      zone_id: event.zoneId,
      colors: event.colors
    })
      .pipe(takeUntilDestroyed(this.destroyRef))
      .subscribe({
        next: () => {
          this.messageService.add({ severity: 'success', summary: 'Applied', detail: 'Zone colors updated' });
          this.loadDevices();
        },
        error: () => this.messageService.add({ severity: 'error', summary: 'Error', detail: 'Failed to set zone colors' })
      });
  }

  onSetMode(modeId: number): void {
    if (!this.selectedDevice) return;
    this.openRgbService.setMode({ device_id: this.selectedDevice.id, mode_id: modeId })
      .pipe(takeUntilDestroyed(this.destroyRef))
      .subscribe({
        next: () => {
          this.messageService.add({ severity: 'success', summary: 'Applied', detail: 'Mode changed' });
          this.loadDevices();
        },
        error: () => this.messageService.add({ severity: 'error', summary: 'Error', detail: 'Failed to set mode' })
      });
  }

  applyGlobalColor(): void {
    this.openRgbService.setAllColors(hexToRgb(this.globalColor))
      .pipe(takeUntilDestroyed(this.destroyRef))
      .subscribe({
        next: () => {
          this.messageService.add({ severity: 'success', summary: 'Applied', detail: 'Global color applied to all devices' });
          this.loadDevices();
        },
        error: () => this.messageService.add({ severity: 'error', summary: 'Error', detail: 'Failed to apply global color' })
      });
  }
}
