import { RGBColor } from '../models/openrgb.models';

/** Converts a CSS hex color string (with or without leading `#`) to an RGBColor object.
 *  Returns `{ r: 0, g: 0, b: 0 }` for invalid input. */
export function hexToRgb(hex: string): RGBColor {
  const result = /^#?([a-f\d]{2})([a-f\d]{2})([a-f\d]{2})$/i.exec(hex);
  return result
    ? { r: parseInt(result[1], 16), g: parseInt(result[2], 16), b: parseInt(result[3], 16) }
    : { r: 0, g: 0, b: 0 };
}

export function rgbToHex(color: RGBColor): string {
  return '#' + [color.r, color.g, color.b]
    .map(v => v.toString(16).padStart(2, '0'))
    .join('');
}
