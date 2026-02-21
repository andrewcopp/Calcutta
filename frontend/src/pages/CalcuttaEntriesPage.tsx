import { useState } from 'react';
import { Link, useNavigate, useParams } from 'react-router-dom';
import { Calcutta } from '../types/calcutta';
import { Alert } from '../components/ui/Alert';
import { ErrorState } from '../components/ui/ErrorState';
import { PageContainer, PageHeader } from '../components/ui/Page';
import { LeaderboardSkeleton } from '../components/skeletons/LeaderboardSkeleton';
import { StatisticsTab } from './CalcuttaEntries/StatisticsTab';
import { InvestmentTab } from './CalcuttaEntries/InvestmentTab';
import { ReturnsTab } from './CalcuttaEntries/ReturnsTab';
import { OwnershipTab } from './CalcuttaEntries/OwnershipTab';
import { Tabs, TabsList, TabsTrigger, TabsContent } from '../components/ui/Tabs';
import { Breadcrumb } from '../components/ui/Breadcrumb';
import { Badge } from '../components/ui/Badge';
import { Button } from '../components/ui/Button';
import { BiddingOverlay } from '../components/BiddingOverlay';
import { useCalcuttaDashboard } from '../hooks/useCalcuttaDashboard';
import { useCalcuttaEntriesData } from '../hooks/useCalcuttaEntriesData';
import { useUser } from '../contexts/useUser';
import { calcuttaService } from '../services/calcuttaService';
import { generateMockBiddingData } from '../utils/mockBiddingData';
import { formatDollarsFromCents } from '../utils/format';

export function CalcuttaEntriesPage() {
  const { calcuttaId } = useParams<{ calcuttaId: string }>();
  const [activeTab, setActiveTab] = useState('leaderboard');
  const [isCreatingEntry, setIsCreatingEntry] = useState(false);
  const navigate = useNavigate();
  const { user } = useUser();

  const dashboardQuery = useCalcuttaDashboard(calcuttaId);
  const dashboardData = dashboardQuery.data;

  const calcutta: Calcutta | undefined = dashboardData?.calcutta;
  if (dashboardData && !calcutta) {
    console.warn('CalcuttaEntriesPage: dashboard loaded but calcutta is missing', { calcuttaId });
  }
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

  const mockData = biddingOpen
    ? generateMockBiddingData(tournamentTeams, dashboardData?.totalEntries ?? 5)
    : null;

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
    try {
      const entry = await calcuttaService.createEntry(calcuttaId, `${user.firstName} ${user.lastName}`);
      navigate(`/calcuttas/${calcuttaId}/entries/${entry.id}/bid`);
    } catch {
      setIsCreatingEntry(false);
    }
  };

  const renderLeaderboard = (leaderboardEntries: typeof entries) => (
    <div className="grid gap-4">
      {leaderboardEntries.map((entry, index) => {
        const displayPosition = entry.finishPosition || index + 1;
        const isInTheMoney = Boolean(entry.inTheMoney);
        const payoutText = entry.payoutCents ? `(${formatDollarsFromCents(entry.payoutCents)})` : '';

        const rowClass = isInTheMoney
          ? displayPosition === 1
            ? 'bg-gradient-to-r from-yellow-50 to-yellow-100 border-2 border-yellow-400'
            : displayPosition === 2
              ? 'bg-gradient-to-r from-slate-50 to-slate-200 border-2 border-slate-400'
              : displayPosition === 3
                ? 'bg-gradient-to-r from-amber-50 to-amber-100 border-2 border-amber-500'
                : 'bg-gradient-to-r from-slate-50 to-blue-50 border-2 border-slate-300'
          : 'bg-white';

        const pointsClass = isInTheMoney
          ? displayPosition === 1
            ? 'text-yellow-700'
            : displayPosition === 2
              ? 'text-slate-700'
              : displayPosition === 3
                ? 'text-amber-700'
                : 'text-slate-700'
          : 'text-blue-600';

        return (
          <div
            key={entry.id}
            className={`block p-4 rounded-lg shadow ${rowClass}`}
          >
            <div className="flex justify-between items-center">
              <div>
                <h2 className="text-xl font-semibold">
                  {displayPosition}. {entry.name}
                  {isInTheMoney && payoutText && <span className="ml-2 text-sm text-gray-700">{payoutText}</span>}
                </h2>
              </div>
              <div className="text-right">
                <p className={`text-2xl font-bold ${pointsClass}`}>
                  {entry.totalPoints ? entry.totalPoints.toFixed(2) : '0.00'} pts
                </p>
              </div>
            </div>
          </div>
        );
      })}
    </div>
  );

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

      {dashboardData && (
        <div className="mb-4">
          {biddingOpen ? (
            <Badge variant="success">Bidding Open</Badge>
          ) : (
            <Badge variant="secondary">Bidding Closed</Badge>
          )}
        </div>
      )}

      {biddingOpen && (
        <div className="mb-6 p-4 border border-gray-200 rounded-lg bg-white shadow-sm">
          <div className="flex items-center justify-between">
            <div>
              <h3 className="text-sm font-medium text-gray-500 mb-1">Your Entry</h3>
              {currentUserEntry ? (
                <div className="flex items-center gap-3">
                  {currentUserEntry.status === 'final' ? (
                    <Badge variant="success">Entry Submitted</Badge>
                  ) : (
                    <Badge variant="warning">Awaiting Submission</Badge>
                  )}
                  <span className="text-sm text-gray-600">{currentUserEntry.name}</span>
                </div>
              ) : (
                <Badge variant="warning">Awaiting Submission</Badge>
              )}
            </div>
            <div className="flex items-center gap-4">
              <span className="text-sm text-gray-500">
                {dashboardData!.totalEntries} {dashboardData!.totalEntries === 1 ? 'entry' : 'entries'} submitted
              </span>
              {currentUserEntry ? (
                <Link to={`/calcuttas/${calcuttaId}/entries/${currentUserEntry.id}/bid`}>
                  <Button size="sm">
                    {currentUserEntry.status === 'final' ? 'Edit Bids' : 'Place Bids'}
                  </Button>
                </Link>
              ) : (
                <Button size="sm" onClick={handleCreateEntry} disabled={isCreatingEntry}>
                  {isCreatingEntry ? 'Creating...' : 'Create Entry'}
                </Button>
              )}
            </div>
          </div>
        </div>
      )}

      <Tabs value={activeTab} onValueChange={setActiveTab}>
        <TabsList>
          <TabsTrigger value="leaderboard">Leaderboard</TabsTrigger>
          <TabsTrigger value="investment">Investments</TabsTrigger>
          <TabsTrigger value="ownership">Ownerships</TabsTrigger>
          <TabsTrigger value="returns">Returns</TabsTrigger>
          <TabsTrigger value="statistics">Statistics</TabsTrigger>
        </TabsList>

        <TabsContent value="leaderboard">
          {biddingOpen && mockData ? (
            <BiddingOverlay tournamentStartingAt={dashboardData!.tournamentStartingAt!}>
              {renderLeaderboard(mockData.entries as typeof entries)}
            </BiddingOverlay>
          ) : (
            <div className="grid gap-4">
              {entries.map((entry, index) => {
                const displayPosition = entry.finishPosition || index + 1;
                const isInTheMoney = Boolean(entry.inTheMoney);
                const payoutText = entry.payoutCents ? `(${formatDollarsFromCents(entry.payoutCents)})` : '';

                const rowClass = isInTheMoney
                  ? displayPosition === 1
                    ? 'bg-gradient-to-r from-yellow-50 to-yellow-100 border-2 border-yellow-400'
                    : displayPosition === 2
                      ? 'bg-gradient-to-r from-slate-50 to-slate-200 border-2 border-slate-400'
                      : displayPosition === 3
                        ? 'bg-gradient-to-r from-amber-50 to-amber-100 border-2 border-amber-500'
                        : 'bg-gradient-to-r from-slate-50 to-blue-50 border-2 border-slate-300'
                  : 'bg-white';

                const pointsClass = isInTheMoney
                  ? displayPosition === 1
                    ? 'text-yellow-700'
                    : displayPosition === 2
                      ? 'text-slate-700'
                      : displayPosition === 3
                        ? 'text-amber-700'
                        : 'text-slate-700'
                  : 'text-blue-600';

                return (
                  <Link
                    key={entry.id}
                    to={`/calcuttas/${calcuttaId}/entries/${entry.id}`}
                    className={`block p-4 rounded-lg shadow hover:shadow-md transition-shadow ${rowClass}`}
                  >
                    <div className="flex justify-between items-center">
                      <div>
                        <h2 className="text-xl font-semibold">
                          {displayPosition}. {entry.name}
                          {isInTheMoney && payoutText && <span className="ml-2 text-sm text-gray-700">{payoutText}</span>}
                        </h2>
                      </div>
                      <div className="text-right">
                        <p className={`text-2xl font-bold ${pointsClass}`}>
                          {entry.totalPoints ? entry.totalPoints.toFixed(2) : '0.00'} pts
                        </p>
                      </div>
                    </div>
                  </Link>
                );
              })}
            </div>
          )}
        </TabsContent>

        <TabsContent value="investment">
          {biddingOpen && mockData ? (
            <BiddingOverlay tournamentStartingAt={dashboardData!.tournamentStartingAt!}>
              <InvestmentTab
                entries={mockData.entries as typeof entries}
                schools={schools}
                tournamentTeams={tournamentTeams}
                allEntryTeams={mockData.entryTeams}
              />
            </BiddingOverlay>
          ) : (
            <InvestmentTab entries={entries} schools={schools} tournamentTeams={tournamentTeams} allEntryTeams={allEntryTeams} />
          )}
        </TabsContent>

        <TabsContent value="ownership">
          {biddingOpen ? (
            <BiddingOverlay tournamentStartingAt={dashboardData!.tournamentStartingAt!}>
              <OwnershipTab
                entries={mockData!.entries as typeof entries}
                schools={schools}
                tournamentTeams={tournamentTeams}
                allEntryTeams={mockData!.entryTeams}
                allCalcuttaPortfolios={[]}
                allCalcuttaPortfolioTeams={[]}
                isFetching={false}
              />
            </BiddingOverlay>
          ) : (
            <OwnershipTab
              entries={entries}
              schools={schools}
              tournamentTeams={tournamentTeams}
              allEntryTeams={allEntryTeams}
              allCalcuttaPortfolios={allCalcuttaPortfolios}
              allCalcuttaPortfolioTeams={allCalcuttaPortfolioTeams}
              isFetching={dashboardQuery.isFetching}
            />
          )}
        </TabsContent>

        <TabsContent value="returns">
          {biddingOpen ? (
            <BiddingOverlay tournamentStartingAt={dashboardData!.tournamentStartingAt!}>
              <ReturnsTab
                entries={mockData!.entries as typeof entries}
                schools={schools}
                tournamentTeams={tournamentTeams}
                allCalcuttaPortfolios={[]}
                allCalcuttaPortfolioTeams={[]}
              />
            </BiddingOverlay>
          ) : (
            <ReturnsTab
              entries={entries}
              schools={schools}
              tournamentTeams={tournamentTeams}
              allCalcuttaPortfolios={allCalcuttaPortfolios}
              allCalcuttaPortfolioTeams={allCalcuttaPortfolioTeams}
            />
          )}
        </TabsContent>

        <TabsContent value="statistics">
          {biddingOpen ? (
            <BiddingOverlay tournamentStartingAt={dashboardData!.tournamentStartingAt!}>
              <StatisticsTab
                calcuttaId={calcuttaId}
                totalEntries={dashboardData?.totalEntries ?? 0}
                totalInvestment={0}
                totalReturns={0}
                averageReturn={0}
                returnsStdDev={0}
                seedInvestmentData={[]}
                teamROIData={[]}
              />
            </BiddingOverlay>
          ) : (
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
          )}
        </TabsContent>
      </Tabs>
    </PageContainer>
  );
}
