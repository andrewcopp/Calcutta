import { useCallback, useMemo, useState } from 'react';
import { Link, useParams } from 'react-router-dom';
import { CalcuttaEntryTeam, CalcuttaPortfolio, CalcuttaPortfolioTeam } from '../types/calcutta';
import { calcuttaService } from '../services/calcuttaService';
import { useQuery } from '@tanstack/react-query';
import { InvestmentsTab } from './EntryTeams/InvestmentsTab';
import { OwnershipsTab } from './EntryTeams/OwnershipsTab';
import { ReturnsTab } from './EntryTeams/ReturnsTab';
import { StatisticsTab } from './EntryTeams/StatisticsTab';
import { TabsNav } from '../components/TabsNav';

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

  const [activeTab, setActiveTab] = useState<'investments' | 'ownerships' | 'returns' | 'statistics'>('ownerships');
  const [sortBy, setSortBy] = useState<'points' | 'ownership' | 'bid'>('points');
  const [investmentsSortBy, setInvestmentsSortBy] = useState<'total' | 'seed' | 'region' | 'team'>('total');
  const [showAllTeams, setShowAllTeams] = useState(false);
  const [ownershipShowAllTeams, setOwnershipShowAllTeams] = useState(false);
  const [returnsShowAllTeams, setReturnsShowAllTeams] = useState(false);

  const tabs = useMemo(
    () =>
      [
        { id: 'investments' as const, label: 'Investments' },
        { id: 'ownerships' as const, label: 'Ownerships' },
        { id: 'returns' as const, label: 'Returns' },
        { id: 'statistics' as const, label: 'Statistics' },
      ] as const,
    []
  );

  const entryTeamsQuery = useQuery({
    queryKey: ['entryTeamsPage', calcuttaId, entryId],
    enabled: Boolean(entryId && calcuttaId),
    staleTime: 30_000,
    queryFn: async () => {
      if (!entryId || !calcuttaId) {
        throw new Error('Missing required parameters');
      }

      const calcutta = await calcuttaService.getCalcutta(calcuttaId);

      const [teamsData, schoolsData, portfoliosData, allEntriesData, tournamentTeamsData] = await Promise.all([
        calcuttaService.getEntryTeams(entryId, calcuttaId),
        calcuttaService.getSchools(),
        calcuttaService.getPortfoliosByEntry(entryId),
        calcuttaService.getCalcuttaEntries(calcuttaId),
        calcuttaService.getTournamentTeams(calcutta.tournamentId),
      ]);

      const currentEntry = allEntriesData.find((e) => e.id === entryId);
      const entryName = currentEntry?.name || '';

      const entryNameMap = new Map(allEntriesData.map((entry) => [entry.id, entry.name]));
      const schoolMap = new Map(schoolsData.map((school) => [school.id, school]));

      const teamsWithSchools = teamsData.map((team) => ({
        ...team,
        team: team.team
          ? {
              ...team.team,
              school: schoolMap.get(team.team.schoolId),
            }
          : undefined,
      }));

      let portfolioTeamsWithSchools: CalcuttaPortfolioTeam[] = [];
      if (portfoliosData.length > 0) {
        const portfolioTeamsResults = await Promise.all(portfoliosData.map((portfolio) => calcuttaService.getPortfolioTeams(portfolio.id)));
        const allPortfolioTeams = portfolioTeamsResults.flat();

        portfolioTeamsWithSchools = allPortfolioTeams.map((team) => ({
          ...team,
          team: team.team
            ? {
                ...team.team,
                school: schoolMap.get(team.team.schoolId),
              }
            : undefined,
        }));
      }

      const allPortfoliosResults = await Promise.all(allEntriesData.map((entry) => calcuttaService.getPortfoliosByEntry(entry.id)));
      const allPortfoliosFlat = allPortfoliosResults.flat();

      const allPortfoliosWithEntryNames = allPortfoliosFlat.map((portfolio) => ({
        ...portfolio,
        entryName: entryNameMap.get(portfolio.entryId),
      }));

      const allPortfolioTeamsResults = await Promise.all(allPortfoliosFlat.map((portfolio) => calcuttaService.getPortfolioTeams(portfolio.id)));
      const allCalcuttaPortfolioTeams = allPortfolioTeamsResults.flat();

      const allCalcuttaPortfolioTeamsWithSchools = allCalcuttaPortfolioTeams.map((team) => ({
        ...team,
        team: team.team
          ? {
              ...team.team,
              school: schoolMap.get(team.team.schoolId),
            }
          : undefined,
      }));

      const allEntryTeamsResults = await Promise.all(allEntriesData.map((entry) => calcuttaService.getEntryTeams(entry.id, calcuttaId)));
      const allEntryTeams = allEntryTeamsResults.flat();

      return {
        entryName,
        teams: teamsWithSchools,
        schools: schoolsData,
        portfolios: portfoliosData,
        portfolioTeams: portfolioTeamsWithSchools,
        tournamentTeams: tournamentTeamsData,
        allEntryTeams,
        allCalcuttaPortfolios: allPortfoliosWithEntryNames,
        allCalcuttaPortfolioTeams: allCalcuttaPortfolioTeamsWithSchools,
      };
    },
  });

  const entryName = entryTeamsQuery.data?.entryName || '';
  const teams = entryTeamsQuery.data?.teams || [];
  const schools = entryTeamsQuery.data?.schools || [];
  const portfolios = entryTeamsQuery.data?.portfolios || [];
  const portfolioTeams = entryTeamsQuery.data?.portfolioTeams || [];
  const tournamentTeams = entryTeamsQuery.data?.tournamentTeams || [];
  const allEntryTeams = entryTeamsQuery.data?.allEntryTeams || [];
  const allCalcuttaPortfolios = entryTeamsQuery.data?.allCalcuttaPortfolios || [];
  const allCalcuttaPortfolioTeams = entryTeamsQuery.data?.allCalcuttaPortfolioTeams || [];

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
    return <div className="error">Missing required parameters</div>;
  }

  if (entryTeamsQuery.isLoading) {
    return <div>Loading...</div>;
  }

  if (entryTeamsQuery.isError) {
    const message = entryTeamsQuery.error instanceof Error ? entryTeamsQuery.error.message : 'Failed to fetch data';
    return <div className="error">{message}</div>;
  }

  const ownershipLoading = entryTeamsQuery.isFetching;

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

  const compareDesc = (a: number, b: number) => b - a;

  const compareTeams = (a: CalcuttaEntryTeam, b: CalcuttaEntryTeam) => {
    const portfolioTeamA = getPortfolioTeamData(a.teamId);
    const portfolioTeamB = getPortfolioTeamData(b.teamId);

    const pointsA = portfolioTeamA?.actualPoints || 0;
    const pointsB = portfolioTeamB?.actualPoints || 0;
    const ownershipA = portfolioTeamA?.ownershipPercentage || 0;
    const ownershipB = portfolioTeamB?.ownershipPercentage || 0;
    const bidA = a.bid || 0;
    const bidB = b.bid || 0;

    if (sortBy === 'points') {
      const byPoints = compareDesc(pointsA, pointsB);
      if (byPoints !== 0) return byPoints;
      const byOwnership = compareDesc(ownershipA, ownershipB);
      if (byOwnership !== 0) return byOwnership;
      return compareDesc(bidA, bidB);
    }

    if (sortBy === 'ownership') {
      const byOwnership = compareDesc(ownershipA, ownershipB);
      if (byOwnership !== 0) return byOwnership;
      const byPoints = compareDesc(pointsA, pointsB);
      if (byPoints !== 0) return byPoints;
      return compareDesc(bidA, bidB);
    }

    const byBid = compareDesc(bidA, bidB);
    if (byBid !== 0) return byBid;
    const byPoints = compareDesc(pointsA, pointsB);
    if (byPoints !== 0) return byPoints;
    return compareDesc(ownershipA, ownershipB);
  };

  const sortedTeams = [...teams].sort(compareTeams);


  return (
    <div className="container mx-auto px-4 py-8">
      <div className="mb-6">
        <Link to={`/calcuttas/${calcuttaId}`} className="text-blue-600 hover:text-blue-800">← Back to Entries</Link>
      </div>
      <h1 className="text-3xl font-bold mb-6">{entryName || 'Entry'}</h1>

      <TabsNav tabs={tabs} activeTab={activeTab} onTabChange={setActiveTab} />

      {activeTab === 'investments' && (
        <InvestmentsTab
          entryId={entryId!}
          teams={teams}
          tournamentTeams={tournamentTeams}
          allEntryTeams={allEntryTeams}
          schools={schools}
          investmentsSortBy={investmentsSortBy}
          setInvestmentsSortBy={setInvestmentsSortBy}
          showAllTeams={showAllTeams}
          setShowAllTeams={setShowAllTeams}
        />
      )}

      {activeTab === 'ownerships' && (
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
      )}

      {activeTab === 'returns' && (
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
          portfolios={portfolios}
          getPortfolioTeamData={getPortfolioTeamData}
        />
      )}

      {activeTab === 'statistics' && (
        <StatisticsTab portfolios={portfolios} portfolioTeams={portfolioTeams} PortfolioScoresComponent={PortfolioScores} />
      )}
    </div>
  );
}
 