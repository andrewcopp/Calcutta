import { useState } from 'react';
import { useParams } from 'react-router-dom';
import { useQuery } from '@tanstack/react-query';
import { Alert } from '../components/ui/Alert';
import { Badge } from '../components/ui/Badge';
import { Card } from '../components/ui/Card';
import { ErrorState } from '../components/ui/ErrorState';
import { EntryTeamsSkeleton } from '../components/skeletons/EntryTeamsSkeleton';
import { PageContainer, PageHeader } from '../components/ui/Page';
import { Breadcrumb } from '../components/ui/Breadcrumb';
import { EntryRosterCard } from '../components/EntryRosterCard';
import { DashboardSummary } from './CalcuttaEntries/DashboardSummary';
import { InvestmentsTab } from './EntryTeams/InvestmentsTab';
import { OwnershipsTab } from './EntryTeams/OwnershipsTab';
import { ReturnsTab } from './EntryTeams/ReturnsTab';
import { StatisticsTab } from './EntryTeams/StatisticsTab';
import { Tabs, TabsList, TabsTrigger, TabsContent } from '../components/ui/Tabs';
import { useCalcuttaDashboard } from '../hooks/useCalcuttaDashboard';
import { useEntryTeamsData } from '../hooks/useEntryTeamsData';
import { useEntryOwnershipData } from '../hooks/useEntryOwnershipData';
import { calcuttaService } from '../services/calcuttaService';
import { queryKeys } from '../queryKeys';
import { formatDate } from '../utils/format';

export function EntryTeamsPage() {
  const { entryId, calcuttaId } = useParams<{ entryId: string; calcuttaId: string }>();

  const [activeTab, setActiveTab] = useState('entry');
  const [sortBy, setSortBy] = useState<'points' | 'ownership' | 'bidPoints'>('points');
  const [investmentsSortBy, setInvestmentsSortBy] = useState<'total' | 'seed' | 'region' | 'team'>('total');
  const [showAllTeams, setShowAllTeams] = useState(false);
  const [ownershipShowAllTeams, setOwnershipShowAllTeams] = useState(false);
  const [returnsShowAllTeams, setReturnsShowAllTeams] = useState(false);

  const dashboardQuery = useCalcuttaDashboard(calcuttaId);

  const biddingOpen = dashboardQuery.data?.biddingOpen ?? false;
  const currentUserEntry = dashboardQuery.data?.currentUserEntry;
  const isOwnEntry = Boolean(currentUserEntry && currentUserEntry.id === entryId);

  const {
    calcuttaName,
    entryName,
    teams,
    schools,
    portfolios,
    portfolioTeams,
    tournamentTeams,
    allEntryTeams,
    allCalcuttaPortfolios,
    allCalcuttaPortfolioTeams,
  } = useEntryTeamsData(dashboardQuery.data, entryId);

  const ownEntryTeamsQuery = useQuery({
    queryKey: queryKeys.calcuttas.entryTeams(calcuttaId, entryId),
    enabled: Boolean(biddingOpen && isOwnEntry && calcuttaId && entryId),
    queryFn: () => calcuttaService.getEntryTeams(entryId!, calcuttaId!),
  });

  const { getPortfolioTeamData, getInvestorRanking, ownershipTeamsData } = useEntryOwnershipData({
    activeTab,
    entryId,
    teams,
    schools,
    tournamentTeams,
    portfolios,
    allCalcuttaPortfolioTeams,
    ownershipShowAllTeams,
    sortBy,
  });

  if (!entryId || !calcuttaId) {
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

  if (biddingOpen && !isOwnEntry) {
    return (
      <PageContainer>
        <Breadcrumb
          items={[
            { label: 'My Pools', href: '/pools' },
            { label: calcuttaName, href: `/pools/${calcuttaId}` },
            { label: 'Portfolio' },
          ]}
        />
        <PageHeader title="Portfolio" />
        <Card className="text-center">
          <p className="text-muted-foreground">This portfolio is hidden while bidding is open.</p>
        </Card>
      </PageContainer>
    );
  }

  const ownershipLoading = dashboardQuery.isFetching;

  const entryTeams = biddingOpen && isOwnEntry ? (ownEntryTeamsQuery.data ?? []) : teams;
  const entryTitle = isOwnEntry ? 'Your Portfolio' : entryName || 'Portfolio';

  return (
    <PageContainer>
      <Breadcrumb
        items={[
          { label: 'My Pools', href: '/pools' },
          { label: calcuttaName, href: `/pools/${calcuttaId}` },
          { label: entryName || 'Portfolio' },
        ]}
      />

      <PageHeader title={entryName || 'Portfolio'} />

      {isOwnEntry && !biddingOpen && dashboardQuery.data?.tournamentStartingAt && (
        <div className="mb-4 flex items-center gap-2">
          <Badge variant="secondary">Portfolios Revealed</Badge>
          <span className="text-sm text-muted-foreground">
            {formatDate(dashboardQuery.data.tournamentStartingAt, true)}
          </span>
        </div>
      )}

      {isOwnEntry &&
        !biddingOpen &&
        (() => {
          const totalSpent = teams.reduce((sum, et) => sum + et.bidPoints, 0);
          const budgetPoints = dashboardQuery.data?.calcutta?.budgetPoints ?? 100;
          return (
            <Card variant="accent" padding="compact" className="mb-6">
              <div className="flex items-center gap-3">
                <h3 className="text-lg font-semibold text-foreground">Your Portfolio</h3>
                <Badge variant={currentUserEntry?.status === 'accepted' ? 'success' : 'secondary'}>
                  {currentUserEntry?.status === 'accepted' ? 'Bids locked' : 'In Progress'}
                </Badge>
                <span className="text-sm text-muted-foreground">
                  {teams.length} teams &middot; {totalSpent} / {budgetPoints} credits
                </span>
              </div>
            </Card>
          );
        })()}

      {isOwnEntry && !biddingOpen && currentUserEntry && dashboardQuery.data && (
        <DashboardSummary
          currentEntry={currentUserEntry}
          entries={dashboardQuery.data.entries}
          portfolios={allCalcuttaPortfolios}
          portfolioTeams={allCalcuttaPortfolioTeams}
          tournamentTeams={tournamentTeams}
        />
      )}

      <Tabs value={activeTab} onValueChange={setActiveTab}>
        <TabsList>
          <TabsTrigger value="entry">Portfolio</TabsTrigger>
          {!biddingOpen && <TabsTrigger value="investments">Bids</TabsTrigger>}
          {!biddingOpen && <TabsTrigger value="ownerships">Shares</TabsTrigger>}
          {!biddingOpen && <TabsTrigger value="returns">Scoring</TabsTrigger>}
          {!biddingOpen && <TabsTrigger value="statistics">Stats</TabsTrigger>}
        </TabsList>

        <TabsContent value="entry">
          <EntryRosterCard
            entryId={entryId!}
            calcuttaId={calcuttaId!}
            entryStatus={currentUserEntry?.status ?? 'incomplete'}
            entryTeams={entryTeams}
            budgetPoints={dashboardQuery.data?.calcutta?.budgetPoints ?? 100}
            canEdit={biddingOpen && isOwnEntry}
            title={entryTitle}
          />
        </TabsContent>

        {!biddingOpen && (
          <TabsContent value="investments">
            <InvestmentsTab
              entryId={entryId!}
              tournamentTeams={tournamentTeams}
              allEntryTeams={allEntryTeams}
              schools={schools}
              investmentsSortBy={investmentsSortBy}
              setInvestmentsSortBy={setInvestmentsSortBy}
              showAllTeams={showAllTeams}
              setShowAllTeams={setShowAllTeams}
            />
          </TabsContent>
        )}

        {!biddingOpen && (
          <TabsContent value="ownerships">
            <OwnershipsTab
              ownershipShowAllTeams={ownershipShowAllTeams}
              setOwnershipShowAllTeams={setOwnershipShowAllTeams}
              sortBy={sortBy}
              setSortBy={setSortBy}
              ownershipLoading={ownershipLoading}
              ownershipTeamsData={ownershipTeamsData}
              getPortfolioTeamData={getPortfolioTeamData}
              getInvestorRanking={getInvestorRanking}
              allCalcuttaPortfolioTeams={allCalcuttaPortfolioTeams}
              allCalcuttaPortfolios={allCalcuttaPortfolios}
              portfolios={portfolios}
            />
          </TabsContent>
        )}

        {!biddingOpen && (
          <TabsContent value="returns">
            <ReturnsTab
              entryId={entryId!}
              returnsShowAllTeams={returnsShowAllTeams}
              setReturnsShowAllTeams={setReturnsShowAllTeams}
              sortBy={sortBy}
              setSortBy={setSortBy}
              tournamentTeams={tournamentTeams}
              allCalcuttaPortfolioTeams={allCalcuttaPortfolioTeams}
              teams={teams}
              schools={schools}
              getPortfolioTeamData={getPortfolioTeamData}
            />
          </TabsContent>
        )}

        {!biddingOpen && (
          <TabsContent value="statistics">
            <StatisticsTab
              portfolios={portfolios}
              portfolioTeams={portfolioTeams}
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
