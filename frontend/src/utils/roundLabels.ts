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

export function getRoundName(round: number, maxRound: number): string {
  if (round === 0) return 'Start';
  const nameMap = maxRound >= 7 ? ROUND_NAMES_7 : ROUND_NAMES_6;
  return nameMap[round] ?? `Round ${round}`;
}

export function getRoundOptions(rounds: number[]): RoundOption[] {
  if (rounds.length === 0) return [];

  const maxRound = Math.max(...rounds);
  const nameMap = maxRound >= 7 ? ROUND_NAMES_7 : ROUND_NAMES_6;

  const options: RoundOption[] = [{ label: 'Current', value: null }];

  // Add round options in reverse chronological order (After Championship → ... → Start)
  const sorted = [...rounds].filter((r) => r > 0).sort((a, b) => b - a);

  for (const round of sorted) {
    const name = nameMap[round] ?? `Round ${round}`;
    options.push({ label: `After ${name}`, value: round });
  }

  if (rounds.includes(0)) {
    options.push({ label: 'Start', value: 0 });
  }

  return options;
}
