import { useCallback, useMemo } from 'react';
import { Investment, OwnershipSummary, OwnershipDetail } from '../schemas/pool';
import type { TournamentTeam } from '../schemas/tournament';

interface InvestorRanking {
  rank: number;
  total: number;
}

export interface PortfolioOwnershipData {
  getOwnershipDetailData: (teamId: string) => OwnershipDetail | undefined;
  getInvestorRanking: (teamId: string) => InvestorRanking;
  ownershipTeamsData: Investment[];
}

export function usePortfolioOwnershipData({
  activeTab,
  portfolioId,
  teams,
  schools,
  tournamentTeams,
  ownershipSummaries,
  allOwnershipDetails,
  ownershipShowAllTeams,
  sortBy,
}: {
  activeTab: string;
  portfolioId: string | undefined;
  teams: Investment[];
  schools: { id: string; name: string }[];
  tournamentTeams: TournamentTeam[];
  ownershipSummaries: OwnershipSummary[];
  allOwnershipDetails: OwnershipDetail[];
  ownershipShowAllTeams: boolean;
  sortBy: 'points' | 'ownership' | 'credits';
}): PortfolioOwnershipData {
  const getOwnershipDetailData = useCallback(
    (teamId: string) => {
      const currentOwnershipSummaryId = ownershipSummaries[0]?.id;
      if (!currentOwnershipSummaryId) return undefined;
      return allOwnershipDetails.find((pt) => pt.teamId === teamId && pt.ownershipSummaryId === currentOwnershipSummaryId);
    },
    [allOwnershipDetails, ownershipSummaries],
  );

  const getInvestorRanking = useCallback(
    (teamId: string): InvestorRanking => {
      const allInvestors = allOwnershipDetails.filter((pt) => pt.teamId === teamId);
      const sortedInvestors = [...allInvestors].sort((a, b) => b.ownershipPercentage - a.ownershipPercentage);

      const userOwnershipSummary = ownershipSummaries[0];
      const userRank = userOwnershipSummary ? sortedInvestors.findIndex((pt) => pt.ownershipSummaryId === userOwnershipSummary.id) + 1 : 0;

      return {
        rank: userRank,
        total: allInvestors.length,
      };
    },
    [allOwnershipDetails, ownershipSummaries],
  );

  const ownershipTeamsData = useMemo(() => {
    if (activeTab !== 'ownerships') return [];
    if (!portfolioId) return [];

    let teamsToShow: Investment[];

    if (ownershipShowAllTeams) {
      const schoolMap = new Map(schools.map((s) => [s.id, s]));
      teamsToShow = tournamentTeams.map((tt) => {
        const existingTeam = teams.find((t) => t.teamId === tt.id);
        if (existingTeam) return existingTeam;

        return {
          id: `synthetic-${tt.id}`,
          portfolioId: portfolioId,
          teamId: tt.id,
          credits: 0,
          createdAt: new Date().toISOString(),
          updatedAt: new Date().toISOString(),
          team: {
            ...tt,
            school: schoolMap.get(tt.schoolId),
          },
        } as Investment;
      });
    } else {
      teamsToShow = teams.filter((team) => {
        const ownershipDetail = getOwnershipDetailData(team.teamId);
        return ownershipDetail && ownershipDetail.ownershipPercentage > 0;
      });
    }

    teamsToShow = teamsToShow.slice().sort((a, b) => {
      const detailA = getOwnershipDetailData(a.teamId);
      const detailB = getOwnershipDetailData(b.teamId);

      const pointsA = detailA?.actualReturns || 0;
      const pointsB = detailB?.actualReturns || 0;
      const ownershipA = detailA?.ownershipPercentage || 0;
      const ownershipB = detailB?.ownershipPercentage || 0;
      const creditsA = a.credits;
      const creditsB = b.credits;

      if (sortBy === 'points') {
        if (pointsB !== pointsA) return pointsB - pointsA;
        if (ownershipB !== ownershipA) return ownershipB - ownershipA;
        return creditsB - creditsA;
      }

      if (sortBy === 'ownership') {
        if (ownershipB !== ownershipA) return ownershipB - ownershipA;
        if (pointsB !== pointsA) return pointsB - pointsA;
        return creditsB - creditsA;
      }

      if (creditsB !== creditsA) return creditsB - creditsA;
      if (pointsB !== pointsA) return pointsB - pointsA;
      return ownershipB - ownershipA;
    });

    return teamsToShow;
  }, [activeTab, portfolioId, getOwnershipDetailData, ownershipShowAllTeams, schools, sortBy, teams, tournamentTeams]);

  return {
    getOwnershipDetailData,
    getInvestorRanking,
    ownershipTeamsData,
  };
}
