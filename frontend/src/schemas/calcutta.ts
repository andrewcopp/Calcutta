import { z } from 'zod';
import { SchoolSchema } from './school';
import { TournamentTeamSchema } from './tournament';

export const CalcuttaAbilitiesSchema = z.object({
  canEditSettings: z.boolean(),
  canInviteUsers: z.boolean(),
  canEditEntries: z.boolean(),
  canManageCoManagers: z.boolean(),
});

export type CalcuttaAbilities = z.infer<typeof CalcuttaAbilitiesSchema>;

export const CalcuttaSchema = z.object({
  id: z.string(),
  name: z.string(),
  tournamentId: z.string(),
  ownerId: z.string(),
  minTeams: z.number(),
  maxTeams: z.number(),
  maxBidPoints: z.number(),
  budgetPoints: z.number(),
  abilities: CalcuttaAbilitiesSchema.optional(),
  createdAt: z.string(),
  updatedAt: z.string(),
});

export type Calcutta = z.infer<typeof CalcuttaSchema>;

export const CalcuttaEntrySchema = z.object({
  id: z.string(),
  name: z.string(),
  calcuttaId: z.string(),
  status: z.enum(['incomplete', 'accepted']).optional(),
  totalPoints: z.number().optional(),
  finishPosition: z.number().optional(),
  inTheMoney: z.boolean().optional(),
  payoutCents: z.number().optional(),
  createdAt: z.string(),
  updatedAt: z.string(),
});

export type CalcuttaEntry = z.infer<typeof CalcuttaEntrySchema>;

const EntryTeamNestedTeamSchema = z.object({
  id: z.string(),
  schoolId: z.string(),
  seed: z.number().optional(),
  region: z.string().optional(),
  byes: z.number().optional(),
  wins: z.number().optional(),
  isEliminated: z.boolean().optional(),
  school: SchoolSchema.optional(),
});

export const CalcuttaEntryTeamSchema = z.object({
  id: z.string(),
  entryId: z.string(),
  teamId: z.string(),
  bid: z.number(),
  createdAt: z.string(),
  updatedAt: z.string(),
  team: EntryTeamNestedTeamSchema.optional(),
});

export type CalcuttaEntryTeam = z.infer<typeof CalcuttaEntryTeamSchema>;

export const CalcuttaPortfolioSchema = z.object({
  id: z.string(),
  entryId: z.string(),
  maximumPoints: z.number(),
  createdAt: z.string(),
  updatedAt: z.string(),
});

export type CalcuttaPortfolio = z.infer<typeof CalcuttaPortfolioSchema>;

const PortfolioTeamNestedTeamSchema = z.object({
  id: z.string(),
  schoolId: z.string(),
  school: SchoolSchema.optional(),
});

export const CalcuttaPortfolioTeamSchema = z.object({
  id: z.string(),
  portfolioId: z.string(),
  teamId: z.string(),
  ownershipPercentage: z.number(),
  actualPoints: z.number(),
  expectedPoints: z.number(),
  createdAt: z.string(),
  updatedAt: z.string(),
  team: PortfolioTeamNestedTeamSchema.optional(),
});

export type CalcuttaPortfolioTeam = z.infer<typeof CalcuttaPortfolioTeamSchema>;

export const CalcuttaDashboardSchema = z.object({
  calcutta: CalcuttaSchema,
  tournamentStartingAt: z.string().optional(),
  biddingOpen: z.boolean(),
  totalEntries: z.number(),
  currentUserEntry: CalcuttaEntrySchema.optional(),
  abilities: CalcuttaAbilitiesSchema.optional(),
  entries: z.array(CalcuttaEntrySchema),
  entryTeams: z.array(CalcuttaEntryTeamSchema),
  portfolios: z.array(CalcuttaPortfolioSchema),
  portfolioTeams: z.array(CalcuttaPortfolioTeamSchema),
  schools: z.array(SchoolSchema),
  tournamentTeams: z.array(TournamentTeamSchema),
});

export type CalcuttaDashboard = z.infer<typeof CalcuttaDashboardSchema>;

export const CalcuttaRankingSchema = z.object({
  rank: z.number(),
  totalEntries: z.number(),
  points: z.number(),
});

export type CalcuttaRanking = z.infer<typeof CalcuttaRankingSchema>;

export const CalcuttaWithRankingSchema = CalcuttaSchema.extend({
  hasEntry: z.boolean(),
  tournamentStartingAt: z.string().optional(),
  ranking: CalcuttaRankingSchema.optional(),
});

export type CalcuttaWithRanking = z.infer<typeof CalcuttaWithRankingSchema>;

export const PayoutsResponseSchema = z.object({
  payouts: z.array(
    z.object({
      position: z.number(),
      amountCents: z.number(),
    }),
  ),
});

export type PayoutsResponse = z.infer<typeof PayoutsResponseSchema>;
