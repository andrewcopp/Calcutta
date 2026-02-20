import { useState } from 'react';
import { Link, useParams } from 'react-router-dom';
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
import { useUser } from '../contexts/useUser';
import { useCalcuttaDashboard } from '../hooks/useCalcuttaDashboard';
import { useCalcuttaEntriesData } from '../hooks/useCalcuttaEntriesData';
import { formatDollarsFromCents } from '../utils/format';

export function CalcuttaEntriesPage() {
  const { calcuttaId } = useParams<{ calcuttaId: string }>();
  const { user } = useUser();
  const [activeTab, setActiveTab] = useState('leaderboard');

  const dashboardQuery = useCalcuttaDashboard(calcuttaId);
  const dashboardData = dashboardQuery.data;

  const calcutta: Calcutta | undefined = dashboardData?.calcutta;
  if (dashboardData && !calcutta) {
    console.warn('CalcuttaEntriesPage: dashboard loaded but calcutta is missing', { calcuttaId });
  }
  const calcuttaName = calcutta?.name ?? '';

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
          user?.id === calcutta?.ownerId ? (
            <Link to={`/calcuttas/${calcuttaId}/settings`}>
              <Button variant="outline" size="sm">Settings</Button>
            </Link>
          ) : undefined
        }
      />

      {calcutta && (
        <div className="mb-4">
          {calcutta.biddingOpen ? (
            <Badge variant="success">Bidding Open</Badge>
          ) : calcutta.biddingLockedAt ? (
            <Badge variant="warning">Bidding Locked</Badge>
          ) : (
            <Badge variant="secondary">Bidding Closed</Badge>
          )}
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

        <TabsContent value="returns">
          <ReturnsTab
            entries={entries}
            schools={schools}
            tournamentTeams={tournamentTeams}
            allCalcuttaPortfolios={allCalcuttaPortfolios}
            allCalcuttaPortfolioTeams={allCalcuttaPortfolioTeams}
          />
        </TabsContent>

        <TabsContent value="investment">
          <InvestmentTab entries={entries} schools={schools} tournamentTeams={tournamentTeams} allEntryTeams={allEntryTeams} />
        </TabsContent>

        <TabsContent value="ownership">
          <OwnershipTab
            entries={entries}
            schools={schools}
            tournamentTeams={tournamentTeams}
            allEntryTeams={allEntryTeams}
            allCalcuttaPortfolios={allCalcuttaPortfolios}
            allCalcuttaPortfolioTeams={allCalcuttaPortfolioTeams}
            isFetching={dashboardQuery.isFetching}
          />
        </TabsContent>

        <TabsContent value="leaderboard">
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
                      {index > 0 && entries[0].totalPoints > 0 && (
                        <p className="text-xs text-gray-500">
                          {((entry.totalPoints || 0) - (entries[0].totalPoints || 0)).toFixed(2)} pts
                        </p>
                      )}
                    </div>
                  </div>
                </Link>
              );
            })}
          </div>
        </TabsContent>
      </Tabs>
    </PageContainer>
  );
}
