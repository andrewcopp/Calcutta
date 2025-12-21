export const queryKeys = {
  analytics: {
    all: () => ['analytics'] as const,
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
  },
};
