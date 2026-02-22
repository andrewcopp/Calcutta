import type { TournamentTeam } from './tournament';

export interface Calcutta {
  id: string;
  name: string;
  tournamentId: string;
  ownerId: string;
  minTeams: number;
  maxTeams: number;
  maxBidPoints: number;
  budgetPoints: number;
  abilities?: CalcuttaAbilities;
  createdAt: string;
  updatedAt: string;
}

export interface CalcuttaEntry {
  id: string;
  name: string;
  calcuttaId: string;
  status?: 'incomplete' | 'accepted';
  totalPoints?: number;
  finishPosition?: number;
  inTheMoney?: boolean;
  payoutCents?: number;
  createdAt: string;
  updatedAt: string;
}

export interface CalcuttaEntryTeam {
  id: string;
  entryId: string;
  teamId: string;
  bid: number;
  createdAt: string;
  updatedAt: string;
  team?: {
    id: string;
    schoolId: string;
    seed?: number;
    region?: string;
    byes?: number;
    wins?: number;
    isEliminated?: boolean;
    school?: {
      id: string;
      name: string;
    };
  };
}

export interface CalcuttaPortfolio {
  id: string;
  entryId: string;
  maximumPoints: number;
  createdAt: string;
  updatedAt: string;
}

export interface CalcuttaPortfolioTeam {
  id: string;
  portfolioId: string;
  teamId: string;
  ownershipPercentage: number;
  actualPoints: number;
  expectedPoints: number;
  createdAt: string;
  updatedAt: string;
  team?: {
    id: string;
    schoolId: string;
    school?: {
      id: string;
      name: string;
    };
  };
}

export interface CalcuttaAbilities {
  canEditSettings: boolean;
  canInviteUsers: boolean;
  canEditEntries: boolean;
  canManageCoManagers: boolean;
}

export interface CalcuttaDashboard {
  calcutta: Calcutta;
  tournamentStartingAt?: string;
  biddingOpen: boolean;
  totalEntries: number;
  currentUserEntry?: CalcuttaEntry;
  abilities?: CalcuttaAbilities;
  entries: CalcuttaEntry[];
  entryTeams: CalcuttaEntryTeam[];
  portfolios: CalcuttaPortfolio[];
  portfolioTeams: CalcuttaPortfolioTeam[];
  schools: { id: string; name: string }[];
  tournamentTeams: TournamentTeam[];
}

export interface CalcuttaRanking {
  rank: number;
  totalEntries: number;
  points: number;
}

export interface CalcuttaWithRanking extends Calcutta {
  hasEntry: boolean;
  tournamentStartingAt?: string;
  ranking?: CalcuttaRanking;
}
