export interface BracketTeam {
  teamId: string;
  schoolId: string;
  name: string;
  seed: number;
  region: string;
}

export interface BracketGame {
  gameId: string;
  round: string;
  region: string;
  team1?: BracketTeam;
  team2?: BracketTeam;
  winner?: BracketTeam;
  nextGameId?: string;
  nextGameSlot?: number;
  sortOrder: number;
  canSelect: boolean;
}

export interface BracketStructure {
  tournamentId: string;
  regions: string[];
  games: BracketGame[];
}

export type BracketRound = 
  | 'first_four'
  | 'round_of_64'
  | 'round_of_32'
  | 'sweet_16'
  | 'elite_8'
  | 'final_four'
  | 'championship';

export const ROUND_LABELS: Record<BracketRound, string> = {
  first_four: 'First Four',
  round_of_64: 'First Round',
  round_of_32: 'Second Round',
  sweet_16: 'Sweet 16',
  elite_8: 'Elite 8',
  final_four: 'Final Four',
  championship: 'Championship',
};

export const ROUND_ORDER: BracketRound[] = [
  'first_four',
  'round_of_64',
  'round_of_32',
  'sweet_16',
  'elite_8',
  'final_four',
  'championship',
];
