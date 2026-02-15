import React from 'react';
import { BrowserRouter as Router, Routes, Route, Navigate, useLocation } from 'react-router-dom';
import { CalcuttaListPage } from './pages/CalcuttaListPage';
import { CalcuttaEntriesPage } from './pages/CalcuttaEntriesPage';
import { CalcuttaTeamsPage } from './pages/CalcuttaTeamsPage';
import { EntryTeamsPage } from './pages/EntryTeamsPage';
import { BiddingPage } from './pages/BiddingPage';
import { TournamentListPage } from './pages/TournamentListPage';
import { TournamentViewPage } from './pages/TournamentViewPage';
import { TournamentEditPage } from './pages/TournamentEditPage';
import { TournamentCreatePage } from './pages/TournamentCreatePage';
import { TournamentAddTeamsPage } from './pages/TournamentAddTeamsPage';
import { TournamentBracketPage } from './pages/TournamentBracketPage';
import { AdminPage } from './pages/AdminPage';
import { AdminBundlesPage } from './pages/AdminBundlesPage';
import { AdminApiKeysPage } from './pages/AdminApiKeysPage';
import { AdminUsersPage } from './pages/AdminUsersPage';
import { HallOfFamePage } from './pages/HallOfFamePage';
import { HomePage } from './pages/HomePage';
import { LoginPage } from './pages/LoginPage';
import { RulesPage } from './pages/RulesPage';
import { CreateCalcuttaPage } from './pages/CreateCalcuttaPage';
import { CalcuttaSettingsPage } from './pages/CalcuttaSettingsPage';
import { LabPage } from './pages/LabPage';
import { ModelDetailPage } from './pages/Lab/ModelDetailPage';
import { EntryDetailPage } from './pages/Lab/EntryDetailPage';
import { EvaluationDetailPage } from './pages/Lab/EvaluationDetailPage';
import { EntryProfilePage } from './pages/Lab/EntryProfilePage';
import { Header } from './components/Header';
import { UserProvider } from './contexts/UserContext';

const AppLayout: React.FC = () => {
  const location = useLocation();
  const hideHeader = location.pathname === '/';

  return (
    <div className={hideHeader ? 'min-h-screen bg-[#070a12]' : 'min-h-screen bg-gray-100'}>
      {!hideHeader && <Header />}
      <Routes>
        <Route path="/" element={<HomePage />} />
        <Route path="/login" element={<LoginPage />} />
        <Route path="/rules" element={<RulesPage />} />
        <Route path="/calcuttas" element={<CalcuttaListPage />} />
        <Route path="/calcuttas/create" element={<CreateCalcuttaPage />} />
        <Route path="/lab" element={<LabPage />} />
        <Route path="/lab/models/:modelId" element={<ModelDetailPage />} />
        <Route path="/lab/models/:modelName/calcutta/:calcuttaId" element={<EntryDetailPage />} />
        <Route path="/lab/entries/:entryId" element={<EntryDetailPage />} />
        <Route path="/lab/models/:modelName/calcuttas/:calcuttaId/evaluations/:evaluationId" element={<EvaluationDetailPage />} />
        <Route path="/lab/models/:modelName/calcutta/:calcuttaId/entry-results/:entryResultId" element={<EntryProfilePage />} />
        {/* Legacy evaluation route redirect */}
        <Route path="/lab/evaluations/:evaluationId" element={<EvaluationDetailPage />} />
        <Route path="/lab/entry-results/:entryResultId" element={<EntryProfilePage />} />
        {/* Legacy lab routes redirect to new lab tabs */}
        <Route path="/lab/candidates/*" element={<Navigate to="/lab?tab=entries" replace />} />
        <Route path="/lab/advancements/*" element={<Navigate to="/lab?tab=models" replace />} />
        <Route path="/lab/investments/*" element={<Navigate to="/lab?tab=models" replace />} />
        <Route path="/lab/entry-runs/*" element={<Navigate to="/lab?tab=entries" replace />} />
        <Route path="/lab/entry-artifacts/*" element={<Navigate to="/lab?tab=entries" replace />} />
        {/* Legacy sandbox routes redirect to lab */}
        <Route path="/sandbox/*" element={<Navigate to="/lab?tab=evaluations" replace />} />
        {/* Legacy runs routes redirect to lab */}
        <Route path="/runs/*" element={<Navigate to="/lab?tab=evaluations" replace />} />
        <Route path="/analytics" element={<Navigate to="/lab" replace />} />
        <Route path="/admin" element={<AdminPage />} />
        <Route path="/admin/api-keys" element={<AdminApiKeysPage />} />
        <Route path="/admin/bundles" element={<AdminBundlesPage />} />
        <Route path="/admin/users" element={<AdminUsersPage />} />
        <Route path="/admin/hall-of-fame" element={<HallOfFamePage />} />
        <Route path="/admin/tournaments" element={<TournamentListPage />} />
        <Route path="/admin/tournaments/create" element={<TournamentCreatePage />} />
        <Route path="/admin/tournaments/:id" element={<TournamentViewPage />} />
        <Route path="/admin/tournaments/:id/edit" element={<TournamentEditPage />} />
        <Route path="/admin/tournaments/:id/teams/add" element={<TournamentAddTeamsPage />} />
        <Route path="/admin/tournaments/:id/bracket" element={<TournamentBracketPage />} />
        <Route path="/calcuttas/:calcuttaId" element={<CalcuttaEntriesPage />} />
        <Route path="/calcuttas/:calcuttaId/settings" element={<CalcuttaSettingsPage />} />
        <Route path="/calcuttas/:calcuttaId/teams" element={<CalcuttaTeamsPage />} />
        <Route path="/calcuttas/:calcuttaId/entries/:entryId" element={<EntryTeamsPage />} />
        <Route path="/calcuttas/:calcuttaId/entries/:entryId/bid" element={<BiddingPage />} />
      </Routes>
    </div>
  );
};

export const App: React.FC = () => {
  return (
    <UserProvider>
      <Router>
        <AppLayout />
      </Router>
    </UserProvider>
  );
};

export default App;
