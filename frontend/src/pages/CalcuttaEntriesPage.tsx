import { useState } from 'react';
import { Link, useNavigate, useParams } from 'react-router-dom';

import { Calcutta } from '../types/calcutta';
import { Alert } from '../components/ui/Alert';
import { ErrorState } from '../components/ui/ErrorState';
import { PageContainer, PageHeader } from '../components/ui/Page';
import { LeaderboardSkeleton } from '../components/skeletons/LeaderboardSkeleton';
import { LeaderboardTab } from './CalcuttaEntries/LeaderboardTab';
import { StatisticsTab } from './CalcuttaEntries/StatisticsTab';
import { InvestmentsTab } from './CalcuttaEntries/InvestmentsTab';
import { ReturnsTab } from './CalcuttaEntries/ReturnsTab';
import { OwnershipsTab } from './CalcuttaEntries/OwnershipsTab';
import { BiddingOpenView } from './CalcuttaEntries/BiddingOpenView';
import { DashboardSummary } from './CalcuttaEntries/DashboardSummary';
import { Tabs, TabsList, TabsTrigger, TabsContent } from '../components/ui/Tabs';
import { Breadcrumb } from '../components/ui/Breadcrumb';
import { Badge } from '../components/ui/Badge';
import { Button } from '../components/ui/Button';

import { useCalcuttaDashboard } from '../hooks/useCalcuttaDashboard';
import { useCalcuttaEntriesData } from '../hooks/useCalcuttaEntriesData';
import { useUser } from '../contexts/UserContext';
import { calcuttaService } from '../services/calcuttaService';
import { toast } from '../lib/toast';

import { formatDate } from '../utils/format';

export function CalcuttaEntriesPage() {
  const { calcuttaId } = useParams<{ calcuttaId: string }>();
  const [activeTab, setActiveTab] = useState('leaderboard');
  const [isCreatingEntry, setIsCreatingEntry] = useState(false);
  const [createEntryError, setCreateEntryError] = useState<string | null>(null);
  const navigate = useNavigate();
  const { user } = useUser();

  const dashboardQuery = useCalcuttaDashboard(calcuttaId);
  const dashboardData = dashboardQuery.data;

  const calcutta: Calcutta | undefined = dashboardData?.calcutta;
  const calcuttaName = calcutta?.name ?? '';

  const biddingOpen = dashboardData?.biddingOpen ?? false;
  const currentUserEntry = dashboardData?.currentUserEntry;

  const {
    entries,
    totalEntries,
    allCalcuttaPortfolios,
    allCalcuttaPortfolioTeams,
    allEntryTeams,
    seedInvestmentData,
    schools,
    tournamentTeams,
    totalInvestment,
    totalReturns,
    averageReturn,
    returnsStdDev,
    teamROIData,
  } = useCalcuttaEntriesData(dashboardData);

  if (!calcuttaId) {
    return (
      <PageContainer>
        <Alert variant="error">Missing required parameters</Alert>
      </PageContainer>
    );
  }

  if (dashboardQuery.isLoading) {
    return (
      <PageContainer>
        <PageHeader title="Loading..." />
        <LeaderboardSkeleton />
      </PageContainer>
    );
  }

  if (dashboardQuery.isError) {
    return (
      <PageContainer>
        <ErrorState error={dashboardQuery.error} onRetry={() => dashboardQuery.refetch()} />
      </PageContainer>
    );
  }

  const handleCreateEntry = async () => {
    if (!user || !calcuttaId) return;
    setIsCreatingEntry(true);
    setCreateEntryError(null);
    try {
      const entry = await calcuttaService.createEntry(calcuttaId, `${user.firstName} ${user.lastName}`);
      toast.success('Entry created!');
      navigate(`/calcuttas/${calcuttaId}/entries/${entry.id}/bid`);
    } catch (err) {
      setCreateEntryError(err instanceof Error ? err.message : 'Failed to create entry');
      setIsCreatingEntry(false);
    }
  };

  if (biddingOpen) {
    return (
      <BiddingOpenView
        calcuttaId={calcuttaId}
        calcuttaName={calcuttaName}
        currentUserEntry={currentUserEntry}
        canEditSettings={dashboardData?.abilities?.canEditSettings}
        tournamentStartingAt={dashboardData?.tournamentStartingAt}
        totalEntries={dashboardData!.totalEntries}
        isCreatingEntry={isCreatingEntry}
        createEntryError={createEntryError}
        onCreateEntry={handleCreateEntry}
      />
    );
  }

  return (
    <PageContainer>
      <Breadcrumb
        items={[
          { label: 'Calcuttas', href: '/calcuttas' },
          { label: calcuttaName },
        ]}
      />

      <PageHeader
        title={calcuttaName}
        actions={
          dashboardData?.abilities?.canEditSettings ? (
            <Link to={`/calcuttas/${calcuttaId}/settings`}>
              <Button variant="outline" size="sm">Settings</Button>
            </Link>
          ) : undefined
        }
      />

      {dashboardData?.tournamentStartingAt && (
        <div className="mb-4 flex items-center gap-2">
          <Badge variant="secondary">Portfolios Revealed</Badge>
          <span className="text-sm text-gray-500">{formatDate(dashboardData.tournamentStartingAt, true)}</span>
        </div>
      )}

      {currentUserEntry && (() => {
        const userTeams = allEntryTeams.filter(et => et.entryId === currentUserEntry.id);
        const totalSpent = userTeams.reduce((sum, et) => sum + et.bid, 0);
        const budgetPoints = dashboardData?.calcutta?.budgetPoints ?? 100;
        return (
          <Link
            to={`/calcuttas/${calcuttaId}/entries/${currentUserEntry.id}`}
            className="block mb-6 p-4 border border-gray-200 rounded-lg bg-white shadow-sm hover:shadow-md transition-shadow"
          >
            <div className="flex items-center justify-between">
              <div className="flex items-center gap-3">
                <h3 className="text-lg font-semibold text-gray-900">Your Entry</h3>
                <Badge variant={currentUserEntry.status === 'accepted' ? 'success' : 'secondary'}>
                  {currentUserEntry.status === 'accepted' ? 'Accepted' : 'Incomplete'}
                </Badge>
                <span className="text-sm text-gray-500">{userTeams.length} teams &middot; {totalSpent} / {budgetPoints} pts</span>
              </div>
              <svg className="h-5 w-5 text-gray-400" fill="none" viewBox="0 0 24 24" strokeWidth="2" stroke="currentColor">
                <path strokeLinecap="round" strokeLinejoin="round" d="M8.25 4.5l7.5 7.5-7.5 7.5" />
              </svg>
            </div>
          </Link>
        );
      })()}

      {currentUserEntry && !biddingOpen && (
        <DashboardSummary
          currentEntry={currentUserEntry}
          entries={entries}
          portfolios={allCalcuttaPortfolios}
          portfolioTeams={allCalcuttaPortfolioTeams}
          tournamentTeams={tournamentTeams}
        />
      )}

      <Tabs value={activeTab} onValueChange={setActiveTab}>
        <TabsList>
          <TabsTrigger value="leaderboard">Standings</TabsTrigger>
          <TabsTrigger value="investment">Bids</TabsTrigger>
          <TabsTrigger value="ownership">Shares</TabsTrigger>
          <TabsTrigger value="returns">Scoring</TabsTrigger>
          <TabsTrigger value="statistics">Pool Stats</TabsTrigger>
        </TabsList>

        <TabsContent value="leaderboard">
          <LeaderboardTab calcuttaId={calcuttaId} entries={entries} />
        </TabsContent>

        <TabsContent value="investment">
          <InvestmentsTab entries={entries} schools={schools} tournamentTeams={tournamentTeams} allEntryTeams={allEntryTeams} />
        </TabsContent>

        <TabsContent value="ownership">
          <OwnershipsTab
            entries={entries}
            schools={schools}
            tournamentTeams={tournamentTeams}
            allEntryTeams={allEntryTeams}
            allCalcuttaPortfolios={allCalcuttaPortfolios}
            allCalcuttaPortfolioTeams={allCalcuttaPortfolioTeams}
            isFetching={dashboardQuery.isFetching}
          />
        </TabsContent>

        <TabsContent value="returns">
          <ReturnsTab
            entries={entries}
            schools={schools}
            tournamentTeams={tournamentTeams}
            allCalcuttaPortfolios={allCalcuttaPortfolios}
            allCalcuttaPortfolioTeams={allCalcuttaPortfolioTeams}
          />
        </TabsContent>

        <TabsContent value="statistics">
          <StatisticsTab
            calcuttaId={calcuttaId}
            totalEntries={totalEntries}
            totalInvestment={totalInvestment}
            totalReturns={totalReturns}
            averageReturn={averageReturn}
            returnsStdDev={returnsStdDev}
            seedInvestmentData={seedInvestmentData}
            teamROIData={teamROIData}
          />
        </TabsContent>
      </Tabs>
    </PageContainer>
  );
}
