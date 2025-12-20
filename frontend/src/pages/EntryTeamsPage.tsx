import { useEffect, useState } from 'react';
import { Link, useParams } from 'react-router-dom';
import { CalcuttaEntryTeam, CalcuttaPortfolio, CalcuttaPortfolioTeam, School } from '../types/calcutta';
import { calcuttaService } from '../services/calcuttaService';
import { PieChart, Pie, Cell, ResponsiveContainer, Tooltip } from 'recharts';

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

// Add a new component for the ownership pie chart
const OwnershipPieChart: React.FC<{ 
  portfolioTeams: CalcuttaPortfolioTeam[]; 
  portfolios: (CalcuttaPortfolio & { entryName?: string })[];
  currentPortfolioId?: string;
  rank?: number;
  totalInvestors?: number;
  sizePx?: number;
  showRank?: boolean;
}> = ({ 
  portfolioTeams,
  portfolios,
  currentPortfolioId,
  rank,
  totalInvestors,
  sizePx = 220,
  showRank = true,
}) => {
  // Transform portfolio teams data for the pie chart
  const data = portfolioTeams.map(pt => {
    const portfolio = portfolios.find(p => p.id === pt.portfolioId);
    const isCurrentPortfolio = pt.portfolioId === currentPortfolioId;
    return {
      name: portfolio?.entryName || `Portfolio ${portfolio?.id.slice(0, 4) || '??'}`,
      value: pt.ownershipPercentage * 100,
      portfolioId: pt.portfolioId,
      isCurrentPortfolio
    };
  });

  // Generate colors for each segment
  const COLORS = ['#0088FE', '#00C49F', '#FFBB28', '#FF8042', '#8884D8', '#82CA9D'];
  const HIGHLIGHT_COLOR = '#FF0000'; // Red color for the current portfolio

  return (
    <div className="flex flex-col items-center">
      <div style={{ height: sizePx, width: sizePx }}>
        <ResponsiveContainer width="100%" height="100%">
          <PieChart>
            <Pie
              data={data}
              cx="50%"
              cy="50%"
              innerRadius={Math.max(20, Math.floor(sizePx * 0.22))}
              outerRadius={Math.max(40, Math.floor(sizePx * 0.42))}
              paddingAngle={2}
              dataKey="value"
            >
              {data.map((entry, index) => (
                <Cell 
                  key={`cell-${index}`} 
                  fill={entry.isCurrentPortfolio ? HIGHLIGHT_COLOR : COLORS[index % COLORS.length]} 
                  stroke={entry.isCurrentPortfolio ? "#000" : undefined}
                  strokeWidth={entry.isCurrentPortfolio ? 2 : 0}
                />
              ))}
            </Pie>
            <Tooltip
              formatter={(value: number) => `${value.toFixed(2)}%`}
              labelFormatter={(label) => label}
            />
          </PieChart>
        </ResponsiveContainer>
      </div>
      {currentPortfolioId && (
        <div className="mt-2 text-xs text-center">
          <div className="flex items-center justify-center">
            <div className="w-3 h-3 bg-red-500 mr-1"></div>
            <span>Your Portfolio</span>
          </div>
        </div>
      )}
      {showRank && rank !== undefined && totalInvestors !== undefined && totalInvestors > 0 && (
        <div className="mt-1 text-xs text-gray-600">
          Investor Rank: {rank} / {totalInvestors}
        </div>
      )}
    </div>
  );
};

export function EntryTeamsPage() {
  const { entryId, calcuttaId } = useParams<{ entryId: string; calcuttaId: string }>();
  const [teams, setTeams] = useState<CalcuttaEntryTeam[]>([]);
  const [schools, setSchools] = useState<School[]>([]);
  const [portfolios, setPortfolios] = useState<CalcuttaPortfolio[]>([]);
  const [portfolioTeams, setPortfolioTeams] = useState<CalcuttaPortfolioTeam[]>([]);
  const [allCalcuttaPortfolioTeams, setAllCalcuttaPortfolioTeams] = useState<CalcuttaPortfolioTeam[]>([]);
  const [allCalcuttaPortfolios, setAllCalcuttaPortfolios] = useState<(CalcuttaPortfolio & { entryName?: string })[]>([]);
  const [totalBidByTeamId, setTotalBidByTeamId] = useState<Record<string, number>>({});
  const [entryName, setEntryName] = useState<string>('');
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);
  const [activeTab, setActiveTab] = useState<'teams' | 'statistics'>('teams');
  const [sortBy, setSortBy] = useState<'points' | 'ownership' | 'bid'>('points');

  useEffect(() => {
    const fetchData = async () => {
      if (!entryId || !calcuttaId) {
        setError('Missing required parameters');
        setLoading(false);
        return;
      }
      
      try {
        const [teamsData, schoolsData, portfoliosData, allEntriesData] = await Promise.all([
          calcuttaService.getEntryTeams(entryId, calcuttaId),
          calcuttaService.getSchools(),
          calcuttaService.getPortfoliosByEntry(entryId),
          calcuttaService.getCalcuttaEntries(calcuttaId)
        ]);

        const currentEntry = allEntriesData.find(e => e.id === entryId);
        setEntryName(currentEntry?.name || '');

        // Create a map of entryId to entryName
        const entryNameMap = new Map(allEntriesData.map(entry => [entry.id, entry.name]));

        // Create a map of schools by ID for quick lookup
        const schoolMap = new Map(schoolsData.map(school => [school.id, school]));

        // Associate schools with teams
        const teamsWithSchools = teamsData.map(team => ({
          ...team,
          team: team.team ? {
            ...team.team,
            school: schoolMap.get(team.team.schoolId)
          } : undefined
        }));

        setTeams(teamsWithSchools);
        setSchools(schoolsData);
        setPortfolios(portfoliosData);

        // Fetch portfolio teams for the current entry's portfolios
        if (portfoliosData.length > 0) {
          const portfolioTeamsPromises = portfoliosData.map(portfolio => 
            calcuttaService.getPortfolioTeams(portfolio.id)
          );
          
          const portfolioTeamsResults = await Promise.all(portfolioTeamsPromises);
          const allPortfolioTeams = portfolioTeamsResults.flat();
          
          // Associate schools with portfolio teams
          const portfolioTeamsWithSchools = allPortfolioTeams.map(team => ({
            ...team,
            team: team.team ? {
              ...team.team,
              school: schoolMap.get(team.team.schoolId)
            } : undefined
          }));
          
          setPortfolioTeams(portfolioTeamsWithSchools);
        }

        // Fetch all portfolios in the Calcutta and associate entry names
        const allPortfoliosPromises = allEntriesData.map(entry => 
          calcuttaService.getPortfoliosByEntry(entry.id)
        );
        const allPortfoliosResults = await Promise.all(allPortfoliosPromises);
        const allPortfoliosFlat = allPortfoliosResults.flat();
        // Add entryName to each portfolio object
        const allPortfoliosWithEntryNames = allPortfoliosFlat.map(portfolio => ({
          ...portfolio,
          entryName: entryNameMap.get(portfolio.entryId)
        })); 
        setAllCalcuttaPortfolios(allPortfoliosWithEntryNames);

        // Fetch portfolio teams for all portfolios
        const allPortfolioTeamsPromises = allPortfoliosFlat.map(portfolio =>
          calcuttaService.getPortfolioTeams(portfolio.id)
        );
        const allPortfolioTeamsResults = await Promise.all(allPortfolioTeamsPromises);
        const allCalcuttaPortfolioTeams = allPortfolioTeamsResults.flat();

        // Associate schools with all portfolio teams
        const allCalcuttaPortfolioTeamsWithSchools = allCalcuttaPortfolioTeams.map(team => ({
          ...team,
          team: team.team ? {
            ...team.team,
            school: schoolMap.get(team.team.schoolId)
          } : undefined
        }));

        setAllCalcuttaPortfolioTeams(allCalcuttaPortfolioTeamsWithSchools);

        const allEntryTeamsPromises = allEntriesData.map((entry) =>
          calcuttaService.getEntryTeams(entry.id, calcuttaId)
        );
        const allEntryTeamsResults = await Promise.all(allEntryTeamsPromises);
        const allEntryTeamsFlat = allEntryTeamsResults.flat();
        const bidTotals: Record<string, number> = {};
        for (const t of allEntryTeamsFlat) {
          bidTotals[t.teamId] = (bidTotals[t.teamId] ?? 0) + (t.bid ?? 0);
        }
        setTotalBidByTeamId(bidTotals);
        
        setLoading(false);
      } catch (err) {
        setError('Failed to fetch data');
        setLoading(false);
      }
    };

    fetchData();
  }, [entryId, calcuttaId]);

  if (loading) {
    return <div>Loading...</div>;
  }

  if (error) {
    return <div className="error">{error}</div>;
  }

  // Helper function to find portfolio team data for a given team ID
  const getPortfolioTeamData = (teamId: string) => {
    // Find the portfolio team belonging to the current entry's portfolio
    const currentPortfolioId = portfolios[0]?.id;
    if (!currentPortfolioId) return undefined;
    return allCalcuttaPortfolioTeams.find(pt => pt.teamId === teamId && pt.portfolioId === currentPortfolioId);
  };

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

  // Helper function to calculate wins (wins + byes - 1)
  const calculateWins = (team: CalcuttaEntryTeam) => {
    if (!team.team) return 0;
    const wins = team.team.wins || 0;
    const byes = team.team.byes || 0;
    return Math.max(0, wins + byes - 1);
  };

  return (
    <div className="container mx-auto px-4 py-8">
      <div className="mb-6">
        <Link to={`/calcuttas/${calcuttaId}`} className="text-blue-600 hover:text-blue-800">← Back to Entries</Link>
      </div>
      <h1 className="text-3xl font-bold mb-6">{entryName || 'Entry'}</h1>

      <div className="mb-8 flex gap-2 border-b border-gray-200">
        <button
          type="button"
          onClick={() => setActiveTab('teams')}
          className={`px-4 py-2 -mb-px border-b-2 font-medium transition-colors ${
            activeTab === 'teams'
              ? 'border-blue-600 text-blue-600'
              : 'border-transparent text-gray-600 hover:text-gray-900'
          }`}
        >
          Teams
        </button>
        <button
          type="button"
          onClick={() => setActiveTab('statistics')}
          className={`px-4 py-2 -mb-px border-b-2 font-medium transition-colors ${
            activeTab === 'statistics'
              ? 'border-blue-600 text-blue-600'
              : 'border-transparent text-gray-600 hover:text-gray-900'
          }`}
        >
          Statistics
        </button>
      </div>

      {activeTab === 'teams' && (
        <>
          <div className="mb-4 flex items-center justify-end">
            <label className="text-sm text-gray-600">
              Sort by
              <select
                className="ml-2 rounded-md border border-gray-300 bg-white px-3 py-2 text-sm text-gray-900 shadow-sm"
                value={sortBy}
                onChange={(e) => setSortBy(e.target.value as 'points' | 'ownership' | 'bid')}
              >
                <option value="points">Points</option>
                <option value="ownership">Ownership</option>
                <option value="bid">Bid Amount</option>
              </select>
            </label>
          </div>
          <div className="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-3 xl:grid-cols-4 gap-6">
            {sortedTeams.map((team) => {
              const portfolioTeam = getPortfolioTeamData(team.teamId);
              const wins = calculateWins(team);
              const investorRanking = getInvestorRanking(team.teamId);
              const teamPortfolioTeams = allCalcuttaPortfolioTeams.filter(pt => pt.teamId === team.teamId);
              const currentPortfolioId = portfolios[0]?.id;
              const totalBids = totalBidByTeamId[team.teamId] ?? team.bid;
              const ownershipPct = portfolioTeam ? portfolioTeam.ownershipPercentage * 100 : undefined;
              const pointsEarned = portfolioTeam ? portfolioTeam.actualPoints : undefined;
              
              return (
                <div
                  key={team.id}
                  className="bg-white rounded-lg shadow p-4 flex flex-col"
                >
                  <div className="flex items-start justify-between gap-3">
                    <div className="min-w-0">
                      <h2 className="text-lg font-semibold leading-snug truncate">
                        {team.team?.school?.name || 'Unknown School'}
                      </h2>
                      <div className="mt-1 text-sm text-gray-600">
                        Investor Rank: {investorRanking.rank} / {investorRanking.total}
                      </div>
                    </div>
                    <div className="text-right">
                      {ownershipPct !== undefined && (
                        <div className="text-sm text-gray-600">
                          Ownership
                          <div className="text-base font-semibold text-gray-900">{ownershipPct.toFixed(2)}%</div>
                        </div>
                      )}
                    </div>
                  </div>

                  <div className="mt-4 flex justify-center">
                    <OwnershipPieChart
                      portfolioTeams={teamPortfolioTeams}
                      portfolios={allCalcuttaPortfolios}
                      currentPortfolioId={currentPortfolioId}
                      sizePx={220}
                      showRank={false}
                    />
                  </div>

                  <div className="mt-4 grid grid-cols-2 gap-x-4 gap-y-3 text-sm">
                    <div>
                      <div className="text-gray-500">Bid</div>
                      <div className="font-medium text-gray-900">${team.bid}</div>
                    </div>
                    <div>
                      <div className="text-gray-500">Total Bids</div>
                      <div className="font-medium text-gray-900">${team.bid} / ${totalBids}</div>
                    </div>
                    <div>
                      <div className="text-gray-500">Wins</div>
                      <div className="font-medium text-gray-900">{wins}</div>
                    </div>
                    <div>
                      <div className="text-gray-500">Points Earned</div>
                      <div className="font-medium text-gray-900">{pointsEarned !== undefined ? pointsEarned.toFixed(2) : '—'}</div>
                    </div>
                  </div>
                </div>
              );
            })}
          </div>
        </>
      )}

      {activeTab === 'statistics' && (
        <>
          {portfolios.length > 0 && <PortfolioScores portfolio={portfolios[0]} teams={portfolioTeams} />}
        </>
      )}
    </div>
  );
}
 