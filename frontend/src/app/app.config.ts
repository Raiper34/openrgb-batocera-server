import { ApplicationConfig, provideBrowserGlobalErrorListeners, provideZonelessChangeDetection, isDevMode } from '@angular/core';
import { provideRouter } from '@angular/router';
import { provideHttpClient } from '@angular/common/http';
import { provideAnimationsAsync } from '@angular/platform-browser/animations/async';
import { providePrimeNG } from 'primeng/config';
import Aura from '@primeuix/themes/aura';
import { definePreset } from '@primeuix/themes';
import { provideServiceWorker } from '@angular/service-worker';

import { routes } from './app.routes';

/**
 * Custom navy-dark preset — maps PrimeNG's surface scale to the app's
 * navy/indigo palette so all PrimeNG components (buttons, dialogs, selects,
 * inputs …) inherit the same surface colours as the hand-crafted layout.
 *
 * Mapping (dark mode only):
 *   surface-950 → #1a1a2e  body / page ground
 *   surface-900 → #16213e  card / section background
 *   surface-800 → #1e1e3f  hover background
 *   surface-700 → #2a2a4a  border colour
 *   surface-600 → #3a3a5a  subtle accent
 */
const NavyPreset = definePreset(Aura, {
  semantic: {
    colorScheme: {
      dark: {
        surface: {
          0:   '#ffffff',
          50:  '#f1f5f9',
          100: '#e2e8f0',
          200: '#cbd5e1',
          300: '#94a3b8',
          400: '#64748b',
          500: '#475569',
          600: '#3a3a5a',
          700: '#2a2a4a',
          800: '#1e1e3f',
          900: '#16213e',
          950: '#1a1a2e',
        },
      },
    },
  },
});

export const appConfig: ApplicationConfig = {
  providers: [
    provideBrowserGlobalErrorListeners(),
    provideZonelessChangeDetection(),
    provideRouter(routes),
    provideHttpClient(),
    provideAnimationsAsync(),
    providePrimeNG({
      theme: {
        preset: NavyPreset,
        options: {
          darkModeSelector: '.dark-mode'
        }
      }
    }),
    provideServiceWorker('ngsw-worker.js', {
      enabled: !isDevMode(),
      registrationStrategy: 'registerWhenStable:30000'
    })
  ]
};
