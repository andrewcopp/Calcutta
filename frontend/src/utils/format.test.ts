import { describe, it, expect } from 'vitest';
import { formatDollarsFromCents } from './format';

describe('formatDollarsFromCents', () => {
  it('returns "$0" for undefined', () => {
    // GIVEN undefined input
    // WHEN formatting
    const result = formatDollarsFromCents(undefined);

    // THEN returns "$0"
    expect(result).toBe('$0');
  });

  it('returns "$0" for zero', () => {
    // GIVEN zero cents
    // WHEN formatting
    const result = formatDollarsFromCents(0);

    // THEN returns "$0"
    expect(result).toBe('$0');
  });

  it('formats whole dollars without decimals', () => {
    // GIVEN 1000 cents (exactly $10)
    // WHEN formatting
    const result = formatDollarsFromCents(1000);

    // THEN returns "$10" without decimals
    expect(result).toBe('$10');
  });

  it('formats cents with zero-padded decimals', () => {
    // GIVEN 1005 cents ($10.05)
    // WHEN formatting
    const result = formatDollarsFromCents(1005);

    // THEN returns "$10.05" with padded cents
    expect(result).toBe('$10.05');
  });

  it('handles negative values', () => {
    // GIVEN -1050 cents (-$10.50)
    // WHEN formatting
    const result = formatDollarsFromCents(-1050);

    // THEN returns "-$10.50"
    expect(result).toBe('-$10.50');
  });
});
