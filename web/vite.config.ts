import { createReadStream, stat } from 'node:fs';
import path from 'node:path';
import react from '@vitejs/plugin-react';
import { defineConfig, type Plugin } from 'vite';
import { configDefaults } from 'vitest/config';

// Serve the committed geodata assets (geodata/assets/) at /geodata/ in
// dev and preview, so the look-dev routes read the same bytes the Go
// server will serve in production without duplicating 1.3 MB into
// web/public. The production build does not include them: which assets
// ship with the product is a Phase 4 decision.
function geodata(): Plugin {
  const root = path.resolve(import.meta.dirname, '../geodata/assets');
  const mime: Record<string, string> = {
    '.png': 'image/png',
    '.json': 'application/json',
    '.geojson': 'application/json',
  };
  const serve = (
    req: { url?: string },
    res: {
      statusCode: number;
      setHeader: (k: string, v: string | number) => void;
      end: () => void;
    },
    next: () => void,
  ) => {
    const rel = decodeURIComponent((req.url ?? '').split('?')[0]).replace(/^\/+/, '');
    const file = path.join(root, rel);
    if (!file.startsWith(root) || rel.includes('..')) {
      res.statusCode = 403;
      res.end();
      return;
    }
    stat(file, (err, st) => {
      if (err || !st.isFile()) {
        next();
        return;
      }
      res.statusCode = 200;
      res.setHeader('Content-Type', mime[path.extname(file)] ?? 'application/octet-stream');
      res.setHeader('Content-Length', st.size);
      createReadStream(file).pipe(res as never);
    });
  };
  return {
    name: 'geodata',
    configureServer(server) {
      server.middlewares.use('/geodata', serve as never);
    },
    configurePreviewServer(server) {
      server.middlewares.use('/geodata', serve as never);
    },
  };
}

// The Go server serves the built files and the API on one origin in
// production. In development, Vite proxies /api to the Go server so the
// browser sees the same single origin.
export default defineConfig({
  plugins: [react(), geodata()],
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
  test: {
    // Pure logic (formatters, search, nearest station, timeline decoding)
    // runs in Node. Component tests can opt into a DOM per file later.
    environment: 'node',
    include: ['src/**/*.test.{ts,tsx}'],
    exclude: [...configDefaults.exclude, '**/._*'],
  },
});
