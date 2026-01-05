import { useMemo, useState } from 'react';
import { Link, useParams } from 'react-router-dom';
import { CalcuttaEntry, CalcuttaPortfolio, CalcuttaPortfolioTeam, CalcuttaEntryTeam, TournamentTeam, School } from '../types/calcutta';
import { calcuttaService } from '../services/calcuttaService';
import { useQuery } from '@tanstack/react-query';
import { Alert } from '../components/ui/Alert';
import { StatisticsTab } from './CalcuttaEntries/StatisticsTab';
import { InvestmentTab } from './CalcuttaEntries/InvestmentTab';
import { ReturnsTab } from './CalcuttaEntries/ReturnsTab';
import { OwnershipTab } from './CalcuttaEntries/OwnershipTab';
import { TabsNav } from '../components/TabsNav';
import { queryKeys } from '../queryKeys';

export function CalcuttaEntriesPage() {
  const { calcuttaId } = useParams<{ calcuttaId: string }>();
  const [activeTab, setActiveTab] = useState<'leaderboard' | 'statistics' | 'investment' | 'ownership' | 'returns'>('leaderboard');

  const formatDollarsFromCents = (cents?: number) => {
    if (!cents) return '$0';
    const abs = Math.abs(cents);
    const dollars = Math.floor(abs / 100);
    const remainder = abs % 100;
    const sign = cents < 0 ? '-' : '';
    if (remainder === 0) return `${sign}$${dollars}`;
    return `${sign}$${dollars}.${remainder.toString().padStart(2, '0')}`;
  };

  const tabs = useMemo(
    () =>
      [
        { id: 'leaderboard' as const, label: 'Leaderboard' },
        { id: 'investment' as const, label: 'Investments' },
        { id: 'ownership' as const, label: 'Ownerships' },
        { id: 'returns' as const, label: 'Returns' },
        { id: 'statistics' as const, label: 'Statistics' },
      ] as const,
    []
  );

  const calcuttaEntriesQuery = useQuery({
    queryKey: queryKeys.calcuttas.entriesPage(calcuttaId),
    enabled: Boolean(calcuttaId),
    staleTime: 30_000,
    queryFn: async () => {
      if (!calcuttaId) throw new Error('Missing calcuttaId');

      const calcutta = await calcuttaService.getCalcutta(calcuttaId);
      const [entriesData, schoolsData, tournamentTeamsData] = await Promise.all([
        calcuttaService.getCalcuttaEntries(calcuttaId),
        calcuttaService.getSchools(),
        calcuttaService.getTournamentTeams(calcutta.tournamentId),
      ]);

      const schoolMap = new Map(schoolsData.map((school) => [school.id, school]));

      const portfoliosByEntry = await Promise.all(
        entriesData.map(async (entry) => {
          const portfolios = await calcuttaService.getPortfoliosByEntry(entry.id);
          return { entry, portfolios };
        })
      );

      const allCalcuttaPortfolios: (CalcuttaPortfolio & { entryName?: string })[] = portfoliosByEntry.flatMap(({ entry, portfolios }) =>
        portfolios.map((portfolio) => ({
          ...portfolio,
          entryName: entry.name,
        }))
      );

      const portfolioTeamsByPortfolio = await Promise.all(
        allCalcuttaPortfolios.map(async (portfolio) => {
          const teams = await calcuttaService.getPortfolioTeams(portfolio.id);
          return { portfolio, teams };
        })
      );

      const entryPointsMap = new Map<string, number>();
      const allCalcuttaPortfolioTeams: CalcuttaPortfolioTeam[] = portfolioTeamsByPortfolio.flatMap(({ portfolio, teams }) => {
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

      const entriesWithPoints = entriesData
        .map((entry) => ({
          ...entry,
          totalPoints: entryPointsMap.get(entry.id) || 0,
        }))
        .sort((a, b) => {
          const diff = (b.totalPoints || 0) - (a.totalPoints || 0);
          if (diff !== 0) return diff;

          return new Date(b.created).getTime() - new Date(a.created).getTime();
        });

      const entryTeamsByEntry = await Promise.all(entriesData.map((entry) => calcuttaService.getEntryTeams(entry.id, calcuttaId)));
      const allEntryTeams: CalcuttaEntryTeam[] = entryTeamsByEntry.flat();

      const seedMap = new Map<number, number>();
      for (const team of allEntryTeams) {
        if (!team.team?.seed || !team.bid) continue;
        const seed = team.team.seed;
        seedMap.set(seed, (seedMap.get(seed) || 0) + team.bid);
      }

      const seedInvestmentData = Array.from(seedMap.entries())
        .map(([seed, totalInvestment]) => ({ seed, totalInvestment }))
        .sort((a, b) => a.seed - b.seed);

      return {
        calcuttaName: calcutta.name,
        entries: entriesWithPoints,
        totalEntries: entriesData.length,
        schools: schoolsData,
        tournamentTeams: tournamentTeamsData,
        allCalcuttaPortfolios,
        allCalcuttaPortfolioTeams,
        allEntryTeams,
        seedInvestmentData,
      };
    },
  });

  const entries = calcuttaEntriesQuery.data?.entries || [];
  const totalEntries = calcuttaEntriesQuery.data?.totalEntries || 0;
  const schools: School[] = calcuttaEntriesQuery.data?.schools || [];
  const calcuttaName = calcuttaEntriesQuery.data?.calcuttaName || '';
  const allEntryTeams = calcuttaEntriesQuery.data?.allEntryTeams || [];
  const seedInvestmentData = calcuttaEntriesQuery.data?.seedInvestmentData || [];
  const tournamentTeams = calcuttaEntriesQuery.data?.tournamentTeams || [];
  const allCalcuttaPortfolios = calcuttaEntriesQuery.data?.allCalcuttaPortfolios || [];
  const allCalcuttaPortfolioTeams = calcuttaEntriesQuery.data?.allCalcuttaPortfolioTeams || [];

  const totalInvestment = useMemo(() => allEntryTeams.reduce((sum, et) => sum + (et.bid || 0), 0), [allEntryTeams]);

  const totalReturns = useMemo(() => entries.reduce((sum, e) => sum + (e.totalPoints || 0), 0), [entries]);

  const averageReturn = useMemo(() => (totalEntries > 0 ? totalReturns / totalEntries : 0), [totalEntries, totalReturns]);

  const returnsStdDev = useMemo(() => {
    if (totalEntries <= 1) return 0;

    const mean = averageReturn;
    const variance = entries.reduce((acc, e) => {
      const v = (e.totalPoints || 0) - mean;
      return acc + v * v;
    }, 0);
    return Math.sqrt(variance / totalEntries);
  }, [entries, totalEntries, averageReturn]);

  const schoolNameById = useMemo(() => new Map(schools.map((school) => [school.id, school.name])), [schools]);

  const teamROIData = useMemo(() => {
    const roiTeams = tournamentTeams.map((team) => {
      const schoolName = schoolNameById.get(team.schoolId) || 'Unknown School';
      
      // Calculate total investment for this team
      const teamInvestment = allEntryTeams
        .filter((et) => et.teamId === team.id)
        .reduce((sum, et) => sum + (et.bid || 0), 0);
      
      // Calculate total points for this team
      const teamPoints = allCalcuttaPortfolioTeams
        .filter((pt) => pt.teamId === team.id)
        .reduce((sum, pt) => sum + (pt.actualPoints || 0), 0);
      
      // Calculate ROI with +$1 to avoid division by zero
      const roi = teamPoints / (teamInvestment + 1);
      
      return {
        teamId: team.id,
        seed: team.seed,
        region: team.region,
        teamName: schoolName,
        points: teamPoints,
        investment: teamInvestment,
        roi: roi,
      };
    });
    
    // Sort by ROI descending (highest ROI first)
    return roiTeams.sort((a, b) => b.roi - a.roi);
  }, [tournamentTeams, schoolNameById, allEntryTeams, allCalcuttaPortfolioTeams]);

  if (!calcuttaId) {
    return <Alert variant="error">Missing required parameters</Alert>;
  }

  if (calcuttaEntriesQuery.isLoading) {
    return <div>Loading...</div>;
  }

  if (calcuttaEntriesQuery.isError) {
    const message = calcuttaEntriesQuery.error instanceof Error ? calcuttaEntriesQuery.error.message : 'Failed to fetch data';
    return <Alert variant="error">{message}</Alert>;
  }

  return (
    <div className="container mx-auto px-4 py-8">
      <div className="mb-6">
        <Link to="/calcuttas" className="text-blue-600 hover:text-blue-800">← Back to Calcuttas</Link>
      </div>
      <h1 className="text-3xl font-bold mb-6">{calcuttaName}</h1>

      <TabsNav tabs={tabs} activeTab={activeTab} onTabChange={setActiveTab} />
      

      {activeTab === 'statistics' && (
        <StatisticsTab
          calcuttaId={calcuttaId}
          totalEntries={totalEntries}
          totalInvestment={totalInvestment}
          totalReturns={totalReturns}
          averageReturn={averageReturn}
          returnsStdDev={returnsStdDev}
          seedInvestmentData={seedInvestmentData}
          teamROIData={teamROIData}
        />
      )}

      {activeTab === 'returns' && (
        <ReturnsTab
          entries={entries}
          schools={schools}
          tournamentTeams={tournamentTeams}
          allCalcuttaPortfolios={allCalcuttaPortfolios}
          allCalcuttaPortfolioTeams={allCalcuttaPortfolioTeams}
        />
      )}

      {activeTab === 'investment' && (
        <InvestmentTab entries={entries} schools={schools} tournamentTeams={tournamentTeams} allEntryTeams={allEntryTeams} />
      )}

      {activeTab === 'ownership' && (
        <OwnershipTab
          entries={entries}
          schools={schools}
          tournamentTeams={tournamentTeams}
          allEntryTeams={allEntryTeams}
          allCalcuttaPortfolios={allCalcuttaPortfolios}
          allCalcuttaPortfolioTeams={allCalcuttaPortfolioTeams}
          isFetching={calcuttaEntriesQuery.isFetching}
        />
      )}

      {activeTab === 'leaderboard' && (
        <>
          <div className="grid gap-4">
            {entries.map((entry, index) => {
              const displayPosition = entry.finishPosition || index + 1;
              const isInTheMoney = Boolean(entry.inTheMoney);
              const payoutText = entry.payoutCents ? `(${formatDollarsFromCents(entry.payoutCents)})` : '';

              const rowClass = isInTheMoney
                ? displayPosition === 1
                  ? 'bg-gradient-to-r from-yellow-50 to-yellow-100 border-2 border-yellow-400'
                  : displayPosition === 2
                    ? 'bg-gradient-to-r from-slate-50 to-slate-200 border-2 border-slate-400'
                    : displayPosition === 3
                      ? 'bg-gradient-to-r from-amber-50 to-amber-100 border-2 border-amber-500'
                      : 'bg-gradient-to-r from-slate-50 to-blue-50 border-2 border-slate-300'
                : 'bg-white';

              const pointsClass = isInTheMoney
                ? displayPosition === 1
                  ? 'text-yellow-700'
                  : displayPosition === 2
                    ? 'text-slate-700'
                    : displayPosition === 3
                      ? 'text-amber-700'
                      : 'text-slate-700'
                : 'text-blue-600';

              return (
                <Link
                  key={entry.id}
                  to={`/calcuttas/${calcuttaId}/entries/${entry.id}`}
                  className={`block p-4 rounded-lg shadow hover:shadow-md transition-shadow ${rowClass}`}
                >
                  <div className="flex justify-between items-center">
                    <div>
                      <h2 className="text-xl font-semibold">
                        {displayPosition}. {entry.name}
                        {isInTheMoney && payoutText && <span className="ml-2 text-sm text-gray-700">{payoutText}</span>}
                      </h2>
                      <p className="text-gray-600">Created: {new Date(entry.created).toLocaleDateString()}</p>
                    </div>
                    <div className="text-right">
                      <p className={`text-2xl font-bold ${pointsClass}`}>
                        {entry.totalPoints ? entry.totalPoints.toFixed(2) : '0.00'} pts
                      </p>
                    </div>
                  </div>
                </Link>
              );
            })}
          </div>
        </>
      )}
    </div>
  );
}
 