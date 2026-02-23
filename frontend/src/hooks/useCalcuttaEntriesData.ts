import { useMemo } from 'react';
import {
  CalcuttaPortfolio,
  CalcuttaPortfolioTeam,
  CalcuttaEntryTeam,
  CalcuttaEntry,
  CalcuttaDashboard,
} from '../types/calcutta';
import type { TournamentTeam } from '../types/tournament';

interface SeedInvestmentDatum {
  seed: number;
  totalInvestment: number;
}

interface TeamROIDatum {
  teamId: string;
  seed: number;
  region: string;
  teamName: string;
  points: number;
  investment: number;
  roi: number;
}

type EnrichedEntry = CalcuttaEntry & { totalPoints: number };

export interface CalcuttaEntriesData {
  entries: EnrichedEntry[];
  totalEntries: number;
  allCalcuttaPortfolios: (CalcuttaPortfolio & { entryName?: string })[];
  allCalcuttaPortfolioTeams: CalcuttaPortfolioTeam[];
  allEntryTeams: CalcuttaEntryTeam[];
  seedInvestmentData: SeedInvestmentDatum[];
  schools: { id: string; name: string }[];
  tournamentTeams: TournamentTeam[];
  totalInvestment: number;
  totalReturns: number;
  averageReturn: number;
  returnsStdDev: number;
  teamROIData: TeamROIDatum[];
}

export function useCalcuttaEntriesData(dashboardData: CalcuttaDashboard | undefined): CalcuttaEntriesData {
  const { entries, totalEntries, allCalcuttaPortfolios, allCalcuttaPortfolioTeams, allEntryTeams, seedInvestmentData } =
    useMemo(() => {
      if (!dashboardData) {
        return {
          entries: [] as EnrichedEntry[],
          totalEntries: 0,
          allCalcuttaPortfolios: [] as (CalcuttaPortfolio & { entryName?: string })[],
          allCalcuttaPortfolioTeams: [] as CalcuttaPortfolioTeam[],
          allEntryTeams: [] as CalcuttaEntryTeam[],
          seedInvestmentData: [] as SeedInvestmentDatum[],
        };
      }

      const { entries: rawEntries, entryTeams, portfolios, portfolioTeams, schools } = dashboardData;
      const schoolMap = new Map(schools.map((s) => [s.id, s]));
      const entryNameMap = new Map(rawEntries.map((e) => [e.id, e.name]));

      // Enrich portfolios with entry names
      const enrichedPortfolios: (CalcuttaPortfolio & { entryName?: string })[] = portfolios.map((p) => ({
        ...p,
        entryName: entryNameMap.get(p.entryId),
      }));

      // Compute points per entry from portfolio teams
      const entryPointsMap = new Map<string, number>();
      const portfolioToEntry = new Map(portfolios.map((p) => [p.id, p.entryId]));
      for (const pt of portfolioTeams) {
        const entryId = portfolioToEntry.get(pt.portfolioId);
        if (entryId) {
          entryPointsMap.set(entryId, (entryPointsMap.get(entryId) || 0) + pt.actualPoints);
        }
      }

      // Enrich portfolio teams with school info
      const enrichedPortfolioTeams: CalcuttaPortfolioTeam[] = portfolioTeams.map((pt) => {
        const school = pt.team?.schoolId ? schoolMap.get(pt.team.schoolId) : undefined;
        return {
          ...pt,
          team: pt.team
            ? {
                ...pt.team,
                school: school ? { id: school.id, name: school.name } : pt.team.school,
              }
            : pt.team,
        };
      });

      // Sort entries by points
      const sortedEntries: EnrichedEntry[] = rawEntries
        .map((entry) => ({
          ...entry,
          totalPoints: entry.totalPoints || entryPointsMap.get(entry.id) || 0,
        }))
        .sort((a, b) => {
          const diff = (b.totalPoints || 0) - (a.totalPoints || 0);
          if (diff !== 0) return diff;
          return new Date(b.createdAt).getTime() - new Date(a.createdAt).getTime();
        });

      // Compute seed investment data
      const seedMap = new Map<number, number>();
      for (const team of entryTeams) {
        if (!team.team?.seed || !team.bid) continue;
        const seed = team.team.seed;
        seedMap.set(seed, (seedMap.get(seed) || 0) + team.bid);
      }
      const seedData = Array.from(seedMap.entries())
        .map(([seed, totalInvestment]) => ({ seed, totalInvestment }))
        .sort((a, b) => a.seed - b.seed);

      return {
        entries: sortedEntries,
        totalEntries: rawEntries.length,
        allCalcuttaPortfolios: enrichedPortfolios,
        allCalcuttaPortfolioTeams: enrichedPortfolioTeams,
        allEntryTeams: entryTeams,
        seedInvestmentData: seedData,
      };
    }, [dashboardData]);

  const schools = useMemo(() => dashboardData?.schools ?? [], [dashboardData?.schools]);
  const tournamentTeams = useMemo(() => dashboardData?.tournamentTeams ?? [], [dashboardData?.tournamentTeams]);

  const totalInvestment = useMemo(() => allEntryTeams.reduce((sum, et) => sum + et.bid, 0), [allEntryTeams]);

  const totalReturns = useMemo(() => entries.reduce((sum, e) => sum + (e.totalPoints || 0), 0), [entries]);

  const averageReturn = useMemo(
    () => (totalEntries > 0 ? totalReturns / totalEntries : 0),
    [totalEntries, totalReturns],
  );

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
      const teamInvestment = allEntryTeams.filter((et) => et.teamId === team.id).reduce((sum, et) => sum + et.bid, 0);

      // Calculate total points for this team
      const teamPoints = allCalcuttaPortfolioTeams
        .filter((pt) => pt.teamId === team.id)
        .reduce((sum, pt) => sum + pt.actualPoints, 0);

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

  return {
    entries,
    totalEntries,
    allCalcuttaPortfolios,
    allCalcuttaPortfolioTeams,
    allEntryTeams,
    seedInvestmentData,
    schools,
    tournamentTeams,
    totalInvestment,
    totalReturns,
    averageReturn,
    returnsStdDev,
    teamROIData,
  };
}
