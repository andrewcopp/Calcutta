import { describe, it, expect } from 'vitest';
import { getRoiColor, formatRoi } from '../../../utils/labFormatters';

describe('getRoiColor', () => {
  it('returns green bold for roi >= 2.0', () => {
    // GIVEN roi of 2.5
    // WHEN getting color
    const result = getRoiColor(2.5);

    // THEN returns green bold
    expect(result).toBe('text-green-700 font-bold');
  });

  it('returns green for roi >= 1.5', () => {
    // GIVEN roi of 1.7
    // WHEN getting color
    const result = getRoiColor(1.7);

    // THEN returns green-600
    expect(result).toBe('text-green-600');
  });

  it('returns gray for roi >= 1.0', () => {
    // GIVEN roi of 1.2
    // WHEN getting color
    const result = getRoiColor(1.2);

    // THEN returns gray-900
    expect(result).toBe('text-gray-900');
  });

  it('returns yellow for roi >= 0.5', () => {
    // GIVEN roi of 0.7
    // WHEN getting color
    const result = getRoiColor(0.7);

    // THEN returns yellow-600
    expect(result).toBe('text-yellow-600');
  });

  it('returns red for roi < 0.5', () => {
    // GIVEN roi of 0.3
    // WHEN getting color
    const result = getRoiColor(0.3);

    // THEN returns red-600
    expect(result).toBe('text-red-600');
  });
});

describe('formatRoi', () => {
  it('returns em dash for NaN', () => {
    // GIVEN NaN
    // WHEN formatting
    const result = formatRoi(NaN);

    // THEN returns em dash
    expect(result).toBe('\u2014');
  });

  it('returns em dash for Infinity', () => {
    // GIVEN Infinity
    // WHEN formatting
    const result = formatRoi(Infinity);

    // THEN returns em dash
    expect(result).toBe('\u2014');
  });

  it('formats with 2 decimals and x suffix', () => {
    // GIVEN a valid roi
    // WHEN formatting
    const result = formatRoi(1.567);

    // THEN returns formatted string with 2 decimals
    expect(result).toBe('1.57x');
  });
});
