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


import { useCalcuttaDashboard } from '../hooks/useCalcuttaDashboard';
import { useCalcuttaEntriesData } from '../hooks/useCalcuttaEntriesData';
import { useUser } from '../contexts/useUser';
import { calcuttaService } from '../services/calcuttaService';

import { formatDollarsFromCents, formatDate } from '../utils/format';

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

  if (biddingOpen) {
    const statusLabelMap: Record<string, string> = { empty: 'Empty', invalid: 'Invalid', valid: 'Valid', accepted: 'Accepted', rejected: 'Rejected' };
    const statusVariantMap: Record<string, string> = { empty: 'secondary', invalid: 'warning', valid: 'default', accepted: 'success', rejected: 'destructive' };
    const entryStatusLabel = !currentUserEntry
      ? 'Not Started'
      : statusLabelMap[currentUserEntry.status] ?? currentUserEntry.status;
    const entryStatusVariant = !currentUserEntry
      ? 'secondary'
      : statusVariantMap[currentUserEntry.status] ?? 'secondary';

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

        {!currentUserEntry ? (
          <div className="p-4 border border-gray-200 rounded-lg bg-white shadow-sm">
            <div className="flex items-center justify-between">
              <div className="flex items-center gap-3">
                <h3 className="text-lg font-semibold text-gray-900">Your Entry</h3>
                <Badge variant={entryStatusVariant as 'secondary' | 'success' | 'warning'}>{entryStatusLabel}</Badge>
              </div>
              <Button onClick={handleCreateEntry} disabled={isCreatingEntry} size="sm">
                {isCreatingEntry ? 'Creating...' : 'Create Entry'}
              </Button>
            </div>
          </div>
        ) : (
          <Link
            to={`/calcuttas/${calcuttaId}/entries/${currentUserEntry.id}`}
            className="block p-4 border border-gray-200 rounded-lg bg-white shadow-sm hover:shadow-md transition-shadow"
          >
            <div className="flex items-center justify-between">
              <div className="flex items-center gap-3">
                <h3 className="text-lg font-semibold text-gray-900">{currentUserEntry.name}</h3>
                <Badge variant={entryStatusVariant as 'secondary' | 'success' | 'warning'}>{entryStatusLabel}</Badge>
              </div>
              <svg className="h-5 w-5 text-gray-400" fill="none" viewBox="0 0 24 24" strokeWidth="2" stroke="currentColor">
                <path strokeLinecap="round" strokeLinejoin="round" d="M8.25 4.5l7.5 7.5-7.5 7.5" />
              </svg>
            </div>
          </Link>
        )}

        <div className="mt-6 p-6 border border-blue-200 rounded-lg bg-blue-50 text-center">
          <p className="text-lg font-semibold text-blue-900 mb-2">
            Tournament hasn't started yet
          </p>
          <p className="text-blue-700">
            Come back once the tournament starts for the full leaderboard, ownership breakdowns, and live scoring.
          </p>
          {dashboardData?.tournamentStartingAt && (
            <p className="mt-3 text-sm text-blue-600">
              Portfolios revealed {formatDate(dashboardData.tournamentStartingAt, true)}
            </p>
          )}
        </div>

        <div className="mt-4 text-sm text-gray-500 text-center">
          {dashboardData!.totalEntries} {dashboardData!.totalEntries === 1 ? 'entry' : 'entries'} submitted
        </div>
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
            className="block p-4 border border-gray-200 rounded-lg bg-white shadow-sm hover:shadow-md transition-shadow"
          >
            <div className="flex items-center justify-between">
              <div className="flex items-center gap-3">
                <h3 className="text-lg font-semibold text-gray-900">Your Entry</h3>
                <Badge variant={(({ empty: 'secondary', invalid: 'warning', valid: 'default', accepted: 'success', rejected: 'destructive' } as Record<string, 'secondary' | 'success' | 'warning' | 'default' | 'destructive'>)[currentUserEntry.status]) ?? 'secondary'}>
                  {({ empty: 'Empty', invalid: 'Invalid', valid: 'Valid', accepted: 'Accepted', rejected: 'Rejected' } as Record<string, string>)[currentUserEntry.status] ?? currentUserEntry.status}
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

      <Tabs value={activeTab} onValueChange={setActiveTab}>
        <TabsList>
          <TabsTrigger value="leaderboard">Leaderboard</TabsTrigger>
          <TabsTrigger value="investment">Investments</TabsTrigger>
          <TabsTrigger value="ownership">Ownerships</TabsTrigger>
          <TabsTrigger value="returns">Returns</TabsTrigger>
          <TabsTrigger value="statistics">Statistics</TabsTrigger>
        </TabsList>

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
                    </div>
                  </div>
                </Link>
              );
            })}
          </div>
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
