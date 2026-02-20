export interface Calcutta {
  id: string;
  name: string;
  tournamentId: string;
  ownerId: string;
  minTeams: number;
  maxTeams: number;
  maxBid: number;
  budgetPoints: number;
  biddingOpen: boolean;
  biddingLockedAt?: string;
  created: string;
  updated: string;
}

export interface CalcuttaEntry {
  id: string;
  name: string;
  calcuttaId: string;
  totalPoints?: number;
  finishPosition?: number;
  inTheMoney?: boolean;
  payoutCents?: number;
  created: string;
  updated: string;
}

export interface CalcuttaEntryTeam {
  id: string;
  entryId: string;
  teamId: string;
  bid: number;
  created: string;
  updated: string;
  team?: {
    id: string;
    schoolId: string;
    seed?: number;
    byes?: number;
    wins?: number;
    eliminated?: boolean;
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
  created: string;
  updated: string;
}

export interface CalcuttaPortfolioTeam {
  id: string;
  portfolioId: string;
  teamId: string;
  ownershipPercentage: number;
  actualPoints: number;
  expectedPoints: number;
  created: string;
  updated: string;
  team?: {
    id: string;
    schoolId: string;
    school?: {
      id: string;
      name: string;
    };
  };
}

export interface Tournament {
  id: string;
  name: string;
  rounds: number;
  winner?: string;
  finalFourTopLeft?: string;
  finalFourBottomLeft?: string;
  finalFourTopRight?: string;
  finalFourBottomRight?: string;
  startingAt?: string;
  created: string;
  updated: string;
}

export interface TournamentTeam {
  id: string;
  schoolId: string;
  tournamentId: string;
  seed: number;
  region: string;
  byes: number;
  wins: number;
  eliminated: boolean;
  created: string;
  updated: string;
  school?: { id: string; name: string };
}

export interface Competition {
  id: string;
  name: string;
}

export interface Season {
  id: string;
  year: number;
}

export interface TournamentModerator {
  id: string;
  email: string;
  firstName: string;
  lastName: string;
}

export interface CalcuttaDashboard {
  calcutta: Calcutta;
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