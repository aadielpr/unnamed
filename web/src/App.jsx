import { Route } from '@solidjs/router';
import Home from './routes/Home';

function App() {
  return (
    <>
      <Route path="/" component={Home} />
    </>
  );
}

export default App;
