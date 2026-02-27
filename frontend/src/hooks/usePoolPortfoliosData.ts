import { useMemo } from 'react';
import {
  OwnershipSummary,
  OwnershipDetail,
  Investment,
  Portfolio,
  PoolDashboard,
} from '../schemas/pool';
import type { TournamentTeam } from '../schemas/tournament';

type EnrichedPortfolio = Portfolio & { totalReturns: number };

export interface PoolPortfoliosData {
  portfolios: EnrichedPortfolio[];
  allOwnershipSummaries: (OwnershipSummary & { portfolioName?: string })[];
  allOwnershipDetails: OwnershipDetail[];
  allInvestments: Investment[];
  schools: { id: string; name: string }[];
  tournamentTeams: TournamentTeam[];
}

export function usePoolPortfoliosData(dashboardData: PoolDashboard | undefined): PoolPortfoliosData {
  const { portfolios, allOwnershipSummaries, allOwnershipDetails, allInvestments } =
    useMemo(() => {
      if (!dashboardData) {
        return {
          portfolios: [] as EnrichedPortfolio[],
          allOwnershipSummaries: [] as (OwnershipSummary & { portfolioName?: string })[],
          allOwnershipDetails: [] as OwnershipDetail[],
          allInvestments: [] as Investment[],
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

      return {
        portfolios: sortedPortfolios,
        allOwnershipSummaries: enrichedOwnershipSummaries,
        allOwnershipDetails: enrichedOwnershipDetails,
        allInvestments: investments,
      };
    }, [dashboardData]);

  const schools = useMemo(() => dashboardData?.schools ?? [], [dashboardData?.schools]);
  const tournamentTeams = useMemo(() => dashboardData?.tournamentTeams ?? [], [dashboardData?.tournamentTeams]);

  return {
    portfolios,
    allOwnershipSummaries,
    allOwnershipDetails,
    allInvestments,
    schools,
    tournamentTeams,
  };
}
