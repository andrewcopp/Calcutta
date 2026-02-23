import type { ScoringRule } from '../schemas/calcutta';

export interface RoundOption {
  label: string;
  value: number | null; // null = "Current"
}

const ROUND_NAMES_6: Record<number, string> = {
  1: 'Round of 64',
  2: 'Round of 32',
  3: 'Sweet 16',
  4: 'Elite 8',
  5: 'Final Four',
  6: 'Championship',
};

const ROUND_NAMES_7: Record<number, string> = {
  1: 'First Four',
  2: 'Round of 64',
  3: 'Round of 32',
  4: 'Sweet 16',
  5: 'Elite 8',
  6: 'Final Four',
  7: 'Championship',
};

export function getRoundOptions(scoringRules: ScoringRule[]): RoundOption[] {
  if (scoringRules.length === 0) return [{ label: 'Current', value: null }];

  const maxWinIndex = Math.max(...scoringRules.map((r) => r.winIndex));
  const nameMap = maxWinIndex >= 7 ? ROUND_NAMES_7 : ROUND_NAMES_6;

  const options: RoundOption[] = [{ label: 'Current', value: null }];

  // Add round options in reverse chronological order (After Championship → ... → Start)
  const winIndices = scoringRules.map((r) => r.winIndex).sort((a, b) => b - a);

  for (const idx of winIndices) {
    const name = nameMap[idx] ?? `Round ${idx}`;
    options.push({ label: `After ${name}`, value: idx });
  }

  options.push({ label: 'Start', value: 0 });

  return options;
}
