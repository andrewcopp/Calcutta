import { CalcuttaEntry, CalcuttaEntryTeam, TournamentTeam } from '../types/calcutta';

const MOCK_NAMES = ['Alpha', 'Bravo', 'Charlie', 'Delta', 'Echo', 'Foxtrot', 'Golf', 'Hotel'];

export function generateMockBiddingData(
  tournamentTeams: TournamentTeam[],
  entryCount: number,
): { entries: CalcuttaEntry[]; entryTeams: CalcuttaEntryTeam[] } {
  const count = Math.max(entryCount, 3);
  const now = new Date().toISOString();

  const entries: CalcuttaEntry[] = Array.from({ length: count }, (_, i) => ({
    id: `mock-entry-${i}`,
    name: MOCK_NAMES[i % MOCK_NAMES.length],
    calcuttaId: 'mock',
    status: 'final',
    totalPoints: Math.round((100 - i * 8 + Math.random() * 10) * 100) / 100,
    finishPosition: i + 1,
    inTheMoney: i < 3,
    payoutCents: i < 3 ? (3 - i) * 5000 : 0,
    created: now,
    updated: now,
  }));

  const entryTeams: CalcuttaEntryTeam[] = [];
  for (const entry of entries) {
    const teamSubset = tournamentTeams.slice(0, Math.min(8, tournamentTeams.length));
    for (const team of teamSubset) {
      entryTeams.push({
        id: `mock-et-${entry.id}-${team.id}`,
        entryId: entry.id,
        teamId: team.id,
        bid: Math.floor(Math.random() * 80) + 20,
        created: now,
        updated: now,
      });
    }
  }

  return { entries, entryTeams };
}
