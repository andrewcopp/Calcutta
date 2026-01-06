import React, { useMemo, useState } from 'react';
import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query';
import { useNavigate, useSearchParams } from 'react-router-dom';

import { ApiError } from '../api/apiClient';
import { Alert } from '../components/ui/Alert';
import { Button } from '../components/ui/Button';
import { Card } from '../components/ui/Card';
import { LoadingState } from '../components/ui/LoadingState';
import { PageContainer, PageHeader } from '../components/ui/Page';
import { Select } from '../components/ui/Select';
import { calcuttaService } from '../services/calcuttaService';
import { analyticsService } from '../services/analyticsService';
import { modelCatalogsService } from '../services/modelCatalogsService';
import {
  suiteCalcuttaEvaluationsService,
  type CreateSuiteCalcuttaEvaluationRequest,
  type SuiteCalcuttaEvaluation,
} from '../services/suiteCalcuttaEvaluationsService';
import type { Calcutta } from '../types/calcutta';

export function SandboxPage() {
  const navigate = useNavigate();
  const queryClient = useQueryClient();
  const [searchParams, setSearchParams] = useSearchParams();
  const selectedSuiteId = searchParams.get('cohortId') || searchParams.get('suiteId') || '';

  const [createCalcuttaId, setCreateCalcuttaId] = useState<string>('');
  const [createSuiteId, setCreateSuiteId] = useState<string>(() => selectedSuiteId);
  const [createSuiteName, setCreateSuiteName] = useState<string>('');
  const [createOptimizerKey, setCreateOptimizerKey] = useState<string>('');
  const [createOptimizerSelection, setCreateOptimizerSelection] = useState<string>('');
  const [createNSims, setCreateNSims] = useState<number>(25000);
  const [createSeed, setCreateSeed] = useState<number>(42);
  const [createStartingStateKey, setCreateStartingStateKey] = useState<string>('post_first_four');
  const [createExcludedEntryName, setCreateExcludedEntryName] = useState<string>('');

  const { data: calcuttas = [] } = useQuery<Calcutta[]>({
    queryKey: ['calcuttas', 'all'],
    queryFn: calcuttaService.getAllCalcuttas,
  });

  const latestRunsQuery = useQuery<{ tournament_id: string; game_outcome_run_id?: string | null; market_share_run_id?: string | null } | null>({
    queryKey: ['analytics', 'latest-prediction-runs', createCalcuttaId],
    queryFn: async () => {
      if (!createCalcuttaId) return null;
      return analyticsService.getLatestPredictionRunsForCalcutta<{
        tournament_id: string;
        game_outcome_run_id?: string | null;
        market_share_run_id?: string | null;
      }>(createCalcuttaId);
    },
    enabled: Boolean(createCalcuttaId),
  });

  const entryOptimizersQuery = useQuery({
    queryKey: ['model-catalogs', 'entry-optimizers'],
    queryFn: () => modelCatalogsService.listEntryOptimizers(),
  });

  const calcuttaNameById = useMemo(() => {
    const m = new Map<string, string>();
    for (const c of calcuttas) {
      m.set(c.id, c.name);
    }
    return m;
  }, [calcuttas]);

  const allEvaluationsQuery = useQuery({
    queryKey: ['simulation-runs', 'list', 'all'],
    queryFn: () => suiteCalcuttaEvaluationsService.list({ limit: 200, offset: 0 }),
  });

  const listQuery = useQuery({
    queryKey: ['simulation-runs', 'list', selectedSuiteId],
    queryFn: () => suiteCalcuttaEvaluationsService.list({ suiteId: selectedSuiteId || undefined, limit: 200, offset: 0 }),
  });

  const items: SuiteCalcuttaEvaluation[] = listQuery.data?.items ?? [];

  const suites = useMemo(() => {
    const byId = new Map<string, { id: string; name: string }>();
    const allItems: SuiteCalcuttaEvaluation[] = allEvaluationsQuery.data?.items ?? [];
    for (const it of allItems) {
      if (!byId.has(it.suite_id)) {
        byId.set(it.suite_id, { id: it.suite_id, name: it.suite_name || it.suite_id });
      }
    }
    return Array.from(byId.values()).sort((a, b) => a.name.localeCompare(b.name));
  }, [allEvaluationsQuery.data?.items]);

  const createMutation = useMutation({
    mutationFn: async (req: CreateSuiteCalcuttaEvaluationRequest) => suiteCalcuttaEvaluationsService.create(req),
    onSuccess: async (res) => {
      await queryClient.invalidateQueries({ queryKey: ['simulation-runs'] });
      navigate(`/sandbox/evaluations/${encodeURIComponent(res.id)}${selectedSuiteId ? `?cohortId=${encodeURIComponent(selectedSuiteId)}` : ''}`);
    },
  });

  const showError = (err: unknown) => {
    if (err instanceof ApiError) {
      if (err.status === 403) {
        return 'You do not have permission to view simulation runs (403).';
      }
      return `Request failed (${err.status}): ${err.message}`;
    }
    return err instanceof Error ? err.message : 'Unknown error';
  };

  const formatDateTime = (v: string) => {
    const d = new Date(v);
    if (Number.isNaN(d.getTime())) return v;
    return d.toLocaleString();
  };

  const formatFloat = (v: number | null | undefined, digits: number) => {
    if (v == null || Number.isNaN(v)) return '—';
    return v.toFixed(digits);
  };

  const formatCurrency = (cents: number | null | undefined) => {
    if (cents == null || Number.isNaN(cents)) return '—';
    return `$${(cents / 100).toFixed(2)}`;
  };

  return (
    <PageContainer>
      <PageHeader title="Sandbox" subtitle="Browse historical simulation runs and drill into results." />

      <Card className="mb-6">
        <h2 className="text-xl font-semibold mb-4">Trigger Simulation Run</h2>

        <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
          <div>
            <div className="text-sm text-gray-500 mb-1">Calcutta</div>
            <Select
              value={createCalcuttaId}
              onChange={(e) => setCreateCalcuttaId(e.target.value)}
              className="w-full"
            >
              <option value="">-- Select calcutta --</option>
              {calcuttas
                .slice()
                .sort((a, b) => a.name.localeCompare(b.name))
                .map((c) => (
                  <option key={c.id} value={c.id}>
                    {c.name}
                  </option>
                ))}
            </Select>
            {createCalcuttaId && latestRunsQuery.isLoading ? (
              <LoadingState label="Loading latest runs..." layout="inline" className="mt-1" size="sm" />
            ) : null}
            {createCalcuttaId && !latestRunsQuery.isLoading && latestRunsQuery.data ? (
              <div className="text-xs text-gray-600 mt-1">
                latest: go={latestRunsQuery.data.game_outcome_run_id ? latestRunsQuery.data.game_outcome_run_id.slice(0, 8) : '—'} · ms=
                {latestRunsQuery.data.market_share_run_id ? latestRunsQuery.data.market_share_run_id.slice(0, 8) : '—'}
              </div>
            ) : null}
          </div>

          <div>
            <div className="text-sm text-gray-500 mb-1">Cohort (optional)</div>
            <Select
              value={createSuiteId}
              onChange={(e) => setCreateSuiteId(e.target.value)}
              className="w-full"
            >
              <option value="">-- Create/resolve by name --</option>
              {suites.map((s) => (
                <option key={s.id} value={s.id}>
                  {s.name}
                </option>
              ))}
            </Select>
            <div className="text-xs text-gray-600 mt-1">If empty, you must specify cohort name + optimizer key.</div>
          </div>

          {!createSuiteId ? (
            <>
              <div>
                <div className="text-sm text-gray-500 mb-1">Cohort name</div>
                <input
                  value={createSuiteName}
                  onChange={(e) => setCreateSuiteName(e.target.value)}
                  className="w-full border border-gray-300 rounded px-3 py-2 text-sm"
                  placeholder="e.g. 2026-latest"
                />
              </div>
              <div>
                <div className="text-sm text-gray-500 mb-1">Optimizer key</div>
                {!entryOptimizersQuery.isLoading && !entryOptimizersQuery.isError && (entryOptimizersQuery.data?.items?.length ?? 0) > 0 ? (
                  <>
                    <Select
                      value={createOptimizerSelection}
                      onChange={(e) => {
                        const v = e.target.value;
                        setCreateOptimizerSelection(v);
                        if (v === '__custom__') {
                          setCreateOptimizerKey('');
                        } else {
                          setCreateOptimizerKey(v);
                        }
                      }}
                      className="w-full"
                    >
                      <option value="">-- Select optimizer --</option>
                      {entryOptimizersQuery.data!.items
                        .slice()
                        .sort((a, b) => a.display_name.localeCompare(b.display_name))
                        .map((m) => (
                          <option key={m.id} value={m.id}>
                            {m.display_name}{m.deprecated ? ' (deprecated)' : ''}
                          </option>
                        ))}
                      <option value="__custom__">Custom…</option>
                    </Select>
                    {createOptimizerSelection === '__custom__' ? (
                      <input
                        value={createOptimizerKey}
                        onChange={(e) => setCreateOptimizerKey(e.target.value)}
                        className="w-full border border-gray-300 rounded px-3 py-2 text-sm mt-2"
                        placeholder="e.g. minlp_v1"
                      />
                    ) : null}
                  </>
                ) : (
                  <input
                    value={createOptimizerKey}
                    onChange={(e) => setCreateOptimizerKey(e.target.value)}
                    className="w-full border border-gray-300 rounded px-3 py-2 text-sm"
                    placeholder="e.g. minlp_v1"
                  />
                )}
              </div>
            </>
          ) : null}

          <div>
            <div className="text-sm text-gray-500 mb-1">nSims</div>
            <input
              type="number"
              value={createNSims}
              onChange={(e) => setCreateNSims(Number(e.target.value))}
              className="w-full border border-gray-300 rounded px-3 py-2 text-sm"
              min={1}
            />
          </div>
          <div>
            <div className="text-sm text-gray-500 mb-1">Seed</div>
            <input
              type="number"
              value={createSeed}
              onChange={(e) => setCreateSeed(Number(e.target.value))}
              className="w-full border border-gray-300 rounded px-3 py-2 text-sm"
            />
          </div>
          <div>
            <div className="text-sm text-gray-500 mb-1">Starting state</div>
            <Select
              value={createStartingStateKey}
              onChange={(e) => setCreateStartingStateKey(e.target.value)}
              className="w-full"
            >
              <option value="post_first_four">post_first_four</option>
              <option value="current">current</option>
            </Select>
          </div>
          <div>
            <div className="text-sm text-gray-500 mb-1">Excluded entry name (optional)</div>
            <input
              value={createExcludedEntryName}
              onChange={(e) => setCreateExcludedEntryName(e.target.value)}
              className="w-full border border-gray-300 rounded px-3 py-2 text-sm"
              placeholder="e.g. Andrew"
            />
          </div>
        </div>

        <div className="mt-4 flex items-center gap-3">
          <button
            className="px-4 py-2 rounded bg-blue-600 text-white text-sm disabled:bg-gray-300"
            disabled={
              createMutation.isPending ||
              !createCalcuttaId ||
              !latestRunsQuery.data?.game_outcome_run_id ||
              !latestRunsQuery.data?.market_share_run_id ||
              (!createSuiteId && (!createSuiteName.trim() || !createOptimizerKey.trim()))
            }
            onClick={() => {
              const latest = latestRunsQuery.data;
              if (!latest?.game_outcome_run_id || !latest?.market_share_run_id) return;

              const req: CreateSuiteCalcuttaEvaluationRequest = {
                calcuttaId: createCalcuttaId,
                suiteId: createSuiteId || undefined,
                suiteName: !createSuiteId ? createSuiteName.trim() : undefined,
                optimizerKey: !createSuiteId ? createOptimizerKey.trim() : undefined,
                gameOutcomeRunId: latest.game_outcome_run_id,
                marketShareRunId: latest.market_share_run_id,
                nSims: createNSims,
                seed: createSeed,
                startingStateKey: createStartingStateKey,
                excludedEntryName: createExcludedEntryName.trim() ? createExcludedEntryName.trim() : undefined,
              };

              createMutation.mutate(req);
            }}
          >
            Trigger
          </button>

          {createMutation.isPending ? <LoadingState label="Creating..." layout="inline" size="sm" /> : null}
          {createMutation.isError ? (
            <Alert variant="error" className="ml-2">
              {showError(createMutation.error)}
            </Alert>
          ) : null}
        </div>

        <div className="mt-3 text-sm text-gray-600">
          This action submits a new simulation run request (async) to <code className="text-gray-800">/api/simulation-runs</code>.
        </div>
      </Card>

      <Card className="mb-6">
        <div className="flex items-center gap-4">
          <label htmlFor="suite-select" className="text-lg font-semibold whitespace-nowrap">
            Cohort:
          </label>
          <Select
            id="suite-select"
            value={selectedSuiteId}
            onChange={(e) => {
              const suiteId = e.target.value;
              setSearchParams(suiteId ? { cohortId: suiteId } : {}, { replace: true });
            }}
            className="flex-1 max-w-2xl"
          >
            <option value="">All cohorts</option>
            {suites.map((s) => (
              <option key={s.id} value={s.id}>
                {s.name}
              </option>
            ))}
          </Select>
        </div>
        <div className="mt-3 text-sm text-gray-600">
          This view is backed by <code className="text-gray-800">/api/simulation-runs</code> and shows each simulation run.
        </div>
      </Card>

      <Card>
        <h2 className="text-xl font-semibold mb-4">Simulation Runs</h2>

        {listQuery.isLoading ? <LoadingState label="Loading simulation runs..." layout="inline" /> : null}
        {listQuery.isError ? (
          <Alert variant="error" className="mt-3">
            <div className="font-semibold mb-1">Failed to load simulation runs</div>
            <div className="mb-3">{showError(listQuery.error)}</div>
            <Button size="sm" onClick={() => listQuery.refetch()}>
              Retry
            </Button>
          </Alert>
        ) : null}

        {!listQuery.isLoading && !listQuery.isError && items.length === 0 ? (
          <Alert variant="info" className="mt-3">
            No simulation runs found.
          </Alert>
        ) : null}

        {!listQuery.isLoading && !listQuery.isError && items.length > 0 ? (
          <div className="overflow-x-auto">
            <table className="min-w-full divide-y divide-gray-200">
              <thead className="bg-gray-50">
                <tr>
                  <th className="px-3 py-2 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">Cohort</th>
                  <th className="px-3 py-2 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">Calcutta</th>
                  <th className="px-3 py-2 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">Status</th>
                  <th className="px-3 py-2 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">Created</th>
                </tr>
              </thead>
              <tbody className="bg-white divide-y divide-gray-200">
                {items.map((it) => {
                  const detailUrl = `/sandbox/evaluations/${encodeURIComponent(it.id)}${
                    selectedSuiteId ? `?cohortId=${encodeURIComponent(selectedSuiteId)}` : ''
                  }`;

                  const showHeadline = it.status === 'succeeded' && it.our_mean_normalized_payout != null;

                  const showRealized =
                    it.status === 'succeeded' &&
                    (it.realized_finish_position != null || it.realized_total_points != null || it.realized_payout_cents != null);

                  return (
                    <tr
                      key={it.id}
                      className="hover:bg-gray-50 cursor-pointer"
                      onClick={() => navigate(detailUrl)}
                    >
                      <td className="px-3 py-2 text-sm text-gray-900">
                        <div className="font-medium">{it.suite_name || it.suite_id}</div>
                        <div className="text-xs text-gray-600">
                          {it.optimizer_key} · n={it.n_sims} · seed={it.seed}
                        </div>
                        {showHeadline ? (
                          <div className="text-xs text-gray-600 mt-1">
                            our: rank={it.our_rank ?? '—'} · mean={formatFloat(it.our_mean_normalized_payout, 4)} · pTop1=
                            {formatFloat(it.our_p_top1, 4)} · pInMoney={formatFloat(it.our_p_in_money, 4)}
                          </div>
                        ) : null}

                        {showRealized ? (
                          <div className="text-xs text-gray-600 mt-1">
                            realized: finish={it.realized_finish_position ?? '—'}
                            {it.realized_is_tied ? ' (tied)' : ''} · payout={formatCurrency(it.realized_payout_cents)} · points=
                            {formatFloat(it.realized_total_points, 2)}
                            {it.realized_in_the_money != null ? ` · ITM=${it.realized_in_the_money ? 'yes' : 'no'}` : ''}
                          </div>
                        ) : null}
                      </td>
                      <td className="px-3 py-2 text-sm text-gray-700">{calcuttaNameById.get(it.calcutta_id) || it.calcutta_id}</td>
                      <td className="px-3 py-2 text-sm text-gray-700">{it.status}</td>
                      <td className="px-3 py-2 text-sm text-gray-700">{formatDateTime(it.created_at)}</td>
                    </tr>
                  );
                })}
              </tbody>
            </table>
          </div>
        ) : null}
      </Card>
    </PageContainer>
  );
}
