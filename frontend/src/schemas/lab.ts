import { z } from 'zod';

export type SortDir = 'asc' | 'desc';

export const InvestmentModelSchema = z.object({
  id: z.string(),
  name: z.string(),
  kind: z.string(),
  paramsJson: z.record(z.string(), z.unknown()),
  notes: z.string().nullable().optional(),
  createdAt: z.string(),
  updatedAt: z.string(),
  nEntries: z.number(),
  nEvaluations: z.number(),
});

export type InvestmentModel = z.infer<typeof InvestmentModelSchema>;

export const EnrichedPredictionSchema = z.object({
  teamId: z.string(),
  schoolName: z.string(),
  seed: z.number(),
  region: z.string(),
  predictedBidPoints: z.number(),
  expectedPoints: z.number(),
  expectedRoi: z.number(),
  rationalPoints: z.number(),
  edgePercent: z.number(),
});

export type EnrichedPrediction = z.infer<typeof EnrichedPredictionSchema>;

export const EnrichedBidSchema = z.object({
  teamId: z.string(),
  schoolName: z.string(),
  seed: z.number(),
  region: z.string(),
  bidPoints: z.number(),
  rationalPoints: z.number(),
  edgePercent: z.number(),
  expectedRoi: z.number().nullable().optional(),
});

export type EnrichedBid = z.infer<typeof EnrichedBidSchema>;

export const EntryDetailSchema = z.object({
  id: z.string(),
  investmentModelId: z.string(),
  calcuttaId: z.string(),
  gameOutcomeKind: z.string(),
  gameOutcomeParamsJson: z.record(z.string(), z.unknown()),
  optimizerKind: z.string(),
  optimizerParamsJson: z.record(z.string(), z.unknown()),
  startingStateKey: z.string(),
  hasPredictions: z.boolean(),
  predictions: z.array(EnrichedPredictionSchema).optional(),
  bids: z.array(EnrichedBidSchema),
  createdAt: z.string(),
  updatedAt: z.string(),
  modelName: z.string(),
  modelKind: z.string(),
  calcuttaName: z.string(),
  nEvaluations: z.number(),
});

export type EntryDetail = z.infer<typeof EntryDetailSchema>;

export const EvaluationDetailSchema = z.object({
  id: z.string(),
  entryId: z.string(),
  nSims: z.number(),
  seed: z.number(),
  meanNormalizedPayout: z.number().nullable().optional(),
  medianNormalizedPayout: z.number().nullable().optional(),
  pTop1: z.number().nullable().optional(),
  pInMoney: z.number().nullable().optional(),
  ourRank: z.number().nullable().optional(),
  simulatedCalcuttaId: z.string().nullable().optional(),
  createdAt: z.string(),
  updatedAt: z.string(),
  modelName: z.string(),
  modelKind: z.string(),
  calcuttaId: z.string(),
  calcuttaName: z.string(),
  startingStateKey: z.string(),
});

export type EvaluationDetail = z.infer<typeof EvaluationDetailSchema>;

export const ListEvaluationsResponseSchema = z.object({
  items: z.array(EvaluationDetailSchema),
});

export type ListEvaluationsResponse = z.infer<typeof ListEvaluationsResponseSchema>;

export const EvaluationEntryResultSchema = z.object({
  id: z.string(),
  entryName: z.string(),
  meanNormalizedPayout: z.number().nullable().optional(),
  pTop1: z.number().nullable().optional(),
  pInMoney: z.number().nullable().optional(),
  rank: z.number(),
});

export type EvaluationEntryResult = z.infer<typeof EvaluationEntryResultSchema>;

export const EvaluationEntryBidSchema = z.object({
  teamId: z.string(),
  schoolName: z.string(),
  seed: z.number(),
  region: z.string(),
  bidPoints: z.number(),
  ownership: z.number(),
});

export type EvaluationEntryBid = z.infer<typeof EvaluationEntryBidSchema>;

export const EvaluationEntryProfileSchema = z.object({
  entryName: z.string(),
  meanNormalizedPayout: z.number().nullable().optional(),
  pTop1: z.number().nullable().optional(),
  pInMoney: z.number().nullable().optional(),
  rank: z.number(),
  totalBidPoints: z.number(),
  bids: z.array(EvaluationEntryBidSchema),
});

export type EvaluationEntryProfile = z.infer<typeof EvaluationEntryProfileSchema>;

export const LeaderboardEntrySchema = z.object({
  investmentModelId: z.string(),
  modelName: z.string(),
  modelKind: z.string(),
  nEntries: z.number(),
  nEntriesWithPredictions: z.number(),
  nEvaluations: z.number(),
  nCalcuttasWithEntries: z.number(),
  nCalcuttasWithEvaluations: z.number(),
  avgMeanPayout: z.number().nullable().optional(),
  avgMedianPayout: z.number().nullable().optional(),
  avgPTop1: z.number().nullable().optional(),
  avgPInMoney: z.number().nullable().optional(),
  firstEvalAt: z.string().nullable().optional(),
  lastEvalAt: z.string().nullable().optional(),
});

export type LeaderboardEntry = z.infer<typeof LeaderboardEntrySchema>;

export const LeaderboardResponseSchema = z.object({
  items: z.array(LeaderboardEntrySchema),
});

export type LeaderboardResponse = z.infer<typeof LeaderboardResponseSchema>;

export const StartPipelineRequestSchema = z.object({
  calcuttaIds: z.array(z.string()).optional(),
  budgetPoints: z.number().optional(),
  optimizerKind: z.string().optional(),
  nSims: z.number().optional(),
  seed: z.number().optional(),
  excludedEntryName: z.string().optional(),
  forceRerun: z.boolean().optional(),
});

export type StartPipelineRequest = z.infer<typeof StartPipelineRequestSchema>;

export const StartPipelineResponseSchema = z.object({
  pipelineRunId: z.string(),
  nCalcuttas: z.number(),
  status: z.string(),
});

export type StartPipelineResponse = z.infer<typeof StartPipelineResponseSchema>;

export const CalcuttaPipelineStatusSchema = z.object({
  calcuttaId: z.string(),
  calcuttaName: z.string(),
  calcuttaYear: z.number(),
  stage: z.string(),
  status: z.string(),
  progress: z.number(),
  progressMessage: z.string().nullable().optional(),
  hasPredictions: z.boolean(),
  hasEntry: z.boolean(),
  hasEvaluation: z.boolean(),
  entryId: z.string().nullable().optional(),
  evaluationId: z.string().nullable().optional(),
  meanPayout: z.number().nullable().optional(),
  ourRank: z.number().nullable().optional(),
  errorMessage: z.string().nullable().optional(),
});

export type CalcuttaPipelineStatus = z.infer<typeof CalcuttaPipelineStatusSchema>;

export const ModelPipelineProgressSchema = z.object({
  modelId: z.string(),
  modelName: z.string(),
  activePipelineRunId: z.string().nullable().optional(),
  totalCalcuttas: z.number(),
  predictionsCount: z.number(),
  entriesCount: z.number(),
  evaluationsCount: z.number(),
  avgMeanPayout: z.number().nullable().optional(),
  calcuttas: z.array(CalcuttaPipelineStatusSchema),
});

export type ModelPipelineProgress = z.infer<typeof ModelPipelineProgressSchema>;

export const EvaluationEntryResultsResponseSchema = z.object({
  items: z.array(EvaluationEntryResultSchema),
});

export const EvaluationTopHoldingSchema = z.object({
  schoolName: z.string(),
  seed: z.number(),
  bidPoints: z.number(),
});

export type EvaluationTopHolding = z.infer<typeof EvaluationTopHoldingSchema>;

export const EvaluationBaselineComparisonSchema = z.object({
  meanPayoutDelta: z.number(),
  pTop1Delta: z.number(),
  interpretation: z.string(),
});

export type EvaluationBaselineComparison = z.infer<typeof EvaluationBaselineComparisonSchema>;

export const EvaluationSummarySchema = z.object({
  percentileRank: z.number(),
  vsBaseline: EvaluationBaselineComparisonSchema.nullable().optional(),
  nEntries: z.number(),
  topHoldings: z.array(EvaluationTopHoldingSchema),
  keyInsight: z.string(),
});

export type EvaluationSummary = z.infer<typeof EvaluationSummarySchema>;
