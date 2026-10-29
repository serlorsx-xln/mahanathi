import react from '@vitejs/plugin-react';
import { defineConfig } from 'vite';

// The Go server serves the built files and the API on one origin in
// production. In development, Vite proxies /api to the Go server so the
// browser sees the same single origin.
export default defineConfig({
  plugins: [react()],
  server: {
    port: 5173,
    strictPort: true,
    proxy: {
      '/api': 'http://127.0.0.1:8080',
      '/healthz': 'http://127.0.0.1:8080',
    },
  },
  preview: {
    port: 4173,
    strictPort: true,
  },
  build: {
    target: 'es2022',
    sourcemap: true,
  },
});
