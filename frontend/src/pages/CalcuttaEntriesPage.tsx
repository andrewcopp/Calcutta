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

interface OwnershipSlice {
  name: string;
  value: number;
  entryId: string;
}

interface OwnershipTeamCard {
  teamId: string;
  seed: number;
  region: string;
  teamName: string;
  totalSpend: number;
  slices: OwnershipSlice[];
  topOwners: OwnershipSlice[];
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
  const [activeTab, setActiveTab] = useState<'leaderboard' | 'statistics' | 'investment' | 'ownership' | 'returns'>('leaderboard');
  const [tournamentTeams, setTournamentTeams] = useState<TournamentTeam[]>([]);
  const [allCalcuttaPortfolios, setAllCalcuttaPortfolios] = useState<(CalcuttaPortfolio & { entryName?: string })[]>([]);
  const [allCalcuttaPortfolioTeams, setAllCalcuttaPortfolioTeams] = useState<CalcuttaPortfolioTeam[]>([]);
  const [investmentHover, setInvestmentHover] = useState<{ entryName: string; amount: number; x: number; y: number } | null>(null);
  const [investmentSortBy, setInvestmentSortBy] = useState<'total' | 'seed' | 'region' | 'team'>('total');
  const [ownershipSortBy, setOwnershipSortBy] = useState<'seed' | 'region' | 'team' | 'investment'>('seed');
  const [ownershipLoading, setOwnershipLoading] = useState(false);
  const [ownershipTeamsData, setOwnershipTeamsData] = useState<OwnershipTeamCard[]>([]);
  const [returnsHover, setReturnsHover] = useState<{ entryName: string; amount: number; x: number; y: number } | null>(null);
  const [returnsSortBy, setReturnsSortBy] = useState<'points' | 'seed' | 'region' | 'team'>('points');

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

  const portfolioEntryIdById = useMemo(() => {
    const map = new Map<string, string>();
    allCalcuttaPortfolios.forEach((p) => {
      map.set(p.id, p.entryId);
    });
    return map;
  }, [allCalcuttaPortfolios]);

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
      .sort((a, b) => {
        if (b.totalInvestment !== a.totalInvestment) return b.totalInvestment - a.totalInvestment;
        const seedA = a.seed ?? 999;
        const seedB = b.seed ?? 999;
        return seedA - seedB;
      });

    const maxTotal = rows.reduce((max, row) => Math.max(max, row.totalInvestment), 0);

    return { rows, maxTotal };
  }, [allEntryTeams, entryNameById, schoolNameById, tournamentTeams, tournamentTeamById]);

  const investmentSortedRows = useMemo(() => {
    const rows = investmentRows.rows.slice();

    const seedA = (seed: number | undefined) => seed ?? 999;
    const regionA = (region: string) => region || '';
    const teamA = (name: string) => name || '';

    rows.sort((a, b) => {
      if (investmentSortBy === 'total') {
        if (b.totalInvestment !== a.totalInvestment) return b.totalInvestment - a.totalInvestment;
        return seedA(a.seed) - seedA(b.seed);
      }

      if (investmentSortBy === 'seed') {
        const seedDiff = seedA(a.seed) - seedA(b.seed);
        if (seedDiff !== 0) return seedDiff;
        const regionDiff = regionA(a.region).localeCompare(regionA(b.region));
        if (regionDiff !== 0) return regionDiff;
        return teamA(a.teamName).localeCompare(teamA(b.teamName));
      }

      if (investmentSortBy === 'region') {
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
  }, [investmentRows.rows, investmentSortBy]);

  useEffect(() => {
    if (activeTab !== 'ownership') {
      setOwnershipLoading(false);
      return;
    }

    setOwnershipLoading(true);
    const handle = window.setTimeout(() => {
      const byTeamSpend = new Map<string, number>();
      for (const entryTeam of allEntryTeams) {
        const amount = entryTeam.bid || 0;
        if (amount <= 0) continue;
        byTeamSpend.set(entryTeam.teamId, (byTeamSpend.get(entryTeam.teamId) || 0) + amount);
      }

      const cards: OwnershipTeamCard[] = tournamentTeams.map((team) => {
        const teamId = team.id;
        const spend = byTeamSpend.get(teamId) || 0;
        const teamPortfolioTeams = allCalcuttaPortfolioTeams.filter((pt) => pt.teamId === teamId && pt.ownershipPercentage > 0);

        const slices: OwnershipSlice[] = teamPortfolioTeams
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

        return {
          teamId,
          seed: team.seed,
          region: team.region,
          teamName: schoolNameById.get(team.schoolId) || 'Unknown School',
          totalSpend: spend,
          slices,
          topOwners: slices.slice(0, 3),
        };
      });

      setOwnershipTeamsData(cards);
      setOwnershipLoading(false);
    }, 0);

    return () => {
      window.clearTimeout(handle);
      setOwnershipLoading(false);
    };
  }, [activeTab, allCalcuttaPortfolios, allCalcuttaPortfolioTeams, allEntryTeams, entryNameById, schoolNameById, tournamentTeams]);

  const ownershipSortedTeams = useMemo(() => {
    const cards = ownershipTeamsData.slice();

    const seedA = (seed: number) => seed;
    const regionA = (region: string) => region || '';
    const teamA = (name: string) => name || '';

    cards.sort((a, b) => {
      if (ownershipSortBy === 'investment') {
        if (b.totalSpend !== a.totalSpend) return b.totalSpend - a.totalSpend;
        const seedDiff = seedA(a.seed) - seedA(b.seed);
        if (seedDiff !== 0) return seedDiff;
        return teamA(a.teamName).localeCompare(teamA(b.teamName));
      }

      if (ownershipSortBy === 'seed') {
        const seedDiff = seedA(a.seed) - seedA(b.seed);
        if (seedDiff !== 0) return seedDiff;
        const regionDiff = regionA(a.region).localeCompare(regionA(b.region));
        if (regionDiff !== 0) return regionDiff;
        return teamA(a.teamName).localeCompare(teamA(b.teamName));
      }

      if (ownershipSortBy === 'region') {
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

    return cards;
  }, [ownershipSortBy, ownershipTeamsData]);

  const returnsRows = useMemo(() => {
    const byTeam = new Map<
      string,
      {
        teamId: string;
        seed: number | undefined;
        region: string;
        teamName: string;
        eliminated: boolean;
        pointsSegments: { entryId: string; entryName: string; amount: number }[];
        possibleSegments: { entryId: string; entryName: string; amount: number }[];
        totalPoints: number;
        totalPossible: number;
      }
    >();

    for (const tt of tournamentTeams) {
      byTeam.set(tt.id, {
        teamId: tt.id,
        seed: tt.seed,
        region: tt.region,
        teamName: schoolNameById.get(tt.schoolId) || 'Unknown School',
        eliminated: tt.eliminated === true,
        pointsSegments: [],
        possibleSegments: [],
        totalPoints: 0,
        totalPossible: 0,
      });
    }

    const teamEntryAgg = new Map<string, Map<string, { actual: number; expected: number }>>();
    for (const pt of allCalcuttaPortfolioTeams) {
      const entryId = portfolioEntryIdById.get(pt.portfolioId);
      if (!entryId) continue;
      if (!teamEntryAgg.has(pt.teamId)) teamEntryAgg.set(pt.teamId, new Map());
      const byEntry = teamEntryAgg.get(pt.teamId)!;
      const current = byEntry.get(entryId) || { actual: 0, expected: 0 };
      current.actual += pt.actualPoints || 0;
      current.expected += pt.expectedPoints || 0;
      byEntry.set(entryId, current);
    }

    for (const [teamId, byEntry] of teamEntryAgg.entries()) {
      const row = byTeam.get(teamId);
      if (!row) continue;

      const pointsSegments: { entryId: string; entryName: string; amount: number }[] = [];
      const possibleSegments: { entryId: string; entryName: string; amount: number }[] = [];

      for (const [entryId, agg] of byEntry.entries()) {
        const entryName = entryNameById.get(entryId) || 'Unknown Entry';
        const actual = agg.actual;
        const possible = row.eliminated ? actual : Math.max(agg.expected, actual);

        if (actual > 0) {
          pointsSegments.push({ entryId, entryName, amount: actual });
        }
        if (possible > 0) {
          possibleSegments.push({ entryId, entryName, amount: possible });
        }
      }

      pointsSegments.sort((a, b) => b.amount - a.amount);
      possibleSegments.sort((a, b) => b.amount - a.amount);

      row.pointsSegments = pointsSegments;
      row.possibleSegments = possibleSegments;
      row.totalPoints = pointsSegments.reduce((sum, s) => sum + s.amount, 0);
      row.totalPossible = possibleSegments.reduce((sum, s) => sum + s.amount, 0);
    }

    const rows = Array.from(byTeam.values()).sort((a, b) => {
      const seedA = a.seed ?? 999;
      const seedB = b.seed ?? 999;
      if (seedA !== seedB) return seedA - seedB;
      const regionDiff = (a.region || '').localeCompare(b.region || '');
      if (regionDiff !== 0) return regionDiff;
      return (a.teamName || '').localeCompare(b.teamName || '');
    });

    const maxTotal = rows.reduce((max, r) => Math.max(max, r.totalPossible, r.totalPoints), 0);

    return { rows, maxTotal };
  }, [allCalcuttaPortfolioTeams, entryNameById, portfolioEntryIdById, schoolNameById, tournamentTeams]);

  const returnsSortedRows = useMemo(() => {
    const rows = returnsRows.rows.slice();
    const seedA = (seed: number | undefined) => seed ?? 999;
    const regionA = (region: string) => region || '';
    const teamA = (name: string) => name || '';

    rows.sort((a, b) => {
      if (returnsSortBy === 'points') {
        if (b.totalPoints !== a.totalPoints) return b.totalPoints - a.totalPoints;
        if (b.totalPossible !== a.totalPossible) return b.totalPossible - a.totalPossible;
        const seedDiff = seedA(a.seed) - seedA(b.seed);
        if (seedDiff !== 0) return seedDiff;
        return teamA(a.teamName).localeCompare(teamA(b.teamName));
      }

      if (returnsSortBy === 'seed') {
        const seedDiff = seedA(a.seed) - seedA(b.seed);
        if (seedDiff !== 0) return seedDiff;
        const regionDiff = regionA(a.region).localeCompare(regionA(b.region));
        if (regionDiff !== 0) return regionDiff;
        return teamA(a.teamName).localeCompare(teamA(b.teamName));
      }

      if (returnsSortBy === 'region') {
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
  }, [returnsRows.rows, returnsSortBy]);

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
              isAnimationActive={false}
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
          Investments
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

      {activeTab === 'returns' && (
        <div className="bg-white rounded-lg shadow p-6">
          <div className="mb-4 flex items-center justify-end">
            <label className="text-sm text-gray-600">
              Sort by
              <select
                className="ml-2 rounded-md border border-gray-300 bg-white px-3 py-2 text-sm text-gray-900 shadow-sm"
                value={returnsSortBy}
                onChange={(e) => setReturnsSortBy(e.target.value as 'points' | 'seed' | 'region' | 'team')}
              >
                <option value="points">Points</option>
                <option value="seed">Seed</option>
                <option value="region">Region</option>
                <option value="team">Team</option>
              </select>
            </label>
          </div>
          <div className="mt-2 overflow-x-auto">
            <table className="min-w-full table-fixed border-separate border-spacing-y-2">
              <thead>
                <tr className="text-left text-xs uppercase tracking-wide text-gray-500">
                  <th className="px-2 py-2 w-14">Seed</th>
                  <th className="px-2 py-2 w-20">Region</th>
                  <th className="px-2 py-2 w-44">Team</th>
                  <th className="px-2 py-2"></th>
                  <th className="px-2 py-2 w-32 text-right">Total Points</th>
                </tr>
              </thead>
              <tbody>
                {returnsSortedRows.map((row) => {
                  const pointsWidthPct = returnsRows.maxTotal > 0 ? (row.totalPoints / returnsRows.maxTotal) * 100 : 0;
                  const possibleWidthPct = returnsRows.maxTotal > 0 ? (row.totalPossible / returnsRows.maxTotal) * 100 : 0;

                  return (
                    <tr key={row.teamId} className="bg-gray-50">
                      <td className="px-2 py-3 font-medium text-gray-900 rounded-l-md whitespace-nowrap">{row.seed ?? '—'}</td>
                      <td className="px-2 py-3 text-gray-700 whitespace-nowrap">{row.region}</td>
                      <td className="px-2 py-3 text-gray-900 font-medium whitespace-nowrap truncate">{row.teamName}</td>
                      <td className="px-2 py-3">
                        <div className="space-y-2">
                          <div>
                            <div className="h-6 w-full rounded bg-gray-200 overflow-hidden">
                              <div className="h-full flex" style={{ width: `${pointsWidthPct.toFixed(2)}%` }}>
                                {row.pointsSegments.map((seg) => {
                                  const segWidthPct = row.totalPoints > 0 ? (seg.amount / row.totalPoints) * 100 : 0;
                                  const color = entryColorById.get(seg.entryId) || '#94A3B8';
                                  const isActive = returnsHover?.entryName === seg.entryName && returnsHover?.amount === seg.amount;

                                  return (
                                    <div
                                      key={`${row.teamId}-points-${seg.entryId}`}
                                      className="h-full"
                                      style={{
                                        width: `${segWidthPct.toFixed(2)}%`,
                                        backgroundColor: color,
                                        boxSizing: 'border-box',
                                        border: isActive ? '2px solid #111827' : '2px solid transparent',
                                      }}
                                      onMouseEnter={(e) => {
                                        setReturnsHover({
                                          entryName: seg.entryName,
                                          amount: seg.amount,
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
                                  );
                                })}
                              </div>
                            </div>
                          </div>

                          <div>
                            <div className="h-6 w-full rounded bg-gray-200 overflow-hidden">
                              <div className="h-full flex" style={{ width: `${possibleWidthPct.toFixed(2)}%` }}>
                                {row.possibleSegments.map((seg) => {
                                  const segWidthPct = row.totalPossible > 0 ? (seg.amount / row.totalPossible) * 100 : 0;
                                  const base = entryColorById.get(seg.entryId) || '#94A3B8';
                                  const color = desaturate(base);
                                  const isActive = returnsHover?.entryName === seg.entryName && returnsHover?.amount === seg.amount;

                                  return (
                                    <div
                                      key={`${row.teamId}-possible-${seg.entryId}`}
                                      className="h-full"
                                      style={{
                                        width: `${segWidthPct.toFixed(2)}%`,
                                        backgroundColor: color,
                                        boxSizing: 'border-box',
                                        border: isActive ? '2px solid #111827' : '2px solid transparent',
                                      }}
                                      onMouseEnter={(e) => {
                                        setReturnsHover({
                                          entryName: seg.entryName,
                                          amount: seg.amount,
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
                                  );
                                })}
                              </div>
                            </div>
                          </div>
                        </div>
                      </td>
                      <td className="px-2 py-3 text-right font-medium text-gray-900 rounded-r-md whitespace-nowrap">
                        {row.totalPoints.toFixed(2)}
                      </td>
                    </tr>
                  );
                })}
              </tbody>
            </table>
          </div>

          {returnsHover && (
            <div
              className="fixed z-50 pointer-events-none rounded bg-gray-900 px-3 py-2 text-xs text-white shadow"
              style={{ left: returnsHover.x + 12, top: returnsHover.y + 12 }}
            >
              <div className="font-medium">{returnsHover.entryName}</div>
              <div>{returnsHover.amount.toFixed(2)}</div>
            </div>
          )}
        </div>
      )}

      {activeTab === 'investment' && (
        <div className="bg-white rounded-lg shadow p-6">
          <div className="mb-4 flex items-center justify-end">
            <label className="text-sm text-gray-600">
              Sort by
              <select
                className="ml-2 rounded-md border border-gray-300 bg-white px-3 py-2 text-sm text-gray-900 shadow-sm"
                value={investmentSortBy}
                onChange={(e) => setInvestmentSortBy(e.target.value as 'total' | 'seed' | 'region' | 'team')}
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
                  <th className="px-2 py-2 w-24 text-right">Total Investment</th>
                </tr>
              </thead>
              <tbody>
                {investmentSortedRows.map((row) => {
                  const barWidthPct = investmentRows.maxTotal > 0 ? (row.totalInvestment / investmentRows.maxTotal) * 100 : 0;
                  return (
                    <tr key={row.teamId} className="bg-gray-50">
                      <td className="px-2 py-3 font-medium text-gray-900 rounded-l-md whitespace-nowrap">{row.seed ?? '—'}</td>
                      <td className="px-2 py-3 text-gray-700 whitespace-nowrap">{row.region}</td>
                      <td className="px-2 py-3 text-gray-900 font-medium whitespace-nowrap truncate">{row.teamName}</td>
                      <td className="px-2 py-3">
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
      )}

      {activeTab === 'ownership' && (
        <>
          <div className="mb-4 flex items-center justify-end">
            <label className="text-sm text-gray-600">
              Sort by
              <select
                className="ml-2 rounded-md border border-gray-300 bg-white px-3 py-2 text-sm text-gray-900 shadow-sm"
                value={ownershipSortBy}
                onChange={(e) => setOwnershipSortBy(e.target.value as 'seed' | 'region' | 'team' | 'investment')}
              >
                <option value="seed">Seed</option>
                <option value="region">Region</option>
                <option value="team">Team</option>
                <option value="investment">Investment</option>
              </select>
            </label>
          </div>

          {ownershipLoading ? (
            <div className="bg-white rounded-lg shadow p-6 text-gray-600">Loading ownership…</div>
          ) : (
            <div className="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-3 xl:grid-cols-4 gap-6">
              {ownershipSortedTeams.map((team) => (
              <div key={team.teamId} className="bg-white rounded-lg shadow p-4 flex flex-col">
                <div className="flex items-start justify-between gap-3">
                  <div className="min-w-0">
                    <h2 className="text-lg font-semibold leading-snug truncate">{team.teamName}</h2>
                    <div className="mt-1 text-sm text-gray-600">
                      {team.seed} Seed - {team.region}
                    </div>
                  </div>
                  <div className="text-right">
                    <div className="text-sm text-gray-600 whitespace-nowrap">
                      Total Investment
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
                    {Array.from({ length: 3 }).map((_, idx) => {
                      const owner = team.topOwners[idx];
                      const name = owner?.name ?? '--';
                      const pct = owner ? `${owner.value.toFixed(2)}%` : '--';
                      return (
                        <div key={idx} className="flex items-center justify-between gap-3 text-sm">
                          <div className="min-w-0 truncate text-gray-700 flex items-center gap-2">
                            <div className="w-4 shrink-0 text-gray-500">{idx + 1}</div>
                            <div className="truncate">{name}</div>
                          </div>
                          <div className="font-medium text-gray-900">{pct}</div>
                        </div>
                      );
                    })}
                  </div>
                </div>
              </div>
              ))}
            </div>
          )}
        </>
      )}

      {activeTab === 'leaderboard' && (
        <>
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
 