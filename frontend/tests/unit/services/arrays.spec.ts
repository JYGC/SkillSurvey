import { describe, it, expect } from 'vitest';
import { sortByProperty } from '@/services/arrays';

describe('sortByProperty', () => {
  it('sorts ascending by string property', () => {
    const arr = [{ name: 'C' }, { name: 'A' }, { name: 'B' }];
    sortByProperty(arr, e => e.name);
    expect(arr.map(e => e.name)).toEqual(['A', 'B', 'C']);
  });

  it('sorts ascending by number property', () => {
    const arr = [{ n: 3 }, { n: 1 }, { n: 2 }];
    sortByProperty(arr, e => e.n);
    expect(arr.map(e => e.n)).toEqual([1, 2, 3]);
  });

  it('returns the mutated input array', () => {
    const arr = [{ n: 2 }, { n: 1 }];
    const result = sortByProperty(arr, e => e.n);
    expect(result).toBe(arr);
  });
});
