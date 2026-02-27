import { z } from 'zod';
import { SchoolSchema } from './school';

export const TournamentSchema = z.object({
  id: z.string(),
  name: z.string(),
  rounds: z.number(),
  winner: z.string().optional(),
  finalFourTopLeft: z.string().optional(),
  finalFourBottomLeft: z.string().optional(),
  finalFourTopRight: z.string().optional(),
  finalFourBottomRight: z.string().optional(),
  startingAt: z.string().optional(),
  createdAt: z.string(),
  updatedAt: z.string(),
});

export type Tournament = z.infer<typeof TournamentSchema>;

export const KenPomStatsSchema = z.object({
  netRtg: z.number(),
  oRtg: z.number(),
  dRtg: z.number(),
  adjT: z.number(),
});

export type KenPomStats = z.infer<typeof KenPomStatsSchema>;

export const TournamentTeamSchema = z.object({
  id: z.string(),
  schoolId: z.string(),
  tournamentId: z.string(),
  seed: z.number(),
  region: z.string(),
  byes: z.number(),
  wins: z.number(),
  isEliminated: z.boolean(),
  createdAt: z.string(),
  updatedAt: z.string(),
  school: SchoolSchema.optional(),
  kenPom: KenPomStatsSchema.optional(),
});

export type TournamentTeam = z.infer<typeof TournamentTeamSchema>;

export const CompetitionSchema = z.object({
  id: z.string(),
  name: z.string(),
});

export type Competition = z.infer<typeof CompetitionSchema>;

export const SeasonSchema = z.object({
  id: z.string(),
  year: z.number(),
});

export type Season = z.infer<typeof SeasonSchema>;

export const PredictionBatchSchema = z.object({
  id: z.string(),
  probabilitySourceKey: z.string(),
  throughRound: z.number(),
  createdAt: z.string(),
});

export type PredictionBatch = z.infer<typeof PredictionBatchSchema>;

export const TeamPredictionSchema = z.object({
  teamId: z.string(),
  schoolName: z.string(),
  seed: z.number(),
  region: z.string(),
  wins: z.number(),
  byes: z.number(),
  isEliminated: z.boolean(),
  pRound1: z.number(),
  pRound2: z.number(),
  pRound3: z.number(),
  pRound4: z.number(),
  pRound5: z.number(),
  pRound6: z.number(),
  pRound7: z.number(),
  expectedPoints: z.number(),
});

export type TeamPrediction = z.infer<typeof TeamPredictionSchema>;

export const TournamentPredictionsSchema = z.object({
  tournamentId: z.string(),
  batchId: z.string(),
  throughRound: z.number(),
  teams: z.array(TeamPredictionSchema),
});

export type TournamentPredictions = z.infer<typeof TournamentPredictionsSchema>;

export const TournamentModeratorSchema = z.object({
  id: z.string(),
  email: z.string(),
  firstName: z.string(),
  lastName: z.string(),
});

export type TournamentModerator = z.infer<typeof TournamentModeratorSchema>;

export const TournamentModeratorsResponseSchema = z.object({
  items: z.array(TournamentModeratorSchema),
});

