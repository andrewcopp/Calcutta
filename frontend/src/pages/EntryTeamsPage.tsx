import { useState } from 'react';
import { useParams } from 'react-router-dom';
import { useQuery } from '@tanstack/react-query';
import { Alert } from '../components/ui/Alert';
import { ErrorState } from '../components/ui/ErrorState';
import { EntryTeamsSkeleton } from '../components/skeletons/EntryTeamsSkeleton';
import { PageContainer, PageHeader } from '../components/ui/Page';
import { Breadcrumb } from '../components/ui/Breadcrumb';
import { EntryTab } from './EntryTeams/EntryTab';
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

export function EntryTeamsPage() {
  const { entryId, calcuttaId } = useParams<{ entryId: string; calcuttaId: string }>();

  const [activeTab, setActiveTab] = useState('entry');
  const [sortBy, setSortBy] = useState<'points' | 'ownership' | 'bid'>('points');
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
            { label: 'Calcuttas', href: '/calcuttas' },
            { label: calcuttaName, href: `/calcuttas/${calcuttaId}` },
            { label: 'Entry' },
          ]}
        />
        <PageHeader title="Entry" />
        <div className="p-6 border border-gray-200 rounded-lg bg-white shadow-sm text-center">
          <p className="text-gray-600">This entry is hidden while bidding is open.</p>
        </div>
      </PageContainer>
    );
  }

  const ownershipLoading = dashboardQuery.isFetching;

  const entryTeams = biddingOpen && isOwnEntry ? ownEntryTeamsQuery.data ?? [] : teams;
  const entryTitle = isOwnEntry ? 'Your Entry' : (entryName || 'Entry');

  return (
    <PageContainer>
      <Breadcrumb
        items={[
          { label: 'Calcuttas', href: '/calcuttas' },
          { label: calcuttaName, href: `/calcuttas/${calcuttaId}` },
          { label: entryName || 'Entry' },
        ]}
      />

      <PageHeader title={entryName || 'Entry'} />

      <Tabs value={activeTab} onValueChange={setActiveTab}>
        <TabsList>
          <TabsTrigger value="entry">Entry</TabsTrigger>
          {!biddingOpen && <TabsTrigger value="investments">Investments</TabsTrigger>}
          {!biddingOpen && <TabsTrigger value="ownerships">Ownerships</TabsTrigger>}
          {!biddingOpen && <TabsTrigger value="returns">Returns</TabsTrigger>}
          {!biddingOpen && <TabsTrigger value="statistics">Statistics</TabsTrigger>}
        </TabsList>

        <TabsContent value="entry">
          <EntryTab
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
            <StatisticsTab portfolios={portfolios} portfolioTeams={portfolioTeams} />
          </TabsContent>
        )}
      </Tabs>
    </PageContainer>
  );
}
