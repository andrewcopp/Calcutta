import React from 'react';
import { BrowserRouter as Router, Routes, Route, Navigate, useLocation } from 'react-router-dom';
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
import { AdminBundlesPage } from './pages/AdminBundlesPage';
import { AdminApiKeysPage } from './pages/AdminApiKeysPage';
import { AdminUsersPage } from './pages/AdminUsersPage';
import { HallOfFamePage } from './pages/HallOfFamePage';
import { HomePage } from './pages/HomePage';
import { LoginPage } from './pages/LoginPage';
import { RulesPage } from './pages/RulesPage';
import { CreateCalcuttaPage } from './pages/CreateCalcuttaPage';
import { RunsPage } from './pages/RunsPage';
import { RunRankingsPage } from './pages/RunRankingsPage';
import { RunReturnsPage } from './pages/RunReturnsPage';
import { RunInvestmentsPage } from './pages/RunInvestmentsPage';
import { EntryPortfolioPage } from './pages/EntryPortfolioPage';
import { LabPage } from './pages/LabPage';
import { LabAdvancementAlgorithmDetailPage } from './pages/LabAdvancementAlgorithmDetailPage';
import { LabAdvancementTournamentDetailPage } from './pages/LabAdvancementTournamentDetailPage';
import { LabInvestmentAlgorithmDetailPage } from './pages/LabInvestmentAlgorithmDetailPage';
import { LabInvestmentCalcuttaDetailPage } from './pages/LabInvestmentCalcuttaDetailPage';
import { LabEntriesPage } from './pages/LabEntriesPage';
import { LabEntriesSuiteDetailPage } from './pages/LabEntriesSuiteDetailPage';
import { LabEntryReportPage } from './pages/LabEntryReportPage';
import { SandboxPage } from './pages/SandboxPage';
import { SandboxCohortsListPage } from './pages/SandboxCohortsListPage';
import { SandboxCohortDetailPage } from './pages/SandboxCohortDetailPage';
import { SimulationRunDetailPage } from './pages/SimulationRunDetailPage';
import { SimulationRunEntryDetailPage } from './pages/SimulationRunEntryDetailPage';
import { Header } from './components/Header';
import { UserProvider } from './contexts/UserContext';

const RunsRedirect: React.FC = () => {
  const year = new Date().getFullYear();
  return <Navigate to={`/runs/${year}`} replace />;
};

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
        <Route path="/lab/entries" element={<LabEntriesPage />} />
        <Route path="/lab/entries/suites/:suiteId" element={<LabEntriesSuiteDetailPage />} />
        <Route path="/lab/entries/scenarios/:scenarioId" element={<LabEntryReportPage />} />
        <Route path="/lab/advancements/algorithms/:algorithmId" element={<LabAdvancementAlgorithmDetailPage />} />
        <Route
          path="/lab/advancements/algorithms/:algorithmId/tournaments/:tournamentId"
          element={<LabAdvancementTournamentDetailPage />}
        />
        <Route path="/lab/investments/algorithms/:algorithmId" element={<LabInvestmentAlgorithmDetailPage />} />
        <Route
          path="/lab/investments/algorithms/:algorithmId/calcuttas/:calcuttaId"
          element={<LabInvestmentCalcuttaDetailPage />}
        />
        <Route path="/sandbox" element={<Navigate to="/sandbox/cohorts" replace />} />
        <Route path="/sandbox/legacy" element={<SandboxPage />} />
        <Route path="/sandbox/cohorts" element={<SandboxCohortsListPage />} />
        <Route path="/sandbox/cohorts/:cohortId" element={<SandboxCohortDetailPage />} />
        <Route path="/sandbox/suites" element={<SandboxCohortsListPage />} />
        <Route path="/sandbox/suites/:suiteId" element={<SandboxCohortDetailPage />} />
        <Route path="/sandbox/evaluations/:id" element={<SimulationRunDetailPage />} />
        <Route path="/sandbox/evaluations/:id/entries/:snapshotEntryId" element={<SimulationRunEntryDetailPage />} />
        <Route path="/runs" element={<RunsRedirect />} />
        <Route path="/runs/:year" element={<RunsPage />} />
        <Route path="/runs/:year/:runId" element={<RunRankingsPage />} />
        <Route path="/runs/:year/:runId/returns" element={<RunReturnsPage />} />
        <Route path="/runs/:year/:runId/investments" element={<RunInvestmentsPage />} />
        <Route path="/runs/:year/:runId/entries/:entryKey" element={<EntryPortfolioPage />} />
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
        <Route path="/calcuttas/:calcuttaId/teams" element={<CalcuttaTeamsPage />} />
        <Route path="/calcuttas/:calcuttaId/entries/:entryId" element={<EntryTeamsPage />} />
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
