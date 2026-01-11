import React, { useMemo, useState } from 'react';
import { useQuery } from '@tanstack/react-query';
import { Link, useParams, useSearchParams } from 'react-router-dom';

import { ApiError } from '../api/apiClient';
import { Alert } from '../components/ui/Alert';
import { Button } from '../components/ui/Button';
import { Card } from '../components/ui/Card';
import { LoadingState } from '../components/ui/LoadingState';
import { PageContainer, PageHeader } from '../components/ui/Page';
import { simulatedCalcuttasService } from '../services/simulatedCalcuttasService';
import { simulatedEntriesService, type SimulatedEntryListItem } from '../services/simulatedEntriesService';
import { simulationRunsService, type SimulationRunResult } from '../services/simulationRunsService';

type EntryPerformance = SimulationRunResult['entries'][number];

export function SimulatedCalcuttaDetailPage() {
  const { id } = useParams<{ id: string }>();
  const [searchParams] = useSearchParams();
  const cohortId = searchParams.get('cohortId') || '';

	const [sortKey, setSortKey] = useState<'mean' | 'p_top1' | 'p_in_money' | 'finish_position'>('mean');
	const [sortDir, setSortDir] = useState<'asc' | 'desc'>('desc');

  const showError = (err: unknown) => {
    if (err instanceof ApiError) {
      if (err.status === 403) return 'You do not have permission to view simulated calcuttas (403).';
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

  const detailQuery = useQuery({
    queryKey: ['simulated-calcuttas', 'get', id],
    queryFn: () => simulatedCalcuttasService.get(id || ''),
    enabled: Boolean(id),
  });

  const entriesQuery = useQuery({
    queryKey: ['simulated-entries', 'list', id],
    queryFn: () => simulatedEntriesService.list(id || ''),
    enabled: Boolean(id),
  });

  const entries: SimulatedEntryListItem[] = useMemo(() => entriesQuery.data?.items ?? [], [entriesQuery.data?.items]);

	const runsQuery = useQuery({
		queryKey: ['simulation-runs', 'list', cohortId],
		queryFn: async () => {
			if (!cohortId) return { items: [] };
			return simulationRunsService.list({ cohortId, limit: 200, offset: 0 });
		},
		enabled: Boolean(cohortId),
	});

	const latestRunId = useMemo(() => {
		const items = runsQuery.data?.items ?? [];
		if (!id) return '';
		const match = items.find((r) => (r.simulated_calcutta_id || '') === id);
		return match?.id || '';
	}, [id, runsQuery.data?.items]);

	const latestRunStatus = useMemo(() => {
		const items = runsQuery.data?.items ?? [];
		if (!latestRunId) return '';
		return items.find((r) => r.id === latestRunId)?.status || '';
	}, [latestRunId, runsQuery.data?.items]);

	const resultQuery = useQuery<SimulationRunResult>({
		queryKey: ['simulation-runs', 'result', latestRunId],
		queryFn: () => simulationRunsService.getResult({ cohortId, id: latestRunId }),
		enabled: Boolean(cohortId && latestRunId) && latestRunStatus === 'succeeded',
	});

	const perfByName = useMemo(() => {
		const m = new Map<string, EntryPerformance>();
		for (const it of resultQuery.data?.entries ?? []) {
			m.set(it.entry_name, it);
		}
		return m;
	}, [resultQuery.data?.entries]);

	const fmtPct = (v: number | null | undefined) => {
		if (v == null || !Number.isFinite(v)) return '—';
		return `${(v * 100).toFixed(1)}%`;
	};
	const fmtFloat = (v: number | null | undefined, digits: number) => {
		if (v == null || !Number.isFinite(v)) return '—';
		return v.toFixed(digits);
	};
	const fmtFinish = (pos?: number | null, tied?: boolean | null) => {
		if (pos == null) return '—';
		return `${pos}${tied ? ' (tied)' : ''}`;
	};

	const rows = useMemo(() => {
		const base = entries.map((e) => {
			const perf = perfByName.get(e.display_name);
			return {
				id: e.id,
				displayName: e.display_name,
				rank: perf?.rank ?? null,
				mean: perf?.mean_normalized_payout ?? null,
				pTop1: perf?.p_top1 ?? null,
				pInMoney: perf?.p_in_money ?? null,
				finishPosition: perf?.finish_position ?? null,
				isTied: perf?.is_tied ?? null,
			};
		});

		const dirMult = sortDir === 'asc' ? 1 : -1;
		base.sort((a, b) => {
			const aVal =
				sortKey === 'mean'
					? a.mean
					: sortKey === 'p_top1'
						? a.pTop1
						: sortKey === 'p_in_money'
							? a.pInMoney
							: a.finishPosition;
			const bVal =
				sortKey === 'mean'
					? b.mean
					: sortKey === 'p_top1'
						? b.pTop1
						: sortKey === 'p_in_money'
							? b.pInMoney
							: b.finishPosition;

			// finish_position sorts ascending by default
			const mult = sortKey === 'finish_position' ? -dirMult : dirMult;
			const av = aVal == null ? NaN : aVal;
			const bv = bVal == null ? NaN : bVal;
			if (!Number.isFinite(av) && !Number.isFinite(bv)) return 0;
			if (!Number.isFinite(av)) return 1;
			if (!Number.isFinite(bv)) return -1;
			return mult * (av - bv);
		});
		return base;
	}, [entries, perfByName, sortDir, sortKey]);

	const toggleSort = (k: typeof sortKey) => {
		if (k === sortKey) {
			setSortDir((d) => (d === 'asc' ? 'desc' : 'asc'));
			return;
		}
		setSortKey(k);
		setSortDir(k === 'finish_position' ? 'asc' : 'desc');
	};

  const backUrl = cohortId ? `/sandbox/cohorts/${encodeURIComponent(cohortId)}` : '/sandbox/cohorts';

  return (
    <PageContainer className="max-w-none">
      <PageHeader
        title="Simulated Calcutta"
        subtitle={id}
        leftActions={
          <Link to={backUrl} className="text-blue-600 hover:text-blue-800">
            ← Back to Sandbox
          </Link>
        }
      />

      {!id ? <Alert variant="error">Missing simulated calcutta ID.</Alert> : null}

      {id && detailQuery.isLoading ? <LoadingState label="Loading simulated calcutta..." /> : null}
      {id && detailQuery.isError ? (
        <Alert variant="error">
          <div className="font-semibold mb-1">Failed to load simulated calcutta</div>
          <div className="mb-3">{showError(detailQuery.error)}</div>
          <Button size="sm" onClick={() => detailQuery.refetch()}>
            Retry
          </Button>
        </Alert>
      ) : null}

      {detailQuery.data ? (
        <div className="space-y-6">
          <Card>
            <h2 className="text-xl font-semibold mb-4">Details</h2>
            <div className="grid grid-cols-1 md:grid-cols-2 gap-4 text-sm">
              <div>
                <div className="text-gray-500">Name</div>
                <div className="text-gray-900 font-medium break-words">{detailQuery.data.simulated_calcutta.name}</div>
              </div>
              <div>
                <div className="text-gray-500">Tournament</div>
                <div className="text-gray-900 break-words">{detailQuery.data.simulated_calcutta.tournament_id}</div>
              </div>
              <div>
                <div className="text-gray-500">Starting state</div>
                <div className="text-gray-900">{detailQuery.data.simulated_calcutta.starting_state_key}</div>
              </div>
              <div>
                <div className="text-gray-500">Excluded entry name</div>
                <div className="text-gray-900">{detailQuery.data.simulated_calcutta.excluded_entry_name || '—'}</div>
              </div>
              <div>
                <div className="text-gray-500">Created</div>
                <div className="text-gray-900">{formatDateTime(detailQuery.data.simulated_calcutta.created_at)}</div>
              </div>
              <div>
                <div className="text-gray-500">Updated</div>
                <div className="text-gray-900">{formatDateTime(detailQuery.data.simulated_calcutta.updated_at)}</div>
              </div>
            </div>
          </Card>

          <Card>
            <h2 className="text-xl font-semibold mb-4">Simulated Entries</h2>

            {entriesQuery.isLoading ? <LoadingState label="Loading entries..." layout="inline" /> : null}
            {entriesQuery.isError ? (
              <Alert variant="error" className="mt-3">
                <div className="font-semibold mb-1">Failed to load entries</div>
                <div className="mb-3">{showError(entriesQuery.error)}</div>
                <Button size="sm" onClick={() => entriesQuery.refetch()}>
                  Retry
                </Button>
              </Alert>
            ) : null}

            {!entriesQuery.isLoading && !entriesQuery.isError && entries.length === 0 ? (
              <Alert variant="info">No simulated entries found.</Alert>
            ) : null}

            {!entriesQuery.isLoading && !entriesQuery.isError && entries.length > 0 ? (
              <div className="overflow-x-auto">
                <table className="min-w-full divide-y divide-gray-200">
                  <thead className="bg-gray-50">
                    <tr>
                      <th className="px-3 py-2 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">Entry</th>
                      <th className="px-3 py-2 text-right text-xs font-medium text-gray-500 uppercase tracking-wider">Rank</th>
                      <th className="px-3 py-2 text-right text-xs font-medium text-gray-500 uppercase tracking-wider">
									<button type="button" className="hover:text-gray-900" onClick={() => toggleSort('mean')}>
										Normalized Mean Payout
									</button>
								</th>
                      <th className="px-3 py-2 text-right text-xs font-medium text-gray-500 uppercase tracking-wider">
									<button type="button" className="hover:text-gray-900" onClick={() => toggleSort('p_top1')}>
										P(1st)
									</button>
								</th>
                      <th className="px-3 py-2 text-right text-xs font-medium text-gray-500 uppercase tracking-wider">
									<button type="button" className="hover:text-gray-900" onClick={() => toggleSort('p_in_money')}>
										P(In Money)
									</button>
								</th>
                      <th className="px-3 py-2 text-right text-xs font-medium text-gray-500 uppercase tracking-wider">
									<button type="button" className="hover:text-gray-900" onClick={() => toggleSort('finish_position')}>
										Real Finish
									</button>
								</th>
                    </tr>
                  </thead>
                  <tbody className="bg-white divide-y divide-gray-200">
                    {rows.map((r) => (
                      <tr key={r.id} className="hover:bg-gray-50">
                        <td className="px-3 py-2 text-sm text-gray-900">
                          <Link
                            to={`/sandbox/simulated-calcuttas/${encodeURIComponent(id || '')}/entries/${encodeURIComponent(r.id)}${
                              cohortId ? `?cohortId=${encodeURIComponent(cohortId)}` : ''
                            }`}
                            className="text-blue-600 hover:text-blue-800"
                          >
                            <div className="font-medium">{r.displayName}</div>
                            <div className="text-xs text-gray-500">{r.id.slice(0, 8)}</div>
                          </Link>
                        </td>
                        <td className="px-3 py-2 text-sm text-right text-gray-700">{r.rank ?? '—'}</td>
                        <td className="px-3 py-2 text-sm text-right text-gray-900 font-medium">{fmtFloat(r.mean, 4)}</td>
                        <td className="px-3 py-2 text-sm text-right text-gray-700">{fmtPct(r.pTop1)}</td>
                        <td className="px-3 py-2 text-sm text-right text-gray-700">{fmtPct(r.pInMoney)}</td>
                        <td className="px-3 py-2 text-sm text-right text-gray-700">{fmtFinish(r.finishPosition, r.isTied)}</td>
                      </tr>
                    ))}
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
