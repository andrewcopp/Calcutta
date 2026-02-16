import React from 'react';
import { BrowserRouter as Router, Routes, Route, useLocation } from 'react-router-dom';
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
import { RouteErrorBoundary } from './components/RouteErrorBoundary';
import { ProtectedRoute } from './components/ProtectedRoute';
import { UserProvider } from './contexts/UserContext';

const AppLayout: React.FC = () => {
  const location = useLocation();
  const hideHeader = location.pathname === '/';

  return (
    <div className={hideHeader ? 'min-h-screen bg-[#070a12]' : 'min-h-screen bg-gray-100'}>
      {!hideHeader && <Header />}
      <Routes>
        <Route path="/" element={<RouteErrorBoundary><HomePage /></RouteErrorBoundary>} />
        <Route path="/login" element={<RouteErrorBoundary><LoginPage /></RouteErrorBoundary>} />
        <Route path="/rules" element={<RouteErrorBoundary><RulesPage /></RouteErrorBoundary>} />
        <Route path="/calcuttas" element={<RouteErrorBoundary><CalcuttaListPage /></RouteErrorBoundary>} />
        <Route path="/calcuttas/create" element={<ProtectedRoute><CreateCalcuttaPage /></ProtectedRoute>} />
        <Route path="/lab" element={<RouteErrorBoundary><LabPage /></RouteErrorBoundary>} />
        <Route path="/lab/models/:modelId" element={<RouteErrorBoundary><ModelDetailPage /></RouteErrorBoundary>} />
        <Route path="/lab/models/:modelName/calcutta/:calcuttaId" element={<RouteErrorBoundary><EntryDetailPage /></RouteErrorBoundary>} />
        <Route path="/lab/entries/:entryId" element={<RouteErrorBoundary><EntryDetailPage /></RouteErrorBoundary>} />
        <Route path="/lab/models/:modelName/calcuttas/:calcuttaId/evaluations/:evaluationId" element={<RouteErrorBoundary><EvaluationDetailPage /></RouteErrorBoundary>} />
        <Route path="/lab/models/:modelName/calcutta/:calcuttaId/entry-results/:entryResultId" element={<RouteErrorBoundary><EntryProfilePage /></RouteErrorBoundary>} />
        {/* Legacy evaluation route redirect */}
        <Route path="/lab/evaluations/:evaluationId" element={<RouteErrorBoundary><EvaluationDetailPage /></RouteErrorBoundary>} />
        <Route path="/lab/entry-results/:entryResultId" element={<RouteErrorBoundary><EntryProfilePage /></RouteErrorBoundary>} />
        <Route path="/admin" element={<ProtectedRoute><RouteErrorBoundary><AdminPage /></RouteErrorBoundary></ProtectedRoute>} />
        <Route path="/admin/api-keys" element={<ProtectedRoute><RouteErrorBoundary><AdminApiKeysPage /></RouteErrorBoundary></ProtectedRoute>} />
        <Route path="/admin/bundles" element={<ProtectedRoute><RouteErrorBoundary><AdminBundlesPage /></RouteErrorBoundary></ProtectedRoute>} />
        <Route path="/admin/users" element={<ProtectedRoute><RouteErrorBoundary><AdminUsersPage /></RouteErrorBoundary></ProtectedRoute>} />
        <Route path="/admin/hall-of-fame" element={<ProtectedRoute><RouteErrorBoundary><HallOfFamePage /></RouteErrorBoundary></ProtectedRoute>} />
        <Route path="/admin/tournaments" element={<ProtectedRoute><RouteErrorBoundary><TournamentListPage /></RouteErrorBoundary></ProtectedRoute>} />
        <Route path="/admin/tournaments/create" element={<ProtectedRoute><RouteErrorBoundary><TournamentCreatePage /></RouteErrorBoundary></ProtectedRoute>} />
        <Route path="/admin/tournaments/:id" element={<ProtectedRoute><RouteErrorBoundary><TournamentViewPage /></RouteErrorBoundary></ProtectedRoute>} />
        <Route path="/admin/tournaments/:id/edit" element={<ProtectedRoute><RouteErrorBoundary><TournamentEditPage /></RouteErrorBoundary></ProtectedRoute>} />
        <Route path="/admin/tournaments/:id/teams/add" element={<ProtectedRoute><RouteErrorBoundary><TournamentAddTeamsPage /></RouteErrorBoundary></ProtectedRoute>} />
        <Route path="/admin/tournaments/:id/bracket" element={<ProtectedRoute><RouteErrorBoundary><TournamentBracketPage /></RouteErrorBoundary></ProtectedRoute>} />
        <Route path="/calcuttas/:calcuttaId" element={<RouteErrorBoundary><CalcuttaEntriesPage /></RouteErrorBoundary>} />
        <Route path="/calcuttas/:calcuttaId/settings" element={<ProtectedRoute><RouteErrorBoundary><CalcuttaSettingsPage /></RouteErrorBoundary></ProtectedRoute>} />
        <Route path="/calcuttas/:calcuttaId/teams" element={<RouteErrorBoundary><CalcuttaTeamsPage /></RouteErrorBoundary>} />
        <Route path="/calcuttas/:calcuttaId/entries/:entryId" element={<RouteErrorBoundary><EntryTeamsPage /></RouteErrorBoundary>} />
        <Route path="/calcuttas/:calcuttaId/entries/:entryId/bid" element={<ProtectedRoute><RouteErrorBoundary><BiddingPage /></RouteErrorBoundary></ProtectedRoute>} />
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
