import React, { useMemo, useState } from 'react';
import { useQuery } from '@tanstack/react-query';
import { Link, useParams } from 'react-router-dom';
import { Alert } from '../components/ui/Alert';
import { Button } from '../components/ui/Button';
import { Card } from '../components/ui/Card';
import { LoadingState } from '../components/ui/LoadingState';
import { PageContainer, PageHeader } from '../components/ui/Page';
import { Select } from '../components/ui/Select';
import { analyticsService } from '../services/analyticsService';

type EntryReportTeam = {
  team_id: string;
  school_name: string;
  seed: number;
  region: string;
  expected_points: number;
  expected_market: number;
  predicted_roi: number;
  our_bid: number;
  observed_roi: number;
};

type EntryReportResponse = {
  suite_scenario_id: string;
  suite_id: string;
  calcutta_id: string;
  calcutta_name: string;
  season: string;
  tournament_name: string;
  advancement_algorithm_id: string;
  advancement_algorithm_name: string;
  investment_algorithm_id: string;
  investment_algorithm_name: string;
  optimizer_key: string;
  strategy_generation_run_id?: string | null;
  game_outcome_run_id?: string | null;
  market_share_run_id?: string | null;
  budget_points: number;
  min_teams: number;
  max_teams: number;
  max_bid_points: number;
  assumed_entries: number;
  excluded_entry_name?: string | null;
  scoring_rules: { win_index: number; points_awarded: number }[];
  teams: EntryReportTeam[];
};

export function LabEntryReportPage() {
  const { scenarioId } = useParams<{ scenarioId: string }>();

  const reportQuery = useQuery<EntryReportResponse | null>({
    queryKey: ['lab', 'entries', 'scenario', scenarioId],
    queryFn: async () => {
      if (!scenarioId) return null;
      return analyticsService.getLabEntryReport<EntryReportResponse>(scenarioId);
    },
    enabled: Boolean(scenarioId),
  });

  const data = reportQuery.data;

  const [sortKey, setSortKey] = useState<'seed' | 'predicted_roi' | 'expected_points' | 'expected_market' | 'observed_roi' | 'our_bid'>('predicted_roi');
  const [sortDir, setSortDir] = useState<'asc' | 'desc'>('desc');

  const predictedROI = (t: EntryReportTeam) => {
    if (!Number.isFinite(t.expected_points) || !Number.isFinite(t.expected_market)) return NaN;
    return t.expected_points / (t.expected_market + 1);
  };

  const sortedTeams = useMemo(() => {
    const teams = data?.teams ?? [];
    const mult = sortDir === 'asc' ? 1 : -1;
    return teams.slice().sort((a, b) => {
      const av = sortKey === 'predicted_roi' ? predictedROI(a) : a[sortKey];
      const bv = sortKey === 'predicted_roi' ? predictedROI(b) : b[sortKey];
      if (!Number.isFinite(av) && !Number.isFinite(bv)) return 0;
      if (!Number.isFinite(av)) return 1;
      if (!Number.isFinite(bv)) return -1;
      return mult * (av - bv);
    });
  }, [data?.teams, sortKey, sortDir]);

  const fmt = (n: number, digits = 1) => {
    if (n === Infinity) return '∞';
    if (n === -Infinity) return '-∞';
    if (!Number.isFinite(n)) return '—';
    return n.toFixed(digits);
  };

  const scoringText = (rules: { win_index: number; points_awarded: number }[]) => {
    if (!rules.length) return '—';
    return rules
      .slice()
      .sort((a, b) => a.win_index - b.win_index)
      .map((r) => `${r.win_index}:${r.points_awarded}`)
      .join(' ');
  };

  return (
    <PageContainer className="max-w-none">
      <PageHeader
        title="Entry"
        subtitle={data ? `${data.season} — ${data.calcutta_name}` : scenarioId}
        leftActions={
          data ? (
            <Link to={`/lab/entries/suites/${encodeURIComponent(data.suite_id)}`} className="text-blue-600 hover:text-blue-800">
              ← Back to Algorithm Combo
            </Link>
          ) : (
            <Link to="/lab/entries" className="text-blue-600 hover:text-blue-800">
              ← Back to Entries
            </Link>
          )
        }
      />

      <div className="space-y-6">
        <Card>
          <h2 className="text-xl font-semibold mb-2">Summary</h2>

          {reportQuery.isLoading ? <LoadingState label="Loading entry report..." layout="inline" /> : null}

          {reportQuery.isError ? (
            <Alert variant="error" className="mt-3">
              <div className="font-semibold mb-1">Failed to load entry report</div>
              <div className="mb-3">{reportQuery.error instanceof Error ? reportQuery.error.message : 'An error occurred'}</div>
              <Button size="sm" onClick={() => reportQuery.refetch()}>
                Retry
              </Button>
            </Alert>
          ) : null}

          {data ? (
            <div className="text-sm text-gray-700">
              <div className="grid grid-cols-1 md:grid-cols-2 gap-2">
                <div>
                  <span className="font-medium text-gray-900">Advancement:</span> {data.advancement_algorithm_name}
                </div>
                <div>
                  <span className="font-medium text-gray-900">Investment:</span> {data.investment_algorithm_name}
                </div>
                <div>
                  <span className="font-medium text-gray-900">Optimizer:</span> {data.optimizer_key}
                </div>
                <div>
                  <span className="font-medium text-gray-900">Assumed entries:</span> {data.assumed_entries}
                  {data.excluded_entry_name ? ` (excluded: ${data.excluded_entry_name})` : ''}
                </div>
              </div>

              <div className="mt-3 grid grid-cols-1 md:grid-cols-2 gap-2 text-gray-600">
                <div>
                  rules: minTeams={data.min_teams} maxTeams={data.max_teams} maxBid={data.max_bid_points} budget={data.budget_points}
                </div>
                <div>scoring: {scoringText(data.scoring_rules)}</div>
              </div>

              <div className="mt-3 text-xs text-gray-500">
                game_outcome_run_id={data.game_outcome_run_id ?? '—'} | market_share_run_id={data.market_share_run_id ?? '—'} | strategy_generation_run_id={data.strategy_generation_run_id ?? '—'}
              </div>
            </div>
          ) : null}
        </Card>

        <Card>
          <div className="flex flex-col md:flex-row md:items-end md:justify-between gap-3 mb-4">
            <h2 className="text-xl font-semibold">Teams</h2>

            <div className="flex flex-col md:flex-row gap-2">
              <label className="text-sm text-gray-700">
                Sort
                <Select className="ml-2" value={sortKey} onChange={(e) => setSortKey(e.target.value as typeof sortKey)}>
                  <option value="seed">Seed</option>
                  <option value="expected_points">Expected Points</option>
                  <option value="expected_market">Expected Market</option>
                  <option value="predicted_roi">Predicted ROI</option>
                  <option value="our_bid">Our Bid</option>
                  <option value="observed_roi">Observed ROI</option>
                </Select>
              </label>

              <label className="text-sm text-gray-700">
                Dir
                <Select className="ml-2" value={sortDir} onChange={(e) => setSortDir(e.target.value as typeof sortDir)}>
                  <option value="asc">Asc</option>
                  <option value="desc">Desc</option>
                </Select>
              </label>
            </div>
          </div>

          {!reportQuery.isLoading && !reportQuery.isError && (data?.teams?.length ?? 0) === 0 ? (
            <Alert variant="info">No team rows found.</Alert>
          ) : null}

          {!reportQuery.isLoading && !reportQuery.isError && (data?.teams?.length ?? 0) > 0 ? (
            <div className="overflow-x-auto">
              <table className="min-w-full divide-y divide-gray-200">
                <thead className="bg-gray-50">
                  <tr>
                    <th className="px-3 py-2 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">Team</th>
                    <th className="px-3 py-2 text-right text-xs font-medium text-gray-500 uppercase tracking-wider">Seed</th>
                    <th className="px-3 py-2 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">Region</th>
                    <th className="px-3 py-2 text-right text-xs font-medium text-gray-500 uppercase tracking-wider">Returns</th>
                    <th className="px-3 py-2 text-right text-xs font-medium text-gray-500 uppercase tracking-wider">Market</th>
                    <th className="px-3 py-2 text-right text-xs font-medium text-gray-500 uppercase tracking-wider">Pred ROI</th>
                    <th className="px-3 py-2 text-right text-xs font-medium text-gray-500 uppercase tracking-wider bg-blue-50">Our Bid</th>
                    <th className="px-3 py-2 text-right text-xs font-medium text-gray-500 uppercase tracking-wider bg-blue-50">Obs ROI</th>
                  </tr>
                </thead>
                <tbody className="bg-white divide-y divide-gray-200">
                  {sortedTeams.map((t) => (
                    <tr key={t.team_id} className="hover:bg-gray-50">
                      <td className="px-3 py-2 text-sm text-gray-900">{t.school_name}</td>
                      <td className="px-3 py-2 text-sm text-right text-gray-700">{t.seed}</td>
                      <td className="px-3 py-2 text-sm text-gray-700">{t.region}</td>
                      <td className="px-3 py-2 text-sm text-right text-gray-700">{fmt(t.expected_points, 1)}</td>
                      <td className="px-3 py-2 text-sm text-right text-gray-700">{fmt(t.expected_market, 1)}</td>
                      <td className="px-3 py-2 text-sm text-right text-gray-700">{fmt(predictedROI(t), 2)}</td>
                      <td className="px-3 py-2 text-sm text-right font-semibold text-blue-700 bg-blue-50">{t.our_bid > 0 ? fmt(t.our_bid, 0) : '—'}</td>
                      <td className="px-3 py-2 text-sm text-right font-semibold text-blue-700 bg-blue-50">{t.our_bid > 0 ? fmt(t.observed_roi, 2) : '—'}</td>
                    </tr>
                  ))}
                </tbody>
              </table>
            </div>
          ) : null}
        </Card>
      </div>
    </PageContainer>
  );
}
