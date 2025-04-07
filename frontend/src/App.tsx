import React from 'react';
import { BrowserRouter as Router, Routes, Route } from 'react-router-dom';
import { CalcuttaListPage } from './pages/CalcuttaListPage';
import { CalcuttaEntriesPage } from './pages/CalcuttaEntriesPage';
import { EntryTeamsPage } from './pages/EntryTeamsPage';
import { TournamentListPage } from './pages/TournamentListPage';
import { TournamentViewPage } from './pages/TournamentViewPage';
import { TournamentEditPage } from './pages/TournamentEditPage';
import { TournamentCreatePage } from './pages/TournamentCreatePage';
import { AdminPage } from './pages/AdminPage';
import { Header } from './components/Header';
import './App.css';

export const App: React.FC = () => {
  return (
    <Router>
      <div className="min-h-screen bg-gray-100">
        <Header />
        <Routes>
          <Route path="/" element={<CalcuttaListPage />} />
          <Route path="/admin" element={<AdminPage />} />
          <Route path="/admin/tournaments" element={<TournamentListPage />} />
          <Route path="/admin/tournaments/create" element={<TournamentCreatePage />} />
          <Route path="/admin/tournaments/:id" element={<TournamentViewPage />} />
          <Route path="/admin/tournaments/:id/edit" element={<TournamentEditPage />} />
          <Route path="/calcuttas/:calcuttaId" element={<CalcuttaEntriesPage />} />
          <Route path="/calcuttas/:calcuttaId/entries/:entryId" element={<EntryTeamsPage />} />
        </Routes>
      </div>
    </Router>
  );
};

export default App;
