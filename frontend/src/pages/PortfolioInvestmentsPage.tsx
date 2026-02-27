import { useState } from 'react';
import { useNavigate, useParams } from 'react-router-dom';
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query';
import { Alert } from '../components/ui/Alert';
import { Badge } from '../components/ui/Badge';
import { Button } from '../components/ui/Button';
import { Card } from '../components/ui/Card';
import { ErrorState } from '../components/ui/ErrorState';
import { EntryTeamsSkeleton } from '../components/skeletons/EntryTeamsSkeleton';
import { PageContainer, PageHeader } from '../components/ui/Page';
import { Breadcrumb } from '../components/ui/Breadcrumb';
import { PortfolioRosterCard } from '../components/PortfolioRosterCard';
import { DashboardSummary } from './PoolPortfolios/DashboardSummary';
import { InvestmentsTab } from './PortfolioInvestments/InvestmentsTab';
import { OwnershipsTab } from './PortfolioInvestments/OwnershipsTab';
import { ReturnsTab } from './PortfolioInvestments/ReturnsTab';
import { StatisticsTab } from './PortfolioInvestments/StatisticsTab';
import { Tabs, TabsList, TabsTrigger, TabsContent } from '../components/ui/Tabs';
import { usePoolDashboard } from '../hooks/usePoolDashboard';
import { usePortfolioInvestmentsData } from '../hooks/usePortfolioInvestmentsData';
import { usePortfolioOwnershipData } from '../hooks/usePortfolioOwnershipData';
import { poolService } from '../services/poolService';
import { queryKeys } from '../queryKeys';
import { toast } from '../lib/toast';
import { formatDate } from '../utils/format';

export function PortfolioInvestmentsPage() {
  const { portfolioId, poolId } = useParams<{ portfolioId: string; poolId: string }>();
  const navigate = useNavigate();
  const queryClient = useQueryClient();

  const [activeTab, setActiveTab] = useState('entry');
  const [sortBy, setSortBy] = useState<'points' | 'ownership' | 'credits'>('points');
  const [investmentsSortBy, setInvestmentsSortBy] = useState<'total' | 'seed' | 'region' | 'team'>('total');
  const [showAllTeams, setShowAllTeams] = useState(false);
  const [ownershipShowAllTeams, setOwnershipShowAllTeams] = useState(false);
  const [returnsShowAllTeams, setReturnsShowAllTeams] = useState(false);

  const dashboardQuery = usePoolDashboard(poolId);

  const investingOpen = dashboardQuery.data?.investingOpen ?? false;
  const currentUserPortfolio = dashboardQuery.data?.currentUserPortfolio;
  const isOwnPortfolio = Boolean(currentUserPortfolio && currentUserPortfolio.id === portfolioId);

  const {
    poolName,
    portfolioName,
    teams,
    schools,
    ownershipSummaries,
    ownershipDetails,
    tournamentTeams,
    allInvestments,
    allOwnershipSummaries,
    allOwnershipDetails,
  } = usePortfolioInvestmentsData(dashboardQuery.data, portfolioId);

  const ownInvestmentsQuery = useQuery({
    queryKey: queryKeys.pools.investments(poolId, portfolioId),
    enabled: Boolean(investingOpen && isOwnPortfolio && poolId && portfolioId),
    queryFn: () => poolService.getInvestments(portfolioId!, poolId!),
  });

  const { getOwnershipDetailData, getInvestorRanking, ownershipTeamsData } = usePortfolioOwnershipData({
    activeTab,
    portfolioId,
    teams,
    schools,
    tournamentTeams,
    ownershipSummaries,
    allOwnershipDetails,
    ownershipShowAllTeams,
    sortBy,
  });

  const deletePortfolioMutation = useMutation({
    mutationFn: async () => {
      if (!poolId || !portfolioId) throw new Error('Missing IDs');
      return poolService.deletePortfolio(poolId, portfolioId);
    },
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: queryKeys.pools.dashboard(poolId) });
      toast.success('Portfolio deleted');
      navigate(`/pools/${poolId}`);
    },
  });

  const handleDeletePortfolio = () => {
    const confirmed = window.confirm('Are you sure you want to delete your portfolio? This cannot be undone.');
    if (!confirmed) return;
    deletePortfolioMutation.mutate();
  };

  if (!portfolioId || !poolId) {
    return (
      <PageContainer>
        <Alert variant="error">Missing required parameters</Alert>
      </PageContainer>
    );
  }

  if (dashboardQuery.isLoading) {
    return (
      <PageContainer>
        <EntryTeamsSkeleton />
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

  if (investingOpen && !isOwnPortfolio) {
    return (
      <PageContainer>
        <Breadcrumb
          items={[
            { label: 'My Pools', href: '/pools' },
            { label: poolName, href: `/pools/${poolId}` },
            { label: 'Portfolio' },
          ]}
        />
        <PageHeader title="Portfolio" />
        <Card className="text-center">
          <p className="text-muted-foreground">Scouting reports stay sealed until tip-off.</p>
        </Card>
      </PageContainer>
    );
  }

  const ownershipLoading = dashboardQuery.isFetching;

  const portfolioInvestments = investingOpen && isOwnPortfolio ? (ownInvestmentsQuery.data ?? []) : teams;
  const portfolioTitle = isOwnPortfolio ? 'Your Portfolio' : portfolioName || 'Portfolio';

  return (
    <PageContainer>
      <Breadcrumb
        items={[
          { label: 'My Pools', href: '/pools' },
          { label: poolName, href: `/pools/${poolId}` },
          { label: portfolioName || 'Portfolio' },
        ]}
      />

      <PageHeader title={portfolioName || 'Portfolio'} />

      {isOwnPortfolio && !investingOpen && dashboardQuery.data?.tournamentStartingAt && (
        <div className="mb-4 flex items-center gap-2">
          <Badge variant="secondary">Portfolios Revealed</Badge>
          <span className="text-sm text-muted-foreground">
            {formatDate(dashboardQuery.data.tournamentStartingAt, true)}
          </span>
        </div>
      )}

      {isOwnPortfolio &&
        !investingOpen &&
        (() => {
          const totalSpent = teams.reduce((sum, et) => sum + et.credits, 0);
          const budgetCredits = dashboardQuery.data?.pool?.budgetCredits ?? 100;
          return (
            <Card variant="accent" padding="compact" className="mb-6">
              <div className="flex items-center gap-3">
                <h3 className="text-lg font-semibold text-foreground">Your Portfolio</h3>
                <span className="text-sm text-muted-foreground">
                  {teams.length} teams &middot; {totalSpent} / {budgetCredits} credits
                </span>
              </div>
            </Card>
          );
        })()}

      {isOwnPortfolio && !investingOpen && currentUserPortfolio && dashboardQuery.data && (
        <DashboardSummary
          currentPortfolio={currentUserPortfolio}
          portfolios={dashboardQuery.data.portfolios}
          ownershipSummaries={allOwnershipSummaries}
          ownershipDetails={allOwnershipDetails}
          tournamentTeams={tournamentTeams}
        />
      )}

      <Tabs value={activeTab} onValueChange={setActiveTab}>
        <TabsList>
          <TabsTrigger value="entry">Portfolio</TabsTrigger>
          {!investingOpen && <TabsTrigger value="investments">Investments</TabsTrigger>}
          {!investingOpen && <TabsTrigger value="ownerships">Ownership</TabsTrigger>}
          {!investingOpen && <TabsTrigger value="returns">Returns</TabsTrigger>}
          {!investingOpen && <TabsTrigger value="statistics">Stats</TabsTrigger>}
        </TabsList>

        <TabsContent value="entry">
          <PortfolioRosterCard
            portfolioId={portfolioId!}
            poolId={poolId!}
            investments={portfolioInvestments}
            budgetCredits={dashboardQuery.data?.pool?.budgetCredits ?? 100}
            canEdit={investingOpen && isOwnPortfolio}
            title={portfolioTitle}
          />
          {investingOpen && isOwnPortfolio && (
            <div className="mt-4">
              {deletePortfolioMutation.isError && (
                <Alert variant="error" className="mb-2">
                  {deletePortfolioMutation.error instanceof Error ? deletePortfolioMutation.error.message : 'Failed to delete portfolio'}
                </Alert>
              )}
              <Button
                variant="destructive"
                size="sm"
                onClick={handleDeletePortfolio}
                disabled={deletePortfolioMutation.isPending}
                loading={deletePortfolioMutation.isPending}
              >
                Delete Portfolio
              </Button>
            </div>
          )}
        </TabsContent>

        {!investingOpen && (
          <TabsContent value="investments">
            <InvestmentsTab
              portfolioId={portfolioId!}
              tournamentTeams={tournamentTeams}
              allInvestments={allInvestments}
              schools={schools}
              investmentsSortBy={investmentsSortBy}
              setInvestmentsSortBy={setInvestmentsSortBy}
              showAllTeams={showAllTeams}
              setShowAllTeams={setShowAllTeams}
            />
          </TabsContent>
        )}

        {!investingOpen && (
          <TabsContent value="ownerships">
            <OwnershipsTab
              ownershipShowAllTeams={ownershipShowAllTeams}
              setOwnershipShowAllTeams={setOwnershipShowAllTeams}
              sortBy={sortBy}
              setSortBy={setSortBy}
              ownershipLoading={ownershipLoading}
              ownershipTeamsData={ownershipTeamsData}
              getOwnershipDetailData={getOwnershipDetailData}
              getInvestorRanking={getInvestorRanking}
              allOwnershipDetails={allOwnershipDetails}
              allOwnershipSummaries={allOwnershipSummaries}
              ownershipSummaries={ownershipSummaries}
            />
          </TabsContent>
        )}

        {!investingOpen && (
          <TabsContent value="returns">
            <ReturnsTab
              portfolioId={portfolioId!}
              returnsShowAllTeams={returnsShowAllTeams}
              setReturnsShowAllTeams={setReturnsShowAllTeams}
              sortBy={sortBy}
              setSortBy={setSortBy}
              tournamentTeams={tournamentTeams}
              allOwnershipDetails={allOwnershipDetails}
              teams={teams}
              schools={schools}
              getOwnershipDetailData={getOwnershipDetailData}
            />
          </TabsContent>
        )}

        {!investingOpen && (
          <TabsContent value="statistics">
            <StatisticsTab
              ownershipSummaries={ownershipSummaries}
              ownershipDetails={ownershipDetails}
              teams={teams}
              tournamentTeams={tournamentTeams}
              schools={schools}
            />
          </TabsContent>
        )}
      </Tabs>
    </PageContainer>
  );
}
