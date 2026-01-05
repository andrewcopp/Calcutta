import React, { useMemo } from 'react';
import { Link, useParams, useSearchParams } from 'react-router-dom';
import { Card } from '../components/ui/Card';
import { PageContainer, PageHeader } from '../components/ui/Page';

type EvaluationRowId = 'return-algorithm' | 'investment-algorithm' | 'entry-optimizer';

type SeasonEvaluationRow = {
  season: number;
  meanPayout: number;
  pTop1: number;
  pPayout: number;
  totalSimulations: number;
};

const formatPercent = (value: number) => `${(value * 100).toFixed(1)}%`;

export function EvaluationDetailPage() {
  const { evaluationId } = useParams<{ evaluationId: string }>();
  const [searchParams] = useSearchParams();

  const returnAlgorithm = searchParams.get('return') || 'KenPom Ratings';
  const investmentAlgorithm = searchParams.get('investment') || 'Ridge Regression 5 Features';
  const entryOptimizer = searchParams.get('optimizer') || 'MINLP';

  const title = useMemo(() => {
    const id = evaluationId as EvaluationRowId | undefined;
    if (id === 'return-algorithm') return 'Return Algorithm';
    if (id === 'investment-algorithm') return 'Investment Algorithm';
    if (id === 'entry-optimizer') return 'Entry Optimizer';
    return 'Evaluation';
  }, [evaluationId]);

  const rows = useMemo<SeasonEvaluationRow[]>(() => {
    const current = new Date().getFullYear();
    const seasons = [current - 4, current - 3, current - 2, current - 1, current];

    return seasons.map((season, idx) => {
      const base = 0.18 + idx * 0.01;
      return {
        season,
        meanPayout: 0.55 + idx * 0.03,
        pTop1: base * 0.25,
        pPayout: base,
        totalSimulations: 5000,
      };
    });
  }, []);

  return (
    <PageContainer>
      <PageHeader
        title={title}
        subtitle={
          <div>
            <div>Return Algorithm: {returnAlgorithm}</div>
            <div>Investment Algorithm: {investmentAlgorithm}</div>
            <div>Entry Optimizer: {entryOptimizer}</div>
          </div>
        }
        actions={
          <Link to="/evaluations" className="text-blue-600 hover:text-blue-800">
            ← Back to Evaluations
          </Link>
        }
      />

      <Card>
        <h2 className="text-xl font-semibold mb-4">Performance by Season Simulation</h2>
        <p className="text-gray-600 mb-6">Metrics shown: Mean Payout, P(Top 1), and P(Payout).</p>

        <div className="overflow-x-auto">
          <table className="min-w-full divide-y divide-gray-200">
            <thead className="bg-gray-50">
              <tr>
                <th className="px-4 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">Season</th>
                <th className="px-4 py-3 text-center text-xs font-medium text-gray-500 uppercase tracking-wider">Mean Payout</th>
                <th className="px-4 py-3 text-center text-xs font-medium text-gray-500 uppercase tracking-wider">P(Top 1)</th>
                <th className="px-4 py-3 text-center text-xs font-medium text-gray-500 uppercase tracking-wider">P(Payout)</th>
                <th className="px-4 py-3 text-center text-xs font-medium text-gray-500 uppercase tracking-wider">Simulations</th>
              </tr>
            </thead>
            <tbody className="bg-white divide-y divide-gray-200">
              {rows.map((r) => (
                <tr key={r.season} className="hover:bg-gray-50">
                  <td className="px-4 py-3 text-sm font-medium text-gray-900">{r.season}</td>
                  <td className="px-4 py-3 text-sm text-center text-gray-700">{r.meanPayout.toFixed(3)}</td>
                  <td className="px-4 py-3 text-sm text-center text-gray-700">{formatPercent(r.pTop1)}</td>
                  <td className="px-4 py-3 text-sm text-center text-gray-700">{formatPercent(r.pPayout)}</td>
                  <td className="px-4 py-3 text-sm text-center text-gray-500">{r.totalSimulations.toLocaleString()}</td>
                </tr>
              ))}
            </tbody>
          </table>
        </div>

        <div className="mt-4 text-sm text-gray-600">This page is UI-only for now (evaluation runs are not wired up yet).</div>
      </Card>
    </PageContainer>
  );
 }
