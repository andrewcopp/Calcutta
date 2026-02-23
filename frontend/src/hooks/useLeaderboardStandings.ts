import { useMemo } from 'react';
import type { CalcuttaDashboard, CalcuttaEntry } from '../schemas/calcutta';
import { pointsForProgress, computeStandings, type Standing } from '../utils/standings';

/**
 * Computes leaderboard standings for a given "through round" cap.
 * When throughRound is null (Current), returns null to signal "use backend entries".
 * When throughRound is a number, recomputes standings from raw dashboard data.
 */
export function useLeaderboardStandings(
  dashboard: CalcuttaDashboard,
  throughRound: number | null,
): CalcuttaEntry[] | null {
  return useMemo(() => {
    if (throughRound === null) return null;

    const { entries, portfolios, portfolioTeams, tournamentTeams, scoringRules, payouts } =
      dashboard;

    // Build team lookup
    const teamByID = new Map(tournamentTeams.map((t) => [t.id, t]));

    // Build portfolio -> entry mapping
    const portfolioToEntry = new Map(portfolios.map((p) => [p.id, p.entryId]));

    // Compute points per entry
    const pointsByEntry: Record<string, number> = {};
    for (const pt of portfolioTeams) {
      const team = teamByID.get(pt.teamId);
      if (!team) continue;

      const entryId = portfolioToEntry.get(pt.portfolioId);
      if (!entryId) continue;

      const cappedProgress = Math.min(team.wins + team.byes, throughRound);
      const teamPoints = pointsForProgress(scoringRules, cappedProgress, 0);
      pointsByEntry[entryId] = (pointsByEntry[entryId] ?? 0) + pt.ownershipPercentage * teamPoints;
    }

    const standings = computeStandings(entries, pointsByEntry, payouts);

    // Map standings back to CalcuttaEntry shape for rendering
    const standingByEntry = new Map(standings.map((s) => [s.entryId, s]));
    const entryById = new Map(entries.map((e) => [e.id, e]));

    return standings.map((s) => {
      const entry = entryById.get(s.entryId)!;
      return {
        ...entry,
        totalPoints: s.totalPoints,
        finishPosition: s.finishPosition,
        inTheMoney: s.inTheMoney,
        payoutCents: s.payoutCents,
        projectedEv: undefined,
      };
    });
  }, [dashboard, throughRound]);
}
