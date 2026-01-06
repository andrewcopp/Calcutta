import React, { useMemo } from 'react';
import { useQuery } from '@tanstack/react-query';
import { Link, useNavigate, useParams } from 'react-router-dom';

import { Card } from '../components/ui/Card';
import { PageContainer, PageHeader } from '../components/ui/Page';
import { analyticsService } from '../services/analyticsService';

type Algorithm = {
  id: string;
  kind: string;
  name: string;
  description?: string | null;
  params_json?: unknown;
  created_at?: string;
};

type GameOutcomeRun = {
  id: string;
  algorithm_id: string;
  tournament_id: string;
  params_json?: unknown;
  git_sha?: string | null;
  created_at: string;
};

type MarketShareRun = {
  id: string;
  algorithm_id: string;
  calcutta_id: string;
  params_json?: unknown;
  git_sha?: string | null;
  created_at: string;
};

type AlgorithmCoverageCalcutta = {
  calcutta_id: string;
  calcutta_name: string;
  tournament_id: string;
  tournament_name: string;
  starting_at?: string | null;
  last_run_at?: string | null;
};

type AlgorithmCoverageDetailResponse = {
  algorithm: {
    id: string;
    name: string;
    description?: string | null;
  };
  covered: number;
  total: number;
  items: AlgorithmCoverageCalcutta[];
};

type TeamPredictedMarketShare = {
  team_id: string;
  school_name: string;
  seed: number;
  region: string;
  rational_share: number;
  predicted_share: number;
  delta_percent: number;
};

export function LabInvestmentAlgorithmDetailPage() {
  const { algorithmId } = useParams<{ algorithmId: string }>();
  const navigate = useNavigate();

  const algorithmsQuery = useQuery<{ items: Algorithm[] } | null>({
    queryKey: ['analytics', 'algorithms', 'market_share'],
    queryFn: async () => {
      const filtered = await analyticsService.listAlgorithms<{ items: Algorithm[] }>('market_share');
      if (filtered?.items?.length) return filtered;
      return analyticsService.listAlgorithms<{ items: Algorithm[] }>();
    },
  });

  const algorithm = useMemo(() => {
    const items = algorithmsQuery.data?.items ?? [];
    return items.find((a) => a.id === algorithmId) ?? null;
  }, [algorithmsQuery.data?.items, algorithmId]);

  const coverageDetailQuery = useQuery<AlgorithmCoverageDetailResponse | null>({
    queryKey: ['analytics', 'coverage-detail', 'market-share', algorithmId],
    queryFn: async () => {
      if (!algorithmId) return null;
      return analyticsService.getMarketShareAlgorithmCoverageDetail<AlgorithmCoverageDetailResponse>(algorithmId);
    },
    enabled: Boolean(algorithmId),
  });

  const calcuttas = coverageDetailQuery.data?.items ?? [];
  const sortedCalcuttas = useMemo(() => {
    return calcuttas
      .slice()
      .sort((a, b) => {
        const aHas = Boolean(a.last_run_at);
        const bHas = Boolean(b.last_run_at);
        if (aHas !== bHas) return aHas ? -1 : 1;
        const aLast = a.last_run_at ?? '';
        const bLast = b.last_run_at ?? '';
        if (bLast !== aLast) return bLast.localeCompare(aLast);
        const aStart = a.starting_at ?? '';
        const bStart = b.starting_at ?? '';
        if (bStart !== aStart) return bStart.localeCompare(aStart);
        return a.calcutta_name.localeCompare(b.calcutta_name);
      });
  }, [calcuttas]);
  const fmtDateTime = (iso?: string | null) => {
    if (!iso) return 'Never';
    const d = new Date(iso);
    if (Number.isNaN(d.getTime())) return '—';
    return d.toLocaleString();
  };
  const calcuttaYear = (startingAt?: string | null) => {
    if (!startingAt) return '—';
    const d = new Date(startingAt);
    if (Number.isNaN(d.getTime())) return '—';
    return String(d.getFullYear());
  };
  const calcuttaYearOrName = (c: AlgorithmCoverageCalcutta) => {
    const byStart = calcuttaYear(c.starting_at);
    if (byStart !== '—') return byStart;
    const m = c.tournament_name.match(/\b(19\d{2}|20\d{2})\b/);
    return m ? m[1] : '—';
  };

  return (
    <PageContainer className="max-w-none">
      <PageHeader
        title="Investments Algorithm"
        subtitle={algorithm ? algorithm.name : algorithmId}
        leftActions={
          <Link to="/lab" className="text-blue-600 hover:text-blue-800">
            ← Back to Lab
          </Link>
        }
      />

      <div className="space-y-6">
        <Card>
          <h2 className="text-xl font-semibold mb-2">Algorithm</h2>
          {algorithmsQuery.isLoading ? <div className="text-gray-500">Loading algorithm...</div> : null}
          {algorithm ? (
            <div className="text-sm text-gray-700">
              <div className="font-medium text-gray-900">{algorithm.name}</div>
              {algorithm.description ? <div className="text-gray-600">{algorithm.description}</div> : null}
              <div className="text-gray-500 mt-1">id={algorithm.id}</div>
            </div>
          ) : !algorithmsQuery.isLoading ? (
            <div className="text-sm text-gray-700">Algorithm not found in registry.</div>
          ) : null}

          {coverageDetailQuery.data ? (
            <div className="text-sm text-gray-700 mt-3">
              <div className="text-gray-600">coverage={coverageDetailQuery.data.covered}/{coverageDetailQuery.data.total}</div>
            </div>
          ) : null}
        </Card>

        <Card>
          <h2 className="text-xl font-semibold mb-4">Calcuttas</h2>

          {coverageDetailQuery.isLoading ? <div className="text-gray-500">Loading coverage...</div> : null}

          {!coverageDetailQuery.isLoading && calcuttas.length === 0 ? (
            <div className="text-gray-700">No calcuttas found.</div>
          ) : null}

          {!coverageDetailQuery.isLoading && calcuttas.length > 0 ? (
            <div className="overflow-x-auto">
              <table className="min-w-full divide-y divide-gray-200">
                <thead className="bg-gray-50">
                  <tr>
                    <th className="px-3 py-2 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">Year</th>
                    <th className="px-3 py-2 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">Calcutta</th>
                    <th className="px-3 py-2 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">Tournament</th>
                    <th className="px-3 py-2 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">Last Run</th>
                  </tr>
                </thead>
                <tbody className="bg-white divide-y divide-gray-200">
                  {sortedCalcuttas.map((c) => {
                    return (
                      <tr
                        key={c.calcutta_id}
                        className="hover:bg-gray-50 cursor-pointer"
                        onClick={() => {
                          navigate(
                            `/lab/investments/algorithms/${encodeURIComponent(algorithmId || '')}/calcuttas/${encodeURIComponent(
                              c.calcutta_id
                            )}`
                          );
                        }}
                      >
                        <td className="px-3 py-2 text-sm text-gray-700">{calcuttaYearOrName(c)}</td>
                        <td className="px-3 py-2 text-sm text-gray-900">{c.calcutta_name}</td>
                        <td className="px-3 py-2 text-sm text-gray-700">{c.tournament_name}</td>
                        <td className="px-3 py-2 text-sm text-gray-700">{fmtDateTime(c.last_run_at)}</td>
                      </tr>
                    );
                  })}
                </tbody>
              </table>
            </div>
          ) : null}
        </Card>
      </div>
    </PageContainer>
  );
}
