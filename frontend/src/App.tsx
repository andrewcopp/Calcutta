import React, { Suspense } from 'react';
import { BrowserRouter as Router, Routes, Route, useLocation } from 'react-router-dom';
import { HomePage } from './pages/HomePage';
import { LoginPage } from './pages/LoginPage';
import { NotFoundPage } from './pages/NotFoundPage';
import { Header } from './components/Header';
import { ErrorBoundary } from './components/ErrorBoundary';
import { RouteErrorBoundary } from './components/RouteErrorBoundary';
import { ProtectedRoute } from './components/ProtectedRoute';
import { PermissionProtectedRoute } from './components/PermissionProtectedRoute';
import { PERMISSIONS } from './constants/permissions';
import { UserProvider } from './contexts/UserContext';
import { LoadingState } from './components/ui/LoadingState';

// Lazy-loaded page components
const CalcuttaListPage = React.lazy(() =>
  import('./pages/CalcuttaListPage').then((m) => ({ default: m.CalcuttaListPage })),
);
const CalcuttaEntriesPage = React.lazy(() =>
  import('./pages/CalcuttaEntriesPage').then((m) => ({ default: m.CalcuttaEntriesPage })),
);
const CalcuttaTeamsPage = React.lazy(() =>
  import('./pages/CalcuttaTeamsPage').then((m) => ({ default: m.CalcuttaTeamsPage })),
);
const EntryTeamsPage = React.lazy(() => import('./pages/EntryTeamsPage').then((m) => ({ default: m.EntryTeamsPage })));
const BiddingPage = React.lazy(() => import('./pages/BiddingPage').then((m) => ({ default: m.BiddingPage })));
const TournamentListPage = React.lazy(() =>
  import('./pages/TournamentListPage').then((m) => ({ default: m.TournamentListPage })),
);
const TournamentViewPage = React.lazy(() =>
  import('./pages/TournamentViewPage').then((m) => ({ default: m.TournamentViewPage })),
);
const TournamentCreatePage = React.lazy(() =>
  import('./pages/TournamentCreatePage').then((m) => ({ default: m.TournamentCreatePage })),
);
const TournamentSetupTeamsPage = React.lazy(() =>
  import('./pages/TournamentSetupTeamsPage').then((m) => ({ default: m.TournamentSetupTeamsPage })),
);
const AdminPage = React.lazy(() => import('./pages/AdminPage').then((m) => ({ default: m.AdminPage })));
const AdminKenPomPage = React.lazy(() =>
  import('./pages/AdminKenPomPage').then((m) => ({ default: m.AdminKenPomPage })),
);
const AdminPredictionsPage = React.lazy(() =>
  import('./pages/AdminPredictionsPage').then((m) => ({ default: m.AdminPredictionsPage })),
);
const AdminTournamentImportsPage = React.lazy(() =>
  import('./pages/AdminTournamentImportsPage').then((m) => ({ default: m.AdminTournamentImportsPage })),
);
const AdminApiKeysPage = React.lazy(() =>
  import('./pages/AdminApiKeysPage').then((m) => ({ default: m.AdminApiKeysPage })),
);
const AdminUsersPage = React.lazy(() => import('./pages/AdminUsersPage').then((m) => ({ default: m.AdminUsersPage })));
const AdminUserProfilePage = React.lazy(() =>
  import('./pages/AdminUserProfilePage').then((m) => ({ default: m.AdminUserProfilePage })),
);
const UserProfilePage = React.lazy(() =>
  import('./pages/UserProfilePage').then((m) => ({ default: m.UserProfilePage })),
);
const HallOfFamePage = React.lazy(() => import('./pages/HallOfFamePage').then((m) => ({ default: m.HallOfFamePage })));
const AcceptInvitePage = React.lazy(() =>
  import('./pages/AcceptInvitePage').then((m) => ({ default: m.AcceptInvitePage })),
);
const ForgotPasswordPage = React.lazy(() =>
  import('./pages/ForgotPasswordPage').then((m) => ({ default: m.ForgotPasswordPage })),
);
const ResetPasswordPage = React.lazy(() =>
  import('./pages/ResetPasswordPage').then((m) => ({ default: m.ResetPasswordPage })),
);
const RulesPage = React.lazy(() => import('./pages/RulesPage').then((m) => ({ default: m.RulesPage })));
const CreateCalcuttaPage = React.lazy(() =>
  import('./pages/CreateCalcuttaPage').then((m) => ({ default: m.CreateCalcuttaPage })),
);
const CalcuttaSettingsPage = React.lazy(() =>
  import('./pages/CalcuttaSettingsPage').then((m) => ({ default: m.CalcuttaSettingsPage })),
);
const LabPage = React.lazy(() => import('./pages/LabPage').then((m) => ({ default: m.LabPage })));
const ModelDetailPage = React.lazy(() =>
  import('./pages/Lab/ModelDetailPage').then((m) => ({ default: m.ModelDetailPage })),
);
const EntryDetailPage = React.lazy(() =>
  import('./pages/Lab/EntryDetailPage').then((m) => ({ default: m.EntryDetailPage })),
);
const EvaluationDetailPage = React.lazy(() =>
  import('./pages/Lab/EvaluationDetailPage').then((m) => ({ default: m.EvaluationDetailPage })),
);
const EntryProfilePage = React.lazy(() =>
  import('./pages/Lab/EntryProfilePage').then((m) => ({ default: m.EntryProfilePage })),
);

const AppLayout: React.FC = () => {
  const location = useLocation();
  const hideHeader = location.pathname === '/';

  return (
    <div className={hideHeader ? 'min-h-screen bg-[#070a12]' : 'min-h-screen bg-background'}>
      {!hideHeader && <Header />}
      <main id="main-content">
        <Suspense fallback={<LoadingState />}>
          <Routes>
            <Route
              path="/"
              element={
                <RouteErrorBoundary>
                  <HomePage />
                </RouteErrorBoundary>
              }
            />
            <Route
              path="/login"
              element={
                <RouteErrorBoundary>
                  <LoginPage />
                </RouteErrorBoundary>
              }
            />
            <Route
              path="/invite"
              element={
                <RouteErrorBoundary>
                  <AcceptInvitePage />
                </RouteErrorBoundary>
              }
            />
            <Route
              path="/forgot-password"
              element={
                <RouteErrorBoundary>
                  <ForgotPasswordPage />
                </RouteErrorBoundary>
              }
            />
            <Route
              path="/reset-password"
              element={
                <RouteErrorBoundary>
                  <ResetPasswordPage />
                </RouteErrorBoundary>
              }
            />
            <Route
              path="/rules"
              element={
                <RouteErrorBoundary>
                  <RulesPage />
                </RouteErrorBoundary>
              }
            />
            <Route
              path="/profile"
              element={
                <ProtectedRoute>
                  <RouteErrorBoundary>
                    <UserProfilePage />
                  </RouteErrorBoundary>
                </ProtectedRoute>
              }
            />
            <Route
              path="/calcuttas"
              element={
                <RouteErrorBoundary>
                  <CalcuttaListPage />
                </RouteErrorBoundary>
              }
            />
            <Route
              path="/calcuttas/create"
              element={
                <ProtectedRoute>
                  <RouteErrorBoundary>
                    <CreateCalcuttaPage />
                  </RouteErrorBoundary>
                </ProtectedRoute>
              }
            />
            <Route
              path="/lab"
              element={
                <PermissionProtectedRoute permission={PERMISSIONS.LAB_READ}>
                  <RouteErrorBoundary>
                    <LabPage />
                  </RouteErrorBoundary>
                </PermissionProtectedRoute>
              }
            />
            <Route
              path="/lab/models/:modelId"
              element={
                <PermissionProtectedRoute permission={PERMISSIONS.LAB_READ}>
                  <RouteErrorBoundary>
                    <ModelDetailPage />
                  </RouteErrorBoundary>
                </PermissionProtectedRoute>
              }
            />
            <Route
              path="/lab/models/:modelName/calcutta/:calcuttaId"
              element={
                <PermissionProtectedRoute permission={PERMISSIONS.LAB_READ}>
                  <RouteErrorBoundary>
                    <EntryDetailPage />
                  </RouteErrorBoundary>
                </PermissionProtectedRoute>
              }
            />
            <Route
              path="/lab/models/:modelName/calcutta/:calcuttaId/evaluations/:evaluationId"
              element={
                <PermissionProtectedRoute permission={PERMISSIONS.LAB_READ}>
                  <RouteErrorBoundary>
                    <EvaluationDetailPage />
                  </RouteErrorBoundary>
                </PermissionProtectedRoute>
              }
            />
            <Route
              path="/lab/models/:modelName/calcutta/:calcuttaId/entry-results/:entryResultId"
              element={
                <PermissionProtectedRoute permission={PERMISSIONS.LAB_READ}>
                  <RouteErrorBoundary>
                    <EntryProfilePage />
                  </RouteErrorBoundary>
                </PermissionProtectedRoute>
              }
            />
            <Route
              path="/admin"
              element={
                <PermissionProtectedRoute permission={PERMISSIONS.ADMIN}>
                  <RouteErrorBoundary>
                    <AdminPage />
                  </RouteErrorBoundary>
                </PermissionProtectedRoute>
              }
            />
            <Route
              path="/admin/api-keys"
              element={
                <PermissionProtectedRoute permission={PERMISSIONS.ADMIN_API_KEYS_WRITE}>
                  <RouteErrorBoundary>
                    <AdminApiKeysPage />
                  </RouteErrorBoundary>
                </PermissionProtectedRoute>
              }
            />
            <Route
              path="/admin/tournament-imports"
              element={
                <PermissionProtectedRoute permission={PERMISSIONS.ADMIN_BUNDLES_EXPORT}>
                  <RouteErrorBoundary>
                    <AdminTournamentImportsPage />
                  </RouteErrorBoundary>
                </PermissionProtectedRoute>
              }
            />
            <Route
              path="/admin/users"
              element={
                <PermissionProtectedRoute permission={PERMISSIONS.ADMIN_USERS_READ}>
                  <RouteErrorBoundary>
                    <AdminUsersPage />
                  </RouteErrorBoundary>
                </PermissionProtectedRoute>
              }
            />
            <Route
              path="/admin/users/:userId"
              element={
                <PermissionProtectedRoute permission={PERMISSIONS.ADMIN_USERS_READ}>
                  <RouteErrorBoundary>
                    <AdminUserProfilePage />
                  </RouteErrorBoundary>
                </PermissionProtectedRoute>
              }
            />
            <Route
              path="/admin/hall-of-fame"
              element={
                <PermissionProtectedRoute permission={PERMISSIONS.ADMIN_HOF_READ}>
                  <RouteErrorBoundary>
                    <HallOfFamePage />
                  </RouteErrorBoundary>
                </PermissionProtectedRoute>
              }
            />
            <Route
              path="/admin/kenpom"
              element={
                <PermissionProtectedRoute permission={PERMISSIONS.TOURNAMENT_GAME_WRITE}>
                  <RouteErrorBoundary>
                    <AdminKenPomPage />
                  </RouteErrorBoundary>
                </PermissionProtectedRoute>
              }
            />
            <Route
              path="/admin/predictions"
              element={
                <PermissionProtectedRoute permission={PERMISSIONS.TOURNAMENT_GAME_WRITE}>
                  <RouteErrorBoundary>
                    <AdminPredictionsPage />
                  </RouteErrorBoundary>
                </PermissionProtectedRoute>
              }
            />
            <Route
              path="/admin/tournaments"
              element={
                <PermissionProtectedRoute permission={PERMISSIONS.TOURNAMENT_GAME_WRITE}>
                  <RouteErrorBoundary>
                    <TournamentListPage />
                  </RouteErrorBoundary>
                </PermissionProtectedRoute>
              }
            />
            <Route
              path="/admin/tournaments/create"
              element={
                <PermissionProtectedRoute permission={PERMISSIONS.TOURNAMENT_GAME_WRITE}>
                  <RouteErrorBoundary>
                    <TournamentCreatePage />
                  </RouteErrorBoundary>
                </PermissionProtectedRoute>
              }
            />
            <Route
              path="/admin/tournaments/:id"
              element={
                <PermissionProtectedRoute permission={PERMISSIONS.TOURNAMENT_GAME_WRITE}>
                  <RouteErrorBoundary>
                    <TournamentViewPage />
                  </RouteErrorBoundary>
                </PermissionProtectedRoute>
              }
            />
            <Route
              path="/admin/tournaments/:id/teams/setup"
              element={
                <PermissionProtectedRoute permission={PERMISSIONS.TOURNAMENT_GAME_WRITE}>
                  <RouteErrorBoundary>
                    <TournamentSetupTeamsPage />
                  </RouteErrorBoundary>
                </PermissionProtectedRoute>
              }
            />
            <Route
              path="/calcuttas/:calcuttaId"
              element={
                <RouteErrorBoundary>
                  <CalcuttaEntriesPage />
                </RouteErrorBoundary>
              }
            />
            <Route
              path="/calcuttas/:calcuttaId/settings"
              element={
                <ProtectedRoute>
                  <RouteErrorBoundary>
                    <CalcuttaSettingsPage />
                  </RouteErrorBoundary>
                </ProtectedRoute>
              }
            />
            <Route
              path="/calcuttas/:calcuttaId/teams"
              element={
                <RouteErrorBoundary>
                  <CalcuttaTeamsPage />
                </RouteErrorBoundary>
              }
            />
            <Route
              path="/calcuttas/:calcuttaId/entries/:entryId"
              element={
                <RouteErrorBoundary>
                  <EntryTeamsPage />
                </RouteErrorBoundary>
              }
            />
            <Route
              path="/calcuttas/:calcuttaId/entries/:entryId/bid"
              element={
                <ProtectedRoute>
                  <RouteErrorBoundary>
                    <BiddingPage />
                  </RouteErrorBoundary>
                </ProtectedRoute>
              }
            />
            <Route
              path="*"
              element={
                <RouteErrorBoundary>
                  <NotFoundPage />
                </RouteErrorBoundary>
              }
            />
          </Routes>
        </Suspense>
      </main>
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
