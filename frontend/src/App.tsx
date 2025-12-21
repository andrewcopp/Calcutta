import React from 'react';
import { BrowserRouter as Router, Routes, Route } from 'react-router-dom';
import { CalcuttaListPage } from './pages/CalcuttaListPage';
import { CalcuttaEntriesPage } from './pages/CalcuttaEntriesPage';
import { CalcuttaTeamsPage } from './pages/CalcuttaTeamsPage';
import { EntryTeamsPage } from './pages/EntryTeamsPage';
import { TournamentListPage } from './pages/TournamentListPage';
import { TournamentViewPage } from './pages/TournamentViewPage';
import { TournamentEditPage } from './pages/TournamentEditPage';
import { TournamentCreatePage } from './pages/TournamentCreatePage';
import { TournamentAddTeamsPage } from './pages/TournamentAddTeamsPage';
import { TournamentBracketPage } from './pages/TournamentBracketPage';
import { AdminPage } from './pages/AdminPage';
import { AnalyticsPage } from './pages/AnalyticsPage';
import { HomePage } from './pages/HomePage';
import { RulesPage } from './pages/RulesPage';
import { CreateCalcuttaPage } from './pages/CreateCalcuttaPage';
import { Header } from './components/Header';
import { UserProvider } from './contexts/UserContext';
import './App.css';

export const App: React.FC = () => {
  return (
    <UserProvider>
      <Router>
        <div className="min-h-screen bg-gray-100">
          <Header />
          <Routes>
            <Route path="/" element={<HomePage />} />
            <Route path="/rules" element={<RulesPage />} />
            <Route path="/calcuttas" element={<CalcuttaListPage />} />
            <Route path="/calcuttas/create" element={<CreateCalcuttaPage />} />
            <Route path="/admin" element={<AdminPage />} />
            <Route path="/admin/analytics" element={<AnalyticsPage />} />
            <Route path="/admin/tournaments" element={<TournamentListPage />} />
            <Route path="/admin/tournaments/create" element={<TournamentCreatePage />} />
            <Route path="/admin/tournaments/:id" element={<TournamentViewPage />} />
            <Route path="/admin/tournaments/:id/edit" element={<TournamentEditPage />} />
            <Route path="/admin/tournaments/:id/teams/add" element={<TournamentAddTeamsPage />} />
            <Route path="/admin/tournaments/:id/bracket" element={<TournamentBracketPage />} />
            <Route path="/calcuttas/:calcuttaId" element={<CalcuttaEntriesPage />} />
            <Route path="/calcuttas/:calcuttaId/teams" element={<CalcuttaTeamsPage />} />
            <Route path="/calcuttas/:calcuttaId/entries/:entryId" element={<EntryTeamsPage />} />
          </Routes>
        </div>
      </Router>
    </UserProvider>
  );
};

export default App;
