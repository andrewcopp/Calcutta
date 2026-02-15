export const queryKeys = {
  analytics: {
    all: () => ['analytics'] as const,
    seedInvestmentDistribution: () => ['analytics', 'seedInvestmentDistribution'] as const,
    bestInvestments: (limit?: number) => ['analytics', 'bestInvestments', limit ?? null] as const,
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
  },
  bracket: {
    validation: (tournamentId: string | null | undefined) => ['bracketValidation', tournamentId ?? null] as const,
    detail: (tournamentId: string | null | undefined) => ['bracket', tournamentId ?? null] as const,
  },
  calcuttas: {
    all: (userId?: string | null) => ['calcuttas', userId ?? null] as const,
    teamsPage: (calcuttaId: string | null | undefined) => ['calcuttaTeamsPage', calcuttaId ?? null] as const,
    entriesPage: (calcuttaId: string | null | undefined) => ['calcuttaEntriesPage', calcuttaId ?? null] as const,
    entryTeamsPage: (calcuttaId: string | null | undefined, entryId: string | null | undefined) =>
      ['entryTeamsPage', calcuttaId ?? null, entryId ?? null] as const,
    settings: (calcuttaId: string | null | undefined) => ['calcuttaSettings', calcuttaId ?? null] as const,
    payouts: (calcuttaId: string | null | undefined) => ['calcuttaPayouts', calcuttaId ?? null] as const,
  },
};
