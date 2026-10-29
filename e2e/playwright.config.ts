import { defineConfig, devices } from '@playwright/test';

const port = 4173;

// For now the suite runs against Vite's preview of the production build.
// From Phase 4 it runs against the Go binary serving recorded fixtures, so
// the tests exercise the same single origin that ships.
export default defineConfig({
  testDir: './tests',
  // exFAT volumes create macOS ._ side files next to every spec.
  testIgnore: '**/._*',
  fullyParallel: true,
  forbidOnly: !!process.env['CI'],
  retries: process.env['CI'] ? 1 : 0,
  reporter: process.env['CI'] ? [['github'], ['html', { open: 'never' }]] : 'list',
  use: {
    baseURL: `http://127.0.0.1:${port}`,
    locale: 'th-TH',
    timezoneId: 'Asia/Bangkok',
    trace: 'retain-on-failure',
  },
  projects: [
    {
      name: 'desktop',
      use: { ...devices['Desktop Chrome'], viewport: { width: 1440, height: 900 } },
    },
    { name: 'mobile', use: { ...devices['Pixel 7'], viewport: { width: 390, height: 844 } } },
  ],
  webServer: {
    command: `pnpm --dir ../web build && pnpm --dir ../web exec vite preview --host 127.0.0.1 --port ${port} --strictPort`,
    url: `http://127.0.0.1:${port}`,
    reuseExistingServer: !process.env['CI'],
    timeout: 120_000,
  },
});
