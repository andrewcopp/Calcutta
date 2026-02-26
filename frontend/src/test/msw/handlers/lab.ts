import { http, HttpResponse } from 'msw';

const BASE = 'http://localhost:8080/api/v1';

const validModel = {
  id: 'model-1',
  name: 'test-model',
  kind: 'regression',
  paramsJson: { lr: 0.01 },
  notes: null,
  createdAt: '2026-01-01T00:00:00Z',
  updatedAt: '2026-01-01T00:00:00Z',
  nEntries: 5,
  nEvaluations: 3,
};

const validEntry = {
  id: 'entry-1',
  investmentModelId: 'model-1',
  calcuttaId: 'calc-1',
  gameOutcomeKind: 'kenpom',
  gameOutcomeParamsJson: {},
  optimizerKind: 'milp',
  optimizerParamsJson: {},
  startingStateKey: 'pre_tournament',
  hasPredictions: true,
  predictions: [
    {
      teamId: 'team-1',
      schoolName: 'Duke',
      seed: 1,
      region: 'East',
      predictedBidPoints: 15,
      expectedPoints: 50,
      expectedRoi: 2.5,
      rationalPoints: 10,
      edgePercent: 0.3,
    },
  ],
  bids: [
    {
      teamId: 'team-1',
      schoolName: 'Duke',
      seed: 1,
      region: 'East',
      bidPoints: 20,
      rationalPoints: 10,
      edgePercent: 0.5,
    },
  ],
  createdAt: '2026-01-01T00:00:00Z',
  updatedAt: '2026-01-01T00:00:00Z',
  modelName: 'test-model',
  modelKind: 'regression',
  calcuttaName: 'Test Pool',
  nEvaluations: 3,
};

const validEvaluation = {
  id: 'eval-1',
  entryId: 'entry-1',
  nSims: 1000,
  seed: 42,
  meanNormalizedPayout: 1.25,
  medianNormalizedPayout: 1.1,
  pTop1: 0.15,
  pInMoney: 0.6,
  ourRank: 2,
  simulatedCalcuttaId: null,
  createdAt: '2026-01-01T00:00:00Z',
  updatedAt: '2026-01-01T00:00:00Z',
  modelName: 'test-model',
  modelKind: 'regression',
  calcuttaId: 'calc-1',
  calcuttaName: 'Test Pool',
  startingStateKey: 'pre_tournament',
};

export const labHandlers = [
  http.get(`${BASE}/lab/models/leaderboard`, () => {
    return HttpResponse.json({
      items: [
        {
          investmentModelId: 'model-1',
          modelName: 'test-model',
          modelKind: 'regression',
          nEntries: 5,
          nEntriesWithPredictions: 5,
          nEvaluations: 3,
          nCalcuttasWithEntries: 2,
          nCalcuttasWithEvaluations: 2,
          avgMeanPayout: 1.25,
          avgMedianPayout: 1.1,
          avgPTop1: 0.15,
          avgPInMoney: 0.6,
          firstEvalAt: '2026-01-01T00:00:00Z',
          lastEvalAt: '2026-01-15T00:00:00Z',
        },
      ],
    });
  }),

  http.get(`${BASE}/lab/models/:name/calcutta/:calcuttaId/entry`, () => {
    return HttpResponse.json(validEntry);
  }),

  http.get(`${BASE}/lab/models/:id/pipeline/progress`, () => {
    return HttpResponse.json({
      modelId: 'model-1',
      modelName: 'test-model',
      activePipelineRunId: null,
      totalCalcuttas: 2,
      predictionsCount: 2,
      entriesCount: 2,
      evaluationsCount: 2,
      avgMeanPayout: 1.25,
      calcuttas: [
        {
          calcuttaId: 'calc-1',
          calcuttaName: 'Test Pool',
          calcuttaYear: 2025,
          stage: 'complete',
          status: 'succeeded',
          progress: 1.0,
          hasPredictions: true,
          hasEntry: true,
          hasEvaluation: true,
          entryId: 'entry-1',
          evaluationId: 'eval-1',
          meanPayout: 1.25,
          ourRank: 2,
        },
      ],
    });
  }),

  http.get(`${BASE}/lab/models/:id`, () => {
    return HttpResponse.json(validModel);
  }),

  http.get(`${BASE}/lab/evaluations/:id/entries`, () => {
    return HttpResponse.json({
      items: [
        {
          id: 'result-1',
          entryName: 'Test Entry',
          meanNormalizedPayout: 1.25,
          pTop1: 0.15,
          pInMoney: 0.6,
          rank: 1,
        },
      ],
    });
  }),

  http.get(`${BASE}/lab/evaluations/:id`, () => {
    return HttpResponse.json(validEvaluation);
  }),

  http.get(`${BASE}/lab/evaluations`, () => {
    return HttpResponse.json({ items: [validEvaluation] });
  }),

  http.get(`${BASE}/lab/entry-results/:id`, () => {
    return HttpResponse.json({
      entryName: 'Test Entry',
      meanNormalizedPayout: 1.25,
      pTop1: 0.15,
      pInMoney: 0.6,
      rank: 1,
      totalBidPoints: 100,
      bids: [
        {
          teamId: 'team-1',
          schoolName: 'Duke',
          seed: 1,
          region: 'East',
          bidPoints: 20,
          ownership: 0.5,
        },
      ],
    });
  }),

  http.post(`${BASE}/lab/models/:id/pipeline/start`, () => {
    return HttpResponse.json({
      pipelineRunId: 'run-1',
      nCalcuttas: 2,
      status: 'running',
    });
  }),

  http.post(`${BASE}/lab/pipeline-runs/:id/cancel`, () => {
    return HttpResponse.json({ status: 'cancelled' });
  }),
];
