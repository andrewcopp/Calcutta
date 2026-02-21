import { useCallback, useMemo, useState } from 'react';
import { useParams } from 'react-router-dom';
import { useQuery } from '@tanstack/react-query';
import { CalcuttaEntryTeam } from '../types/calcutta';
import { Alert } from '../components/ui/Alert';
import { ErrorState } from '../components/ui/ErrorState';
import { EntryTeamsSkeleton } from '../components/skeletons/EntryTeamsSkeleton';
import { PageContainer, PageHeader } from '../components/ui/Page';
import { Breadcrumb } from '../components/ui/Breadcrumb';
import { EntryRosterCard } from '../components/EntryRosterCard';
import { InvestmentsTab } from './EntryTeams/InvestmentsTab';
import { OwnershipsTab } from './EntryTeams/OwnershipsTab';
import { ReturnsTab } from './EntryTeams/ReturnsTab';
import { StatisticsTab } from './EntryTeams/StatisticsTab';
import { PortfolioScores } from './EntryTeams/PortfolioScores';
import { Tabs, TabsList, TabsTrigger, TabsContent } from '../components/ui/Tabs';
import { useCalcuttaDashboard } from '../hooks/useCalcuttaDashboard';
import { useEntryTeamsData } from '../hooks/useEntryTeamsData';
import { calcuttaService } from '../services/calcuttaService';
import { queryKeys } from '../queryKeys';

export function EntryTeamsPage() {
  const { entryId, calcuttaId } = useParams<{ entryId: string; calcuttaId: string }>();

  const [activeTab, setActiveTab] = useState('ownerships');
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

  // Helper function to find portfolio team data for a given team ID
  const getPortfolioTeamData = useCallback(
    (teamId: string) => {
      // Find the portfolio team belonging to the current entry's portfolio
      const currentPortfolioId = portfolios[0]?.id;
      if (!currentPortfolioId) return undefined;
      return allCalcuttaPortfolioTeams.find((pt) => pt.teamId === teamId && pt.portfolioId === currentPortfolioId);
    },
    [allCalcuttaPortfolioTeams, portfolios]
  );

  const ownershipTeamsData = useMemo(() => {
    if (activeTab !== 'ownerships') return [];
    if (!entryId) return [];

    let teamsToShow: CalcuttaEntryTeam[];

    if (ownershipShowAllTeams) {
      const schoolMap = new Map(schools.map((s) => [s.id, s]));
      teamsToShow = tournamentTeams.map((tt) => {
        const existingTeam = teams.find((t) => t.teamId === tt.id);
        if (existingTeam) return existingTeam;

        return {
          id: `synthetic-${tt.id}`,
          entryId: entryId,
          teamId: tt.id,
          bid: 0,
          created: new Date().toISOString(),
          updated: new Date().toISOString(),
          team: {
            ...tt,
            school: schoolMap.get(tt.schoolId),
          },
        } as CalcuttaEntryTeam;
      });
    } else {
      teamsToShow = teams.filter((team) => {
        const portfolioTeam = getPortfolioTeamData(team.teamId);
        return portfolioTeam && portfolioTeam.ownershipPercentage > 0;
      });
    }

    teamsToShow = teamsToShow.slice().sort((a, b) => {
      const portfolioTeamA = getPortfolioTeamData(a.teamId);
      const portfolioTeamB = getPortfolioTeamData(b.teamId);

      const pointsA = portfolioTeamA?.actualPoints || 0;
      const pointsB = portfolioTeamB?.actualPoints || 0;
      const ownershipA = portfolioTeamA?.ownershipPercentage || 0;
      const ownershipB = portfolioTeamB?.ownershipPercentage || 0;
      const bidA = a.bid;
      const bidB = b.bid;

      if (sortBy === 'points') {
        if (pointsB !== pointsA) return pointsB - pointsA;
        if (ownershipB !== ownershipA) return ownershipB - ownershipA;
        return bidB - bidA;
      }

      if (sortBy === 'ownership') {
        if (ownershipB !== ownershipA) return ownershipB - ownershipA;
        if (pointsB !== pointsA) return pointsB - pointsA;
        return bidB - bidA;
      }

      if (bidB !== bidA) return bidB - bidA;
      if (pointsB !== pointsA) return pointsB - pointsA;
      return ownershipB - ownershipA;
    });

    return teamsToShow;
  }, [activeTab, entryId, getPortfolioTeamData, ownershipShowAllTeams, schools, sortBy, teams, tournamentTeams]);

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

  if (biddingOpen && isOwnEntry) {
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
        {ownEntryTeamsQuery.data && (
          <EntryRosterCard
            entryId={entryId}
            calcuttaId={calcuttaId}
            entryName={entryName || 'Entry'}
            entryStatus={currentUserEntry?.status ?? 'draft'}
            entryTeams={ownEntryTeamsQuery.data}
            budgetPoints={dashboardQuery.data?.calcutta?.budgetPoints ?? 100}
          />
        )}
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

  // Helper function to calculate investor ranking for a team
  const getInvestorRanking = (teamId: string) => {
    // Get all portfolio teams for this team across all portfolios in the Calcutta
    const allInvestors = allCalcuttaPortfolioTeams.filter(pt => pt.teamId === teamId);

    // Sort investors by ownership percentage (descending)
    const sortedInvestors = [...allInvestors].sort((a, b) => b.ownershipPercentage - a.ownershipPercentage);

    // Find current user's rank
    const userPortfolio = portfolios[0];
    const userRank = userPortfolio ?
      sortedInvestors.findIndex(pt => pt.portfolioId === userPortfolio.id) + 1 :
      0;

    return {
      rank: userRank,
      total: allInvestors.length
    };
  };

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
          <TabsTrigger value="investments">Investments</TabsTrigger>
          <TabsTrigger value="ownerships">Ownerships</TabsTrigger>
          <TabsTrigger value="returns">Returns</TabsTrigger>
          <TabsTrigger value="statistics">Statistics</TabsTrigger>
        </TabsList>

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

        <TabsContent value="statistics">
          <StatisticsTab portfolios={portfolios} portfolioTeams={portfolioTeams} PortfolioScoresComponent={PortfolioScores} />
        </TabsContent>
      </Tabs>
    </PageContainer>
  );
}
