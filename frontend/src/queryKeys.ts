export const queryKeys = {
  admin: {
    users: (status?: string) => ['admin', 'users', status ?? null] as const,
    userDetail: (userId: string | null | undefined) => ['admin', 'users', 'detail', userId ?? null] as const,
    apiKeys: () => ['admin', 'apiKeys'] as const,
    stubs: () => ['admin', 'users', 'stubs'] as const,
    mergeCandidates: (userId: string) => ['admin', 'users', 'merge-candidates', userId] as const,
    mergeHistory: (userId: string) => ['admin', 'users', 'merges', userId] as const,
  },
  profile: {
    me: () => ['profile', 'me'] as const,
  },
  hallOfFame: {
    bestTeams: (limit?: number) => ['hallOfFame', 'bestTeams', limit ?? null] as const,
    bestInvestments: (limit?: number) => ['hallOfFame', 'bestInvestments', limit ?? null] as const,
    bestEntries: (limit?: number) => ['hallOfFame', 'bestEntries', limit ?? null] as const,
    bestCareers: (limit?: number) => ['hallOfFame', 'bestCareers', limit ?? null] as const,
  },
  schools: {
    all: () => ['schools'] as const,
  },
  tournaments: {
    all: () => ['tournaments'] as const,
    detail: (id: string | null | undefined) => ['tournaments', 'detail', id ?? null] as const,
    teams: (id: string | null | undefined) => ['tournaments', 'teams', id ?? null] as const,
    moderators: (id: string | null | undefined) => ['tournaments', 'moderators', id ?? null] as const,
    predictionBatches: (id: string | null | undefined) => ['tournaments', 'predictionBatches', id ?? null] as const,
    predictions: (id: string | null | undefined, batchId?: string | null) =>
      ['tournaments', 'predictions', id ?? null, batchId ?? null] as const,
    competitions: () => ['tournaments', 'competitions'] as const,
    seasons: () => ['tournaments', 'seasons'] as const,
  },
  bracket: {
    validation: (tournamentId: string | null | undefined) => ['tournaments', 'bracket', 'validation', tournamentId ?? null] as const,
    detail: (tournamentId: string | null | undefined) => ['tournaments', 'bracket', 'detail', tournamentId ?? null] as const,
  },
  pools: {
    all: (userId?: string | null) => ['pools', userId ?? null] as const,
    listWithRankings: (userId?: string | null) => ['poolsWithRankings', userId ?? null] as const,
    dashboard: (poolId: string | null | undefined) => ['poolDashboard', poolId ?? null] as const,
    investments: (poolId: string | null | undefined, portfolioId: string | null | undefined) =>
      ['poolInvestments', poolId ?? null, portfolioId ?? null] as const,
    settings: (poolId: string | null | undefined) => ['poolSettings', poolId ?? null] as const,
    payouts: (poolId: string | null | undefined) => ['poolPayouts', poolId ?? null] as const,
  },
  investing: {
    page: (poolId: string | null | undefined, portfolioId: string | null | undefined) =>
      ['investingPage', poolId ?? null, portfolioId ?? null] as const,
  },
  lab: {
    leaderboard: () => ['lab', 'models', 'leaderboard'] as const,
    models: {
      detail: (modelId: string | null | undefined) => ['lab', 'models', modelId ?? null] as const,
      pipelineProgress: (modelId: string | null | undefined) =>
        ['lab', 'models', modelId ?? null, 'pipeline-progress'] as const,
    },
    entries: {
      byModelAndPool: (modelId: string | null | undefined, poolId: string | null | undefined) =>
        ['lab', 'entries', 'by-model-pool', modelId ?? null, poolId ?? null] as const,
    },
    evaluations: {
      detail: (evaluationId: string | null | undefined) => ['lab', 'evaluations', evaluationId ?? null] as const,
      entries: (evaluationId: string | null | undefined) =>
        ['lab', 'evaluations', evaluationId ?? null, 'entries'] as const,
      summary: (evaluationId: string | null | undefined) =>
        ['lab', 'evaluations', evaluationId ?? null, 'summary'] as const,
      byEntry: (entryId: string | null | undefined) => ['lab', 'evaluations', { entryId: entryId ?? null }] as const,
    },
    entryResults: {
      profile: (entryResultId: string | null | undefined) => ['lab', 'entry-results', entryResultId ?? null] as const,
    },
  },
};
