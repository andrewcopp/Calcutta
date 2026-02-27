import { z } from 'zod';
import { SchoolSchema } from './school';
import { TournamentTeamSchema } from './tournament';

export const PoolAbilitiesSchema = z.object({
  canEditSettings: z.boolean(),
  canInviteUsers: z.boolean(),
  canEditEntries: z.boolean(),
  canManageCoManagers: z.boolean(),
});

export type PoolAbilities = z.infer<typeof PoolAbilitiesSchema>;

export const PoolSchema = z.object({
  id: z.string(),
  name: z.string(),
  tournamentId: z.string(),
  ownerId: z.string(),
  minTeams: z.number(),
  maxTeams: z.number(),
  maxInvestmentCredits: z.number(),
  budgetCredits: z.number(),
  abilities: PoolAbilitiesSchema.optional(),
  createdAt: z.string(),
  updatedAt: z.string(),
});

export type Pool = z.infer<typeof PoolSchema>;

export const PortfolioSchema = z.object({
  id: z.string(),
  name: z.string(),
  poolId: z.string(),
  status: z.enum(['draft', 'submitted']).optional(),
  totalReturns: z.number().optional(),
  finishPosition: z.number().optional(),
  inTheMoney: z.boolean().optional(),
  payoutCents: z.number().optional(),
  expectedValue: z.number().optional(),
  projectedFavorites: z.number().optional(),
  createdAt: z.string(),
  updatedAt: z.string(),
});

export type Portfolio = z.infer<typeof PortfolioSchema>;

const InvestmentNestedTeamSchema = z.object({
  id: z.string(),
  schoolId: z.string(),
  seed: z.number().optional(),
  region: z.string().optional(),
  byes: z.number().optional(),
  wins: z.number().optional(),
  isEliminated: z.boolean().optional(),
  school: SchoolSchema.optional(),
});

export const InvestmentSchema = z.object({
  id: z.string(),
  portfolioId: z.string(),
  teamId: z.string(),
  credits: z.number(),
  createdAt: z.string(),
  updatedAt: z.string(),
  team: InvestmentNestedTeamSchema.optional(),
});

export type Investment = z.infer<typeof InvestmentSchema>;

export const OwnershipSummarySchema = z.object({
  id: z.string(),
  portfolioId: z.string(),
  maximumReturns: z.number(),
  createdAt: z.string(),
  updatedAt: z.string(),
});

export type OwnershipSummary = z.infer<typeof OwnershipSummarySchema>;

const OwnershipDetailNestedTeamSchema = z.object({
  id: z.string(),
  schoolId: z.string(),
  school: SchoolSchema.optional(),
});

export const OwnershipDetailSchema = z.object({
  id: z.string(),
  ownershipSummaryId: z.string(),
  teamId: z.string(),
  ownershipPercentage: z.number(),
  actualReturns: z.number(),
  expectedReturns: z.number(),
  createdAt: z.string(),
  updatedAt: z.string(),
  team: OwnershipDetailNestedTeamSchema.optional(),
});

export type OwnershipDetail = z.infer<typeof OwnershipDetailSchema>;

export const RoundStandingPortfolioSchema = z.object({
  portfolioId: z.string(),
  totalReturns: z.number(),
  finishPosition: z.number(),
  isTied: z.boolean(),
  payoutCents: z.number(),
  inTheMoney: z.boolean(),
  expectedValue: z.number().optional(),
  projectedFavorites: z.number().optional(),
});

export type RoundStandingPortfolio = z.infer<typeof RoundStandingPortfolioSchema>;

export const RoundStandingGroupSchema = z.object({
  round: z.number(),
  entries: z.array(RoundStandingPortfolioSchema),
});

export type RoundStandingGroup = z.infer<typeof RoundStandingGroupSchema>;

export const FinalFourTeamSchema = z.object({
  teamId: z.string(),
  schoolId: z.string(),
  seed: z.number(),
  region: z.string(),
});

export type FinalFourTeam = z.infer<typeof FinalFourTeamSchema>;

export const FinalFourOutcomeSchema = z.object({
  semifinal1Winner: FinalFourTeamSchema,
  semifinal2Winner: FinalFourTeamSchema,
  champion: FinalFourTeamSchema,
  runnerUp: FinalFourTeamSchema,
  entries: z.array(RoundStandingPortfolioSchema),
});

export type FinalFourOutcome = z.infer<typeof FinalFourOutcomeSchema>;

export const ScoringRuleSchema = z.object({
  winIndex: z.number(),
  pointsAwarded: z.number(),
});

export type ScoringRule = z.infer<typeof ScoringRuleSchema>;

export const PoolDashboardSchema = z.object({
  pool: PoolSchema,
  tournamentStartingAt: z.string().optional(),
  investingOpen: z.boolean(),
  totalPortfolios: z.number(),
  currentUserPortfolio: PortfolioSchema.optional(),
  abilities: PoolAbilitiesSchema.optional(),
  scoringRules: z.array(ScoringRuleSchema),
  portfolios: z.array(PortfolioSchema),
  investments: z.array(InvestmentSchema),
  ownershipSummaries: z.array(OwnershipSummarySchema),
  ownershipDetails: z.array(OwnershipDetailSchema),
  schools: z.array(SchoolSchema),
  tournamentTeams: z.array(TournamentTeamSchema),
  roundStandings: z.array(RoundStandingGroupSchema),
  finalFourOutcomes: z.array(FinalFourOutcomeSchema).optional(),
});

export type PoolDashboard = z.infer<typeof PoolDashboardSchema>;

export const PoolRankingSchema = z.object({
  rank: z.number(),
  totalPortfolios: z.number(),
  points: z.number(),
});

export type PoolRanking = z.infer<typeof PoolRankingSchema>;

export const PoolWithRankingSchema = PoolSchema.extend({
  hasPortfolio: z.boolean(),
  tournamentStartingAt: z.string().optional(),
  ranking: PoolRankingSchema.optional(),
});

export type PoolWithRanking = z.infer<typeof PoolWithRankingSchema>;

export const PayoutsResponseSchema = z.object({
  items: z.array(
    z.object({
      position: z.number(),
      amountCents: z.number(),
    }),
  ),
});

export type PayoutsResponse = z.infer<typeof PayoutsResponseSchema>;
