import { useMemo } from 'react';
import { CalcuttaEntryTeam, CalcuttaPortfolio, CalcuttaPortfolioTeam, CalcuttaDashboard, TournamentTeam } from '../types/calcutta';

export interface EntryTeamsData {
  calcuttaName: string;
  entryName: string;
  teams: CalcuttaEntryTeam[];
  schools: { id: string; name: string }[];
  portfolios: CalcuttaPortfolio[];
  portfolioTeams: CalcuttaPortfolioTeam[];
  tournamentTeams: TournamentTeam[];
  allEntryTeams: CalcuttaEntryTeam[];
  allCalcuttaPortfolios: (CalcuttaPortfolio & { entryName?: string })[];
  allCalcuttaPortfolioTeams: CalcuttaPortfolioTeam[];
}

export function useEntryTeamsData(dashboardData: CalcuttaDashboard | undefined, entryId: string | undefined): EntryTeamsData {
  return useMemo(() => {
    if (!dashboardData || !entryId) {
      return {
        calcuttaName: '',
        entryName: '',
        teams: [] as CalcuttaEntryTeam[],
        schools: [] as { id: string; name: string }[],
        portfolios: [] as CalcuttaPortfolio[],
        portfolioTeams: [] as CalcuttaPortfolioTeam[],
        tournamentTeams: dashboardData?.tournamentTeams ?? [],
        allEntryTeams: [] as CalcuttaEntryTeam[],
        allCalcuttaPortfolios: [] as (CalcuttaPortfolio & { entryName?: string })[],
        allCalcuttaPortfolioTeams: [] as CalcuttaPortfolioTeam[],
      };
    }

    const { calcutta, entries, entryTeams, portfolios, portfolioTeams, schools, tournamentTeams } = dashboardData;
    const schoolMap = new Map(schools.map((s) => [s.id, s]));
    const entryNameMap = new Map(entries.map((e) => [e.id, e.name]));

    const currentEntry = entries.find((e) => e.id === entryId);
    const entryName = currentEntry?.name || '';

    // Filter entry teams for this entry
    const thisEntryTeams = entryTeams.filter((et) => et.entryId === entryId);

    // Enrich entry teams with school info
    const teamsWithSchools: CalcuttaEntryTeam[] = thisEntryTeams.map((team) => ({
      ...team,
      team: team.team
        ? {
            ...team.team,
            school: schoolMap.get(team.team.schoolId),
          }
        : undefined,
    }));

    // Filter portfolios for this entry
    const thisEntryPortfolios = portfolios.filter((p) => p.entryId === entryId);

    // Filter portfolio teams for this entry's portfolios
    const thisPortfolioIds = new Set(thisEntryPortfolios.map((p) => p.id));
    const thisPortfolioTeams: CalcuttaPortfolioTeam[] = portfolioTeams
      .filter((pt) => thisPortfolioIds.has(pt.portfolioId))
      .map((pt) => ({
        ...pt,
        team: pt.team
          ? {
              ...pt.team,
              school: schoolMap.get(pt.team.schoolId),
            }
          : undefined,
      }));

    // Enrich all portfolios with entry names
    const allPortfoliosWithNames: (CalcuttaPortfolio & { entryName?: string })[] = portfolios.map((p) => ({
      ...p,
      entryName: entryNameMap.get(p.entryId),
    }));

    // Enrich all portfolio teams with school info
    const allPortfolioTeamsWithSchools: CalcuttaPortfolioTeam[] = portfolioTeams.map((pt) => ({
      ...pt,
      team: pt.team
        ? {
            ...pt.team,
            school: schoolMap.get(pt.team.schoolId),
          }
        : undefined,
    }));

    return {
      calcuttaName: calcutta.name,
      entryName,
      teams: teamsWithSchools,
      schools,
      portfolios: thisEntryPortfolios,
      portfolioTeams: thisPortfolioTeams,
      tournamentTeams,
      allEntryTeams: entryTeams,
      allCalcuttaPortfolios: allPortfoliosWithNames,
      allCalcuttaPortfolioTeams: allPortfolioTeamsWithSchools,
    };
  }, [dashboardData, entryId]);
}
