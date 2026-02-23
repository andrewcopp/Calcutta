import { describe, it, expect } from 'vitest';
import { http, HttpResponse } from 'msw';
import { server } from '../test/msw/server';
import { schoolService } from './schoolService';

const BASE = 'http://localhost:8080/api';

describe('schoolService', () => {
  describe('getSchools', () => {
    it('returns parsed schools from API', async () => {
      const schools = await schoolService.getSchools();

      expect(schools).toEqual([
        { id: 'sch-1', name: 'Duke' },
        { id: 'sch-2', name: 'UNC' },
      ]);
    });

    it('throws when response missing required field', async () => {
      server.use(
        http.get(`${BASE}/schools`, () => {
          return HttpResponse.json([{ id: 'sch-1' }]);
        }),
      );

      await expect(schoolService.getSchools()).rejects.toThrow();
    });

    it('returns empty array when API returns empty list', async () => {
      server.use(
        http.get(`${BASE}/schools`, () => {
          return HttpResponse.json([]);
        }),
      );

      const schools = await schoolService.getSchools();

      expect(schools).toEqual([]);
    });
  });
});
