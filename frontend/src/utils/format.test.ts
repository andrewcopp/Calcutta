import { describe, it, expect } from 'vitest';
import { formatDollarsFromCents, toDatetimeLocalValue } from './format';

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

describe('toDatetimeLocalValue', () => {
  it('formats a standard ISO date to datetime-local format', () => {
    // GIVEN a local date of November 25, 2024, 14:30
    const date = new Date(2024, 10, 25, 14, 30);

    // WHEN converting to datetime-local value
    const result = toDatetimeLocalValue(date.toISOString());

    // THEN returns the expected YYYY-MM-DDTHH:MM string
    expect(result).toBe('2024-11-25T14:30');
  });

  it('pads single-digit months and days', () => {
    // GIVEN a local date of March 5, 2024, 12:00
    const date = new Date(2024, 2, 5, 12, 0);

    // WHEN converting to datetime-local value
    const result = toDatetimeLocalValue(date.toISOString());

    // THEN zero-pads the month and day
    expect(result).toBe('2024-03-05T12:00');
  });

  it('pads single-digit hours and minutes', () => {
    // GIVEN a local date of January 15, 2024, 9:05 AM
    const date = new Date(2024, 0, 15, 9, 5);

    // WHEN converting to datetime-local value
    const result = toDatetimeLocalValue(date.toISOString());

    // THEN zero-pads the hours and minutes
    expect(result).toBe('2024-01-15T09:05');
  });

  it('handles midnight', () => {
    // GIVEN a local date of June 1, 2024, at midnight (00:00)
    const date = new Date(2024, 5, 1, 0, 0);

    // WHEN converting to datetime-local value
    const result = toDatetimeLocalValue(date.toISOString());

    // THEN represents midnight as 00:00
    expect(result).toBe('2024-06-01T00:00');
  });
});
