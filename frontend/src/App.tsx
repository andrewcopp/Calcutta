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
import { ErrorBoundary } from './components/ErrorBoundary';
import { ProtectedRoute } from './components/ProtectedRoute';
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
        <Route path="/calcuttas/create" element={<ProtectedRoute><CreateCalcuttaPage /></ProtectedRoute>} />
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
        <Route path="/admin" element={<ProtectedRoute><AdminPage /></ProtectedRoute>} />
        <Route path="/admin/api-keys" element={<ProtectedRoute><AdminApiKeysPage /></ProtectedRoute>} />
        <Route path="/admin/bundles" element={<ProtectedRoute><AdminBundlesPage /></ProtectedRoute>} />
        <Route path="/admin/users" element={<ProtectedRoute><AdminUsersPage /></ProtectedRoute>} />
        <Route path="/admin/hall-of-fame" element={<ProtectedRoute><HallOfFamePage /></ProtectedRoute>} />
        <Route path="/admin/tournaments" element={<ProtectedRoute><TournamentListPage /></ProtectedRoute>} />
        <Route path="/admin/tournaments/create" element={<ProtectedRoute><TournamentCreatePage /></ProtectedRoute>} />
        <Route path="/admin/tournaments/:id" element={<ProtectedRoute><TournamentViewPage /></ProtectedRoute>} />
        <Route path="/admin/tournaments/:id/edit" element={<ProtectedRoute><TournamentEditPage /></ProtectedRoute>} />
        <Route path="/admin/tournaments/:id/teams/add" element={<ProtectedRoute><TournamentAddTeamsPage /></ProtectedRoute>} />
        <Route path="/admin/tournaments/:id/bracket" element={<ProtectedRoute><TournamentBracketPage /></ProtectedRoute>} />
        <Route path="/calcuttas/:calcuttaId" element={<CalcuttaEntriesPage />} />
        <Route path="/calcuttas/:calcuttaId/settings" element={<ProtectedRoute><CalcuttaSettingsPage /></ProtectedRoute>} />
        <Route path="/calcuttas/:calcuttaId/teams" element={<CalcuttaTeamsPage />} />
        <Route path="/calcuttas/:calcuttaId/entries/:entryId" element={<EntryTeamsPage />} />
        <Route path="/calcuttas/:calcuttaId/entries/:entryId/bid" element={<ProtectedRoute><BiddingPage /></ProtectedRoute>} />
      </Routes>
    </div>
  );
};

export const App: React.FC = () => {
  return (
    <ErrorBoundary>
      <UserProvider>
        <Router>
          <AppLayout />
        </Router>
      </UserProvider>
    </ErrorBoundary>
  );
};

export default App;
