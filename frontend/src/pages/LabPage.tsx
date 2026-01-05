import React, { useEffect, useMemo, useState } from 'react';
import { useQuery } from '@tanstack/react-query';
import { useNavigate, useSearchParams } from 'react-router-dom';
import { TabsNav } from '../components/TabsNav';
import { Card } from '../components/ui/Card';
import { PageContainer, PageHeader } from '../components/ui/Page';
import { analyticsService } from '../services/analyticsService';

type TabType = 'advancements' | 'investments';

type Algorithm = {
  id: string;
  kind: string;
  name: string;
  description?: string | null;
  params_json?: unknown;
  created_at?: string;
};

type AlgorithmCoverageItem = {
  id: string;
  name: string;
  description?: string | null;
  covered: number;
  total: number;
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

type TeamPredictedAdvancement = {
  team_id: string;
  school_name: string;
  seed: number;
  region: string;
  prob_pi: number;
  reach_r64: number;
  reach_r32: number;
  reach_s16: number;
  reach_e8: number;
  reach_ff: number;
  reach_champ: number;
  win_champ: number;
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

export function LabPage() {
  const navigate = useNavigate();
  const [searchParams, setSearchParams] = useSearchParams();

  const [activeTab, setActiveTab] = useState<TabType>(() => {
    const tab = searchParams.get('tab');
    return tab === 'investments' ? 'investments' : 'advancements';
  });
  const [selectedAlgorithmId, setSelectedAlgorithmId] = useState<string>(() => searchParams.get('algorithmId') || '');

  useEffect(() => {
    const next = new URLSearchParams();
    next.set('tab', activeTab);
    if (selectedAlgorithmId) next.set('algorithmId', selectedAlgorithmId);

    if (next.toString() !== searchParams.toString()) {
      setSearchParams(next, { replace: true });
    }
  }, [
    activeTab,
    selectedAlgorithmId,
    searchParams,
    setSearchParams,
  ]);

  const tabs = useMemo(
    () =>
      [
        { id: 'advancements' as const, label: 'Advancements' },
        { id: 'investments' as const, label: 'Investments' },
      ] as const,
    []
  );

  const algorithmsQuery = useQuery<{ items: Algorithm[] } | null>({
    queryKey: ['analytics', 'algorithms', activeTab],
    queryFn: async () => {
      const kind = activeTab === 'advancements' ? 'game_outcomes' : 'market_share';
      const filtered = await analyticsService.listAlgorithms<{ items: Algorithm[] }>(kind);
      if (filtered?.items?.length) return filtered;
      return analyticsService.listAlgorithms<{ items: Algorithm[] }>();
    },
  });

  const coverageQuery = useQuery<{ items: AlgorithmCoverageItem[] } | null>({
    queryKey: ['analytics', 'coverage', activeTab],
    queryFn: async () => {
      if (activeTab === 'advancements') {
        return analyticsService.getGameOutcomesAlgorithmCoverage<{ items: AlgorithmCoverageItem[] }>();
      }
      return analyticsService.getMarketShareAlgorithmCoverage<{ items: AlgorithmCoverageItem[] }>();
    },
  });

  const coverageItems = coverageQuery.data?.items ?? [];
  const algoById = useMemo(() => {
    const m = new Map<string, Algorithm>();
    for (const a of algorithmsQuery.data?.items ?? []) {
      m.set(a.id, a);
    }
    return m;
  }, [algorithmsQuery.data?.items]);

  const rows = useMemo(() => {
    if (coverageItems.length) return coverageItems;
    // Fallback: if coverage endpoint isn't populated for some reason, show algorithms without counts.
    return (algorithmsQuery.data?.items ?? []).map((a) => ({
      id: a.id,
      name: a.name,
      description: a.description,
      covered: 0,
      total: 0,
    }));
  }, [coverageItems, algorithmsQuery.data?.items]);

  const sortedRows = useMemo(() => {
    return rows
      .slice()
      .sort((a, b) => {
        if (b.covered !== a.covered) return b.covered - a.covered;
        if (b.total !== a.total) return b.total - a.total;
        return a.name.localeCompare(b.name);
      });
  }, [rows]);

  const formatCoverage = (covered: number, total: number) => {
    if (!Number.isFinite(total) || total <= 0) return `${covered}`;
    return `${covered}/${total}`;
  };

  return (
    <PageContainer>
      <PageHeader title="Lab" subtitle="Browse registered algorithms and their run outputs." />

      <Card className="mb-6">
        <TabsNav tabs={tabs} activeTab={activeTab} onTabChange={setActiveTab} />
      </Card>

      <Card>
        <h2 className="text-xl font-semibold mb-4">Algorithms</h2>

        {coverageQuery.isLoading || algorithmsQuery.isLoading ? <div className="text-gray-500">Loading algorithms...</div> : null}

        {!coverageQuery.isLoading && !algorithmsQuery.isLoading && rows.length === 0 ? (
          <div className="text-gray-700">No algorithms found.</div>
        ) : null}

        {!coverageQuery.isLoading && !algorithmsQuery.isLoading && rows.length > 0 ? (
          <div className="overflow-x-auto">
            <table className="min-w-full divide-y divide-gray-200">
              <thead className="bg-gray-50">
                <tr>
                  <th className="px-3 py-2 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">Algorithm</th>
                  <th className="px-3 py-2 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">Coverage</th>
                </tr>
              </thead>
              <tbody className="bg-white divide-y divide-gray-200">
                {sortedRows.map((row) => {
                  const detailUrl =
                    activeTab === 'advancements'
                      ? `/lab/advancements/algorithms/${encodeURIComponent(row.id)}`
                      : `/lab/investments/algorithms/${encodeURIComponent(row.id)}`;
                  const alg = algoById.get(row.id);
                  const desc = row.description ?? alg?.description;
                  return (
                    <tr
                      key={row.id}
                      className="hover:bg-gray-50 cursor-pointer"
                      onClick={() => {
                        setSelectedAlgorithmId(row.id);
                        navigate(detailUrl);
                      }}
                    >
                      <td className="px-3 py-2 text-sm text-gray-900">
                        <div className="font-medium">{row.name}</div>
                        {desc ? <div className="text-xs text-gray-600">{desc}</div> : null}
                      </td>
                      <td className="px-3 py-2 text-sm text-gray-700">{formatCoverage(row.covered, row.total)}</td>
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
