import { useEffect, useState, useMemo } from 'react';
import { Link, useParams } from 'react-router-dom';
import { CalcuttaEntryTeam, CalcuttaPortfolio, CalcuttaPortfolioTeam, School, TournamentTeam } from '../types/calcutta';
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
              isAnimationActive={false}
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

const InvestmentsTab: React.FC<{
  entryId: string;
  teams: CalcuttaEntryTeam[];
  tournamentTeams: TournamentTeam[];
  allEntryTeams: CalcuttaEntryTeam[];
  schools: School[];
  investmentsSortBy: 'total' | 'seed' | 'region' | 'team';
  setInvestmentsSortBy: (value: 'total' | 'seed' | 'region' | 'team') => void;
  showAllTeams: boolean;
  setShowAllTeams: (value: boolean) => void;
  investmentHover: { entryName: string; amount: number; x: number; y: number } | null;
  setInvestmentHover: (value: { entryName: string; amount: number; x: number; y: number } | null) => void;
}> = ({
  entryId,
  teams,
  tournamentTeams,
  allEntryTeams,
  schools,
  investmentsSortBy,
  setInvestmentsSortBy,
  showAllTeams,
  setShowAllTeams,
  investmentHover,
  setInvestmentHover,
}) => {
  const schoolNameById = useMemo(() => new Map(schools.map((school) => [school.id, school.name])), [schools]);

  const mixHex = (hexA: string, hexB: string, amountB: number) => {
    const clamp = (n: number) => Math.max(0, Math.min(255, Math.round(n)));
    const norm = (hex: string) => hex.replace('#', '');
    const a = norm(hexA);
    const b = norm(hexB);
    const ar = parseInt(a.substring(0, 2), 16);
    const ag = parseInt(a.substring(2, 4), 16);
    const ab = parseInt(a.substring(4, 6), 16);
    const br = parseInt(b.substring(0, 2), 16);
    const bg = parseInt(b.substring(2, 4), 16);
    const bb = parseInt(b.substring(4, 6), 16);
    const r = clamp(ar * (1 - amountB) + br * amountB);
    const g = clamp(ag * (1 - amountB) + bg * amountB);
    const bl = clamp(ab * (1 - amountB) + bb * amountB);
    return `#${r.toString(16).padStart(2, '0')}${g.toString(16).padStart(2, '0')}${bl.toString(16).padStart(2, '0')}`;
  };

  const desaturate = (hex: string) => mixHex(hex, '#FFFFFF', 0.55);

  const investmentRows = useMemo(() => {
    const byTeam = new Map<
      string,
      {
        teamId: string;
        seed: number | undefined;
        region: string;
        teamName: string;
        totalInvestment: number;
        entryInvestment: number;
        otherInvestments: { entryId: string; entryName: string; amount: number }[];
      }
    >();

    for (const tt of tournamentTeams) {
      byTeam.set(tt.id, {
        teamId: tt.id,
        seed: tt.seed,
        region: tt.region,
        teamName: schoolNameById.get(tt.schoolId) || 'Unknown School',
        totalInvestment: 0,
        entryInvestment: 0,
        otherInvestments: [],
      });
    }

    for (const entryTeam of allEntryTeams) {
      const amount = entryTeam.bid || 0;
      if (amount <= 0) continue;

      const row = byTeam.get(entryTeam.teamId);
      if (!row) continue;

      row.totalInvestment += amount;
      if (entryTeam.entryId === entryId) {
        row.entryInvestment += amount;
      } else {
        row.otherInvestments.push({
          entryId: entryTeam.entryId,
          entryName: 'Other',
          amount,
        });
      }
    }

    let rows = Array.from(byTeam.values());

    if (!showAllTeams) {
      rows = rows.filter((row) => row.entryInvestment > 0);
    }

    const maxTotal = rows.reduce((max, row) => Math.max(max, row.totalInvestment), 0);

    return { rows, maxTotal };
  }, [allEntryTeams, entryId, schoolNameById, showAllTeams, tournamentTeams]);

  const sortedRows = useMemo(() => {
    const rows = investmentRows.rows.slice();
    const seedA = (seed: number | undefined) => seed ?? 999;
    const regionA = (region: string) => region || '';
    const teamA = (name: string) => name || '';

    rows.sort((a, b) => {
      if (investmentsSortBy === 'total') {
        if (b.totalInvestment !== a.totalInvestment) return b.totalInvestment - a.totalInvestment;
        return seedA(a.seed) - seedA(b.seed);
      }

      if (investmentsSortBy === 'seed') {
        const seedDiff = seedA(a.seed) - seedA(b.seed);
        if (seedDiff !== 0) return seedDiff;
        const regionDiff = regionA(a.region).localeCompare(regionA(b.region));
        if (regionDiff !== 0) return regionDiff;
        return teamA(a.teamName).localeCompare(teamA(b.teamName));
      }

      if (investmentsSortBy === 'region') {
        const regionDiff = regionA(a.region).localeCompare(regionA(b.region));
        if (regionDiff !== 0) return regionDiff;
        const seedDiff = seedA(a.seed) - seedA(b.seed);
        if (seedDiff !== 0) return seedDiff;
        return teamA(a.teamName).localeCompare(teamA(b.teamName));
      }

      const nameDiff = teamA(a.teamName).localeCompare(teamA(b.teamName));
      if (nameDiff !== 0) return nameDiff;
      return seedA(a.seed) - seedA(b.seed);
    });

    return rows;
  }, [investmentRows.rows, investmentsSortBy]);

  const ENTRY_COLOR = '#2563EB';
  const OTHER_COLOR = desaturate('#94A3B8');

  return (
    <div className="bg-white rounded-lg shadow p-6">
      <div className="mb-4 flex items-center justify-between gap-4">
        <label className="flex items-center gap-2 text-sm text-gray-700">
          <input
            type="checkbox"
            checked={showAllTeams}
            onChange={(e) => setShowAllTeams(e.target.checked)}
            className="rounded border-gray-300"
          />
          Show All Teams
        </label>
        <label className="text-sm text-gray-600">
          Sort by
          <select
            className="ml-2 rounded-md border border-gray-300 bg-white px-3 py-2 text-sm text-gray-900 shadow-sm"
            value={investmentsSortBy}
            onChange={(e) => setInvestmentsSortBy(e.target.value as 'total' | 'seed' | 'region' | 'team')}
          >
            <option value="total">Total</option>
            <option value="seed">Seed</option>
            <option value="region">Region</option>
            <option value="team">Team</option>
          </select>
        </label>
      </div>

      <div className="mt-4 overflow-x-auto">
        <table className="min-w-full table-fixed border-separate border-spacing-y-2">
          <thead>
            <tr className="text-left text-xs uppercase tracking-wide text-gray-500">
              <th className="px-2 py-2 w-14">Seed</th>
              <th className="px-2 py-2 w-20">Region</th>
              <th className="px-2 py-2 w-44">Team</th>
              <th className="px-2 py-2"></th>
              <th className="px-2 py-2 w-24 text-right">Investment</th>
              <th className="px-2 py-2 w-32 text-right">Total Investment</th>
            </tr>
          </thead>
          <tbody>
            {sortedRows.map((row) => {
              const barWidthPct = investmentRows.maxTotal > 0 ? (row.totalInvestment / investmentRows.maxTotal) * 100 : 0;
              const entryWidthPct = row.totalInvestment > 0 ? (row.entryInvestment / row.totalInvestment) * 100 : 0;
              const otherTotal = row.otherInvestments.reduce((sum, inv) => sum + inv.amount, 0);
              const otherWidthPct = row.totalInvestment > 0 ? (otherTotal / row.totalInvestment) * 100 : 0;

              return (
                <tr key={row.teamId} className="bg-gray-50">
                  <td className="px-2 py-3 font-medium text-gray-900 rounded-l-md whitespace-nowrap">{row.seed ?? '—'}</td>
                  <td className="px-2 py-3 text-gray-700 whitespace-nowrap">{row.region}</td>
                  <td className="px-2 py-3 text-gray-900 font-medium whitespace-nowrap truncate">{row.teamName}</td>
                  <td className="px-2 py-3">
                    <div className="h-6 w-full rounded bg-gray-200 overflow-hidden">
                      <div className="h-full flex" style={{ width: `${barWidthPct.toFixed(2)}%` }}>
                        {row.entryInvestment > 0 && (
                          <div
                            className="h-full"
                            style={{
                              width: `${entryWidthPct.toFixed(2)}%`,
                              backgroundColor: ENTRY_COLOR,
                              boxSizing: 'border-box',
                              border:
                                investmentHover?.entryName === 'Your Entry' && investmentHover?.amount === row.entryInvestment
                                  ? '2px solid #111827'
                                  : '2px solid transparent',
                            }}
                            onMouseEnter={(e) => {
                              setInvestmentHover({
                                entryName: 'Your Entry',
                                amount: row.entryInvestment,
                                x: e.clientX,
                                y: e.clientY,
                              });
                            }}
                            onMouseMove={(e) => {
                              if (investmentHover) {
                                setInvestmentHover({
                                  ...investmentHover,
                                  x: e.clientX,
                                  y: e.clientY,
                                });
                              }
                            }}
                            onMouseLeave={() => setInvestmentHover(null)}
                          />
                        )}
                        {otherTotal > 0 && (
                          <div
                            className="h-full"
                            style={{
                              width: `${otherWidthPct.toFixed(2)}%`,
                              backgroundColor: OTHER_COLOR,
                              boxSizing: 'border-box',
                              border:
                                investmentHover?.entryName === 'Others' && investmentHover?.amount === otherTotal
                                  ? '2px solid #111827'
                                  : '2px solid transparent',
                            }}
                            onMouseEnter={(e) => {
                              setInvestmentHover({
                                entryName: 'Others',
                                amount: otherTotal,
                                x: e.clientX,
                                y: e.clientY,
                              });
                            }}
                            onMouseMove={(e) => {
                              if (investmentHover) {
                                setInvestmentHover({
                                  ...investmentHover,
                                  x: e.clientX,
                                  y: e.clientY,
                                });
                              }
                            }}
                            onMouseLeave={() => setInvestmentHover(null)}
                          />
                        )}
                      </div>
                    </div>
                  </td>
                  <td className="px-2 py-3 text-right font-medium text-gray-900 whitespace-nowrap">
                    ${row.entryInvestment.toFixed(2)}
                  </td>
                  <td className="px-2 py-3 text-right font-medium text-gray-900 rounded-r-md whitespace-nowrap">
                    ${row.totalInvestment.toFixed(2)}
                  </td>
                </tr>
              );
            })}
          </tbody>
        </table>
      </div>

      {investmentHover && (
        <div
          className="fixed z-50 pointer-events-none rounded bg-gray-900 px-3 py-2 text-xs text-white shadow"
          style={{ left: investmentHover.x + 12, top: investmentHover.y + 12 }}
        >
          <div className="font-medium">{investmentHover.entryName}</div>
          <div>${investmentHover.amount.toFixed(2)}</div>
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
  const [tournamentTeams, setTournamentTeams] = useState<TournamentTeam[]>([]);
  const [allEntryTeams, setAllEntryTeams] = useState<CalcuttaEntryTeam[]>([]);
  const [activeTab, setActiveTab] = useState<'investments' | 'ownerships' | 'returns' | 'statistics'>('ownerships');
  const [sortBy, setSortBy] = useState<'points' | 'ownership' | 'bid'>('points');
  const [investmentsSortBy, setInvestmentsSortBy] = useState<'total' | 'seed' | 'region' | 'team'>('total');
  const [showAllTeams, setShowAllTeams] = useState(false);
  const [investmentHover, setInvestmentHover] = useState<{ entryName: string; amount: number; x: number; y: number } | null>(null);
  const [ownershipShowAllTeams, setOwnershipShowAllTeams] = useState(false);
  const [ownershipLoading, setOwnershipLoading] = useState(false);
  const [ownershipTeamsData, setOwnershipTeamsData] = useState<CalcuttaEntryTeam[]>([]);
  const [returnsShowAllTeams, setReturnsShowAllTeams] = useState(false);
  const [returnsHover, setReturnsHover] = useState<{ type: 'entry' | 'others'; points: number; x: number; y: number } | null>(null);

  useEffect(() => {
    const fetchData = async () => {
      if (!entryId || !calcuttaId) {
        setError('Missing required parameters');
        setLoading(false);
        return;
      }
      
      try {
        const calcutta = await calcuttaService.getCalcutta(calcuttaId);
        const [teamsData, schoolsData, portfoliosData, allEntriesData, tournamentTeamsData] = await Promise.all([
          calcuttaService.getEntryTeams(entryId, calcuttaId),
          calcuttaService.getSchools(),
          calcuttaService.getPortfoliosByEntry(entryId),
          calcuttaService.getCalcuttaEntries(calcuttaId),
          calcuttaService.getTournamentTeams(calcutta.tournamentId)
        ]);

        setTournamentTeams(tournamentTeamsData);

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

        // Fetch all entry teams for all entries
        const allEntryTeamsPromises = allEntriesData.map(entry =>
          calcuttaService.getEntryTeams(entry.id, calcuttaId)
        );
        const allEntryTeamsResults = await Promise.all(allEntryTeamsPromises);
        const allEntryTeamsFlat = allEntryTeamsResults.flat();
        setAllEntryTeams(allEntryTeamsFlat);

        setLoading(false);
      } catch (err) {
        setError('Failed to fetch data');
        setLoading(false);
      }
    };

    fetchData();
  }, [entryId, calcuttaId]);

  // Helper function to find portfolio team data for a given team ID
  const getPortfolioTeamData = (teamId: string) => {
    // Find the portfolio team belonging to the current entry's portfolio
    const currentPortfolioId = portfolios[0]?.id;
    if (!currentPortfolioId) return undefined;
    return allCalcuttaPortfolioTeams.find(pt => pt.teamId === teamId && pt.portfolioId === currentPortfolioId);
  };

  useEffect(() => {
    if (activeTab !== 'ownerships') {
      setOwnershipLoading(false);
      return;
    }

    setOwnershipLoading(true);
    const handle = window.setTimeout(() => {
      let teamsToShow: CalcuttaEntryTeam[];
      
      if (ownershipShowAllTeams) {
        // Show all tournament teams, creating synthetic CalcuttaEntryTeam objects for teams not bid on
        const teamIdSet = new Set(teams.map(t => t.teamId));
        const schoolMap = new Map(schools.map(s => [s.id, s]));
        
        teamsToShow = tournamentTeams.map((tt) => {
          const existingTeam = teams.find(t => t.teamId === tt.id);
          if (existingTeam) {
            return existingTeam;
          }
          
          // Create synthetic team for teams not bid on
          return {
            id: `synthetic-${tt.id}`,
            entryId: entryId!,
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
        // Show only teams with ownership > 0
        teamsToShow = teams.filter((team) => {
          const portfolioTeam = getPortfolioTeamData(team.teamId);
          return portfolioTeam && portfolioTeam.ownershipPercentage > 0;
        });
      }

      // Sort the teams based on sortBy
      teamsToShow.sort((a, b) => {
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

        // sortBy === 'bid'
        if (bidB !== bidA) return bidB - bidA;
        if (pointsB !== pointsA) return pointsB - pointsA;
        return ownershipB - ownershipA;
      });

      setOwnershipTeamsData(teamsToShow);
      setOwnershipLoading(false);
    }, 0);

    return () => {
      window.clearTimeout(handle);
      setOwnershipLoading(false);
    };
  }, [activeTab, teams, ownershipShowAllTeams, sortBy, portfolios, allCalcuttaPortfolioTeams, tournamentTeams, schools, entryId]);

  if (loading) {
    return <div>Loading...</div>;
  }

  if (error) {
    return <div className="error">{error}</div>;
  }

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

      <div className="mb-8 flex gap-2 border-b border-gray-200">
        <button
          type="button"
          onClick={() => setActiveTab('investments')}
          className={`px-4 py-2 -mb-px border-b-2 font-medium transition-colors ${
            activeTab === 'investments'
              ? 'border-blue-600 text-blue-600'
              : 'border-transparent text-gray-600 hover:text-gray-900'
          }`}
        >
          Investments
        </button>
        <button
          type="button"
          onClick={() => setActiveTab('ownerships')}
          className={`px-4 py-2 -mb-px border-b-2 font-medium transition-colors ${
            activeTab === 'ownerships'
              ? 'border-blue-600 text-blue-600'
              : 'border-transparent text-gray-600 hover:text-gray-900'
          }`}
        >
          Ownerships
        </button>
        <button
          type="button"
          onClick={() => setActiveTab('returns')}
          className={`px-4 py-2 -mb-px border-b-2 font-medium transition-colors ${
            activeTab === 'returns'
              ? 'border-blue-600 text-blue-600'
              : 'border-transparent text-gray-600 hover:text-gray-900'
          }`}
        >
          Returns
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
          investmentHover={investmentHover}
          setInvestmentHover={setInvestmentHover}
        />
      )}

      {activeTab === 'ownerships' && (
        <>
          <div className="mb-4 flex items-center justify-between gap-4">
            <label className="flex items-center gap-2 text-sm text-gray-700">
              <input
                type="checkbox"
                checked={ownershipShowAllTeams}
                onChange={(e) => setOwnershipShowAllTeams(e.target.checked)}
                className="rounded border-gray-300"
              />
              Show All Teams
            </label>
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
          {ownershipLoading ? (
            <div className="bg-white rounded-lg shadow p-6 text-gray-600">Loading ownerships…</div>
          ) : (
            <div className="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-3 xl:grid-cols-4 gap-6">
              {ownershipTeamsData.map((team) => {
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
                          <div className="min-w-0 truncate text-gray-700 flex items-center gap-2">
                            <div className="w-4 shrink-0 text-gray-500">{idx + 1}</div>
                            <div className="truncate">{owner.name}</div>
                          </div>
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
          )}
        </>
      )}

      {activeTab === 'returns' && (
        <>
          <div className="mb-4 flex items-center justify-between gap-4">
            <label className="flex items-center gap-2 text-sm text-gray-700">
              <input
                type="checkbox"
                checked={returnsShowAllTeams}
                onChange={(e) => setReturnsShowAllTeams(e.target.checked)}
                className="rounded border-gray-300"
              />
              Show All Teams
            </label>
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

          <div className="bg-white rounded-lg shadow p-6">
            <div className="overflow-x-auto">
              <table className="min-w-full table-fixed border-separate border-spacing-y-2">
                <thead>
                  <tr className="text-left text-xs uppercase tracking-wide text-gray-500">
                    <th className="px-2 py-2 w-14">Seed</th>
                    <th className="px-2 py-2 w-20">Region</th>
                    <th className="px-2 py-2 w-44">Team</th>
                    <th className="px-2 py-2"></th>
                    <th className="px-2 py-2 w-28 text-right">Points</th>
                    <th className="px-2 py-2 w-32 text-right">Total Points</th>
                  </tr>
                </thead>
                <tbody>
                  {(() => {
                    // Calculate global max from ALL tournament teams (not just filtered list)
                    const globalMax = tournamentTeams.reduce((max, tt) => {
                      const teamPortfolioTeams = allCalcuttaPortfolioTeams.filter(pt => pt.teamId === tt.id);
                      const totalActualPoints = teamPortfolioTeams.reduce((sum, pt) => sum + (pt.actualPoints || 0), 0);
                      const totalExpectedPoints = teamPortfolioTeams.reduce((sum, pt) => sum + (pt.expectedPoints || 0), 0);
                      const eliminated = tt.eliminated === true;
                      const totalPossiblePoints = eliminated ? totalActualPoints : Math.max(totalExpectedPoints, totalActualPoints);
                      return Math.max(max, totalActualPoints, totalPossiblePoints);
                    }, 0);

                    // Filter teams based on toggle
                    let teamsToShow = returnsShowAllTeams
                      ? tournamentTeams.map((tt) => {
                          const existingTeam = teams.find(t => t.teamId === tt.id);
                          if (existingTeam) return existingTeam;
                          const schoolMap = new Map(schools.map(s => [s.id, s]));
                          return {
                            id: `synthetic-${tt.id}`,
                            entryId: entryId!,
                            teamId: tt.id,
                            bid: 0,
                            created: new Date().toISOString(),
                            updated: new Date().toISOString(),
                            team: { ...tt, school: schoolMap.get(tt.schoolId) },
                          } as CalcuttaEntryTeam;
                        })
                      : teams;

                    // Sort teams based on sortBy
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

                      // sortBy === 'bid'
                      if (bidB !== bidA) return bidB - bidA;
                      if (pointsB !== pointsA) return pointsB - pointsA;
                      return ownershipB - ownershipA;
                    });

                    // Get all portfolio teams for this entry to calculate user's share
                    const currentPortfolioId = portfolios[0]?.id;

                    return teamsToShow.map((team) => {
                      const portfolioTeam = getPortfolioTeamData(team.teamId);
                      const tournamentTeam = tournamentTeams.find(tt => tt.id === team.teamId);
                      const teamPortfolioTeams = allCalcuttaPortfolioTeams.filter(pt => pt.teamId === team.teamId);
                      
                      // Calculate TOTAL points across all entries for this team
                      const totalActualPoints = teamPortfolioTeams.reduce((sum, pt) => sum + (pt.actualPoints || 0), 0);
                      const totalExpectedPoints = teamPortfolioTeams.reduce((sum, pt) => sum + (pt.expectedPoints || 0), 0);
                      const eliminated = team.team?.eliminated === true;
                      const totalPossiblePoints = eliminated ? totalActualPoints : Math.max(totalExpectedPoints, totalActualPoints);

                      // Calculate user's ownership percentage and their share of points
                      const userOwnership = portfolioTeam?.ownershipPercentage ?? 0;
                      const userActualPoints = totalActualPoints * userOwnership;
                      const othersActualPoints = totalActualPoints * (1 - userOwnership);
                      const userPossiblePoints = totalPossiblePoints * userOwnership;
                      const othersPossiblePoints = totalPossiblePoints * (1 - userOwnership);

                      // Calculate widths as percentage of global max
                      const userActualWidthPct = globalMax > 0 ? (userActualPoints / globalMax) * 100 : 0;
                      const othersActualWidthPct = globalMax > 0 ? (othersActualPoints / globalMax) * 100 : 0;
                      const userPossibleWidthPct = globalMax > 0 ? (userPossiblePoints / globalMax) * 100 : 0;
                      const othersPossibleWidthPct = globalMax > 0 ? (othersPossiblePoints / globalMax) * 100 : 0;

                      return (
                        <tr key={team.id} className="bg-gray-50">
                          <td className="px-2 py-3 font-medium text-gray-900 rounded-l-md whitespace-nowrap">{team.team?.seed ?? '—'}</td>
                          <td className="px-2 py-3 text-gray-700 whitespace-nowrap">{tournamentTeam?.region || '—'}</td>
                          <td className="px-2 py-3 text-gray-900 font-medium whitespace-nowrap truncate">{team.team?.school?.name || 'Unknown School'}</td>
                          <td className="px-2 py-3">
                            <div className="h-6 w-full rounded overflow-hidden" style={{ backgroundColor: eliminated ? 'transparent' : '#F3F4F6' }}>
                              <div className="h-full flex">
                                {userActualPoints > 0 && (
                                  <div
                                    className="h-full"
                                    style={{ 
                                      width: `${userActualWidthPct.toFixed(2)}%`,
                                      backgroundColor: '#4F46E5'
                                    }}
                                    onMouseEnter={(e) => {
                                      setReturnsHover({
                                        type: 'entry',
                                        points: userActualPoints,
                                        x: e.clientX,
                                        y: e.clientY,
                                      });
                                    }}
                                    onMouseMove={(e) => {
                                      setReturnsHover((prev) =>
                                        prev
                                          ? {
                                              ...prev,
                                              x: e.clientX,
                                              y: e.clientY,
                                            }
                                          : prev
                                      );
                                    }}
                                    onMouseLeave={() => setReturnsHover(null)}
                                  />
                                )}
                                {othersActualPoints > 0 && (
                                  <div
                                    className="h-full"
                                    style={{ 
                                      width: `${othersActualWidthPct.toFixed(2)}%`,
                                      backgroundColor: '#9CA3AF'
                                    }}
                                    onMouseEnter={(e) => {
                                      setReturnsHover({
                                        type: 'others',
                                        points: othersActualPoints,
                                        x: e.clientX,
                                        y: e.clientY,
                                      });
                                    }}
                                    onMouseMove={(e) => {
                                      setReturnsHover((prev) =>
                                        prev
                                          ? {
                                              ...prev,
                                              x: e.clientX,
                                              y: e.clientY,
                                            }
                                          : prev
                                      );
                                    }}
                                    onMouseLeave={() => setReturnsHover(null)}
                                  />
                                )}
                              </div>
                            </div>
                          </td>
                          <td className="px-2 py-3 text-right font-medium text-gray-900 whitespace-nowrap">
                            {userActualPoints.toFixed(2)}
                          </td>
                          <td className="px-2 py-3 text-right font-medium text-gray-900 rounded-r-md whitespace-nowrap">
                            {totalActualPoints.toFixed(2)}
                          </td>
                        </tr>
                      );
                    });
                  })()}
                </tbody>
              </table>
            </div>
          </div>

          {returnsHover && (
            <div
              className="fixed z-50 pointer-events-none rounded bg-gray-900 px-3 py-2 text-xs text-white shadow"
              style={{ left: returnsHover.x + 12, top: returnsHover.y + 12 }}
            >
              <div className="font-medium">{returnsHover.type === 'entry' ? 'Entry' : 'Others'}</div>
              <div>{returnsHover.points.toFixed(2)} points</div>
            </div>
          )}
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
 