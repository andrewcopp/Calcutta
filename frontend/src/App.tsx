import { BrowserRouter as Router, Routes, Route } from 'react-router-dom';
import { CalcuttaListPage } from './pages/CalcuttaListPage';
import { CalcuttaEntriesPage } from './pages/CalcuttaEntriesPage';
import { EntryTeamsPage } from './pages/EntryTeamsPage';
import './App.css';

function App() {
  return (
    <Router>
      <div className="min-h-screen bg-gray-100">
        <Routes>
          <Route path="/" element={<CalcuttaListPage />} />
          <Route path="/calcuttas/:calcuttaId" element={<CalcuttaEntriesPage />} />
          <Route path="/calcuttas/:calcuttaId/entries/:entryId" element={<EntryTeamsPage />} />
        </Routes>
      </div>
    </Router>
  );
}

export default App;
