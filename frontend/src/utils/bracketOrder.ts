/**
 * Returns seeds in bracket matchup order for a given bracket size.
 *
 * For size 16: [1, 16, 8, 9, 4, 13, 5, 12, 2, 15, 7, 10, 3, 14, 6, 11]
 * Adjacent pairs represent first-round matchups (1v16, 8v9, 4v13, ...).
 */
export function bracketOrder(size: number): number[] {
  if (!Number.isInteger(size) || size < 1 || (size & (size - 1)) !== 0) {
    throw new RangeError(`size must be a positive power of 2, got ${size}`);
  }

  let seeds = [1];
  let currentSize = 1;

  while (currentSize < size) {
    currentSize *= 2;
    seeds = seeds.flatMap((s) => [s, currentSize + 1 - s]);
  }

  return seeds;
}
