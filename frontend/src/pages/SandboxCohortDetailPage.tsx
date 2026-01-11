import React, { useMemo, useState } from 'react';
import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query';
import { Link, useNavigate, useParams } from 'react-router-dom';

import { ApiError } from '../api/apiClient';
import { Alert } from '../components/ui/Alert';
import { Button } from '../components/ui/Button';
import { Card } from '../components/ui/Card';
import { LoadingState } from '../components/ui/LoadingState';
import { PageContainer, PageHeader } from '../components/ui/Page';
import { Select } from '../components/ui/Select';
import { calcuttaService } from '../services/calcuttaService';
import { cohortsService } from '../services/cohortsService';
import { simulationRunsService, type ListSimulationRunsResponse, type SimulationRun } from '../services/simulationRunsService';
import { simulatedCalcuttasService, type SimulatedCalcuttaListItem } from '../services/simulatedCalcuttasService';
import type { Calcutta } from '../types/calcutta';

export function SandboxCohortDetailPage() {
  const { cohortId } = useParams<{ cohortId?: string }>();
  const queryClient = useQueryClient();
  const navigate = useNavigate();
  const [createFromCalcuttaId, setCreateFromCalcuttaId] = useState<string>('');

  const { data: calcuttas = [] } = useQuery<Calcutta[]>({
    queryKey: ['calcuttas', 'all'],
    queryFn: calcuttaService.getAllCalcuttas,
  });

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

  const fmtFloat = (v: number | null | undefined, digits: number) => {
    if (v == null || !Number.isFinite(v)) return '—';
    return v.toFixed(digits);
  };

  const fmtPct = (v: number | null | undefined) => {
    if (v == null || !Number.isFinite(v)) return '—';
    return `${(v * 100).toFixed(1)}%`;
  };

  const fmtFinish = (pos?: number | null, tied?: boolean | null) => {
    if (pos == null) return '—';
    return `${pos}${tied ? ' (tied)' : ''}`;
  };

  const seasonFromCalcuttaName = (name: string | null | undefined) => {
    if (!name) return '—';
    const m = name.match(/\b(19|20)\d{2}\b/);
    return m ? m[0] : '—';
  };

  const effectiveCohortId = cohortId || '';

  const effectiveCreateFromCalcuttaId = useMemo(() => {
    if (createFromCalcuttaId) return createFromCalcuttaId;
    return calcuttas.length > 0 ? calcuttas[0].id : '';
  }, [calcuttas, createFromCalcuttaId]);

  const calcuttaById = useMemo(() => {
    const m = new Map<string, Calcutta>();
    for (const c of calcuttas) m.set(c.id, c);
    return m;
  }, [calcuttas]);

  const cohortQuery = useQuery({
    queryKey: ['cohorts', 'get', effectiveCohortId],
    queryFn: () => cohortsService.get(effectiveCohortId),
    enabled: Boolean(effectiveCohortId),
  });

  const cohortTitle = cohortQuery.data?.name ? `${cohortQuery.data.name}` : effectiveCohortId ? `Cohort ${effectiveCohortId}` : 'Cohort';

  const simulatedCalcuttasQuery = useQuery({
    queryKey: ['simulated-calcuttas', 'list', effectiveCohortId],
    queryFn: async () => {
      if (!effectiveCohortId) return { items: [] as SimulatedCalcuttaListItem[] };
      return simulatedCalcuttasService.list({
        cohortId: effectiveCohortId,
        limit: 200,
        offset: 0,
      });
    },
    enabled: Boolean(effectiveCohortId),
  });

  const runsQuery = useQuery<ListSimulationRunsResponse>({
    queryKey: ['simulation-runs', 'list', effectiveCohortId],
    queryFn: async () => {
      if (!effectiveCohortId) return { items: [] };
      return simulationRunsService.list({ cohortId: effectiveCohortId, limit: 200, offset: 0 });
    },
    enabled: Boolean(effectiveCohortId),
  });

  const simulatedCalcuttas: SimulatedCalcuttaListItem[] = useMemo(
    () => simulatedCalcuttasQuery.data?.items ?? [],
    [simulatedCalcuttasQuery.data?.items]
  );

  const latestRunBySimulatedCalcuttaId = useMemo(() => {
    const m = new Map<string, SimulationRun>();
    const items = runsQuery.data?.items ?? [];
    // items are ordered newest-first by the API, so first occurrence wins
    for (const r of items) {
      const scID = (r.simulated_calcutta_id || '').trim();
      if (!scID) continue;
      if (!m.has(scID)) m.set(scID, r);
    }
    return m;
  }, [runsQuery.data?.items]);

  const createSimulatedCalcuttaMutation = useMutation({
    mutationFn: async () => {
      if (!effectiveCohortId) throw new Error('Missing cohort ID');
      const src = effectiveCreateFromCalcuttaId;
      if (!src) throw new Error('Missing source calcutta');
      return simulatedCalcuttasService.createFromCalcutta({
        calcuttaId: src,
        startingStateKey: 'post_first_four',
        metadata: { cohort_id: effectiveCohortId },
      });
    },
    onSuccess: async () => {
      await queryClient.invalidateQueries({ queryKey: ['simulated-calcuttas', 'list', effectiveCohortId] });
    },
  });

  return (
    <PageContainer className="max-w-none">
      <PageHeader
        title="Sandbox"
        subtitle={cohortTitle}
        leftActions={
          <Link to="/sandbox/cohorts" className="text-blue-600 hover:text-blue-800">
            ← Back to Cohorts
          </Link>
        }
      />

      {!effectiveCohortId ? <Alert variant="error">Missing cohort ID.</Alert> : null}

      {effectiveCohortId && cohortQuery.isLoading ? <LoadingState label="Loading cohort..." /> : null}
      {effectiveCohortId && cohortQuery.isError ? (
        <Alert variant="error">
          <div className="font-semibold mb-1">Failed to load cohort</div>
          <div className="mb-3">{showError(cohortQuery.error)}</div>
          <Button size="sm" onClick={() => cohortQuery.refetch()}>
            Retry
          </Button>
        </Alert>
      ) : null}

      {cohortQuery.data ? (
        <div className="space-y-6">
          <Card>
            <h2 className="text-xl font-semibold mb-4">Cohort</h2>
            <div className="grid grid-cols-1 md:grid-cols-2 gap-4 text-sm">
              <div>
                <div className="text-gray-500">Name</div>
                <div className="text-gray-900 font-medium">{cohortQuery.data.name || cohortQuery.data.id}</div>
              </div>
              <div>
                <div className="text-gray-500">Optimizer</div>
                <div className="text-gray-900">{cohortQuery.data.optimizer_key}</div>
              </div>
              <div>
                <div className="text-gray-500">nSims</div>
                <div className="text-gray-900">{cohortQuery.data.n_sims}</div>
              </div>
              <div>
                <div className="text-gray-500">Seed</div>
                <div className="text-gray-900">{cohortQuery.data.seed}</div>
              </div>
              <div>
                <div className="text-gray-500">Updated</div>
                <div className="text-gray-900">{formatDateTime(cohortQuery.data.updated_at)}</div>
              </div>
            </div>
          </Card>

          <Card>
            <div className="flex items-end justify-between gap-4 mb-4">
              <h2 className="text-xl font-semibold">Simulated Calcuttas</h2>

              <div className="flex items-center gap-3">
                <div className="text-sm text-gray-500 whitespace-nowrap">Create from</div>
                <Select
                  value={effectiveCreateFromCalcuttaId}
                  onChange={(e) => setCreateFromCalcuttaId(e.target.value)}
                  disabled={calcuttas.length === 0}
                >
                  {calcuttas.map((c) => (
                    <option key={c.id} value={c.id}>
                      {seasonFromCalcuttaName(c.name)} · {c.name}
                    </option>
                  ))}
                </Select>
                <Button
                  size="sm"
                  disabled={createSimulatedCalcuttaMutation.isPending || !effectiveCohortId || !effectiveCreateFromCalcuttaId}
                  onClick={() => createSimulatedCalcuttaMutation.mutate()}
                >
                  Create
                </Button>
              </div>
            </div>

            {simulatedCalcuttasQuery.isLoading ? <LoadingState label="Loading simulated calcuttas..." layout="inline" /> : null}

            {simulatedCalcuttasQuery.isError ? (
              <Alert variant="error" className="mt-3">
                <div className="font-semibold mb-1">Failed to load simulated calcuttas</div>
                <div className="mb-3">{showError(simulatedCalcuttasQuery.error)}</div>
                <Button size="sm" onClick={() => simulatedCalcuttasQuery.refetch()}>
                  Retry
                </Button>
              </Alert>
            ) : null}

            {!simulatedCalcuttasQuery.isLoading && !simulatedCalcuttasQuery.isError && simulatedCalcuttas.length === 0 ? (
              <Alert variant="info" className="mt-3">
                No simulated calcuttas found for this cohort.
              </Alert>
            ) : null}

            {createSimulatedCalcuttaMutation.isError ? (
              <Alert variant="error" className="mt-3">
                {showError(createSimulatedCalcuttaMutation.error)}
              </Alert>
            ) : null}

            {!simulatedCalcuttasQuery.isLoading && !simulatedCalcuttasQuery.isError && simulatedCalcuttas.length > 0 ? (
              <div className="overflow-x-auto">
                <table className="min-w-full divide-y divide-gray-200">
                  <thead className="bg-gray-50">
                    <tr>
                      <th className="px-3 py-2 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">Name</th>
                      <th className="px-3 py-2 text-right text-xs font-medium text-gray-500 uppercase tracking-wider">Normalized Mean Payout</th>
                      <th className="px-3 py-2 text-right text-xs font-medium text-gray-500 uppercase tracking-wider">P(1st)</th>
                      <th className="px-3 py-2 text-right text-xs font-medium text-gray-500 uppercase tracking-wider">P(In Money)</th>
                      <th className="px-3 py-2 text-right text-xs font-medium text-gray-500 uppercase tracking-wider">Real Finish</th>
                    </tr>
                  </thead>
                  <tbody className="bg-white divide-y divide-gray-200">
                    {simulatedCalcuttas.map((sc) => {
                      const href = `/sandbox/simulated-calcuttas/${encodeURIComponent(sc.id)}?cohortId=${encodeURIComponent(effectiveCohortId)}`;
                      const base = sc.base_calcutta_id ? calcuttaById.get(sc.base_calcutta_id) : null;
                      const run = latestRunBySimulatedCalcuttaId.get(sc.id);
                      return (
                        <tr
                          key={sc.id}
                          className="hover:bg-gray-50 cursor-pointer"
                          onClick={() => navigate(href)}
                        >
                          <td className="px-3 py-2 text-sm text-gray-900" title={sc.name}>
                            <div className="font-medium">{sc.name}</div>
                            {base ? <div className="text-xs text-gray-600">base {base.name}</div> : sc.base_calcutta_id ? <div className="text-xs text-gray-600">base {sc.base_calcutta_id.slice(0, 8)}</div> : null}
                          </td>
                          <td className="px-3 py-2 text-sm text-right text-gray-900 font-medium">{fmtFloat(run?.our_mean_normalized_payout, 4)}</td>
                          <td className="px-3 py-2 text-sm text-right text-gray-700">{fmtPct(run?.our_p_top1)}</td>
                          <td className="px-3 py-2 text-sm text-right text-gray-700">{fmtPct(run?.our_p_in_money)}</td>
                          <td className="px-3 py-2 text-sm text-right text-gray-700">{fmtFinish(run?.realized_finish_position, run?.realized_is_tied)}</td>
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
