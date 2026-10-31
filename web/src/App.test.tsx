import { renderToString } from 'react-dom/server';
import { describe, expect, it } from 'vitest';
import { App } from './App';

describe('App', () => {
  it('renders the relief as the product entry - the root is the look-dev surface', () => {
    const html = renderToString(<App />);
    // The judges' QR points at the root: it must carry the look-dev
    // chrome (wordmark, view toggle, ใกล้ฉัน), not the Phase-1 stub.
    expect(html).toContain('มหานที');
    expect(html).toContain('TERRAIN 3D');
    expect(html).toContain('ใกล้ฉัน');
    expect(html).not.toContain('<h1>มหานที</h1>');
  });
});
