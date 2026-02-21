export const queryKeys = {
  admin: {
    users: (status?: string) => ['admin', 'users', status ?? null] as const,
    apiKeys: () => ['admin', 'apiKeys'] as const,
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
    detail: (id: string | null | undefined) => ['tournament', id ?? null] as const,
    teams: (id: string | null | undefined) => ['tournamentTeams', id ?? null] as const,
    moderators: (id: string | null | undefined) => ['tournamentModerators', id ?? null] as const,
    competitions: () => ['competitions'] as const,
    seasons: () => ['seasons'] as const,
  },
  bracket: {
    validation: (tournamentId: string | null | undefined) => ['bracketValidation', tournamentId ?? null] as const,
    detail: (tournamentId: string | null | undefined) => ['bracket', tournamentId ?? null] as const,
  },
  calcuttas: {
    all: (userId?: string | null) => ['calcuttas', userId ?? null] as const,
    listWithRankings: (userId?: string | null) => ['calcuttasWithRankings', userId ?? null] as const,
    dashboard: (calcuttaId: string | null | undefined) => ['calcuttaDashboard', calcuttaId ?? null] as const,
    entryTeams: (calcuttaId: string | null | undefined, entryId: string | null | undefined) =>
      ['calcuttaEntryTeams', calcuttaId ?? null, entryId ?? null] as const,
    settings: (calcuttaId: string | null | undefined) => ['calcuttaSettings', calcuttaId ?? null] as const,
    payouts: (calcuttaId: string | null | undefined) => ['calcuttaPayouts', calcuttaId ?? null] as const,
  },
  bidding: {
    page: (calcuttaId: string | null | undefined, entryId: string | null | undefined) =>
      ['biddingPage', calcuttaId ?? null, entryId ?? null] as const,
  },
  lab: {
    leaderboard: () => ['lab', 'models', 'leaderboard'] as const,
    models: {
      detail: (modelId: string | null | undefined) => ['lab', 'models', modelId ?? null] as const,
      pipelineProgress: (modelId: string | null | undefined) =>
        ['lab', 'models', modelId ?? null, 'pipeline-progress'] as const,
    },
    entries: {
      detail: (entryId: string | null | undefined) => ['lab', 'entries', entryId ?? null] as const,
      byModelAndCalcutta: (modelName: string | null | undefined, calcuttaId: string | null | undefined) =>
        ['lab', 'entries', 'by-model-calcutta', modelName ?? null, calcuttaId ?? null] as const,
    },
    evaluations: {
      detail: (evaluationId: string | null | undefined) => ['lab', 'evaluations', evaluationId ?? null] as const,
      entries: (evaluationId: string | null | undefined) =>
        ['lab', 'evaluations', evaluationId ?? null, 'entries'] as const,
      byEntry: (entryId: string | null | undefined) =>
        ['lab', 'evaluations', { entry_id: entryId ?? null }] as const,
    },
    entryResults: {
      profile: (entryResultId: string | null | undefined) =>
        ['lab', 'entry-results', entryResultId ?? null] as const,
    },
  },
};
