import { renderToString } from 'react-dom/server';
import { describe, expect, it } from 'vitest';
import { App } from './App';

describe('App', () => {
  it('renders the wordmark as the page heading', () => {
    const html = renderToString(<App />);
    expect(html).toContain('<h1>มหานที</h1>');
  });
});
