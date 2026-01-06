import React, { useMemo } from 'react';
import { useQuery } from '@tanstack/react-query';
import { Link, useNavigate, useParams, useSearchParams } from 'react-router-dom';

import { ApiError } from '../api/apiClient';
import { Alert } from '../components/ui/Alert';
import { Button } from '../components/ui/Button';
import { Card } from '../components/ui/Card';
import { LoadingState } from '../components/ui/LoadingState';
import { PageContainer, PageHeader } from '../components/ui/Page';
import { Select } from '../components/ui/Select';
import { calcuttaService } from '../services/calcuttaService';
import { suiteCalcuttaEvaluationsService, type SuiteCalcuttaEvaluation } from '../services/suiteCalcuttaEvaluationsService';
import { suiteExecutionsService } from '../services/suiteExecutionsService';
import { suitesService } from '../services/suitesService';
import { syntheticCalcuttasService, type SyntheticCalcuttaListItem } from '../services/syntheticCalcuttasService';
import type { Calcutta } from '../types/calcutta';

export function SandboxSuiteDetailPage() {
  const navigate = useNavigate();
  const { suiteId } = useParams<{ suiteId: string }>();
  const [searchParams, setSearchParams] = useSearchParams();
  const selectedExecutionId = searchParams.get('executionId') || '';

  const { data: calcuttas = [] } = useQuery<Calcutta[]>({
    queryKey: ['calcuttas', 'all'],
    queryFn: calcuttaService.getAllCalcuttas,
  });

  const calcuttaNameById = useMemo(() => {
    const m = new Map<string, string>();
    for (const c of calcuttas) {
      m.set(c.id, c.name);
    }
    return m;
  }, [calcuttas]);

  const showError = (err: unknown) => {
    if (err instanceof ApiError) {
      if (err.status === 403) {
        return 'You do not have permission to view cohorts (403).';
      }
      return `Request failed (${err.status}): ${err.message}`;
    }
    return err instanceof Error ? err.message : 'Unknown error';
  };

  const formatDateTime = (v: string | null | undefined) => {
    if (!v) return '—';
    const d = new Date(v);
    if (Number.isNaN(d.getTime())) return v;
    return d.toLocaleString();
  };

  const formatFloat = (v: number | null | undefined, digits: number) => {
    if (v == null || Number.isNaN(v)) return '—';
    return v.toFixed(digits);
  };

  const seasonFromCalcuttaName = (name: string | null | undefined) => {
    if (!name) return '—';
    const m = name.match(/\b(19|20)\d{2}\b/);
    return m ? m[0] : '—';
  };

  const suiteQuery = useQuery({
    queryKey: ['synthetic-calcutta-cohorts', 'get', suiteId],
    queryFn: () => suitesService.get(suiteId!),
    enabled: Boolean(suiteId),
  });

  const suiteTitle = suiteQuery.data?.name ? `${suiteQuery.data.name}` : suiteId ? `Cohort ${suiteId}` : 'Cohort';

  const syntheticCalcuttasQuery = useQuery({
    queryKey: ['synthetic-calcuttas', 'list', suiteId],
    queryFn: () => syntheticCalcuttasService.list({ cohortId: suiteId!, limit: 200, offset: 0 }),
    enabled: Boolean(suiteId),
  });

  const syntheticCalcuttas: SyntheticCalcuttaListItem[] = syntheticCalcuttasQuery.data?.items ?? [];

  const executionsQuery = useQuery({
    queryKey: ['simulation-run-batches', 'list', suiteId],
    queryFn: () => suiteExecutionsService.list({ suiteId: suiteId!, limit: 200, offset: 0 }),
    enabled: Boolean(suiteId),
  });

  const executions = useMemo(() => executionsQuery.data?.items ?? [], [executionsQuery.data?.items]);

  const effectiveExecutionId = useMemo(() => {
    if (selectedExecutionId) return selectedExecutionId;
    if (suiteQuery.data?.latest_execution_id) return suiteQuery.data.latest_execution_id;
    return executions.length > 0 ? executions[0].id : '';
  }, [executions, selectedExecutionId, suiteQuery.data?.latest_execution_id]);

  const evaluationsQuery = useQuery({
    queryKey: ['simulation-runs', 'list', 'simulation-run-batch', effectiveExecutionId],
    queryFn: () =>
      suiteCalcuttaEvaluationsService.list({
        suiteExecutionId: effectiveExecutionId || undefined,
        limit: 200,
        offset: 0,
      }),
    enabled: Boolean(effectiveExecutionId),
  });

  const evals: SuiteCalcuttaEvaluation[] = evaluationsQuery.data?.items ?? [];

  return (
    <PageContainer className="max-w-none">
      <PageHeader
        title="Sandbox"
        subtitle={suiteTitle}
        leftActions={
          <Link to="/sandbox/suites" className="text-blue-600 hover:text-blue-800">
            ← Back to Cohorts
          </Link>
        }
      />

      {!suiteId ? <Alert variant="error">Missing cohort ID.</Alert> : null}

      {suiteId && suiteQuery.isLoading ? <LoadingState label="Loading cohort..." /> : null}
      {suiteId && suiteQuery.isError ? (
        <Alert variant="error">
          <div className="font-semibold mb-1">Failed to load cohort</div>
          <div className="mb-3">{showError(suiteQuery.error)}</div>
          <Button size="sm" onClick={() => suiteQuery.refetch()}>
            Retry
          </Button>
        </Alert>
      ) : null}

      {suiteQuery.data ? (
        <div className="space-y-6">
          <Card>
            <h2 className="text-xl font-semibold mb-4">Cohort</h2>
            <div className="grid grid-cols-1 md:grid-cols-2 gap-4 text-sm">
              <div>
                <div className="text-gray-500">Name</div>
                <div className="text-gray-900 font-medium">{suiteQuery.data.name || suiteQuery.data.id}</div>
              </div>
              <div>
                <div className="text-gray-500">Optimizer</div>
                <div className="text-gray-900">{suiteQuery.data.optimizer_key}</div>
              </div>
              <div>
                <div className="text-gray-500">nSims</div>
                <div className="text-gray-900">{suiteQuery.data.n_sims}</div>
              </div>
              <div>
                <div className="text-gray-500">Seed</div>
                <div className="text-gray-900">{suiteQuery.data.seed}</div>
              </div>
              <div>
                <div className="text-gray-500">Latest simulation run batch</div>
                <div className="text-gray-900">
                  {suiteQuery.data.latest_execution_id
                    ? `${suiteQuery.data.latest_execution_status ?? '—'} · ${suiteQuery.data.latest_execution_id.slice(0, 8)}`
                    : '—'}
                </div>
              </div>
              <div>
                <div className="text-gray-500">Updated</div>
                <div className="text-gray-900">{formatDateTime(suiteQuery.data.updated_at)}</div>
              </div>
            </div>
          </Card>

          <Card>
            <h2 className="text-xl font-semibold mb-4">Synthetic Calcuttas</h2>

            {syntheticCalcuttasQuery.isLoading ? <LoadingState label="Loading synthetic calcuttas..." layout="inline" /> : null}
            {syntheticCalcuttasQuery.isError ? (
              <Alert variant="error" className="mt-3">
                <div className="font-semibold mb-1">Failed to load synthetic calcuttas</div>
                <div className="mb-3">{showError(syntheticCalcuttasQuery.error)}</div>
                <Button size="sm" onClick={() => syntheticCalcuttasQuery.refetch()}>
                  Retry
                </Button>
              </Alert>
            ) : null}

            {!syntheticCalcuttasQuery.isLoading && !syntheticCalcuttasQuery.isError && syntheticCalcuttas.length === 0 ? (
              <Alert variant="info" className="mt-3">
                No synthetic calcuttas found for this cohort.
              </Alert>
            ) : null}

            {!syntheticCalcuttasQuery.isLoading && !syntheticCalcuttasQuery.isError && syntheticCalcuttas.length > 0 ? (
              <div className="overflow-x-auto">
                <table className="min-w-full divide-y divide-gray-200">
                  <thead className="bg-gray-50">
                    <tr>
                      <th className="px-3 py-2 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">Calcutta</th>
                      <th className="px-3 py-2 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">Highlighted Entry</th>
                      <th className="px-3 py-2 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">Created</th>
                    </tr>
                  </thead>
                  <tbody className="bg-white divide-y divide-gray-200">
                    {syntheticCalcuttas.map((sc) => {
                      const calcuttaName = calcuttaNameById.get(sc.calcutta_id) || sc.calcutta_id;
                      const highlighted = sc.focus_entry_name || '—';
                      return (
                        <tr key={sc.id} className="hover:bg-gray-50">
                          <td className="px-3 py-2 text-sm text-gray-900" title={calcuttaName}>
                            <div className="font-medium">{seasonFromCalcuttaName(calcuttaName)}</div>
                            <div className="text-xs text-gray-600">{calcuttaName}</div>
                          </td>
                          <td className="px-3 py-2 text-sm text-gray-700">{highlighted}</td>
                          <td className="px-3 py-2 text-sm text-gray-700">{formatDateTime(sc.created_at)}</td>
                        </tr>
                      );
                    })}
                  </tbody>
                </table>
              </div>
            ) : null}
          </Card>

          <Card>
            <div className="flex items-center justify-between gap-4">
              <h2 className="text-xl font-semibold">Simulation Run Batches</h2>
              <div className="flex items-center gap-3">
                <div className="text-sm text-gray-500 whitespace-nowrap">Run batch</div>
                <Select
                  value={effectiveExecutionId}
                  onChange={(e) => {
                    const next = e.target.value;
                    setSearchParams(next ? { executionId: next } : {}, { replace: true });
                  }}
                  disabled={executions.length === 0}
                >
                  {executions.map((ex) => (
                    <option key={ex.id} value={ex.id}>
                      {ex.status} · {ex.created_at.slice(0, 10)} · {ex.id.slice(0, 8)}
                    </option>
                  ))}
                </Select>
              </div>
            </div>

            {executionsQuery.isLoading ? <LoadingState label="Loading run batches..." layout="inline" className="mt-3" /> : null}
            {executionsQuery.isError ? (
              <Alert variant="error" className="mt-3">
                <div className="font-semibold mb-1">Failed to load run batches</div>
                <div className="mb-3">{showError(executionsQuery.error)}</div>
                <Button size="sm" onClick={() => executionsQuery.refetch()}>
                  Retry
                </Button>
              </Alert>
            ) : null}

            {!executionsQuery.isLoading && !executionsQuery.isError && executions.length === 0 ? (
              <Alert variant="info" className="mt-3">
                No simulation run batches found yet. Create one from the pipeline (or use the legacy Sandbox page to trigger individual runs).
              </Alert>
            ) : null}

            {effectiveExecutionId ? (
              <div className="mt-3 text-sm text-gray-600">
                Showing simulation runs for run batch <code className="text-gray-800">{effectiveExecutionId}</code>
              </div>
            ) : null}
          </Card>

          <Card>
            <h2 className="text-xl font-semibold mb-4">Simulation Runs</h2>

            {evaluationsQuery.isLoading ? <LoadingState label="Loading simulation runs..." layout="inline" /> : null}
            {evaluationsQuery.isError ? (
              <Alert variant="error" className="mt-3">
                <div className="font-semibold mb-1">Failed to load simulation runs</div>
                <div className="mb-3">{showError(evaluationsQuery.error)}</div>
                <Button size="sm" onClick={() => evaluationsQuery.refetch()}>
                  Retry
                </Button>
              </Alert>
            ) : null}

            {!evaluationsQuery.isLoading && !evaluationsQuery.isError && evals.length === 0 ? (
              <Alert variant="info" className="mt-3">
                No simulation runs found for this run batch.
              </Alert>
            ) : null}

            {!evaluationsQuery.isLoading && !evaluationsQuery.isError && evals.length > 0 ? (
              <div className="overflow-x-auto">
                <table className="min-w-full divide-y divide-gray-200">
                  <thead className="bg-gray-50">
                    <tr>
                      <th className="px-3 py-2 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">Season</th>
                      <th className="px-3 py-2 text-right text-xs font-medium text-gray-500 uppercase tracking-wider">Rank</th>
                      <th className="px-3 py-2 text-right text-xs font-medium text-gray-500 uppercase tracking-wider">Mean Norm Payout</th>
                      <th className="px-3 py-2 text-right text-xs font-medium text-gray-500 uppercase tracking-wider">P(Top 1)</th>
                      <th className="px-3 py-2 text-right text-xs font-medium text-gray-500 uppercase tracking-wider">P(In Money)</th>
                      <th className="px-3 py-2 text-right text-xs font-medium text-gray-500 uppercase tracking-wider">Finish</th>
                    </tr>
                  </thead>
                  <tbody className="bg-white divide-y divide-gray-200">
                    {evals.map((it) => {
                      const detailUrl = `/sandbox/evaluations/${encodeURIComponent(it.id)}?suiteId=${encodeURIComponent(
                        suiteId || ''
                      )}${effectiveExecutionId ? `&executionId=${encodeURIComponent(effectiveExecutionId)}` : ''}`;

                      const calcuttaName = calcuttaNameById.get(it.calcutta_id) || it.calcutta_id;
                      const season = seasonFromCalcuttaName(calcuttaName);
                      const finish =
                        it.realized_finish_position != null
                          ? `${it.realized_finish_position}${it.realized_is_tied ? ' (tied)' : ''}`
                          : '—';

                      return (
                        <tr
                          key={it.id}
                          className="hover:bg-gray-50 cursor-pointer focus-visible:outline focus-visible:outline-2 focus-visible:outline-blue-500"
                          role="link"
                          tabIndex={0}
                          onClick={() => navigate(detailUrl)}
                          onKeyDown={(e) => {
                            if (e.key === 'Enter' || e.key === ' ') {
                              e.preventDefault();
                              navigate(detailUrl);
                            }
                          }}
                          aria-label={`Open simulation run ${it.id}`}
                        >
                          <td className="px-3 py-2 text-sm text-gray-900" title={calcuttaName}>
                            <div className="font-medium">{season}</div>
                            <div className="text-xs text-gray-500">
                              {it.optimizer_key} · n={it.n_sims} · seed={it.seed}
                            </div>
                          </td>

                          <td className="px-3 py-2 text-sm text-right text-gray-700">{it.our_rank ?? '—'}</td>
                          <td className="px-3 py-2 text-sm text-right text-gray-700">{formatFloat(it.our_mean_normalized_payout, 4)}</td>
                          <td className="px-3 py-2 text-sm text-right text-gray-700">{formatFloat(it.our_p_top1, 4)}</td>
                          <td className="px-3 py-2 text-sm text-right text-gray-700">{formatFloat(it.our_p_in_money, 4)}</td>
                          <td className="px-3 py-2 text-sm text-right text-gray-700">{finish}</td>
                        </tr>
                      );
                    })}
                  </tbody>
                </table>
              </div>
            ) : null}
          </Card>
        </div>
      ) : null}
    </PageContainer>
  );
}
