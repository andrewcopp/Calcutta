// Shared utility types
export type SortDir = 'asc' | 'desc';

// Types for lab.investment_models
export type InvestmentModel = {
  id: string;
  name: string;
  kind: string;
  paramsJson: Record<string, unknown>;
  notes?: string | null;
  createdAt: string;
  updatedAt: string;
  nEntries: number;
  nEvaluations: number;
};

// Types for lab.entries - market predictions (what model thinks market will bid)
export type EnrichedPrediction = {
  teamId: string;
  schoolName: string;
  seed: number;
  region: string;
  predictedBidPoints: number;
  expectedPoints: number;
  expectedRoi: number;
  naivePoints: number;
  edgePercent: number;
};

// Types for lab.entries - optimized bids (our response to predictions)
export type EnrichedBid = {
  teamId: string;
  schoolName: string;
  seed: number;
  region: string;
  bidPoints: number;
  naivePoints: number;
  edgePercent: number;
  expectedRoi?: number | null;
};

export type EntryDetail = {
  id: string;
  investmentModelId: string;
  calcuttaId: string;
  gameOutcomeKind: string;
  gameOutcomeParamsJson: Record<string, unknown>;
  optimizerKind: string;
  optimizerParamsJson: Record<string, unknown>;
  startingStateKey: string;
  hasPredictions: boolean;
  predictions?: EnrichedPrediction[];
  bids: EnrichedBid[];
  createdAt: string;
  updatedAt: string;
  modelName: string;
  modelKind: string;
  calcuttaName: string;
  nEvaluations: number;
};

// Types for lab.evaluations
export type EvaluationDetail = {
  id: string;
  entryId: string;
  nSims: number;
  seed: number;
  meanNormalizedPayout?: number | null;
  medianNormalizedPayout?: number | null;
  pTop1?: number | null;
  pInMoney?: number | null;
  ourRank?: number | null;
  simulatedCalcuttaId?: string | null;
  createdAt: string;
  updatedAt: string;
  modelName: string;
  modelKind: string;
  calcuttaId: string;
  calcuttaName: string;
  startingStateKey: string;
};

export type ListEvaluationsResponse = {
  items: EvaluationDetail[];
};

// Types for evaluation entry results
export type EvaluationEntryResult = {
  id: string;
  entryName: string;
  meanNormalizedPayout?: number | null;
  pTop1?: number | null;
  pInMoney?: number | null;
  rank: number;
};

// Types for evaluation entry bid
export type EvaluationEntryBid = {
  teamId: string;
  schoolName: string;
  seed: number;
  region: string;
  bidPoints: number;
  ownership: number;
};

// Types for evaluation entry profile
export type EvaluationEntryProfile = {
  entryName: string;
  meanNormalizedPayout?: number | null;
  pTop1?: number | null;
  pInMoney?: number | null;
  rank: number;
  totalBidPoints: number;
  bids: EvaluationEntryBid[];
};

// Types for leaderboard
export type LeaderboardEntry = {
  investmentModelId: string;
  modelName: string;
  modelKind: string;
  nEntries: number;
  nEntriesWithPredictions: number;
  nEvaluations: number;
  nCalcuttasWithEntries: number;
  nCalcuttasWithEvaluations: number;
  avgMeanPayout?: number | null;
  avgMedianPayout?: number | null;
  avgPTop1?: number | null;
  avgPInMoney?: number | null;
  firstEvalAt?: string | null;
  lastEvalAt?: string | null;
};

export type LeaderboardResponse = {
  items: LeaderboardEntry[];
};

// Types for pipeline
export type StartPipelineRequest = {
  calcuttaIds?: string[];
  budgetPoints?: number;
  optimizerKind?: string;
  nSims?: number;
  seed?: number;
  excludedEntryName?: string;
  forceRerun?: boolean;
};

export type StartPipelineResponse = {
  pipelineRunId: string;
  nCalcuttas: number;
  status: string;
};

export type CalcuttaPipelineStatus = {
  calcuttaId: string;
  calcuttaName: string;
  calcuttaYear: number;
  stage: string;
  status: string;
  progress: number;
  progressMessage?: string | null;
  hasPredictions: boolean;
  hasEntry: boolean;
  hasEvaluation: boolean;
  entryId?: string | null;
  evaluationId?: string | null;
  meanPayout?: number | null;
  ourRank?: number | null;
  errorMessage?: string | null;
};

export type ModelPipelineProgress = {
  modelId: string;
  modelName: string;
  activePipelineRunId?: string | null;
  totalCalcuttas: number;
  predictionsCount: number;
  entriesCount: number;
  evaluationsCount: number;
  avgMeanPayout?: number | null;
  calcuttas: CalcuttaPipelineStatus[];
};
