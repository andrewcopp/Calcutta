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
  const [activeIndex, setActiveIndex] = useState<number | null>(null);

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

  const CURRENT_FILL = '#1F2937';
  const OTHER_FILL = '#CBD5E1';
  const DIVIDER_STROKE = '#FFFFFF';
  const HOVER_STROKE = '#111827';

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
              onMouseEnter={(_, index) => setActiveIndex(index)}
              onMouseLeave={() => setActiveIndex(null)}
            >
              {data.map((entry, index) => (
                (() => {
                  const isActive = activeIndex === index;
                  return (
                <Cell 
                  key={`cell-${index}`} 
                  fill={entry.isCurrentPortfolio ? CURRENT_FILL : OTHER_FILL}
                  stroke={isActive ? HOVER_STROKE : DIVIDER_STROKE}
                  strokeWidth={isActive ? 2 : 2}
                />
                  );
                })()
              ))}
            </Pie>
            <Tooltip
              formatter={(value: number) => `${value.toFixed(2)}%`}
              labelFormatter={(label) => label}
            />
          </PieChart>
        </ResponsiveContainer>
      </div>
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
  const [entryName, setEntryName] = useState<string>('');
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);
  const [activeTab, setActiveTab] = useState<'bids' | 'ownership' | 'points' | 'statistics'>('ownership');
  const [sortBy, setSortBy] = useState<'points' | 'ownership' | 'bid'>('points');
  const [bidsSortBy, setBidsSortBy] = useState<'bid' | 'seed' | 'alpha'>('bid');

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

  const bidsTotal = teams.reduce((sum, team) => sum + (team.bid ?? 0), 0);

  const compareBidsTeams = (a: CalcuttaEntryTeam, b: CalcuttaEntryTeam) => {
    const nameA = a.team?.school?.name || '';
    const nameB = b.team?.school?.name || '';
    const seedA = a.team?.seed;
    const seedB = b.team?.seed;
    const bidA = a.bid ?? 0;
    const bidB = b.bid ?? 0;

    const compareSeedAsc = () => {
      if (seedA === undefined && seedB === undefined) return 0;
      if (seedA === undefined) return 1;
      if (seedB === undefined) return -1;
      return seedA - seedB;
    };

    if (bidsSortBy === 'bid') {
      const byBid = compareDesc(bidA, bidB);
      if (byBid !== 0) return byBid;
      const bySeed = compareSeedAsc();
      if (bySeed !== 0) return bySeed;
      return nameA.localeCompare(nameB);
    }

    if (bidsSortBy === 'seed') {
      const bySeed = compareSeedAsc();
      if (bySeed !== 0) return bySeed;
      const byBid = compareDesc(bidA, bidB);
      if (byBid !== 0) return byBid;
      return nameA.localeCompare(nameB);
    }

    const byName = nameA.localeCompare(nameB);
    if (byName !== 0) return byName;
    return compareDesc(bidA, bidB);
  };

  const bidsList = [...teams].sort(compareBidsTeams);

  return (
    <div className="container mx-auto px-4 py-8">
      <div className="mb-6">
        <Link to={`/calcuttas/${calcuttaId}`} className="text-blue-600 hover:text-blue-800">← Back to Entries</Link>
      </div>
      <h1 className="text-3xl font-bold mb-6">{entryName || 'Entry'}</h1>

      <div className="mb-8 flex gap-2 border-b border-gray-200">
        <button
          type="button"
          onClick={() => setActiveTab('bids')}
          className={`px-4 py-2 -mb-px border-b-2 font-medium transition-colors ${
            activeTab === 'bids'
              ? 'border-blue-600 text-blue-600'
              : 'border-transparent text-gray-600 hover:text-gray-900'
          }`}
        >
          Bids
        </button>
        <button
          type="button"
          onClick={() => setActiveTab('ownership')}
          className={`px-4 py-2 -mb-px border-b-2 font-medium transition-colors ${
            activeTab === 'ownership'
              ? 'border-blue-600 text-blue-600'
              : 'border-transparent text-gray-600 hover:text-gray-900'
          }`}
        >
          Ownership
        </button>
        <button
          type="button"
          onClick={() => setActiveTab('points')}
          className={`px-4 py-2 -mb-px border-b-2 font-medium transition-colors ${
            activeTab === 'points'
              ? 'border-blue-600 text-blue-600'
              : 'border-transparent text-gray-600 hover:text-gray-900'
          }`}
        >
          Points
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

      {activeTab === 'bids' && (
        <div className="bg-white rounded-lg shadow p-6">
          <div className="flex items-baseline justify-between gap-4">
            <h2 className="text-xl font-semibold">Bids</h2>
            <div className="flex items-center gap-4">
              <label className="text-sm text-gray-600">
                Sort by
                <select
                  className="ml-2 rounded-md border border-gray-300 bg-white px-3 py-2 text-sm text-gray-900 shadow-sm"
                  value={bidsSortBy}
                  onChange={(e) => setBidsSortBy(e.target.value as 'bid' | 'seed' | 'alpha')}
                >
                  <option value="bid">Bid</option>
                  <option value="seed">Seed</option>
                  <option value="alpha">Alphabetical</option>
                </select>
              </label>
              <div className="text-sm text-gray-600">
                Total Spent: <span className="font-medium text-gray-900">${bidsTotal}</span>
              </div>
            </div>
          </div>

          <div className="mt-4 overflow-x-auto">
            <table className="min-w-full border-separate border-spacing-y-2">
              <thead>
                <tr className="text-left text-xs uppercase tracking-wide text-gray-500">
                  <th className="px-3 py-2">Team</th>
                  <th className="px-3 py-2">Seed</th>
                  <th className="px-3 py-2 text-right">Bid</th>
                </tr>
              </thead>
              <tbody>
                {bidsList.map((team) => (
                  <tr key={team.id} className="bg-gray-50">
                    <td className="px-3 py-3 font-medium text-gray-900 rounded-l-md">
                      {team.team?.school?.name || 'Unknown School'}
                    </td>
                    <td className="px-3 py-3 text-gray-700">
                      {team.team?.seed ?? '—'}
                    </td>
                    <td className="px-3 py-3 text-right font-medium text-gray-900 rounded-r-md">
                      ${team.bid}
                    </td>
                  </tr>
                ))}
              </tbody>
            </table>
          </div>
        </div>
      )}

      {activeTab === 'ownership' && (
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
              const investorRanking = getInvestorRanking(team.teamId);
              const teamPortfolioTeams = allCalcuttaPortfolioTeams.filter(pt => pt.teamId === team.teamId);
              const currentPortfolioId = portfolios[0]?.id;
              const ownershipPct = portfolioTeam ? portfolioTeam.ownershipPercentage * 100 : undefined;

              const topOwners: { name: string; pct: number | null }[] = [...teamPortfolioTeams]
                .filter(pt => pt.ownershipPercentage > 0)
                .sort((a, b) => b.ownershipPercentage - a.ownershipPercentage)
                .slice(0, 3)
                .map((pt) => {
                  const portfolio = allCalcuttaPortfolios.find(p => p.id === pt.portfolioId);
                  const name = portfolio?.entryName || `Portfolio ${pt.portfolioId.slice(0, 4)}`;
                  return {
                    name,
                    pct: pt.ownershipPercentage * 100,
                  };
                });

              while (topOwners.length < 3) {
                topOwners.push({ name: '--', pct: null });
              }
              
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

                  <div className="mt-4">
                    <div className="text-sm font-medium text-gray-900">Top Shareholders</div>
                    <div className="mt-2 space-y-2">
                      {topOwners.map((owner, idx) => (
                        <div key={idx} className="flex items-center justify-between gap-3 text-sm">
                          <div className="min-w-0 truncate text-gray-700">{owner.name}</div>
                          <div className="font-medium text-gray-900">
                            {owner.pct === null ? '--' : `${owner.pct.toFixed(2)}%`}
                          </div>
                        </div>
                      ))}
                    </div>
                  </div>
                </div>
              );
            })}
          </div>
        </>
      )}

      {activeTab === 'points' && (
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

          <div className="bg-white rounded-lg shadow p-6 overflow-x-auto">
            <table className="min-w-full">
              <thead>
                <tr className="text-left text-xs uppercase tracking-wide text-gray-500 border-b">
                  <th className="py-3 pr-4">Team</th>
                  <th className="py-3 pr-4">Points</th>
                </tr>
              </thead>
              <tbody className="divide-y">
                {sortedTeams.map((team) => {
                  const portfolioTeam = getPortfolioTeamData(team.teamId);
                  const totalPoints = portfolioTeam?.actualPoints ?? 0;
                  const possiblePointsRaw = portfolioTeam?.expectedPoints ?? 0;
                  const eliminated = team.team?.eliminated === true;
                  const possiblePoints = eliminated ? totalPoints : Math.max(possiblePointsRaw, totalPoints);
                  const ratio = possiblePoints > 0 ? Math.min(1, totalPoints / possiblePoints) : 0;

                  return (
                    <tr key={team.id}>
                      <td className="py-4 pr-4 align-top">
                        <div className="font-medium text-gray-900">
                          {team.team?.school?.name || 'Unknown School'}
                        </div>
                        <div className="mt-1 text-xs text-gray-600">
                          Seed: {team.team?.seed ?? '—'}
                        </div>
                      </td>
                      <td className="py-4 pr-4">
                        <div className="space-y-2">
                          <div>
                            <div className="flex items-center justify-between text-xs text-gray-600">
                              <span>Total Points</span>
                              <span className="font-medium text-gray-900">{totalPoints.toFixed(2)}</span>
                            </div>
                            <div className="mt-1 h-3 w-full rounded bg-indigo-200 relative overflow-hidden">
                              <div
                                className="absolute left-0 top-0 h-full bg-indigo-600"
                                style={{ width: `${(ratio * 100).toFixed(2)}%` }}
                              />
                            </div>
                          </div>

                          <div>
                            <div className="flex items-center justify-between text-xs text-gray-600">
                              <span>Possible Points</span>
                              <span className="font-medium text-gray-900">{possiblePoints.toFixed(2)}</span>
                            </div>
                            <div className="mt-1 h-3 w-full rounded bg-indigo-200" />
                          </div>
                        </div>
                      </td>
                    </tr>
                  );
                })}
              </tbody>
            </table>
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
 