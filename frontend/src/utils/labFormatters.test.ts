import { describe, it, expect } from 'vitest';
import { formatPayoutX, formatPct, getPayoutColor } from './labFormatters';

describe('formatPayoutX', () => {
  it('returns dash for null', () => {
    // GIVEN null input
    // WHEN formatting
    const result = formatPayoutX(null);

    // THEN returns "-"
    expect(result).toBe('-');
  });

  it('returns dash for undefined', () => {
    // GIVEN undefined input
    // WHEN formatting
    const result = formatPayoutX(undefined);

    // THEN returns "-"
    expect(result).toBe('-');
  });

  it('formats with default 3 decimals', () => {
    // GIVEN a numeric value
    // WHEN formatting with default decimals
    const result = formatPayoutX(1.2346);

    // THEN returns value with 3 decimal places and "x"
    expect(result).toBe('1.235x');
  });

  it('formats with custom decimals', () => {
    // GIVEN a numeric value
    // WHEN formatting with 1 decimal
    const result = formatPayoutX(1.2345, 1);

    // THEN returns value with 1 decimal place and "x"
    expect(result).toBe('1.2x');
  });
});

describe('formatPct', () => {
  it('returns dash for null', () => {
    // GIVEN null input
    // WHEN formatting
    const result = formatPct(null);

    // THEN returns "-"
    expect(result).toBe('-');
  });

  it('returns dash for undefined', () => {
    // GIVEN undefined input
    // WHEN formatting
    const result = formatPct(undefined);

    // THEN returns "-"
    expect(result).toBe('-');
  });

  it('multiplies by 100 and appends percent sign', () => {
    // GIVEN 0.123 (12.3%)
    // WHEN formatting with default 1 decimal
    const result = formatPct(0.123);

    // THEN returns "12.3%"
    expect(result).toBe('12.3%');
  });

  it('formats with custom decimals', () => {
    // GIVEN 0.12345
    // WHEN formatting with 2 decimals
    const result = formatPct(0.12345, 2);

    // THEN returns "12.35%"
    expect(result).toBe('12.35%');
  });
});

describe('getPayoutColor', () => {
  it('returns gray for null', () => {
    // GIVEN null payout
    // WHEN getting color
    const result = getPayoutColor(null);

    // THEN returns gray class
    expect(result).toBe('text-gray-400');
  });

  it('returns green bold for payout >= 1.2', () => {
    // GIVEN payout of 1.5
    // WHEN getting color
    const result = getPayoutColor(1.5);

    // THEN returns green bold class
    expect(result).toBe('text-green-700 font-bold');
  });

  it('returns yellow for payout >= 0.9', () => {
    // GIVEN payout of 0.95
    // WHEN getting color
    const result = getPayoutColor(0.95);

    // THEN returns yellow class
    expect(result).toBe('text-yellow-700');
  });

  it('returns red for payout < 0.9', () => {
    // GIVEN payout of 0.5
    // WHEN getting color
    const result = getPayoutColor(0.5);

    // THEN returns red class
    expect(result).toBe('text-red-700');
  });

  it('returns green bold at exactly 1.2 boundary', () => {
    // GIVEN payout of exactly 1.2
    // WHEN getting color
    const result = getPayoutColor(1.2);

    // THEN returns green bold (>= 1.2)
    expect(result).toBe('text-green-700 font-bold');
  });

  it('returns yellow at exactly 0.9 boundary', () => {
    // GIVEN payout of exactly 0.9
    // WHEN getting color
    const result = getPayoutColor(0.9);

    // THEN returns yellow (>= 0.9)
    expect(result).toBe('text-yellow-700');
  });
});
