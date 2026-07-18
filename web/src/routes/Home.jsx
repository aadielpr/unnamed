import { createSignal, onMount } from 'solid-js';

function Home() {
  const [health, setHealth] = createSignal(null);

  onMount(async () => {
    try {
      const res = await fetch('/api/health');
      setHealth(await res.json());
    } catch (err) {
      setHealth({ error: err.message });
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
