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
  createdAt: string;
  updatedAt: string;
}

export interface KenPomStats {
  netRtg: number;
  oRtg: number;
  dRtg: number;
  adjT: number;
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
  createdAt: string;
  updatedAt: string;
  school?: { id: string; name: string };
  kenPom?: KenPomStats;
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
