import { describe, expect, it } from 'vitest';
import { bracketOrder } from './bracketOrder';

describe('bracketOrder', () => {
  describe('exact output', () => {
    it('TestThatBracketOrderOfSize1Returns1', () => {
      expect(bracketOrder(1)).toEqual([1]);
    });

    it('TestThatBracketOrderOfSize2Returns1Then2', () => {
      expect(bracketOrder(2)).toEqual([1, 2]);
    });

    it('TestThatBracketOrderOfSize4Returns1423', () => {
      expect(bracketOrder(4)).toEqual([1, 4, 2, 3]);
    });

    it('TestThatBracketOrderOfSize8ReturnsCorrectOrder', () => {
      expect(bracketOrder(8)).toEqual([1, 8, 4, 5, 3, 6, 2, 7]);
    });

    it('TestThatBracketOrderOfSize16ReturnsCorrectOrder', () => {
      expect(bracketOrder(16)).toEqual([1, 16, 8, 9, 5, 12, 4, 13, 6, 11, 3, 14, 7, 10, 2, 15]);
    });
  });

  describe('matchup property', () => {
    it('TestThatAdjacentPairsSumToSizePlusOneForSize4', () => {
      const seeds = bracketOrder(4);
      for (let i = 0; i < seeds.length; i += 2) {
        expect(seeds[i] + seeds[i + 1]).toBe(5);
      }
    });

    it('TestThatAdjacentPairsSumToSizePlusOneForSize16', () => {
      const seeds = bracketOrder(16);
      for (let i = 0; i < seeds.length; i += 2) {
        expect(seeds[i] + seeds[i + 1]).toBe(17);
      }
    });
  });

  describe('chalk collapse property', () => {
    it('TestThatChalkCollapseOfSize4ProducesSize2Order', () => {
      const seeds = bracketOrder(4);
      const chalked = [];
      for (let i = 0; i < seeds.length; i += 2) {
        chalked.push(Math.min(seeds[i], seeds[i + 1]));
      }
      expect(chalked).toEqual(bracketOrder(2));
    });

    it('TestThatChalkCollapseOfSize16ProducesExpectedAnchors', () => {
      const seeds = bracketOrder(16);
      const chalked = [];
      for (let i = 0; i < seeds.length; i += 2) {
        chalked.push(Math.min(seeds[i], seeds[i + 1]));
      }
      expect(chalked).toEqual([1, 8, 5, 4, 6, 3, 7, 2]);
    });
  });

  describe('completeness', () => {
    it('TestThatBracketOrderOfSize16ContainsEverySeedException', () => {
      const seeds = bracketOrder(16);
      const sorted = [...seeds].sort((a, b) => a - b);
      expect(sorted).toEqual(Array.from({ length: 16 }, (_, i) => i + 1));
    });
  });

  describe('validation', () => {
    it('TestThatSize0Throws', () => {
      expect(() => bracketOrder(0)).toThrow(RangeError);
    });

    it('TestThatNegativeSizeThrows', () => {
      expect(() => bracketOrder(-4)).toThrow(RangeError);
    });

    it('TestThatNonPowerOf2Throws', () => {
      expect(() => bracketOrder(6)).toThrow(RangeError);
    });

    it('TestThatNonIntegerThrows', () => {
      expect(() => bracketOrder(2.5)).toThrow(RangeError);
    });
  });
});
