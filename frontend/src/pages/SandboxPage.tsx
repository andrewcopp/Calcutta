import React, { useMemo, useState } from 'react';
import { useQuery } from '@tanstack/react-query';

import { ApiError } from '../api/apiClient';
import { Card } from '../components/ui/Card';
import { LoadingState } from '../components/ui/LoadingState';
import { PageContainer, PageHeader } from '../components/ui/Page';
import { Select } from '../components/ui/Select';
import { calcuttaService } from '../services/calcuttaService';
import { suiteCalcuttaEvaluationsService, type SuiteCalcuttaEvaluation } from '../services/suiteCalcuttaEvaluationsService';
import type { Calcutta } from '../types/calcutta';

export function SandboxPage() {
  const [selectedSuiteId, setSelectedSuiteId] = useState<string>('');
  const [selectedEvaluationId, setSelectedEvaluationId] = useState<string>('');

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

  const listQuery = useQuery({
    queryKey: ['suite-calcutta-evaluations', 'list', selectedSuiteId],
    queryFn: () => suiteCalcuttaEvaluationsService.list({ suiteId: selectedSuiteId || undefined, limit: 200, offset: 0 }),
  });

  const items: SuiteCalcuttaEvaluation[] = listQuery.data?.items ?? [];

  const suites = useMemo(() => {
    const byId = new Map<string, { id: string; name: string }>();
    for (const it of items) {
      if (!byId.has(it.suite_id)) {
        byId.set(it.suite_id, { id: it.suite_id, name: it.suite_name || it.suite_id });
      }
    }
    return Array.from(byId.values()).sort((a, b) => a.name.localeCompare(b.name));
  }, [items]);

  const selectedEval = useMemo(() => items.find((it) => it.id === selectedEvaluationId) ?? null, [items, selectedEvaluationId]);

  const detailQuery = useQuery({
    queryKey: ['suite-calcutta-evaluations', 'get', selectedEvaluationId],
    queryFn: () => suiteCalcuttaEvaluationsService.get(selectedEvaluationId),
    enabled: Boolean(selectedEvaluationId),
  });

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
              setSelectedSuiteId(e.target.value);
              setSelectedEvaluationId('');
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

      <div className="grid grid-cols-1 lg:grid-cols-3 gap-6">
        <Card className="lg:col-span-2">
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
                    const active = it.id === selectedEvaluationId;
                    return (
                      <tr
                        key={it.id}
                        className={active ? 'bg-blue-50 cursor-pointer' : 'hover:bg-gray-50 cursor-pointer'}
                        onClick={() => setSelectedEvaluationId(it.id)}
                      >
                        <td className="px-3 py-2 text-sm text-gray-900">
                          <div className="font-medium">{it.suite_name || it.suite_id}</div>
                          <div className="text-xs text-gray-600">
                            {it.optimizer_key} · n={it.n_sims} · seed={it.seed}
                          </div>
                        </td>
                        <td className="px-3 py-2 text-sm text-gray-700">
                          {calcuttaNameById.get(it.calcutta_id) || it.calcutta_id}
                        </td>
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

        <Card>
          <h2 className="text-xl font-semibold mb-4">Details</h2>

          {!selectedEvaluationId ? <div className="text-gray-700">Select an evaluation to see details.</div> : null}

          {selectedEvaluationId && detailQuery.isLoading ? <LoadingState label="Loading details..." layout="inline" /> : null}
          {selectedEvaluationId && detailQuery.isError ? <div className="text-red-700">{showError(detailQuery.error)}</div> : null}

          {detailQuery.data ? (
            <div className="space-y-2 text-sm">
              <div>
                <div className="text-gray-500">Suite</div>
                <div className="font-medium text-gray-900">{detailQuery.data.suite_name || detailQuery.data.suite_id}</div>
              </div>
              <div>
                <div className="text-gray-500">Calcutta</div>
                <div className="text-gray-900">{calcuttaNameById.get(detailQuery.data.calcutta_id) || detailQuery.data.calcutta_id}</div>
              </div>
              <div>
                <div className="text-gray-500">Status</div>
                <div className="text-gray-900">{detailQuery.data.status}</div>
              </div>
              <div>
                <div className="text-gray-500">Starting state</div>
                <div className="text-gray-900">{detailQuery.data.starting_state_key}</div>
              </div>
              <div>
                <div className="text-gray-500">Excluded entry</div>
                <div className="text-gray-900">{detailQuery.data.excluded_entry_name || '—'}</div>
              </div>
              <div>
                <div className="text-gray-500">Game outcomes run</div>
                <div className="text-gray-900">{detailQuery.data.game_outcome_run_id || '—'}</div>
              </div>
              <div>
                <div className="text-gray-500">Market share run</div>
                <div className="text-gray-900">{detailQuery.data.market_share_run_id || '—'}</div>
              </div>
              <div>
                <div className="text-gray-500">Created / Updated</div>
                <div className="text-gray-900">
                  {formatDateTime(detailQuery.data.created_at)}
                  <span className="text-gray-500"> · </span>
                  {formatDateTime(detailQuery.data.updated_at)}
                </div>
              </div>
              {detailQuery.data.error_message ? (
                <div>
                  <div className="text-gray-500">Error</div>
                  <div className="text-red-700 break-words">{detailQuery.data.error_message}</div>
                </div>
              ) : null}
            </div>
          ) : null}

          {selectedEval && !detailQuery.data && !detailQuery.isLoading && !detailQuery.isError ? (
            <div className="text-gray-700">
              <div className="text-gray-500">Status</div>
              <div className="text-gray-900">{selectedEval.status}</div>
            </div>
          ) : null}
        </Card>
      </div>
    </PageContainer>
  );
}
