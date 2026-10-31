// The product entry is the relief itself: the root renders the look-dev
// surface (the judges' QR points at the root, not a dev route). The
// Phase 3 evidence routes stay beside it as reachable paths.
import { ColorEvidence } from './routes/ColorEvidence';
import { Specimen } from './routes/Specimen';
import { Lookdev } from './routes/Lookdev';

const path = typeof window === 'undefined' ? '/' : window.location.pathname;

export function App() {
  if (path === '/specimen') {
    return <Specimen />;
  }
  if (path === '/color') {
    return <ColorEvidence />;
  }
  // The root and /lookdev are the same surface - the relief is the product.
  return <Lookdev />;
}
