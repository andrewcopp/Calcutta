import React, { useCallback, useMemo, useState } from 'react';
import { useParams } from 'react-router-dom';
import { CalcuttaEntryTeam, CalcuttaPortfolio, CalcuttaPortfolioTeam } from '../types/calcutta';
import { Alert } from '../components/ui/Alert';
import { EntryTeamsSkeleton } from '../components/skeletons/EntryTeamsSkeleton';
import { PageContainer, PageHeader } from '../components/ui/Page';
import { Breadcrumb } from '../components/ui/Breadcrumb';
import { InvestmentsTab } from './EntryTeams/InvestmentsTab';
import { OwnershipsTab } from './EntryTeams/OwnershipsTab';
import { ReturnsTab } from './EntryTeams/ReturnsTab';
import { StatisticsTab } from './EntryTeams/StatisticsTab';
import { Tabs, TabsList, TabsTrigger, TabsContent } from '../components/ui/Tabs';
import { useCalcuttaDashboard } from '../hooks/useCalcuttaDashboard';

// Add a new section to display portfolio scores
const PortfolioScores: React.FC<{ portfolio: CalcuttaPortfolio; teams: CalcuttaPortfolioTeam[] }> = ({
  portfolio,
  teams,
}) => {
  return (
    <div className="bg-white shadow rounded-lg p-6 mb-6">
      <h3 className="text-lg font-semibold mb-4">Portfolio Scores</h3>
      <div className="grid grid-cols-1 gap-4">
        <div className="flex justify-between items-center">
          <span className="text-gray-600">Maximum Possible Score:</span>
          <span className="font-medium">{portfolio.maximumPoints.toFixed(2)}</span>
        </div>
        <div className="border-t pt-4">
          <h4 className="text-md font-medium mb-2">Team Scores</h4>
          <div className="space-y-2">
            {teams.map((team) => (
              <div key={team.id} className="flex justify-between items-center">
                <span className="text-gray-600">{team.team?.school?.name || 'Unknown Team'}</span>
                <div className="text-right">
                  <div className="text-sm text-gray-500">
                    Expected: {team.expectedPoints.toFixed(2)}
                  </div>
                  <div className="text-sm text-gray-500">
                    Predicted: {team.predictedPoints.toFixed(2)}
                  </div>
                </div>
              </div>
            ))}
          </div>
        </div>
      </div>
    </div>
  );
};

export function EntryTeamsPage() {
  const { entryId, calcuttaId } = useParams<{ entryId: string; calcuttaId: string }>();

  const [activeTab, setActiveTab] = useState('ownerships');
  const [sortBy, setSortBy] = useState<'points' | 'ownership' | 'bid'>('points');
  const [investmentsSortBy, setInvestmentsSortBy] = useState<'total' | 'seed' | 'region' | 'team'>('total');
  const [showAllTeams, setShowAllTeams] = useState(false);
  const [ownershipShowAllTeams, setOwnershipShowAllTeams] = useState(false);
  const [returnsShowAllTeams, setReturnsShowAllTeams] = useState(false);

  const dashboardQuery = useCalcuttaDashboard(calcuttaId);

  const enrichedData = useMemo(() => {
    if (!dashboardQuery.data || !entryId) {
      return {
        calcuttaName: '',
        entryName: '',
        teams: [] as CalcuttaEntryTeam[],
        schools: [] as { id: string; name: string }[],
        portfolios: [] as CalcuttaPortfolio[],
        portfolioTeams: [] as CalcuttaPortfolioTeam[],
        tournamentTeams: dashboardQuery.data?.tournamentTeams ?? [],
        allEntryTeams: [] as CalcuttaEntryTeam[],
        allCalcuttaPortfolios: [] as (CalcuttaPortfolio & { entryName?: string })[],
        allCalcuttaPortfolioTeams: [] as CalcuttaPortfolioTeam[],
      };
    }

    const { calcutta, entries, entryTeams, portfolios, portfolioTeams, schools, tournamentTeams } = dashboardQuery.data;
    const schoolMap = new Map(schools.map((s) => [s.id, s]));
    const entryNameMap = new Map(entries.map((e) => [e.id, e.name]));

    const currentEntry = entries.find((e) => e.id === entryId);
    const entryName = currentEntry?.name || '';

    // Filter entry teams for this entry
    const thisEntryTeams = entryTeams.filter((et) => et.entryId === entryId);

    // Enrich entry teams with school info
    const teamsWithSchools: CalcuttaEntryTeam[] = thisEntryTeams.map((team) => ({
      ...team,
      team: team.team
        ? {
            ...team.team,
            school: schoolMap.get(team.team.schoolId),
          }
        : undefined,
    }));

    // Filter portfolios for this entry
    const thisEntryPortfolios = portfolios.filter((p) => p.entryId === entryId);

    // Filter portfolio teams for this entry's portfolios
    const thisPortfolioIds = new Set(thisEntryPortfolios.map((p) => p.id));
    const thisPortfolioTeams: CalcuttaPortfolioTeam[] = portfolioTeams
      .filter((pt) => thisPortfolioIds.has(pt.portfolioId))
      .map((pt) => ({
        ...pt,
        team: pt.team
          ? {
              ...pt.team,
              school: schoolMap.get(pt.team.schoolId),
            }
          : undefined,
      }));

    // Enrich all portfolios with entry names
    const allPortfoliosWithNames: (CalcuttaPortfolio & { entryName?: string })[] = portfolios.map((p) => ({
      ...p,
      entryName: entryNameMap.get(p.entryId),
    }));

    // Enrich all portfolio teams with school info
    const allPortfolioTeamsWithSchools: CalcuttaPortfolioTeam[] = portfolioTeams.map((pt) => ({
      ...pt,
      team: pt.team
        ? {
            ...pt.team,
            school: schoolMap.get(pt.team.schoolId),
          }
        : undefined,
    }));

    return {
      calcuttaName: calcutta.name,
      entryName,
      teams: teamsWithSchools,
      schools,
      portfolios: thisEntryPortfolios,
      portfolioTeams: thisPortfolioTeams,
      tournamentTeams,
      allEntryTeams: entryTeams,
      allCalcuttaPortfolios: allPortfoliosWithNames,
      allCalcuttaPortfolioTeams: allPortfolioTeamsWithSchools,
    };
  }, [dashboardQuery.data, entryId]);

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
  } = enrichedData;

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
      const bidA = a.bid || 0;
      const bidB = b.bid || 0;

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
    const message = dashboardQuery.error instanceof Error ? dashboardQuery.error.message : 'Failed to fetch data';
    return (
      <PageContainer>
        <Alert variant="error">{message}</Alert>
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
