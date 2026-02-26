import { describe, it, expect } from 'vitest';
import { http, HttpResponse } from 'msw';
import { server } from '../test/msw/server';
import { labService } from './labService';

const BASE = 'http://localhost:8080/api/v1';

describe('labService', () => {
  describe('getModel', () => {
    it('returns parsed investment model', async () => {
      const model = await labService.getModel('model-1');

      expect(model.name).toBe('test-model');
      expect(model.nEntries).toBe(5);
    });

    it('throws when model missing required field', async () => {
      server.use(
        http.get(`${BASE}/lab/models/:id`, () => {
          return HttpResponse.json({ id: 'model-1', name: 'test' });
        }),
      );

      await expect(labService.getModel('model-1')).rejects.toThrow();
    });
  });

  describe('getLeaderboard', () => {
    it('returns parsed leaderboard', async () => {
      const result = await labService.getLeaderboard();

      expect(result.items).toHaveLength(1);
      expect(result.items[0].modelName).toBe('test-model');
      expect(result.items[0].avgMeanPayout).toBe(1.25);
    });
  });

  describe('getEntryByModelAndCalcutta', () => {
    it('returns parsed entry detail', async () => {
      const entry = await labService.getEntryByModelAndCalcutta('test-model', 'calc-1');

      expect(entry.hasPredictions).toBe(true);
      expect(entry.predictions).toHaveLength(1);
      expect(entry.bids).toHaveLength(1);
    });

    it('returns entry with starting state key', async () => {
      const entry = await labService.getEntryByModelAndCalcutta('test-model', 'calc-1', 'pre_tournament');

      expect(entry.startingStateKey).toBe('pre_tournament');
    });
  });

  describe('listEvaluations', () => {
    it('returns parsed evaluations', async () => {
      const result = await labService.listEvaluations();

      expect(result.items).toHaveLength(1);
      expect(result.items[0].meanNormalizedPayout).toBe(1.25);
    });

    it('passes filter params correctly', async () => {
      const result = await labService.listEvaluations({
        calcuttaId: 'calc-1',
        limit: 10,
      });

      expect(result.items).toHaveLength(1);
    });
  });

  describe('getEvaluation', () => {
    it('returns parsed evaluation detail', async () => {
      const eval_ = await labService.getEvaluation('eval-1');

      expect(eval_.nSims).toBe(1000);
      expect(eval_.pTop1).toBe(0.15);
    });

    it('throws when evaluation missing required field', async () => {
      server.use(
        http.get(`${BASE}/lab/evaluations/:id`, () => {
          return HttpResponse.json({ id: 'eval-1' });
        }),
      );

      await expect(labService.getEvaluation('eval-1')).rejects.toThrow();
    });
  });

  describe('getEvaluationEntryResults', () => {
    it('returns items array directly', async () => {
      const results = await labService.getEvaluationEntryResults('eval-1');

      expect(results).toHaveLength(1);
      expect(results[0].entryName).toBe('Test Entry');
    });
  });

  describe('getEvaluationEntryProfile', () => {
    it('returns parsed entry profile', async () => {
      const profile = await labService.getEvaluationEntryProfile('result-1');

      expect(profile.totalBidPoints).toBe(100);
      expect(profile.bids).toHaveLength(1);
    });
  });

  describe('startPipeline', () => {
    it('returns pipeline start response', async () => {
      const result = await labService.startPipeline('model-1');

      expect(result.pipelineRunId).toBe('run-1');
      expect(result.status).toBe('running');
    });

    it('accepts optional request params', async () => {
      const result = await labService.startPipeline('model-1', { nSims: 500, seed: 42 });

      expect(result.nCalcuttas).toBe(2);
    });
  });

  describe('getModelPipelineProgress', () => {
    it('returns parsed pipeline progress', async () => {
      const progress = await labService.getModelPipelineProgress('model-1');

      expect(progress.totalCalcuttas).toBe(2);
      expect(progress.calcuttas).toHaveLength(1);
      expect(progress.calcuttas[0].stage).toBe('complete');
    });
  });

  describe('cancelPipeline', () => {
    it('completes without error', async () => {
      await expect(labService.cancelPipeline('run-1')).resolves.toBeUndefined();
    });
  });
});
