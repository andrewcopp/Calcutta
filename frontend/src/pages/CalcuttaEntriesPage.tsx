import { useMemo, useState } from 'react';
import { Link, useParams } from 'react-router-dom';
import { Calcutta, CalcuttaPortfolio, CalcuttaPortfolioTeam, CalcuttaEntryTeam } from '../types/calcutta';
import { calcuttaService } from '../services/calcuttaService';
import { useQuery } from '@tanstack/react-query';
import { Alert } from '../components/ui/Alert';
import { PageContainer, PageHeader } from '../components/ui/Page';
import { LeaderboardSkeleton } from '../components/skeletons/LeaderboardSkeleton';
import { StatisticsTab } from './CalcuttaEntries/StatisticsTab';
import { InvestmentTab } from './CalcuttaEntries/InvestmentTab';
import { ReturnsTab } from './CalcuttaEntries/ReturnsTab';
import { OwnershipTab } from './CalcuttaEntries/OwnershipTab';
import { Tabs, TabsList, TabsTrigger, TabsContent } from '../components/ui/Tabs';
import { Breadcrumb } from '../components/ui/Breadcrumb';
import { Badge } from '../components/ui/Badge';
import { Button } from '../components/ui/Button';
import { useUser } from '../contexts/useUser';
import { queryKeys } from '../queryKeys';

export function CalcuttaEntriesPage() {
  const { calcuttaId } = useParams<{ calcuttaId: string }>();
  const { user } = useUser();
  const [activeTab, setActiveTab] = useState('leaderboard');

  const formatDollarsFromCents = (cents?: number) => {
    if (!cents) return '$0';
    const abs = Math.abs(cents);
    const dollars = Math.floor(abs / 100);
    const remainder = abs % 100;
    const sign = cents < 0 ? '-' : '';
    if (remainder === 0) return `${sign}$${dollars}`;
    return `${sign}$${dollars}.${remainder.toString().padStart(2, '0')}`;
  };


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
        calcutta,
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

  const calcutta: Calcutta | undefined = calcuttaEntriesQuery.data?.calcutta;
  const entries = useMemo(() => calcuttaEntriesQuery.data?.entries ?? [], [calcuttaEntriesQuery.data?.entries]);
  const totalEntries = calcuttaEntriesQuery.data?.totalEntries || 0;
  const schools = useMemo(() => calcuttaEntriesQuery.data?.schools ?? [], [calcuttaEntriesQuery.data?.schools]);
  const calcuttaName = calcuttaEntriesQuery.data?.calcuttaName || '';
  const allEntryTeams = useMemo(() => calcuttaEntriesQuery.data?.allEntryTeams ?? [], [calcuttaEntriesQuery.data?.allEntryTeams]);
  const seedInvestmentData = calcuttaEntriesQuery.data?.seedInvestmentData || [];
  const tournamentTeams = useMemo(
    () => calcuttaEntriesQuery.data?.tournamentTeams ?? [],
    [calcuttaEntriesQuery.data?.tournamentTeams]
  );
  const allCalcuttaPortfolios = useMemo(
    () => calcuttaEntriesQuery.data?.allCalcuttaPortfolios ?? [],
    [calcuttaEntriesQuery.data?.allCalcuttaPortfolios]
  );
  const allCalcuttaPortfolioTeams = useMemo(
    () => calcuttaEntriesQuery.data?.allCalcuttaPortfolioTeams ?? [],
    [calcuttaEntriesQuery.data?.allCalcuttaPortfolioTeams]
  );

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
    return (
      <PageContainer>
        <Alert variant="error">Missing required parameters</Alert>
      </PageContainer>
    );
  }

  if (calcuttaEntriesQuery.isLoading) {
    return (
      <PageContainer>
        <PageHeader title="Loading..." />
        <LeaderboardSkeleton />
      </PageContainer>
    );
  }

  if (calcuttaEntriesQuery.isError) {
    const message = calcuttaEntriesQuery.error instanceof Error ? calcuttaEntriesQuery.error.message : 'Failed to fetch data';
    return (
      <PageContainer>
        <Alert variant="error">{message}</Alert>
      </PageContainer>
    );
  }

  return (
    <PageContainer>
      <Breadcrumb
        items={[
          { label: 'Calcuttas', href: '/calcuttas' },
          { label: calcuttaName },
        ]}
      />

      <PageHeader
        title={calcuttaName}
        actions={
          user?.id === calcutta?.ownerId ? (
            <Link to={`/calcuttas/${calcuttaId}/settings`}>
              <Button variant="outline" size="sm">Settings</Button>
            </Link>
          ) : undefined
        }
      />

      {calcutta && (
        <div className="mb-4">
          {calcutta.biddingOpen ? (
            <Badge variant="success">Bidding Open</Badge>
          ) : calcutta.biddingLockedAt ? (
            <Badge variant="warning">Bidding Locked</Badge>
          ) : (
            <Badge variant="secondary">Bidding Closed</Badge>
          )}
        </div>
      )}

      <Tabs value={activeTab} onValueChange={setActiveTab}>
        <TabsList>
          <TabsTrigger value="leaderboard">Leaderboard</TabsTrigger>
          <TabsTrigger value="investment">Investments</TabsTrigger>
          <TabsTrigger value="ownership">Ownerships</TabsTrigger>
          <TabsTrigger value="returns">Returns</TabsTrigger>
          <TabsTrigger value="statistics">Statistics</TabsTrigger>
        </TabsList>

        <TabsContent value="statistics">
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
        </TabsContent>

        <TabsContent value="returns">
          <ReturnsTab
            entries={entries}
            schools={schools}
            tournamentTeams={tournamentTeams}
            allCalcuttaPortfolios={allCalcuttaPortfolios}
            allCalcuttaPortfolioTeams={allCalcuttaPortfolioTeams}
          />
        </TabsContent>

        <TabsContent value="investment">
          <InvestmentTab entries={entries} schools={schools} tournamentTeams={tournamentTeams} allEntryTeams={allEntryTeams} />
        </TabsContent>

        <TabsContent value="ownership">
          <OwnershipTab
            entries={entries}
            schools={schools}
            tournamentTeams={tournamentTeams}
            allEntryTeams={allEntryTeams}
            allCalcuttaPortfolios={allCalcuttaPortfolios}
            allCalcuttaPortfolioTeams={allCalcuttaPortfolioTeams}
            isFetching={calcuttaEntriesQuery.isFetching}
          />
        </TabsContent>

        <TabsContent value="leaderboard">
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
                    </div>
                    <div className="text-right">
                      <p className={`text-2xl font-bold ${pointsClass}`}>
                        {entry.totalPoints ? entry.totalPoints.toFixed(2) : '0.00'} pts
                      </p>
                      {index > 0 && entries[0].totalPoints > 0 && (
                        <p className="text-xs text-gray-500">
                          {((entry.totalPoints || 0) - (entries[0].totalPoints || 0)).toFixed(2)} pts
                        </p>
                      )}
                    </div>
                  </div>
                </Link>
              );
            })}
          </div>
        </TabsContent>
      </Tabs>
    </PageContainer>
  );
}
 