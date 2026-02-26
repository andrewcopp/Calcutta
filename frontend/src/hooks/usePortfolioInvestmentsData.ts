import { useMemo } from 'react';
import { Investment, OwnershipSummary, OwnershipDetail, PoolDashboard } from '../schemas/pool';
import type { TournamentTeam } from '../schemas/tournament';

export interface PortfolioInvestmentsData {
  poolName: string;
  portfolioName: string;
  teams: Investment[];
  schools: { id: string; name: string }[];
  ownershipSummaries: OwnershipSummary[];
  ownershipDetails: OwnershipDetail[];
  tournamentTeams: TournamentTeam[];
  allInvestments: Investment[];
  allOwnershipSummaries: (OwnershipSummary & { portfolioName?: string })[];
  allOwnershipDetails: OwnershipDetail[];
}

export function usePortfolioInvestmentsData(
  dashboardData: PoolDashboard | undefined,
  portfolioId: string | undefined,
): PortfolioInvestmentsData {
  return useMemo(() => {
    if (!dashboardData || !portfolioId) {
      return {
        poolName: '',
        portfolioName: '',
        teams: [] as Investment[],
        schools: [] as { id: string; name: string }[],
        ownershipSummaries: [] as OwnershipSummary[],
        ownershipDetails: [] as OwnershipDetail[],
        tournamentTeams: dashboardData?.tournamentTeams ?? [],
        allInvestments: [] as Investment[],
        allOwnershipSummaries: [] as (OwnershipSummary & { portfolioName?: string })[],
        allOwnershipDetails: [] as OwnershipDetail[],
      };
    }

    const { pool, portfolios, investments, ownershipSummaries, ownershipDetails, schools, tournamentTeams } = dashboardData;
    const schoolMap = new Map(schools.map((s) => [s.id, s]));
    const portfolioNameMap = new Map(portfolios.map((e) => [e.id, e.name]));

    const currentPortfolio = portfolios.find((e) => e.id === portfolioId);
    const portfolioName = currentPortfolio?.name || '';

    // Filter investments for this portfolio
    const thisPortfolioInvestments = investments.filter((et) => et.portfolioId === portfolioId);

    // Enrich investments with school info
    const teamsWithSchools: Investment[] = thisPortfolioInvestments.map((team) => ({
      ...team,
      team: team.team
        ? {
            ...team.team,
            school: schoolMap.get(team.team.schoolId),
          }
        : undefined,
    }));

    // Filter ownership summaries for this portfolio
    const thisPortfolioOwnershipSummaries = ownershipSummaries.filter((p) => p.portfolioId === portfolioId);

    // Filter ownership details for this portfolio's ownership summaries
    const thisOwnershipSummaryIds = new Set(thisPortfolioOwnershipSummaries.map((p) => p.id));
    const thisOwnershipDetails: OwnershipDetail[] = ownershipDetails
      .filter((pt) => thisOwnershipSummaryIds.has(pt.ownershipSummaryId))
      .map((pt) => ({
        ...pt,
        team: pt.team
          ? {
              ...pt.team,
              school: schoolMap.get(pt.team.schoolId),
            }
          : undefined,
      }));

    // Enrich all ownership summaries with portfolio names
    const allSummariesWithNames: (OwnershipSummary & { portfolioName?: string })[] = ownershipSummaries.map((p) => ({
      ...p,
      portfolioName: portfolioNameMap.get(p.portfolioId),
    }));

    // Enrich all ownership details with school info
    const allDetailsWithSchools: OwnershipDetail[] = ownershipDetails.map((pt) => ({
      ...pt,
      team: pt.team
        ? {
            ...pt.team,
            school: schoolMap.get(pt.team.schoolId),
          }
        : undefined,
    }));

    return {
      poolName: pool.name,
      portfolioName,
      teams: teamsWithSchools,
      schools,
      ownershipSummaries: thisPortfolioOwnershipSummaries,
      ownershipDetails: thisOwnershipDetails,
      tournamentTeams,
      allInvestments: investments,
      allOwnershipSummaries: allSummariesWithNames,
      allOwnershipDetails: allDetailsWithSchools,
    };
  }, [dashboardData, portfolioId]);
}
