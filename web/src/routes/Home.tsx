import { createSignal, onMount } from 'solid-js';

type HealthResponse = {
  status?: string;
  db?: string;
  error?: string;
};

function Home() {
  const [health, setHealth] = createSignal<HealthResponse | null>(null);

  onMount(async () => {
    try {
      const res = await fetch('/api/health');
      setHealth(await res.json());
    } catch (err) {
      setHealth({ error: err instanceof Error ? err.message : String(err) });
    }
  });

  return (
    <main>
      <h1>EventLens</h1>
      <p>Hello from the SolidJS SPA.</p>
      <pre>{JSON.stringify(health(), null, 2)}</pre>
    </main>
  );
}

export default Home;
