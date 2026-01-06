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
        return 'You do not have permission to view suites (403).';
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
    queryKey: ['suites', 'get', suiteId],
    queryFn: () => suitesService.get(suiteId!),
    enabled: Boolean(suiteId),
  });

  const suiteTitle = suiteQuery.data?.name ? `${suiteQuery.data.name}` : suiteId ? `Suite ${suiteId}` : 'Suite';

  const executionsQuery = useQuery({
    queryKey: ['suite-executions', 'list', suiteId],
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
    queryKey: ['suite-calcutta-evaluations', 'list', 'suite-execution', effectiveExecutionId],
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
            ← Back to Suites
          </Link>
        }
      />

      {!suiteId ? <Alert variant="error">Missing suite ID.</Alert> : null}

      {suiteId && suiteQuery.isLoading ? <LoadingState label="Loading suite..." /> : null}
      {suiteId && suiteQuery.isError ? (
        <Alert variant="error">
          <div className="font-semibold mb-1">Failed to load suite</div>
          <div className="mb-3">{showError(suiteQuery.error)}</div>
          <Button size="sm" onClick={() => suiteQuery.refetch()}>
            Retry
          </Button>
        </Alert>
      ) : null}

      {suiteQuery.data ? (
        <div className="space-y-6">
          <Card>
            <h2 className="text-xl font-semibold mb-4">Suite</h2>
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
                <div className="text-gray-500">Latest execution</div>
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
            <div className="flex items-center justify-between gap-4">
              <h2 className="text-xl font-semibold">Executions</h2>
              <div className="flex items-center gap-3">
                <div className="text-sm text-gray-500 whitespace-nowrap">Execution</div>
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

            {executionsQuery.isLoading ? <LoadingState label="Loading executions..." layout="inline" className="mt-3" /> : null}
            {executionsQuery.isError ? (
              <Alert variant="error" className="mt-3">
                <div className="font-semibold mb-1">Failed to load executions</div>
                <div className="mb-3">{showError(executionsQuery.error)}</div>
                <Button size="sm" onClick={() => executionsQuery.refetch()}>
                  Retry
                </Button>
              </Alert>
            ) : null}

            {!executionsQuery.isLoading && !executionsQuery.isError && executions.length === 0 ? (
              <Alert variant="info" className="mt-3">
                No suite executions found yet. Create one from the pipeline (or use the legacy Sandbox page to trigger individual evaluations).
              </Alert>
            ) : null}

            {effectiveExecutionId ? (
              <div className="mt-3 text-sm text-gray-600">
                Showing evaluations for execution <code className="text-gray-800">{effectiveExecutionId}</code>
              </div>
            ) : null}
          </Card>

          <Card>
            <h2 className="text-xl font-semibold mb-4">Evaluations</h2>

            {evaluationsQuery.isLoading ? <LoadingState label="Loading evaluations..." layout="inline" /> : null}
            {evaluationsQuery.isError ? (
              <Alert variant="error" className="mt-3">
                <div className="font-semibold mb-1">Failed to load evaluations</div>
                <div className="mb-3">{showError(evaluationsQuery.error)}</div>
                <Button size="sm" onClick={() => evaluationsQuery.refetch()}>
                  Retry
                </Button>
              </Alert>
            ) : null}

            {!evaluationsQuery.isLoading && !evaluationsQuery.isError && evals.length === 0 ? (
              <Alert variant="info" className="mt-3">
                No evaluations found for this execution.
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
                          aria-label={`Open evaluation ${it.id}`}
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
