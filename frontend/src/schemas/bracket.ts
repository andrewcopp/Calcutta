import { z } from 'zod';

export const BracketTeamSchema = z.object({
  teamId: z.string(),
  schoolId: z.string(),
  name: z.string(),
  seed: z.number(),
  region: z.string(),
});

export type BracketTeam = z.infer<typeof BracketTeamSchema>;

export const BracketGameSchema = z.object({
  gameId: z.string(),
  round: z.string(),
  region: z.string(),
  team1: BracketTeamSchema.optional(),
  team2: BracketTeamSchema.optional(),
  winner: BracketTeamSchema.optional(),
  nextGameId: z.string().optional(),
  nextGameSlot: z.number().optional(),
  sortOrder: z.number(),
  canSelect: z.boolean(),
});

export type BracketGame = z.infer<typeof BracketGameSchema>;

export const BracketStructureSchema = z.object({
  tournamentId: z.string(),
  regions: z.array(z.string()),
  games: z.array(BracketGameSchema),
});

export type BracketStructure = z.infer<typeof BracketStructureSchema>;

export const BracketValidationSchema = z.object({
  valid: z.boolean(),
  errors: z.array(z.string()),
});

export type BracketValidation = z.infer<typeof BracketValidationSchema>;

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
