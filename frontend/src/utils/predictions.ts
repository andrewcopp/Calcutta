import { TeamPrediction } from '../schemas/tournament';

export function getConditionalProbability(
  team: TeamPrediction,
  round: number,
  throughRound: number,
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

  // Future rounds: team is alive — pRound is already conditional from backend
  const pKey = `pRound${round}` as keyof TeamPrediction;
  const raw = team[pKey] as number;
  return { value: raw, style: '' };
}

export function formatPercent(value: number): string {
  if (value >= 0.9995) return '100%';
  if (value < 0.0005) return '—';
  return `${(value * 100).toFixed(1)}%`;
}
