import { Component, Input, Output, EventEmitter, OnChanges, SimpleChanges } from '@angular/core';
import { FormsModule } from '@angular/forms';
import { ButtonModule } from 'primeng/button';
import { TabsModule } from 'primeng/tabs';
import { SelectModule } from 'primeng/select';
import { ColorPickerModule } from 'primeng/colorpicker';
import { DividerModule } from 'primeng/divider';
import { TooltipModule } from 'primeng/tooltip';
import { TagModule } from 'primeng/tag';
import { CardModule } from 'primeng/card';
import { RGBDevice, RGBColor } from '../../models/openrgb.models';
import { hexToRgb, rgbToHex } from '../../utils/color.utils';

@Component({
  selector: 'app-device-detail',
  standalone: true,
  imports: [
    FormsModule,
    ButtonModule,
    TabsModule,
    SelectModule,
    ColorPickerModule,
    DividerModule,
    TooltipModule,
    TagModule,
    CardModule,
  ],
  templateUrl: './device-detail.html',
  styleUrl: './device-detail.scss'
})
export class DeviceDetailComponent implements OnChanges {
  @Input() device: RGBDevice | null = null;
  @Output() setDeviceColors = new EventEmitter<RGBColor[]>();
  @Output() setZoneColors = new EventEmitter<{ zoneId: number; colors: RGBColor[] }>();
  @Output() setMode = new EventEmitter<number>();

  selectedColor: string = '#ff0000';
  selectedZoneIndex: number = 0;
  selectedModeId: number = 0;
  activeTab: string = '0';

  modeOptions: { label: string; value: number }[] = [];

  ngOnChanges(changes: SimpleChanges): void {
    if (changes['device'] && this.device) {
      this.activeTab = '0';
      this.selectedModeId = this.device.active_mode;
      this.modeOptions = this.device.modes.map((mode, index) => ({
        label: mode.name,
        value: index
      }));
      const firstColor = this.device.colors[0];
      this.selectedColor = firstColor ? rgbToHex(firstColor) : '#ff0000';
    }
  }

  applyColorToDevice(): void {
    if (!this.device) return;
    const color = hexToRgb(this.selectedColor);
    const colors = new Array(this.device.leds.length).fill(color);
    this.setDeviceColors.emit(colors);
  }

  applyColorToZone(): void {
    if (!this.device) return;
    const zone = this.device.zones[this.selectedZoneIndex];
    if (!zone) return;
    const color = hexToRgb(this.selectedColor);
    const colors = new Array(zone.leds_count).fill(color);
    this.setZoneColors.emit({ zoneId: this.selectedZoneIndex, colors });
  }

  applyMode(): void {
    this.setMode.emit(this.selectedModeId);
  }

  getLedColor(index: number): string {
    if (!this.device || !this.device.colors[index]) return '#000000';
    const c = this.device.colors[index];
    return `rgb(${c.r},${c.g},${c.b})`;
  }

  getLedsForZone(zoneIndex: number): number[] {
    if (!this.device) return [];
    let offset = 0;
    for (let i = 0; i < zoneIndex; i++) {
      offset += this.device.zones[i]?.leds_count ?? 0;
    }
    const count = this.device.zones[zoneIndex]?.leds_count ?? 0;
    return Array.from({ length: count }, (_, i) => offset + i);
  }
}
