import React, { useMemo } from 'react';
import { useQuery } from '@tanstack/react-query';
import { useNavigate, useSearchParams } from 'react-router-dom';

import { ApiError } from '../api/apiClient';
import { Card } from '../components/ui/Card';
import { LoadingState } from '../components/ui/LoadingState';
import { PageContainer, PageHeader } from '../components/ui/Page';
import { Select } from '../components/ui/Select';
import { calcuttaService } from '../services/calcuttaService';
import { suiteCalcuttaEvaluationsService, type SuiteCalcuttaEvaluation } from '../services/suiteCalcuttaEvaluationsService';
import type { Calcutta } from '../types/calcutta';

export function SandboxPage() {
  const navigate = useNavigate();
  const [searchParams, setSearchParams] = useSearchParams();
  const selectedSuiteId = searchParams.get('suiteId') || '';

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

  const allEvaluationsQuery = useQuery({
    queryKey: ['suite-calcutta-evaluations', 'list', 'all'],
    queryFn: () => suiteCalcuttaEvaluationsService.list({ limit: 200, offset: 0 }),
  });

  const listQuery = useQuery({
    queryKey: ['suite-calcutta-evaluations', 'list', selectedSuiteId],
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

  const showError = (err: unknown) => {
    if (err instanceof ApiError) {
      if (err.status === 403) {
        return 'You do not have permission to view suite evaluations (403).';
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

  return (
    <PageContainer>
      <PageHeader title="Sandbox" subtitle="Browse historical TestSuite runs and drill into results." />

      <Card className="mb-6">
        <div className="flex items-center gap-4">
          <label htmlFor="suite-select" className="text-lg font-semibold whitespace-nowrap">
            Suite:
          </label>
          <Select
            id="suite-select"
            value={selectedSuiteId}
            onChange={(e) => {
              const suiteId = e.target.value;
              setSearchParams(suiteId ? { suiteId } : {}, { replace: true });
            }}
            className="flex-1 max-w-2xl"
          >
            <option value="">All suites</option>
            {suites.map((s) => (
              <option key={s.id} value={s.id}>
                {s.name}
              </option>
            ))}
          </Select>
        </div>
        <div className="mt-3 text-sm text-gray-600">
          This view is backed by <code className="text-gray-800">/api/suite-calcutta-evaluations</code> and shows each evaluation request/run.
        </div>
      </Card>

      <Card>
        <h2 className="text-xl font-semibold mb-4">Evaluations</h2>

        {listQuery.isLoading ? <LoadingState label="Loading evaluations..." layout="inline" /> : null}
        {listQuery.isError ? <div className="text-red-700">{showError(listQuery.error)}</div> : null}

        {!listQuery.isLoading && !listQuery.isError && items.length === 0 ? (
          <div className="text-gray-700">No suite evaluations found.</div>
        ) : null}

        {!listQuery.isLoading && !listQuery.isError && items.length > 0 ? (
          <div className="overflow-x-auto">
            <table className="min-w-full divide-y divide-gray-200">
              <thead className="bg-gray-50">
                <tr>
                  <th className="px-3 py-2 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">Suite</th>
                  <th className="px-3 py-2 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">Calcutta</th>
                  <th className="px-3 py-2 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">Status</th>
                  <th className="px-3 py-2 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">Created</th>
                </tr>
              </thead>
              <tbody className="bg-white divide-y divide-gray-200">
                {items.map((it) => {
                  const detailUrl = `/sandbox/evaluations/${encodeURIComponent(it.id)}${
                    selectedSuiteId ? `?suiteId=${encodeURIComponent(selectedSuiteId)}` : ''
                  }`;

                  const showHeadline = it.status === 'succeeded' && it.our_mean_normalized_payout != null;

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
