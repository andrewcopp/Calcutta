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
