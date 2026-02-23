import { z } from 'zod';

export const InvestmentLeaderboardRowSchema = z.object({
  tournamentName: z.string(),
  tournamentYear: z.number(),
  calcuttaId: z.string(),
  entryId: z.string(),
  entryName: z.string(),
  teamId: z.string(),
  schoolName: z.string(),
  seed: z.number(),
  investment: z.number(),
  ownershipPercentage: z.number(),
  rawReturns: z.number(),
  normalizedReturns: z.number(),
});

export type InvestmentLeaderboardRow = z.infer<typeof InvestmentLeaderboardRowSchema>;

export const InvestmentLeaderboardResponseSchema = z.object({
  investments: z.array(InvestmentLeaderboardRowSchema),
});

export type InvestmentLeaderboardResponse = z.infer<typeof InvestmentLeaderboardResponseSchema>;

export const BestTeamSchema = z.object({
  tournamentName: z.string(),
  tournamentYear: z.number(),
  calcuttaId: z.string(),
  teamId: z.string(),
  schoolName: z.string(),
  seed: z.number(),
  region: z.string(),
  teamPoints: z.number(),
  totalBid: z.number(),
  calcuttaTotalBid: z.number(),
  calcuttaTotalPoints: z.number(),
  investmentShare: z.number(),
  pointsShare: z.number(),
  rawROI: z.number(),
  normalizedROI: z.number(),
});

export type BestTeam = z.infer<typeof BestTeamSchema>;

export const BestTeamsResponseSchema = z.object({
  teams: z.array(BestTeamSchema),
});

export type BestTeamsResponse = z.infer<typeof BestTeamsResponseSchema>;

export const EntryLeaderboardRowSchema = z.object({
  tournamentName: z.string(),
  tournamentYear: z.number(),
  calcuttaId: z.string(),
  entryId: z.string(),
  entryName: z.string(),
  totalReturns: z.number(),
  totalParticipants: z.number(),
  averageReturns: z.number(),
  normalizedReturns: z.number(),
});

export type EntryLeaderboardRow = z.infer<typeof EntryLeaderboardRowSchema>;

export const EntryLeaderboardResponseSchema = z.object({
  entries: z.array(EntryLeaderboardRowSchema),
});

export type EntryLeaderboardResponse = z.infer<typeof EntryLeaderboardResponseSchema>;

export const CareerLeaderboardRowSchema = z.object({
  entryName: z.string(),
  wins: z.number(),
  years: z.number(),
  bestFinish: z.number(),
  podiums: z.number(),
  inTheMoneys: z.number(),
  top10s: z.number(),
  careerEarningsCents: z.number(),
  activeInLatestCalcutta: z.boolean(),
});

export type CareerLeaderboardRow = z.infer<typeof CareerLeaderboardRowSchema>;

export const CareerLeaderboardResponseSchema = z.object({
  careers: z.array(CareerLeaderboardRowSchema),
});

export type CareerLeaderboardResponse = z.infer<typeof CareerLeaderboardResponseSchema>;
