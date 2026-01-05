import React, { useMemo } from 'react';
import { useQuery } from '@tanstack/react-query';
import { Link, useParams, useSearchParams } from 'react-router-dom';

import { ApiError } from '../api/apiClient';
import { Alert } from '../components/ui/Alert';
import { Button } from '../components/ui/Button';
import { Card } from '../components/ui/Card';
import { LoadingState } from '../components/ui/LoadingState';
import { PageContainer, PageHeader } from '../components/ui/Page';
import { calcuttaService } from '../services/calcuttaService';
import {
  suiteCalcuttaEvaluationsService,
  type SuiteCalcuttaEvaluationPortfolioBid,
  type SuiteCalcuttaEvaluationResult,
} from '../services/suiteCalcuttaEvaluationsService';
import type { Calcutta } from '../types/calcutta';

export function SuiteCalcuttaEvaluationDetailPage() {
  const { id } = useParams<{ id: string }>();
  const [searchParams] = useSearchParams();

  const suiteId = searchParams.get('suiteId') || '';
  const executionId = searchParams.get('executionId') || '';
  const backUrl = suiteId
    ? `/sandbox/suites/${encodeURIComponent(suiteId)}${executionId ? `?executionId=${encodeURIComponent(executionId)}` : ''}`
    : '/sandbox/suites';

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

  const formatROI = (v: number) => {
    if (Number.isNaN(v)) return '—';
    return v.toFixed(3);
  };

  const formatFloat = (v: number | null | undefined, digits: number) => {
    if (v == null || Number.isNaN(v)) return '—';
    return v.toFixed(digits);
  };

  const formatCurrency = (cents: number | null | undefined) => {
    if (cents == null || Number.isNaN(cents)) return '—';
    return `$${(cents / 100).toFixed(2)}`;
  };

  const renderPortfolioTable = (bids: SuiteCalcuttaEvaluationPortfolioBid[]) => {
    if (bids.length === 0) {
      return <Alert variant="info">No portfolio bids found.</Alert>;
    }
    return (
      <div className="overflow-x-auto">
        <table className="min-w-full divide-y divide-gray-200">
          <thead className="bg-gray-50">
            <tr>
              <th className="px-3 py-2 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">Team</th>
              <th className="px-3 py-2 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">Seed</th>
              <th className="px-3 py-2 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">Region</th>
              <th className="px-3 py-2 text-right text-xs font-medium text-gray-500 uppercase tracking-wider">Bid</th>
              <th className="px-3 py-2 text-right text-xs font-medium text-gray-500 uppercase tracking-wider">Expected ROI</th>
            </tr>
          </thead>
          <tbody className="bg-white divide-y divide-gray-200">
            {bids.map((b) => (
              <tr key={b.team_id}>
                <td className="px-3 py-2 text-sm text-gray-900">{b.school_name}</td>
                <td className="px-3 py-2 text-sm text-gray-700">{b.seed}</td>
                <td className="px-3 py-2 text-sm text-gray-700">{b.region}</td>
                <td className="px-3 py-2 text-sm text-gray-900 text-right font-medium">{b.bid_points}</td>
                <td className="px-3 py-2 text-sm text-gray-700 text-right">{formatROI(b.expected_roi)}</td>
              </tr>
            ))}
          </tbody>
        </table>
      </div>
    );
  };

  const detailQuery = useQuery({
    queryKey: ['suite-calcutta-evaluations', 'get', id],
    queryFn: () => suiteCalcuttaEvaluationsService.get(id!),
    enabled: Boolean(id),
  });

  const resultQuery = useQuery<SuiteCalcuttaEvaluationResult>({
    queryKey: ['suite-calcutta-evaluations', 'result', id],
    queryFn: () => suiteCalcuttaEvaluationsService.getResult(id!),
    enabled: Boolean(id) && detailQuery.data?.status === 'succeeded',
  });

  return (
    <PageContainer className="max-w-none">
      <PageHeader
        title="Suite Evaluation"
        subtitle={id}
        actions={
          <Link to={backUrl} className="text-blue-600 hover:text-blue-800">
            ← Back to Sandbox
          </Link>
        }
      />

      {!id ? (
        <Alert variant="error">Missing evaluation ID.</Alert>
      ) : null}

      {id && detailQuery.isLoading ? <LoadingState label="Loading evaluation..." /> : null}
      {id && detailQuery.isError ? (
        <Alert variant="error">
          <div className="font-semibold mb-1">Failed to load evaluation</div>
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
                <div className="text-gray-500">Suite</div>
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
                Result is not available until the evaluation succeeds (current status: {detailQuery.data.status}).
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
                      <div className="text-gray-500">Our Strategy performance</div>
                      {resultQuery.data.our_strategy ? (
                        <div className="mt-1 text-gray-900">
                          <div>
                            rank={resultQuery.data.our_strategy.rank} · mean={resultQuery.data.our_strategy.mean_normalized_payout.toFixed(4)} ·
                            pTop1={resultQuery.data.our_strategy.p_top1.toFixed(4)} · pInMoney={resultQuery.data.our_strategy.p_in_money.toFixed(4)}
                          </div>
                          <div className="text-xs text-gray-600">nSims={resultQuery.data.our_strategy.total_simulations}</div>
                        </div>
                      ) : (
                        <div className="mt-1 text-gray-700">No performance row found for Our Strategy.</div>
                      )}
                    </div>

                    <div>
                      <div className="text-gray-500">Generated entry (portfolio)</div>
                      <div className="mt-2">{renderPortfolioTable(resultQuery.data.portfolio)}</div>
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
