import { useMemo } from 'react';
import {
  OwnershipSummary,
  OwnershipDetail,
  Investment,
  Portfolio,
  PoolDashboard,
} from '../schemas/pool';
import type { TournamentTeam } from '../schemas/tournament';

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

type EnrichedPortfolio = Portfolio & { totalReturns: number };

export interface PoolPortfoliosData {
  portfolios: EnrichedPortfolio[];
  totalPortfolios: number;
  allOwnershipSummaries: (OwnershipSummary & { portfolioName?: string })[];
  allOwnershipDetails: OwnershipDetail[];
  allInvestments: Investment[];
  seedInvestmentData: SeedInvestmentDatum[];
  schools: { id: string; name: string }[];
  tournamentTeams: TournamentTeam[];
  totalInvestment: number;
  totalReturns: number;
  averageReturn: number;
  returnsStdDev: number;
  teamROIData: TeamROIDatum[];
}

export function usePoolPortfoliosData(dashboardData: PoolDashboard | undefined): PoolPortfoliosData {
  const { portfolios, totalPortfolios, allOwnershipSummaries, allOwnershipDetails, allInvestments, seedInvestmentData } =
    useMemo(() => {
      if (!dashboardData) {
        return {
          portfolios: [] as EnrichedPortfolio[],
          totalPortfolios: 0,
          allOwnershipSummaries: [] as (OwnershipSummary & { portfolioName?: string })[],
          allOwnershipDetails: [] as OwnershipDetail[],
          allInvestments: [] as Investment[],
          seedInvestmentData: [] as SeedInvestmentDatum[],
        };
      }

      const { portfolios: rawPortfolios, investments, ownershipSummaries, ownershipDetails, schools } = dashboardData;
      const schoolMap = new Map(schools.map((s) => [s.id, s]));
      const portfolioNameMap = new Map(rawPortfolios.map((e) => [e.id, e.name]));

      // Enrich ownership summaries with portfolio names
      const enrichedOwnershipSummaries: (OwnershipSummary & { portfolioName?: string })[] = ownershipSummaries.map((p) => ({
        ...p,
        portfolioName: portfolioNameMap.get(p.portfolioId),
      }));

      // Compute returns per portfolio from ownership details
      const portfolioReturnsMap = new Map<string, number>();
      const ownershipToPortfolio = new Map(ownershipSummaries.map((p) => [p.id, p.portfolioId]));
      for (const pt of ownershipDetails) {
        const portfolioId = ownershipToPortfolio.get(pt.portfolioId);
        if (portfolioId) {
          portfolioReturnsMap.set(portfolioId, (portfolioReturnsMap.get(portfolioId) || 0) + pt.actualReturns);
        }
      }

      // Enrich ownership details with school info
      const enrichedOwnershipDetails: OwnershipDetail[] = ownershipDetails.map((pt) => {
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

      // Sort portfolios by returns
      const sortedPortfolios: EnrichedPortfolio[] = rawPortfolios
        .map((portfolio) => ({
          ...portfolio,
          totalReturns: portfolio.totalReturns || portfolioReturnsMap.get(portfolio.id) || 0,
        }))
        .sort((a, b) => {
          const diff = (b.totalReturns || 0) - (a.totalReturns || 0);
          if (diff !== 0) return diff;
          return new Date(b.createdAt).getTime() - new Date(a.createdAt).getTime();
        });

      // Compute seed investment data
      const seedMap = new Map<number, number>();
      for (const team of investments) {
        if (!team.team?.seed || !team.credits) continue;
        const seed = team.team.seed;
        seedMap.set(seed, (seedMap.get(seed) || 0) + team.credits);
      }
      const seedData = Array.from(seedMap.entries())
        .map(([seed, totalInvestment]) => ({ seed, totalInvestment }))
        .sort((a, b) => a.seed - b.seed);

      return {
        portfolios: sortedPortfolios,
        totalPortfolios: rawPortfolios.length,
        allOwnershipSummaries: enrichedOwnershipSummaries,
        allOwnershipDetails: enrichedOwnershipDetails,
        allInvestments: investments,
        seedInvestmentData: seedData,
      };
    }, [dashboardData]);

  const schools = useMemo(() => dashboardData?.schools ?? [], [dashboardData?.schools]);
  const tournamentTeams = useMemo(() => dashboardData?.tournamentTeams ?? [], [dashboardData?.tournamentTeams]);

  const totalInvestment = useMemo(() => allInvestments.reduce((sum, et) => sum + et.credits, 0), [allInvestments]);

  const totalReturns = useMemo(() => portfolios.reduce((sum, e) => sum + (e.totalReturns || 0), 0), [portfolios]);

  const averageReturn = useMemo(
    () => (totalPortfolios > 0 ? totalReturns / totalPortfolios : 0),
    [totalPortfolios, totalReturns],
  );

  const returnsStdDev = useMemo(() => {
    if (totalPortfolios <= 1) return 0;

    const mean = averageReturn;
    const variance = portfolios.reduce((acc, e) => {
      const v = (e.totalReturns || 0) - mean;
      return acc + v * v;
    }, 0);
    return Math.sqrt(variance / totalPortfolios);
  }, [portfolios, totalPortfolios, averageReturn]);

  const schoolNameById = useMemo(() => new Map(schools.map((school) => [school.id, school.name])), [schools]);

  const teamROIData = useMemo(() => {
    const roiTeams = tournamentTeams.map((team) => {
      const schoolName = schoolNameById.get(team.schoolId) || 'Unknown School';

      // Calculate total investment for this team
      const teamInvestment = allInvestments.filter((et) => et.teamId === team.id).reduce((sum, et) => sum + et.credits, 0);

      // Calculate total points for this team
      const teamPoints = allOwnershipDetails
        .filter((pt) => pt.teamId === team.id)
        .reduce((sum, pt) => sum + pt.actualReturns, 0);

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
  }, [tournamentTeams, schoolNameById, allInvestments, allOwnershipDetails]);

  return {
    portfolios,
    totalPortfolios,
    allOwnershipSummaries,
    allOwnershipDetails,
    allInvestments,
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
