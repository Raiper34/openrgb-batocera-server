import { Component, Input, Output, EventEmitter } from '@angular/core';
import { RGBDevice } from '../../models/openrgb.models';

@Component({
  selector: 'app-device-list',
  standalone: true,
  imports: [],
  templateUrl: './device-list.html',
  styleUrl: './device-list.scss'
})
export class DeviceListComponent {
  @Input() devices: RGBDevice[] = [];
  @Input() selectedDeviceId: number | null = null;
  @Output() deviceSelected = new EventEmitter<RGBDevice>();

  selectDevice(device: RGBDevice): void {
    this.deviceSelected.emit(device);
  }

  getDeviceTypeLabel(type: number): string {
    const types: Record<number, string> = {
      0: 'Motherboard',
      1: 'DRAM',
      2: 'GPU',
      3: 'Cooler',
      4: 'LED Strip',
      5: 'Keyboard',
      6: 'Mouse',
      7: 'Mousepad',
      8: 'Headset',
      9: 'Headset Stand',
      10: 'Gamepad',
      11: 'Light',
      12: 'Speaker',
      13: 'Virtual',
      14: 'Storage',
      15: 'Case',
      16: 'Microphone',
      17: 'Accessory',
      18: 'Keypad',
    };
    return types[type] ?? 'Unknown';
  }

  getDeviceTypeIcon(type: number): string {
    const icons: Record<number, string> = {
      0: 'pi-server',
      1: 'pi-microchip',
      2: 'pi-desktop',
      3: 'pi-sun',
      4: 'pi-bolt',
      5: 'pi-tablet',
      6: 'pi-mobile',
      7: 'pi-stop',
      8: 'pi-headphones',
      9: 'pi-headphones',
      10: 'pi-gamepad',
      11: 'pi-lightbulb',
      12: 'pi-volume-up',
      13: 'pi-objects-column',
      14: 'pi-database',
      15: 'pi-box',
      16: 'pi-microphone',
      17: 'pi-tag',
      18: 'pi-th',
    };
    return icons[type] ?? 'pi-box';
  }
}
