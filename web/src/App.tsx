// The blank page for Phase 1 stays the product entry: the wordmark only.
// The Phase 3 look-dev surfaces are dev routes beside it, removed when
// the product surfaces arrive in Phase 4.
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
  if (path === '/lookdev') {
    return <Lookdev />;
  }
  return (
    <main>
      <h1>มหานที</h1>
    </main>
  );
}
