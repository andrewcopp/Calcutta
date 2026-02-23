/**
 * Returns seeds in bracket matchup order using NCAA S-curve seeding.
 *
 * For size 16: [1, 16, 8, 9, 5, 12, 4, 13, 6, 11, 3, 14, 7, 10, 2, 15]
 * Adjacent pairs represent first-round matchups (1v16, 8v9, 5v12, ...).
 */
export function bracketOrder(size: number): number[] {
  if (!Number.isInteger(size) || size < 1 || (size & (size - 1)) !== 0) {
    throw new RangeError(`size must be a positive power of 2, got ${size}`);
  }

  if (size === 1) return [1];

  // Phase 1: Build S-curve anchors for the top half of seeds.
  // At each expansion, index 0 keeps natural order [s, comp],
  // all other indices reverse to [comp, s] for maximum separation.
  let anchors = [1];
  let currentSize = 1;

  while (currentSize < size / 2) {
    currentSize *= 2;
    anchors = anchors.flatMap((s, i) => (i === 0 ? [s, currentSize + 1 - s] : [currentSize + 1 - s, s]));
  }

  // Phase 2: Expand each anchor to a matchup pair [s, size+1-s].
  return anchors.flatMap((s) => [s, size + 1 - s]);
}
