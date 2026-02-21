import { useCallback, useMemo } from 'react';
import { CalcuttaEntryTeam, CalcuttaPortfolio, CalcuttaPortfolioTeam } from '../types/calcutta';
import type { TournamentTeam } from '../types/tournament';

interface InvestorRanking {
  rank: number;
  total: number;
}

export interface EntryOwnershipData {
  getPortfolioTeamData: (teamId: string) => CalcuttaPortfolioTeam | undefined;
  getInvestorRanking: (teamId: string) => InvestorRanking;
  ownershipTeamsData: CalcuttaEntryTeam[];
}

export function useEntryOwnershipData({
  activeTab,
  entryId,
  teams,
  schools,
  tournamentTeams,
  portfolios,
  allCalcuttaPortfolioTeams,
  ownershipShowAllTeams,
  sortBy,
}: {
  activeTab: string;
  entryId: string | undefined;
  teams: CalcuttaEntryTeam[];
  schools: { id: string; name: string }[];
  tournamentTeams: TournamentTeam[];
  portfolios: CalcuttaPortfolio[];
  allCalcuttaPortfolioTeams: CalcuttaPortfolioTeam[];
  ownershipShowAllTeams: boolean;
  sortBy: 'points' | 'ownership' | 'bid';
}): EntryOwnershipData {
  const getPortfolioTeamData = useCallback(
    (teamId: string) => {
      const currentPortfolioId = portfolios[0]?.id;
      if (!currentPortfolioId) return undefined;
      return allCalcuttaPortfolioTeams.find((pt) => pt.teamId === teamId && pt.portfolioId === currentPortfolioId);
    },
    [allCalcuttaPortfolioTeams, portfolios]
  );

  const getInvestorRanking = useCallback(
    (teamId: string): InvestorRanking => {
      const allInvestors = allCalcuttaPortfolioTeams.filter((pt) => pt.teamId === teamId);
      const sortedInvestors = [...allInvestors].sort((a, b) => b.ownershipPercentage - a.ownershipPercentage);

      const userPortfolio = portfolios[0];
      const userRank = userPortfolio ? sortedInvestors.findIndex((pt) => pt.portfolioId === userPortfolio.id) + 1 : 0;

      return {
        rank: userRank,
        total: allInvestors.length,
      };
    },
    [allCalcuttaPortfolioTeams, portfolios]
  );

  const ownershipTeamsData = useMemo(() => {
    if (activeTab !== 'ownerships') return [];
    if (!entryId) return [];

    let teamsToShow: CalcuttaEntryTeam[];

    if (ownershipShowAllTeams) {
      const schoolMap = new Map(schools.map((s) => [s.id, s]));
      teamsToShow = tournamentTeams.map((tt) => {
        const existingTeam = teams.find((t) => t.teamId === tt.id);
        if (existingTeam) return existingTeam;

        return {
          id: `synthetic-${tt.id}`,
          entryId: entryId,
          teamId: tt.id,
          bid: 0,
          createdAt: new Date().toISOString(),
          updatedAt: new Date().toISOString(),
          team: {
            ...tt,
            school: schoolMap.get(tt.schoolId),
          },
        } as CalcuttaEntryTeam;
      });
    } else {
      teamsToShow = teams.filter((team) => {
        const portfolioTeam = getPortfolioTeamData(team.teamId);
        return portfolioTeam && portfolioTeam.ownershipPercentage > 0;
      });
    }

    teamsToShow = teamsToShow.slice().sort((a, b) => {
      const portfolioTeamA = getPortfolioTeamData(a.teamId);
      const portfolioTeamB = getPortfolioTeamData(b.teamId);

      const pointsA = portfolioTeamA?.actualPoints || 0;
      const pointsB = portfolioTeamB?.actualPoints || 0;
      const ownershipA = portfolioTeamA?.ownershipPercentage || 0;
      const ownershipB = portfolioTeamB?.ownershipPercentage || 0;
      const bidA = a.bid;
      const bidB = b.bid;

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

      if (bidB !== bidA) return bidB - bidA;
      if (pointsB !== pointsA) return pointsB - pointsA;
      return ownershipB - ownershipA;
    });

    return teamsToShow;
  }, [activeTab, entryId, getPortfolioTeamData, ownershipShowAllTeams, schools, sortBy, teams, tournamentTeams]);

  return {
    getPortfolioTeamData,
    getInvestorRanking,
    ownershipTeamsData,
  };
}
