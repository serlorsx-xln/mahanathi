// Writes web/src/design/tokens.css from the single source tokens.ts.
// Run from web/ with `pnpm tokens:css` (Node 24 strips the TS types).
import { mkdirSync, writeFileSync } from 'node:fs';
import { dirname, join } from 'node:path';
import { fileURLToPath } from 'node:url';
import { renderCss } from '../src/design/tokens.ts';

const here = dirname(fileURLToPath(import.meta.url));
const out = join(here, '..', 'src', 'design', 'tokens.css');
mkdirSync(dirname(out), { recursive: true });
writeFileSync(out, renderCss());
console.log(`wrote ${out}`);
