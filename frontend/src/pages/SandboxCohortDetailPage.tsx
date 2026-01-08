import React, { useMemo, useState } from 'react';
import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query';
import { Link, useParams } from 'react-router-dom';

import { ApiError } from '../api/apiClient';
import { Alert } from '../components/ui/Alert';
import { Button } from '../components/ui/Button';
import { Card } from '../components/ui/Card';
import { LoadingState } from '../components/ui/LoadingState';
import { PageContainer, PageHeader } from '../components/ui/Page';
import { Select } from '../components/ui/Select';
import { calcuttaService } from '../services/calcuttaService';
import { cohortsService } from '../services/cohortsService';
import { syntheticCalcuttasService, type SyntheticCalcuttaListItem } from '../services/syntheticCalcuttasService';
import type { Calcutta } from '../types/calcutta';

export function SandboxCohortDetailPage() {
  const { cohortId } = useParams<{ cohortId?: string }>();
  const queryClient = useQueryClient();
  const [sourceCalcuttaId, setSourceCalcuttaId] = useState<string>('');

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

  const effectiveCohortId = cohortId || '';

  const cohortQuery = useQuery({
    queryKey: ['cohorts', 'get', effectiveCohortId],
    queryFn: () => cohortsService.get(effectiveCohortId),
    enabled: Boolean(effectiveCohortId),
  });

  const cohortTitle = cohortQuery.data?.name ? `${cohortQuery.data.name}` : effectiveCohortId ? `Cohort ${effectiveCohortId}` : 'Cohort';

  const syntheticCalcuttasQuery = useQuery({
    queryKey: ['synthetic-calcuttas', 'list', effectiveCohortId],
    queryFn: () => syntheticCalcuttasService.list({ cohortId: effectiveCohortId, limit: 200, offset: 0 }),
    enabled: Boolean(effectiveCohortId),
  });

  const syntheticCalcuttas: SyntheticCalcuttaListItem[] = useMemo(
    () => syntheticCalcuttasQuery.data?.items ?? [],
    [syntheticCalcuttasQuery.data?.items]
  );

  const effectiveSourceCalcuttaId = useMemo(() => {
    if (sourceCalcuttaId) return sourceCalcuttaId;
    return calcuttas.length > 0 ? calcuttas[0].id : '';
  }, [calcuttas, sourceCalcuttaId]);

  const createSyntheticCalcuttaMutation = useMutation({
    mutationFn: async () => {
      if (!effectiveCohortId) throw new Error('Missing cohort ID');
      const src = effectiveSourceCalcuttaId;
      if (!src) throw new Error('Missing source calcutta');
      return syntheticCalcuttasService.create({ cohortId: effectiveCohortId, sourceCalcuttaId: src });
    },
    onSuccess: async () => {
      await queryClient.invalidateQueries({ queryKey: ['synthetic-calcuttas', 'list', effectiveCohortId] });
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
              <h2 className="text-xl font-semibold">Synthetic Calcuttas</h2>

              <div className="flex items-center gap-3">
                <div className="text-sm text-gray-500 whitespace-nowrap">Source Calcutta</div>
                <Select value={effectiveSourceCalcuttaId} onChange={(e) => setSourceCalcuttaId(e.target.value)} disabled={calcuttas.length === 0}>
                  {calcuttas.map((c) => (
                    <option key={c.id} value={c.id}>
                      {seasonFromCalcuttaName(c.name)} · {c.name}
                    </option>
                  ))}
                </Select>
                <Button
                  size="sm"
                  disabled={createSyntheticCalcuttaMutation.isPending || !effectiveCohortId || !effectiveSourceCalcuttaId}
                  onClick={() => createSyntheticCalcuttaMutation.mutate()}
                >
                  Create
                </Button>
              </div>
            </div>

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

            {createSyntheticCalcuttaMutation.isError ? (
              <Alert variant="error" className="mt-3">
                {showError(createSyntheticCalcuttaMutation.error)}
              </Alert>
            ) : null}

            {!syntheticCalcuttasQuery.isLoading && !syntheticCalcuttasQuery.isError && syntheticCalcuttas.length > 0 ? (
              <div className="overflow-x-auto">
                <table className="min-w-full divide-y divide-gray-200">
                  <thead className="bg-gray-50">
                    <tr>
                      <th className="px-3 py-2 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">Calcutta</th>
                      <th className="px-3 py-2 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">Highlighted Entry</th>
                      <th className="px-3 py-2 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">Our Rank</th>
                      <th className="px-3 py-2 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">Our Mean</th>
                      <th className="px-3 py-2 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">Our pTop1</th>
                      <th className="px-3 py-2 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">Our pITM</th>
                      <th className="px-3 py-2 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">Created</th>
                      <th className="px-3 py-2 text-right text-xs font-medium text-gray-500 uppercase tracking-wider">Open</th>
                    </tr>
                  </thead>
                  <tbody className="bg-white divide-y divide-gray-200">
                    {syntheticCalcuttas.map((sc) => {
                      const calcuttaName = calcuttaNameById.get(sc.calcutta_id) || sc.calcutta_id;
                      const highlighted = sc.focus_entry_name || '—';
                      const href = `/sandbox/synthetic-calcuttas/${encodeURIComponent(sc.id)}?cohortId=${encodeURIComponent(effectiveCohortId)}`;
                      return (
                        <tr key={sc.id} className="hover:bg-gray-50">
                          <td className="px-3 py-2 text-sm text-gray-900" title={calcuttaName}>
                            <div className="font-medium">{seasonFromCalcuttaName(calcuttaName)}</div>
                            <div className="text-xs text-gray-600">{calcuttaName}</div>
                          </td>
                          <td className="px-3 py-2 text-sm text-gray-700">{highlighted}</td>
                          <td className="px-3 py-2 text-sm text-gray-700">
                            {sc.latest_simulation_status === 'succeeded' ? sc.our_rank ?? '—' : sc.latest_simulation_status || '—'}
                          </td>
                          <td className="px-3 py-2 text-sm text-gray-700">
                            {sc.latest_simulation_status === 'succeeded' ? formatFloat(sc.our_mean_normalized_payout, 4) : '—'}
                          </td>
                          <td className="px-3 py-2 text-sm text-gray-700">
                            {sc.latest_simulation_status === 'succeeded' ? formatFloat(sc.our_p_top1, 4) : '—'}
                          </td>
                          <td className="px-3 py-2 text-sm text-gray-700">
                            {sc.latest_simulation_status === 'succeeded' ? formatFloat(sc.our_p_in_money, 4) : '—'}
                          </td>
                          <td className="px-3 py-2 text-sm text-gray-700">{formatDateTime(sc.created_at)}</td>
                          <td className="px-3 py-2 text-sm text-right">
                            <Link to={href} className="text-blue-600 hover:text-blue-800">
                              View
                            </Link>
                          </td>
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
