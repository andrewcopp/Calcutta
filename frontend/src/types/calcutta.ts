export interface Calcutta {
  id: string;
  name: string;
  tournamentId: string;
  ownerId: string;
  created: string;
  updated: string;
}

export interface CalcuttaEntry {
  id: string;
  name: string;
  calcuttaId: string;
  userId?: string;
  totalPoints?: number;
  finishPosition?: number;
  inTheMoney?: boolean;
  payoutCents?: number;
  isTied?: boolean;
  created: string;
  updated: string;
}

export interface School {
  id: string;
  name: string;
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
  deleted?: string;
}

export interface CalcuttaPortfolioTeam {
  id: string;
  portfolioId: string;
  teamId: string;
  ownershipPercentage: number;
  actualPoints: number;
  expectedPoints: number;
  predictedPoints: number;
  created: string;
  updated: string;
  deleted?: string;
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
} 