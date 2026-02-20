import { describe, it, expect } from 'vitest';
import { getEntryColorById } from './entryColors';

describe('getEntryColorById', () => {
  it('returns empty map for empty entries', () => {
    // GIVEN an empty entries array
    // WHEN getting color map
    const result = getEntryColorById([]);

    // THEN map is empty
    expect(result.size).toBe(0);
  });

  it('assigns first color to first entry', () => {
    // GIVEN a single entry and a custom palette
    const entries = [{ id: 'e1' }];
    const palette = ['#FF0000', '#00FF00'];

    // WHEN getting color map
    const result = getEntryColorById(entries, palette);

    // THEN first entry gets first color
    expect(result.get('e1')).toBe('#FF0000');
  });

  it('wraps around palette when entries exceed palette length', () => {
    // GIVEN 3 entries and a 2-color palette
    const entries = [{ id: 'e1' }, { id: 'e2' }, { id: 'e3' }];
    const palette = ['#FF0000', '#00FF00'];

    // WHEN getting color map
    const result = getEntryColorById(entries, palette);

    // THEN third entry wraps back to first color
    expect(result.get('e3')).toBe('#FF0000');
  });
});
