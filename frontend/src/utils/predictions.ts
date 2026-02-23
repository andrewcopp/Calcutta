import { TeamPrediction } from '../schemas/tournament';

export const EXPECTED_TEAMS_AT_ROUND: Record<number, number> = {
  1: 64,
  2: 32,
  3: 16,
  4: 8,
  5: 4,
  6: 2,
  7: 1,
};

export function getConditionalProbability(
  team: TeamPrediction,
  round: number,
  throughRound: number,
  roundSums?: Record<number, number>,
): { value: number; style: string } {
  const progress = team.wins + team.byes;

  if (throughRound === 0) {
    if (team.byes >= round) {
      return { value: 1, style: 'text-green-600 font-semibold' };
    }
    const pKey = `pRound${round}` as keyof TeamPrediction;
    const raw = team[pKey] as number;
    return { value: raw, style: '' };
  }

  // Resolved rounds: team already played through this round
  if (round <= throughRound) {
    if (progress >= round) {
      return { value: 1, style: 'text-green-600 font-semibold' };
    }
    return { value: 0, style: 'text-muted-foreground' };
  }

  // Future rounds: team was eliminated before this checkpoint
  if (progress < throughRound) {
    return { value: 0, style: 'text-muted-foreground' };
  }

  // Future rounds: team is alive — conditional probability
  const pKey = `pRound${round}` as keyof TeamPrediction;
  const pCapKey = `pRound${throughRound}` as keyof TeamPrediction;
  const pRound = team[pKey] as number;
  const pCap = team[pCapKey] as number;

  let conditional: number;
  if (roundSums && roundSums[round] !== undefined) {
    const expected = EXPECTED_TEAMS_AT_ROUND[round] ?? 1;
    conditional = roundSums[round] > 0 ? Math.min((pRound / roundSums[round]) * expected, 1) : 0;
  } else {
    conditional = pCap > 0 ? Math.min(pRound / pCap, 1) : 0;
  }
  return { value: conditional, style: '' };
}

export function formatPercent(value: number): string {
  if (value >= 0.9995) return '100%';
  if (value < 0.0005) return '—';
  return `${(value * 100).toFixed(1)}%`;
}
