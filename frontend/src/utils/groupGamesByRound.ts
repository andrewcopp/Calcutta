import { BracketGame, BracketRound, ROUND_ORDER } from '../schemas/bracket';

export function groupGamesByRound(games: BracketGame[]): Record<BracketRound, BracketGame[]> {
  const grouped = {} as Record<BracketRound, BracketGame[]>;

  ROUND_ORDER.forEach((round) => {
    grouped[round] = games.filter((game) => game.round === round).sort((a, b) => a.sortOrder - b.sortOrder);
  });

  return grouped;
}
