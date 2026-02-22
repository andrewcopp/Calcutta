import React from 'react';
import { BrowserRouter as Router, Routes, Route, useLocation } from 'react-router-dom';
import { CalcuttaListPage } from './pages/CalcuttaListPage';
import { CalcuttaEntriesPage } from './pages/CalcuttaEntriesPage';
import { CalcuttaTeamsPage } from './pages/CalcuttaTeamsPage';
import { EntryTeamsPage } from './pages/EntryTeamsPage';
import { BiddingPage } from './pages/BiddingPage';
import { TournamentListPage } from './pages/TournamentListPage';
import { TournamentViewPage } from './pages/TournamentViewPage';
import { TournamentCreatePage } from './pages/TournamentCreatePage';
import { TournamentSetupTeamsPage } from './pages/TournamentSetupTeamsPage';
import { AdminPage } from './pages/AdminPage';
import { AdminKenPomPage } from './pages/AdminKenPomPage';
import { AdminBundlesPage } from './pages/AdminBundlesPage';
import { AdminApiKeysPage } from './pages/AdminApiKeysPage';
import { AdminUsersPage } from './pages/AdminUsersPage';
import { HallOfFamePage } from './pages/HallOfFamePage';
import { HomePage } from './pages/HomePage';
import { LoginPage } from './pages/LoginPage';
import { AcceptInvitePage } from './pages/AcceptInvitePage';
import { RulesPage } from './pages/RulesPage';
import { CreateCalcuttaPage } from './pages/CreateCalcuttaPage';
import { NotFoundPage } from './pages/NotFoundPage';
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
import { PermissionProtectedRoute } from './components/PermissionProtectedRoute';
import { PERMISSIONS } from './constants/permissions';
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
        <Route path="/invite" element={<RouteErrorBoundary><AcceptInvitePage /></RouteErrorBoundary>} />
        <Route path="/rules" element={<RouteErrorBoundary><RulesPage /></RouteErrorBoundary>} />
        <Route path="/calcuttas" element={<RouteErrorBoundary><CalcuttaListPage /></RouteErrorBoundary>} />
        <Route path="/calcuttas/create" element={<ProtectedRoute><RouteErrorBoundary><CreateCalcuttaPage /></RouteErrorBoundary></ProtectedRoute>} />
        <Route path="/lab" element={<PermissionProtectedRoute permission={PERMISSIONS.LAB_READ}><RouteErrorBoundary><LabPage /></RouteErrorBoundary></PermissionProtectedRoute>} />
        <Route path="/lab/models/:modelId" element={<PermissionProtectedRoute permission={PERMISSIONS.LAB_READ}><RouteErrorBoundary><ModelDetailPage /></RouteErrorBoundary></PermissionProtectedRoute>} />
        <Route path="/lab/models/:modelName/calcutta/:calcuttaId" element={<PermissionProtectedRoute permission={PERMISSIONS.LAB_READ}><RouteErrorBoundary><EntryDetailPage /></RouteErrorBoundary></PermissionProtectedRoute>} />
        <Route path="/lab/models/:modelName/calcutta/:calcuttaId/evaluations/:evaluationId" element={<PermissionProtectedRoute permission={PERMISSIONS.LAB_READ}><RouteErrorBoundary><EvaluationDetailPage /></RouteErrorBoundary></PermissionProtectedRoute>} />
        <Route path="/lab/models/:modelName/calcutta/:calcuttaId/entry-results/:entryResultId" element={<PermissionProtectedRoute permission={PERMISSIONS.LAB_READ}><RouteErrorBoundary><EntryProfilePage /></RouteErrorBoundary></PermissionProtectedRoute>} />
        <Route path="/admin" element={<PermissionProtectedRoute permission={PERMISSIONS.ADMIN}><RouteErrorBoundary><AdminPage /></RouteErrorBoundary></PermissionProtectedRoute>} />
        <Route path="/admin/api-keys" element={<PermissionProtectedRoute permission={PERMISSIONS.ADMIN_API_KEYS_WRITE}><RouteErrorBoundary><AdminApiKeysPage /></RouteErrorBoundary></PermissionProtectedRoute>} />
        <Route path="/admin/bundles" element={<PermissionProtectedRoute permission={PERMISSIONS.ADMIN_BUNDLES_EXPORT}><RouteErrorBoundary><AdminBundlesPage /></RouteErrorBoundary></PermissionProtectedRoute>} />
        <Route path="/admin/users" element={<PermissionProtectedRoute permission={PERMISSIONS.ADMIN_USERS_READ}><RouteErrorBoundary><AdminUsersPage /></RouteErrorBoundary></PermissionProtectedRoute>} />
        <Route path="/admin/hall-of-fame" element={<PermissionProtectedRoute permission={PERMISSIONS.ADMIN_HOF_READ}><RouteErrorBoundary><HallOfFamePage /></RouteErrorBoundary></PermissionProtectedRoute>} />
        <Route path="/admin/kenpom" element={<PermissionProtectedRoute permission={PERMISSIONS.TOURNAMENT_GAME_WRITE}><RouteErrorBoundary><AdminKenPomPage /></RouteErrorBoundary></PermissionProtectedRoute>} />
        <Route path="/admin/tournaments" element={<PermissionProtectedRoute permission={PERMISSIONS.TOURNAMENT_GAME_WRITE}><RouteErrorBoundary><TournamentListPage /></RouteErrorBoundary></PermissionProtectedRoute>} />
        <Route path="/admin/tournaments/create" element={<PermissionProtectedRoute permission={PERMISSIONS.TOURNAMENT_GAME_WRITE}><RouteErrorBoundary><TournamentCreatePage /></RouteErrorBoundary></PermissionProtectedRoute>} />
        <Route path="/admin/tournaments/:id" element={<PermissionProtectedRoute permission={PERMISSIONS.TOURNAMENT_GAME_WRITE}><RouteErrorBoundary><TournamentViewPage /></RouteErrorBoundary></PermissionProtectedRoute>} />
        <Route path="/admin/tournaments/:id/teams/setup" element={<PermissionProtectedRoute permission={PERMISSIONS.TOURNAMENT_GAME_WRITE}><RouteErrorBoundary><TournamentSetupTeamsPage /></RouteErrorBoundary></PermissionProtectedRoute>} />
        <Route path="/calcuttas/:calcuttaId" element={<RouteErrorBoundary><CalcuttaEntriesPage /></RouteErrorBoundary>} />
        <Route path="/calcuttas/:calcuttaId/settings" element={<ProtectedRoute><RouteErrorBoundary><CalcuttaSettingsPage /></RouteErrorBoundary></ProtectedRoute>} />
        <Route path="/calcuttas/:calcuttaId/teams" element={<RouteErrorBoundary><CalcuttaTeamsPage /></RouteErrorBoundary>} />
        <Route path="/calcuttas/:calcuttaId/entries/:entryId" element={<RouteErrorBoundary><EntryTeamsPage /></RouteErrorBoundary>} />
        <Route path="/calcuttas/:calcuttaId/entries/:entryId/bid" element={<ProtectedRoute><RouteErrorBoundary><BiddingPage /></RouteErrorBoundary></ProtectedRoute>} />
        <Route path="*" element={<RouteErrorBoundary><NotFoundPage /></RouteErrorBoundary>} />
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
