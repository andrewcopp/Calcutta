import React, { useMemo, useState } from 'react';
import { useQuery } from '@tanstack/react-query';
import { Link, useNavigate, useParams, useSearchParams } from 'react-router-dom';

import { ApiError } from '../api/apiClient';
import { Alert } from '../components/ui/Alert';
import { Button } from '../components/ui/Button';
import { Card } from '../components/ui/Card';
import { LoadingState } from '../components/ui/LoadingState';
import { PageContainer, PageHeader } from '../components/ui/Page';
import { calcuttaService } from '../services/calcuttaService';
import {
  suiteCalcuttaEvaluationsService,
  type SuiteCalcuttaEvaluationResult,
} from '../services/suiteCalcuttaEvaluationsService';
import type { Calcutta } from '../types/calcutta';

export function SuiteCalcuttaEvaluationDetailPage() {
  const { id } = useParams<{ id: string }>();
  const [searchParams] = useSearchParams();
  const navigate = useNavigate();

  const suiteId = searchParams.get('cohortId') || searchParams.get('suiteId') || '';
  const executionId = searchParams.get('executionId') || '';
  const backUrl = suiteId
    ? `/sandbox/cohorts/${encodeURIComponent(suiteId)}${executionId ? `?executionId=${encodeURIComponent(executionId)}` : ''}`
    : '/sandbox/cohorts';

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

  const fmtFinish = (pos?: number | null, tied?: boolean | null) => {
    if (pos == null) return '—';
    return `${pos}${tied ? ' (tied)' : ''}`;
  };

  const formatCurrency = (cents: number | null | undefined) => {
    if (cents == null || Number.isNaN(cents)) return '—';
    return `$${(cents / 100).toFixed(2)}`;
  };

  const detailQuery = useQuery({
    queryKey: ['simulation-runs', 'get', id],
    queryFn: () => suiteCalcuttaEvaluationsService.get(id!),
    enabled: Boolean(id),
  });

  const resultQuery = useQuery<SuiteCalcuttaEvaluationResult>({
    queryKey: ['simulation-runs', 'result', id],
    queryFn: () => suiteCalcuttaEvaluationsService.getResult(id!),
    enabled: Boolean(id) && detailQuery.data?.status === 'succeeded',
  });

  const [sortKey, setSortKey] = useState<'mean' | 'p_top1' | 'p_in_money' | 'finish_position'>(
    'mean'
  );
  const [sortDir, setSortDir] = useState<'asc' | 'desc'>('desc');

  const sortedEntries = useMemo(() => {
    const entries = resultQuery.data?.entries ?? [];
    const mult = sortDir === 'asc' ? 1 : -1;
    return entries.slice().sort((a, b) => {
      const av =
        sortKey === 'mean'
          ? a.mean_normalized_payout
          : sortKey === 'p_top1'
            ? a.p_top1
            : sortKey === 'p_in_money'
              ? a.p_in_money
              : a.finish_position ?? NaN;
      const bv =
        sortKey === 'mean'
          ? b.mean_normalized_payout
          : sortKey === 'p_top1'
            ? b.p_top1
            : sortKey === 'p_in_money'
              ? b.p_in_money
              : b.finish_position ?? NaN;
      if (!Number.isFinite(av) && !Number.isFinite(bv)) return 0;
      if (!Number.isFinite(av)) return 1;
      if (!Number.isFinite(bv)) return -1;
      return mult * (av - bv);
    });
  }, [resultQuery.data?.entries, sortKey, sortDir]);

  const toggleSort = (key: typeof sortKey) => {
    if (key === sortKey) {
      setSortDir((d) => (d === 'asc' ? 'desc' : 'asc'));
    } else {
      setSortKey(key);
      setSortDir('desc');
    }
  };

  return (
    <PageContainer className="max-w-none">
      <PageHeader
        title="Simulation Run"
        subtitle={id}
        leftActions={
          <Link to={backUrl} className="text-blue-600 hover:text-blue-800">
            ← Back to Sandbox
          </Link>
        }
      />

      {!id ? (
        <Alert variant="error">Missing simulation run ID.</Alert>
      ) : null}

      {id && detailQuery.isLoading ? <LoadingState label="Loading simulation run..." /> : null}
      {id && detailQuery.isError ? (
        <Alert variant="error">
          <div className="font-semibold mb-1">Failed to load simulation run</div>
          <div className="mb-3">{showError(detailQuery.error)}</div>
          <Button size="sm" onClick={() => detailQuery.refetch()}>
            Retry
          </Button>
        </Alert>
      ) : null}

      {detailQuery.data ? (
        <div className="space-y-6">
          <Card>
            <h2 className="text-xl font-semibold mb-4">Provenance</h2>
            <div className="grid grid-cols-1 md:grid-cols-2 gap-4 text-sm">
              <div>
                <div className="text-gray-500">Cohort</div>
                <div className="font-medium text-gray-900">{detailQuery.data.suite_name || detailQuery.data.suite_id}</div>
              </div>
              <div>
                <div className="text-gray-500">Calcutta</div>
                <div className="text-gray-900">
                  {calcuttaNameById.get(detailQuery.data.calcutta_id) || detailQuery.data.calcutta_id}
                </div>
              </div>
              <div>
                <div className="text-gray-500">Status</div>
                <div className="text-gray-900">{detailQuery.data.status}</div>
              </div>
              <div>
                <div className="text-gray-500">Optimizer</div>
                <div className="text-gray-900">{detailQuery.data.optimizer_key}</div>
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
                <div className="text-gray-500">Created</div>
                <div className="text-gray-900">{formatDateTime(detailQuery.data.created_at)}</div>
              </div>
              <div>
                <div className="text-gray-500">Updated</div>
                <div className="text-gray-900">{formatDateTime(detailQuery.data.updated_at)}</div>
              </div>
            </div>

            {detailQuery.data.error_message ? (
              <div className="mt-4">
                <div className="text-gray-500 text-sm">Error</div>
                <Alert variant="error" className="mt-2 break-words">
                  {detailQuery.data.error_message}
                </Alert>
              </div>
            ) : null}
          </Card>

          <Card>
            <h2 className="text-xl font-semibold mb-4">Result</h2>

            {detailQuery.data.status === 'succeeded' && detailQuery.data.our_mean_normalized_payout != null ? (
              <div className="mb-4 text-sm">
                <div className="text-gray-500">Headline (persisted)</div>
                <div className="text-gray-900">
                  rank={detailQuery.data.our_rank ?? '—'} · mean={formatFloat(detailQuery.data.our_mean_normalized_payout, 4)} · pTop1=
                  {formatFloat(detailQuery.data.our_p_top1, 4)} · pInMoney={formatFloat(detailQuery.data.our_p_in_money, 4)}
                  {detailQuery.data.total_simulations != null ? (
                    <span className="text-gray-500"> · nSims={detailQuery.data.total_simulations}</span>
                  ) : null}
                </div>
              </div>
            ) : null}

            {detailQuery.data.status === 'succeeded' &&
            (detailQuery.data.realized_finish_position != null ||
              detailQuery.data.realized_total_points != null ||
              detailQuery.data.realized_payout_cents != null) ? (
              <div className="mb-4 text-sm">
                <div className="text-gray-500">Realized (historical)</div>
                <div className="text-gray-900">
                  finish={detailQuery.data.realized_finish_position ?? '—'}
                  {detailQuery.data.realized_is_tied ? ' (tied)' : ''} · payout={formatCurrency(detailQuery.data.realized_payout_cents)} · points=
                  {formatFloat(detailQuery.data.realized_total_points, 2)}
                  {detailQuery.data.realized_in_the_money != null ? (
                    <span className="text-gray-500"> · ITM={detailQuery.data.realized_in_the_money ? 'yes' : 'no'}</span>
                  ) : null}
                </div>
              </div>
            ) : null}

            {detailQuery.data.status !== 'succeeded' ? (
              <Alert variant="info">
                Result is not available until the simulation run succeeds (current status: {detailQuery.data.status}).
              </Alert>
            ) : null}

            {detailQuery.data.status === 'succeeded' ? (
              <div>
                {resultQuery.isLoading ? <LoadingState label="Loading result..." layout="inline" /> : null}
                {resultQuery.isError ? (
                  <Alert variant="error" className="mt-3">
                    <div className="font-semibold mb-1">Failed to load result</div>
                    <div className="mb-3">{showError(resultQuery.error)}</div>
                    <Button size="sm" onClick={() => resultQuery.refetch()}>
                      Retry
                    </Button>
                  </Alert>
                ) : null}

                {resultQuery.data ? (
                  <div className="space-y-6">
                    <div>
                      <div className="text-gray-500">Entries</div>
                      {sortedEntries.length === 0 ? (
                        <Alert variant="info" className="mt-2">
                          No entry performance rows found.
                        </Alert>
                      ) : (
                        <div className="overflow-x-auto mt-2">
                          <table className="min-w-full divide-y divide-gray-200">
                            <thead className="bg-gray-50">
                              <tr>
                                <th className="px-3 py-2 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">Entry</th>
                                <th
                                  className="px-3 py-2 text-right text-xs font-medium text-gray-500 uppercase tracking-wider cursor-pointer"
                                  onClick={() => toggleSort('mean')}
                                >
                                  Mean Norm Payout
                                </th>
                                <th
                                  className="px-3 py-2 text-right text-xs font-medium text-gray-500 uppercase tracking-wider cursor-pointer"
                                  onClick={() => toggleSort('p_top1')}
                                >
                                  P(Top 1)
                                </th>
                                <th
                                  className="px-3 py-2 text-right text-xs font-medium text-gray-500 uppercase tracking-wider cursor-pointer"
                                  onClick={() => toggleSort('p_in_money')}
                                >
                                  P(In Money)
                                </th>
                                <th
                                  className="px-3 py-2 text-right text-xs font-medium text-gray-500 uppercase tracking-wider cursor-pointer"
                                  onClick={() => toggleSort('finish_position')}
                                >
                                  Finish
                                </th>
                              </tr>
                            </thead>
                            <tbody className="bg-white divide-y divide-gray-200">
                              {sortedEntries.map((e) => {
                                const clickable = Boolean(e.snapshot_entry_id);
                                return (
                                  <tr
                                    key={`${e.entry_name}-${e.rank}`}
                                    className={
                                      clickable
                                        ? 'hover:bg-gray-50 cursor-pointer focus-visible:outline focus-visible:outline-2 focus-visible:outline-blue-500'
                                        : ''
                                    }
                                    role={clickable ? 'link' : undefined}
                                    tabIndex={clickable ? 0 : undefined}
                                    onClick={() => {
                                      if (!clickable) return;
                                      navigate(
                                        `/sandbox/evaluations/${encodeURIComponent(id || '')}/entries/${encodeURIComponent(e.snapshot_entry_id || '')}` +
                                          `${suiteId ? `?cohortId=${encodeURIComponent(suiteId)}${executionId ? `&executionId=${encodeURIComponent(executionId)}` : ''}` : ''}`
                                      );
                                    }}
                                    onKeyDown={(ev) => {
                                      if (!clickable) return;
                                      if (ev.key === 'Enter' || ev.key === ' ') {
                                        ev.preventDefault();
                                        navigate(
                                          `/sandbox/evaluations/${encodeURIComponent(id || '')}/entries/${encodeURIComponent(e.snapshot_entry_id || '')}` +
                                            `${suiteId ? `?cohortId=${encodeURIComponent(suiteId)}${executionId ? `&executionId=${encodeURIComponent(executionId)}` : ''}` : ''}`
                                        );
                                      }
                                    }}
                                  >
                                    <td className="px-3 py-2 text-sm text-gray-900">
                                      <div className="font-medium">{e.entry_name}</div>
                                      <div className="text-xs text-gray-500">rank={e.rank}</div>
                                    </td>
                                    <td className="px-3 py-2 text-sm text-right text-gray-700">{formatFloat(e.mean_normalized_payout, 4)}</td>
                                    <td className="px-3 py-2 text-sm text-right text-gray-700">{formatFloat(e.p_top1, 4)}</td>
                                    <td className="px-3 py-2 text-sm text-right text-gray-700">{formatFloat(e.p_in_money, 4)}</td>
                                    <td className="px-3 py-2 text-sm text-right text-gray-700">{fmtFinish(e.finish_position, e.is_tied)}</td>
                                  </tr>
                                );
                              })}
                            </tbody>
                          </table>
                        </div>
                      )}
                    </div>
                  </div>
                ) : null}
              </div>
            ) : null}
          </Card>
        </div>
      ) : null}
    </PageContainer>
  );
}
