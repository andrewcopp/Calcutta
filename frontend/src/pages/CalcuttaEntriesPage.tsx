import { useEffect, useMemo, useState } from 'react';
import { Link, useParams } from 'react-router-dom';
import { CalcuttaEntry, CalcuttaPortfolio, CalcuttaPortfolioTeam, School, CalcuttaEntryTeam, TournamentTeam } from '../types/calcutta';
import { calcuttaService } from '../services/calcuttaService';
import { BarChart, Bar, XAxis, YAxis, CartesianGrid, Tooltip, ResponsiveContainer, PieChart, Pie, Cell } from 'recharts';

interface TeamInvestment {
  teamId: string;
  teamName: string;
  totalBid: number;
}

interface SeedInvestment {
  seed: number;
  totalInvestment: number;
  teamCount: number;
}

export function CalcuttaEntriesPage() {
  const { calcuttaId } = useParams<{ calcuttaId: string }>();
  const [entries, setEntries] = useState<CalcuttaEntry[]>([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);
  const [totalEntries, setTotalEntries] = useState<number>(0);
  const [seedInvestments, setSeedInvestments] = useState<SeedInvestment[]>([]);
  const [schools, setSchools] = useState<School[]>([]);
  const [calcuttaName, setCalcuttaName] = useState<string>('');
  const [allEntryTeams, setAllEntryTeams] = useState<CalcuttaEntryTeam[]>([]);
  const [seedInvestmentData, setSeedInvestmentData] = useState<{ seed: number; totalInvestment: number }[]>([]);
  const [activeTab, setActiveTab] = useState<'leaderboard' | 'statistics' | 'investment' | 'ownership'>('leaderboard');
  const [tournamentTeams, setTournamentTeams] = useState<TournamentTeam[]>([]);
  const [allCalcuttaPortfolios, setAllCalcuttaPortfolios] = useState<(CalcuttaPortfolio & { entryName?: string })[]>([]);
  const [allCalcuttaPortfolioTeams, setAllCalcuttaPortfolioTeams] = useState<CalcuttaPortfolioTeam[]>([]);
  const [investmentHover, setInvestmentHover] = useState<{ entryName: string; amount: number; x: number; y: number } | null>(null);

  useEffect(() => {
    const fetchData = async () => {
      if (!calcuttaId) return;
      
      try {
        // Fetch calcutta details to get the name
        const calcutta = await calcuttaService.getCalcutta(calcuttaId);
        setCalcuttaName(calcutta.name);

        const [entriesData, schoolsData, tournamentTeamsData] = await Promise.all([
          calcuttaService.getCalcuttaEntries(calcuttaId),
          calcuttaService.getSchools(),
          calcuttaService.getTournamentTeams(calcutta.tournamentId),
        ]);

        setSchools(schoolsData);
        setTournamentTeams(tournamentTeamsData);

        const schoolMap = new Map(schoolsData.map((school) => [school.id, school]));

        const portfoliosByEntry = await Promise.all(
          entriesData.map(async (entry) => {
            const portfolios = await calcuttaService.getPortfoliosByEntry(entry.id);
            return { entry, portfolios };
          })
        );

        const portfoliosFlat: (CalcuttaPortfolio & { entryName?: string })[] = portfoliosByEntry.flatMap(({ entry, portfolios }) =>
          portfolios.map((portfolio) => ({
            ...portfolio,
            entryName: entry.name,
          }))
        );

        setAllCalcuttaPortfolios(portfoliosFlat);

        const portfolioTeamsByPortfolio = await Promise.all(
          portfoliosFlat.map(async (portfolio) => {
            const teams = await calcuttaService.getPortfolioTeams(portfolio.id);
            return { portfolio, teams };
          })
        );

        const entryPointsMap = new Map<string, number>();
        const allPortfolioTeamsFlat: CalcuttaPortfolioTeam[] = portfolioTeamsByPortfolio.flatMap(({ portfolio, teams }) => {
          const sum = teams.reduce((acc, team) => acc + team.actualPoints, 0);
          entryPointsMap.set(portfolio.entryId, (entryPointsMap.get(portfolio.entryId) || 0) + sum);

          return teams.map((team) => {
            const school = team.team?.schoolId ? schoolMap.get(team.team.schoolId) : undefined;
            return {
              ...team,
              team: team.team
                ? {
                    ...team.team,
                    school: school ? { id: school.id, name: school.name } : team.team.school,
                  }
                : team.team,
            };
          });
        });

        setAllCalcuttaPortfolioTeams(allPortfolioTeamsFlat);

        const entriesWithPoints = entriesData.map((entry) => ({
          ...entry,
          totalPoints: entryPointsMap.get(entry.id) || 0,
        }));
        
        // Sort entries by total points in descending order
        const sortedEntries = entriesWithPoints.sort((a, b) => (b.totalPoints || 0) - (a.totalPoints || 0));
        setEntries(sortedEntries);
        
        // Fetch all entry teams to calculate investments
        const entryTeamsByEntry = await Promise.all(
          entriesData.map((entry) => calcuttaService.getEntryTeams(entry.id, calcuttaId))
        );
        const allTeams: CalcuttaEntryTeam[] = entryTeamsByEntry.flat();
        setAllEntryTeams(allTeams);
        
        // Calculate investment by seed
        const seedMap = new Map<number, number>();
        for (const team of allTeams) {
          if (!team.team?.seed || !team.bid) continue;
          const seed = team.team.seed;
          const currentTotal = seedMap.get(seed) || 0;
          seedMap.set(seed, currentTotal + team.bid);
        }
        
        // Convert map to array and sort by seed
        const seedData = Array.from(seedMap.entries())
          .map(([seed, totalInvestment]) => ({ seed, totalInvestment }))
          .sort((a, b) => a.seed - b.seed);
        
        setSeedInvestmentData(seedData);
        setTotalEntries(entriesData.length);
        setLoading(false);
      } catch (err) {
        setError('Failed to fetch data');
        setLoading(false);
      }
    };

    fetchData();
  }, [calcuttaId]);

  const entryNameById = useMemo(() => new Map(entries.map((entry) => [entry.id, entry.name])), [entries]);

  const tournamentTeamById = useMemo(
    () => new Map(tournamentTeams.map((team) => [team.id, team])),
    [tournamentTeams]
  );

  const schoolNameById = useMemo(() => new Map(schools.map((school) => [school.id, school.name])), [schools]);

  const entryColorById = useMemo(() => {
    const palette = [
      '#2563EB',
      '#DC2626',
      '#059669',
      '#7C3AED',
      '#DB2777',
      '#D97706',
      '#0D9488',
      '#4B5563',
      '#9333EA',
      '#16A34A',
      '#EA580C',
      '#0EA5E9',
    ];
    const map = new Map<string, string>();
    entries.forEach((entry, idx) => {
      map.set(entry.id, palette[idx % palette.length]);
    });
    return map;
  }, [entries]);

  const totalSpent = useMemo(() => allEntryTeams.reduce((sum, team) => sum + (team.bid || 0), 0), [allEntryTeams]);

  const investmentRows = useMemo(() => {
    const byTeam = new Map<
      string,
      {
        teamId: string;
        seed: number | undefined;
        region: string;
        teamName: string;
        totalInvestment: number;
        segments: { entryId: string; entryName: string; amount: number }[];
      }
    >();

    for (const tournamentTeam of tournamentTeams) {
      const schoolName = schoolNameById.get(tournamentTeam.schoolId) || 'Unknown School';
      byTeam.set(tournamentTeam.id, {
        teamId: tournamentTeam.id,
        seed: tournamentTeam.seed,
        region: tournamentTeam.region,
        teamName: schoolName,
        totalInvestment: 0,
        segments: [],
      });
    }

    for (const entryTeam of allEntryTeams) {
      const amount = entryTeam.bid || 0;
      if (amount <= 0) continue;

      const teamId = entryTeam.teamId;
      const row = byTeam.get(teamId);
      if (!row) {
        const tournamentTeam = tournamentTeamById.get(teamId);
        const seed = tournamentTeam?.seed;
        const region = tournamentTeam?.region || 'Unknown';
        const schoolName = tournamentTeam ? schoolNameById.get(tournamentTeam.schoolId) : undefined;

        byTeam.set(teamId, {
          teamId,
          seed,
          region,
          teamName: schoolName || 'Unknown School',
          totalInvestment: amount,
          segments: [
            {
              entryId: entryTeam.entryId,
              entryName: entryNameById.get(entryTeam.entryId) || 'Unknown Entry',
              amount,
            },
          ],
        });
        continue;
      }

      row.totalInvestment += amount;
      row.segments.push({
        entryId: entryTeam.entryId,
        entryName: entryNameById.get(entryTeam.entryId) || 'Unknown Entry',
        amount,
      });
    }

    const rows = Array.from(byTeam.values())
      .map((row) => ({
        ...row,
        segments: row.segments
          .slice()
          .sort((a, b) => b.amount - a.amount)
          .filter((seg) => seg.amount > 0),
      }))
      .filter((row) => row.totalInvestment > 0)
      .sort((a, b) => {
        if (b.totalInvestment !== a.totalInvestment) return b.totalInvestment - a.totalInvestment;
        const seedA = a.seed ?? 999;
        const seedB = b.seed ?? 999;
        return seedA - seedB;
      });

    const maxTotal = rows.reduce((max, row) => Math.max(max, row.totalInvestment), 0);

    return { rows, maxTotal };
  }, [allEntryTeams, entryNameById, schoolNameById, tournamentTeams, tournamentTeamById]);

  const ownershipTeams = useMemo(() => {
    const byTeamSpend = new Map<string, number>();
    for (const entryTeam of allEntryTeams) {
      const amount = entryTeam.bid || 0;
      if (amount <= 0) continue;
      byTeamSpend.set(entryTeam.teamId, (byTeamSpend.get(entryTeam.teamId) || 0) + amount);
    }

    const cards = tournamentTeams
      .map((team) => {
        const teamId = team.id;
        const spend = byTeamSpend.get(teamId) || 0;
        const teamPortfolioTeams = allCalcuttaPortfolioTeams.filter((pt) => pt.teamId === teamId && pt.ownershipPercentage > 0);

        const slices = teamPortfolioTeams
          .slice()
          .sort((a, b) => b.ownershipPercentage - a.ownershipPercentage)
          .map((pt) => {
            const portfolio = allCalcuttaPortfolios.find((p) => p.id === pt.portfolioId);
            const entryId = portfolio?.entryId || '';
            const entryName = portfolio?.entryName || (entryId ? entryNameById.get(entryId) : undefined) || 'Unknown Entry';
            return {
              name: entryName,
              value: pt.ownershipPercentage * 100,
              entryId,
            };
          });

        const topOwners = slices.slice(0, 3);

        return {
          teamId,
          seed: team.seed,
          region: team.region,
          teamName: schoolNameById.get(team.schoolId) || 'Unknown School',
          totalSpend: spend,
          slices,
          topOwners,
        };
      })
      .filter((card) => card.totalSpend > 0)
      .sort((a, b) => {
        if (b.totalSpend !== a.totalSpend) return b.totalSpend - a.totalSpend;
        return a.seed - b.seed;
      });

    return cards;
  }, [allCalcuttaPortfolios, allCalcuttaPortfolioTeams, allEntryTeams, entryNameById, schoolNameById, tournamentTeams]);

  const CalcuttaOwnershipPieChart = ({ slices }: { slices: { name: string; value: number; entryId: string }[] }) => {
    const [activeIndex, setActiveIndex] = useState<number | null>(null);

    const DIVIDER_STROKE = '#FFFFFF';
    const HOVER_STROKE = '#111827';

    if (slices.length === 0) {
      return (
        <div className="flex h-[220px] w-[220px] items-center justify-center rounded-full bg-gray-100 text-sm text-gray-500">
          No ownership
        </div>
      );
    }

    return (
      <div style={{ height: 220, width: 220 }}>
        <ResponsiveContainer width="100%" height="100%">
          <PieChart>
            <Pie
              data={slices}
              cx="50%"
              cy="50%"
              innerRadius={48}
              outerRadius={92}
              paddingAngle={2}
              dataKey="value"
              onMouseEnter={(_, index) => setActiveIndex(index)}
              onMouseLeave={() => setActiveIndex(null)}
            >
              {slices.map((slice, index) => {
                const isActive = activeIndex === index;
                const color = entryColorById.get(slice.entryId) || '#94A3B8';
                return (
                  <Cell
                    key={`cell-${index}`}
                    fill={color}
                    stroke={isActive ? HOVER_STROKE : DIVIDER_STROKE}
                    strokeWidth={2}
                  />
                );
              })}
            </Pie>
            <Tooltip formatter={(value: number) => `${value.toFixed(2)}%`} labelFormatter={(label) => label} />
          </PieChart>
        </ResponsiveContainer>
      </div>
    );
  };

  if (loading) {
    return <div>Loading...</div>;
  }

  if (error) {
    return <div className="error">{error}</div>;
  }

  return (
    <div className="container mx-auto px-4 py-8">
      <div className="mb-6">
        <Link to="/calcuttas" className="text-blue-600 hover:text-blue-800">← Back to Calcuttas</Link>
      </div>
      <h1 className="text-3xl font-bold mb-6">{calcuttaName}</h1>

      <div className="mb-8 flex gap-2 border-b border-gray-200">
        <button
          type="button"
          onClick={() => setActiveTab('leaderboard')}
          className={`px-4 py-2 -mb-px border-b-2 font-medium transition-colors ${
            activeTab === 'leaderboard'
              ? 'border-blue-600 text-blue-600'
              : 'border-transparent text-gray-600 hover:text-gray-900'
          }`}
        >
          Leaderboard
        </button>
        <button
          type="button"
          onClick={() => setActiveTab('investment')}
          className={`px-4 py-2 -mb-px border-b-2 font-medium transition-colors ${
            activeTab === 'investment'
              ? 'border-blue-600 text-blue-600'
              : 'border-transparent text-gray-600 hover:text-gray-900'
          }`}
        >
          Investment
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
      

      {activeTab === 'statistics' && (
        <div className="grid grid-cols-1 md:grid-cols-2 gap-6 mb-8">
          <div className="bg-white rounded-lg shadow p-6">
            <h2 className="text-xl font-semibold mb-4">Tournament Statistics</h2>
            <p className="text-gray-600">Total Entries: {totalEntries}</p>
          </div>
          
          <div className="bg-white rounded-lg shadow p-6">
            <h2 className="text-xl font-semibold mb-4">Investment by Seed</h2>
            <div className="h-64">
              <ResponsiveContainer width="100%" height="100%">
                <BarChart data={seedInvestmentData}>
                  <CartesianGrid strokeDasharray="3 3" />
                  <XAxis dataKey="seed" label={{ value: 'Seed', position: 'insideBottom', offset: -5 }} />
                  <YAxis label={{ value: 'Total Investment ($)', angle: -90, position: 'insideLeft' }} />
                  <Tooltip formatter={(value: number) => [`$${value.toFixed(2)}`, 'Total Investment']} />
                  <Bar dataKey="totalInvestment" fill="#4F46E5" />
                </BarChart>
              </ResponsiveContainer>
            </div>
            <div className="mt-4 text-center">
              <Link 
                to={`/calcuttas/${calcuttaId}/teams`}
                className="text-blue-600 hover:text-blue-800 font-medium"
              >
                View All Teams →
              </Link>
            </div>
          </div>
        </div>
      )}

      {activeTab === 'investment' && (
        <div className="bg-white rounded-lg shadow p-6">
          <div className="flex items-baseline justify-between gap-4">
            <h2 className="text-xl font-semibold">Investment</h2>
            <div className="text-sm text-gray-600">
              Total Investment: <span className="font-medium text-gray-900">${totalSpent.toFixed(2)}</span>
            </div>
          </div>

          <div className="mt-4 overflow-x-auto">
            <table className="min-w-full border-separate border-spacing-y-2">
              <thead>
                <tr className="text-left text-xs uppercase tracking-wide text-gray-500">
                  <th className="px-3 py-2">Seed</th>
                  <th className="px-3 py-2">Team</th>
                  <th className="px-3 py-2">Region</th>
                  <th className="px-3 py-2">Investment</th>
                  <th className="px-3 py-2 text-right">Total</th>
                </tr>
              </thead>
              <tbody>
                {investmentRows.rows.map((row) => {
                  const barWidthPct = investmentRows.maxTotal > 0 ? (row.totalInvestment / investmentRows.maxTotal) * 100 : 0;
                  return (
                    <tr key={row.teamId} className="bg-gray-50">
                      <td className="px-3 py-3 font-medium text-gray-900 rounded-l-md whitespace-nowrap">{row.seed ?? '—'}</td>
                      <td className="px-3 py-3 text-gray-900 font-medium whitespace-nowrap">{row.teamName}</td>
                      <td className="px-3 py-3 text-gray-700 whitespace-nowrap">{row.region}</td>
                      <td className="px-3 py-3">
                        <div className="h-6 w-full rounded bg-gray-200 overflow-hidden">
                          <div className="h-full flex" style={{ width: `${barWidthPct.toFixed(2)}%` }}>
                            {row.segments.map((seg) => {
                              const segWidthPct = row.totalInvestment > 0 ? (seg.amount / row.totalInvestment) * 100 : 0;
                              const color = entryColorById.get(seg.entryId) || '#94A3B8';
                              const isActive = investmentHover?.entryName === seg.entryName && investmentHover?.amount === seg.amount;

                              return (
                                <div
                                  key={`${row.teamId}-${seg.entryId}`}
                                  className="h-full"
                                  style={{
                                    width: `${segWidthPct.toFixed(2)}%`,
                                    backgroundColor: color,
                                    boxSizing: 'border-box',
                                    border: isActive ? '2px solid #111827' : '2px solid transparent',
                                  }}
                                  onMouseEnter={(e) => {
                                    setInvestmentHover({
                                      entryName: seg.entryName,
                                      amount: seg.amount,
                                      x: e.clientX,
                                      y: e.clientY,
                                    });
                                  }}
                                  onMouseMove={(e) => {
                                    setInvestmentHover((prev) =>
                                      prev
                                        ? {
                                            ...prev,
                                            x: e.clientX,
                                            y: e.clientY,
                                          }
                                        : prev
                                    );
                                  }}
                                  onMouseLeave={() => setInvestmentHover(null)}
                                />
                              );
                            })}
                          </div>
                        </div>
                      </td>
                      <td className="px-3 py-3 text-right font-medium text-gray-900 rounded-r-md whitespace-nowrap">
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
      )}

      {activeTab === 'ownership' && (
        <>
          <div className="mb-4 flex items-center justify-end">
            <div className="text-sm text-gray-600">
              Total Spent: <span className="font-medium text-gray-900">${totalSpent.toFixed(2)}</span>
            </div>
          </div>

          <div className="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-3 xl:grid-cols-4 gap-6">
            {ownershipTeams.map((team) => (
              <div key={team.teamId} className="bg-white rounded-lg shadow p-4 flex flex-col">
                <div className="flex items-start justify-between gap-3">
                  <div className="min-w-0">
                    <h2 className="text-lg font-semibold leading-snug truncate">{team.teamName}</h2>
                    <div className="mt-1 text-sm text-gray-600">
                      Seed: {team.seed} · {team.region}
                    </div>
                  </div>
                  <div className="text-right">
                    <div className="text-sm text-gray-600">
                      Total Spend
                      <div className="text-base font-semibold text-gray-900">${team.totalSpend.toFixed(2)}</div>
                    </div>
                  </div>
                </div>

                <div className="mt-4 flex justify-center">
                  <CalcuttaOwnershipPieChart slices={team.slices} />
                </div>

                <div className="mt-4">
                  <div className="text-sm font-medium text-gray-900">Top Shareholders</div>
                  <div className="mt-2 space-y-2">
                    {team.topOwners.map((owner, idx) => {
                      const color = entryColorById.get(owner.entryId) || '#94A3B8';
                      return (
                        <div key={idx} className="flex items-center justify-between gap-3 text-sm">
                          <div className="min-w-0 truncate text-gray-700 flex items-center gap-2">
                            <span className="inline-block h-2.5 w-2.5 rounded-sm" style={{ backgroundColor: color }} />
                            <span className="truncate">{owner.name}</span>
                          </div>
                          <div className="font-medium text-gray-900">{owner.value.toFixed(2)}%</div>
                        </div>
                      );
                    })}
                  </div>
                </div>
              </div>
            ))}
          </div>
        </>
      )}

      {activeTab === 'leaderboard' && (
        <>
          <h2 className="text-2xl font-bold mb-6">Leaderboard</h2>
          <div className="grid gap-4">
            {entries.map((entry, index) => (
              <Link
                key={entry.id}
                to={`/calcuttas/${calcuttaId}/entries/${entry.id}`}
                className={`block p-4 rounded-lg shadow hover:shadow-md transition-shadow ${
                  index < 3 
                    ? 'bg-gradient-to-r from-yellow-50 to-yellow-100 border-2 border-yellow-400' 
                    : 'bg-white'
                }`}
              >
                <div className="flex justify-between items-center">
                  <div>
                    <h2 className="text-xl font-semibold">
                      {index + 1}. {entry.name}
                      {index < 3 && (
                        <span className="ml-2 text-yellow-600 text-sm">
                          {index === 0 ? '🥇' : index === 1 ? '🥈' : '🥉'}
                        </span>
                      )}
                    </h2>
                    <p className="text-gray-600">Created: {new Date(entry.created).toLocaleDateString()}</p>
                  </div>
                  <div className="text-right">
                    <p className={`text-2xl font-bold ${index < 3 ? 'text-yellow-600' : 'text-blue-600'}`}>
                      {entry.totalPoints ? entry.totalPoints.toFixed(2) : '0.00'} pts
                    </p>
                  </div>
                </div>
              </Link>
            ))}
          </div>
        </>
      )}
    </div>
  );
}
 