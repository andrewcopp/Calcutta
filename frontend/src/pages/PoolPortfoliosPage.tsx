import { useState } from 'react';
import { Link, useNavigate, useParams, useSearchParams } from 'react-router-dom';

import { Pool } from '../schemas/pool';
import { Alert } from '../components/ui/Alert';
import { Card } from '../components/ui/Card';
import { ErrorState } from '../components/ui/ErrorState';
import { PageContainer, PageHeader } from '../components/ui/Page';
import { LeaderboardSkeleton } from '../components/skeletons/LeaderboardSkeleton';
import { LeaderboardTab } from './PoolPortfolios/LeaderboardTab';
import { RaceTab } from './PoolPortfolios/RaceTab';
import { InvestmentsTab } from './PoolPortfolios/InvestmentsTab';
import { ReturnsTab } from './PoolPortfolios/ReturnsTab';
import { OwnershipsTab } from './PoolPortfolios/OwnershipsTab';
import { FinalFourTab } from './PoolPortfolios/FinalFourTab';
import { BiddingOpenView } from './PoolPortfolios/BiddingOpenView';
import { DashboardSummary } from './PoolPortfolios/DashboardSummary';
import { Tabs, TabsList, TabsTrigger, TabsContent } from '../components/ui/Tabs';
import { Breadcrumb } from '../components/ui/Breadcrumb';
import { Badge } from '../components/ui/Badge';
import { Button } from '../components/ui/Button';

import { usePoolDashboard } from '../hooks/usePoolDashboard';
import { usePoolPortfoliosData } from '../hooks/usePoolPortfoliosData';
import { useUser } from '../contexts/UserContext';
import { poolService } from '../services/poolService';
import { toast } from '../lib/toast';

import { formatDate } from '../utils/format';

export function PoolPortfoliosPage() {
  const { poolId } = useParams<{ poolId: string }>();
  const [searchParams, setSearchParams] = useSearchParams();
  const validTabs = ['leaderboard', 'race', 'bids', 'shares', 'scoring', 'final-four'] as const;
  const tabParam = searchParams.get('tab');
  const activeTab = validTabs.includes(tabParam as (typeof validTabs)[number]) ? tabParam! : 'leaderboard';
  const setActiveTab = (tab: string) => setSearchParams({ tab }, { replace: true });
  const [isCreatingPortfolio, setIsCreatingPortfolio] = useState(false);
  const [createPortfolioError, setCreatePortfolioError] = useState<string | null>(null);
  const navigate = useNavigate();
  const { user } = useUser();

  const dashboardQuery = usePoolDashboard(poolId);
  const dashboardData = dashboardQuery.data;

  const pool: Pool | undefined = dashboardData?.pool;
  const poolName = pool?.name ?? '';

  const investingOpen = dashboardData?.investingOpen ?? false;
  const currentUserPortfolio = dashboardData?.currentUserPortfolio;

  const {
    portfolios,
    allOwnershipSummaries,
    allOwnershipDetails,
    allInvestments,
    schools,
    tournamentTeams,
  } = usePoolPortfoliosData(dashboardData);

  if (!poolId) {
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

  const handleCreatePortfolio = async () => {
    if (!user || !poolId) return;
    setIsCreatingPortfolio(true);
    setCreatePortfolioError(null);
    try {
      const portfolio = await poolService.createPortfolio(poolId, `${user.firstName} ${user.lastName}`);
      toast.success('Portfolio created!');
      navigate(`/pools/${poolId}/portfolios/${portfolio.id}/invest`);
    } catch (err) {
      setCreatePortfolioError(err instanceof Error ? err.message : 'Failed to create portfolio');
      setIsCreatingPortfolio(false);
    }
  };

  if (investingOpen) {
    return (
      <BiddingOpenView
        poolId={poolId}
        poolName={poolName}
        currentUserPortfolio={currentUserPortfolio}
        canEditSettings={dashboardData?.abilities?.canEditSettings}
        tournamentStartingAt={dashboardData?.tournamentStartingAt}
        totalPortfolios={dashboardData!.totalPortfolios}
        isCreatingPortfolio={isCreatingPortfolio}
        createPortfolioError={createPortfolioError}
        onCreatePortfolio={handleCreatePortfolio}
      />
    );
  }

  return (
    <PageContainer>
      <Breadcrumb items={[{ label: 'My Pools', href: '/pools' }, { label: poolName }]} />

      <PageHeader
        title={poolName}
        actions={
          dashboardData?.abilities?.canEditSettings ? (
            <Link to={`/pools/${poolId}/settings`}>
              <Button variant="outline" size="sm">
                Settings
              </Button>
            </Link>
          ) : undefined
        }
      />

      {dashboardData?.tournamentStartingAt && (
        <div className="mb-4 flex items-center gap-2">
          <Badge variant="secondary">Portfolios Revealed</Badge>
          <span className="text-sm text-muted-foreground">{formatDate(dashboardData.tournamentStartingAt, true)}</span>
        </div>
      )}

      {currentUserPortfolio &&
        (() => {
          const userInvestments = allInvestments.filter((et) => et.portfolioId === currentUserPortfolio.id);
          const totalSpent = userInvestments.reduce((sum, et) => sum + et.credits, 0);
          const budgetCredits = dashboardData?.pool?.budgetCredits ?? 100;
          return (
            <Link to={`/pools/${poolId}/portfolios/${currentUserPortfolio.id}`} className="block mb-6">
              <Card variant="accent" padding="compact" className="hover:shadow-md transition-shadow">
                <div className="flex items-center justify-between">
                  <div className="flex items-center gap-3">
                    <h3 className="text-lg font-semibold text-foreground">Your Portfolio</h3>
                    <Badge variant={currentUserPortfolio.status === 'submitted' ? 'success' : 'secondary'}>
                      {currentUserPortfolio.status === 'submitted' ? 'Investments locked' : 'In Progress'}
                    </Badge>
                    <span className="text-sm text-muted-foreground">
                      {userInvestments.length} teams &middot; {totalSpent} / {budgetCredits} credits
                    </span>
                  </div>
                  <svg
                    className="h-5 w-5 text-muted-foreground/60"
                    fill="none"
                    viewBox="0 0 24 24"
                    strokeWidth="2"
                    stroke="currentColor"
                  >
                    <path strokeLinecap="round" strokeLinejoin="round" d="M8.25 4.5l7.5 7.5-7.5 7.5" />
                  </svg>
                </div>
              </Card>
            </Link>
          );
        })()}

      {currentUserPortfolio && !investingOpen && (
        <DashboardSummary
          currentPortfolio={currentUserPortfolio}
          portfolios={portfolios}
          ownershipSummaries={allOwnershipSummaries}
          ownershipDetails={allOwnershipDetails}
          tournamentTeams={tournamentTeams}
        />
      )}

      <Tabs value={activeTab} onValueChange={setActiveTab}>
        <TabsList>
          <TabsTrigger value="leaderboard">Leaderboard</TabsTrigger>
          <TabsTrigger value="race">Race</TabsTrigger>
          <TabsTrigger value="bids">Investments</TabsTrigger>
          <TabsTrigger value="shares">Ownership</TabsTrigger>
          <TabsTrigger value="scoring">Returns</TabsTrigger>
          {dashboardData?.finalFourOutcomes && dashboardData.finalFourOutcomes.length > 0 && (
            <TabsTrigger value="final-four">Final Four</TabsTrigger>
          )}
        </TabsList>

        <TabsContent value="leaderboard">
          <LeaderboardTab poolId={poolId} portfolios={portfolios} dashboard={dashboardData!} />
        </TabsContent>

        <TabsContent value="race">
          <RaceTab portfolios={portfolios} dashboard={dashboardData!} />
        </TabsContent>

        <TabsContent value="bids">
          <InvestmentsTab
            portfolios={portfolios}
            schools={schools}
            tournamentTeams={tournamentTeams}
            allInvestments={allInvestments}
          />
        </TabsContent>

        <TabsContent value="shares">
          <OwnershipsTab
            portfolios={portfolios}
            schools={schools}
            tournamentTeams={tournamentTeams}
            allInvestments={allInvestments}
            allOwnershipSummaries={allOwnershipSummaries}
            allOwnershipDetails={allOwnershipDetails}
            isFetching={dashboardQuery.isFetching}
          />
        </TabsContent>

        <TabsContent value="scoring">
          <ReturnsTab
            portfolios={portfolios}
            schools={schools}
            tournamentTeams={tournamentTeams}
            allOwnershipSummaries={allOwnershipSummaries}
            allOwnershipDetails={allOwnershipDetails}
          />
        </TabsContent>

        <TabsContent value="final-four">
          <FinalFourTab
            portfolios={portfolios}
            dashboard={dashboardData!}
            schools={schools}
            tournamentTeams={tournamentTeams}
          />
        </TabsContent>

      </Tabs>
    </PageContainer>
  );
}
