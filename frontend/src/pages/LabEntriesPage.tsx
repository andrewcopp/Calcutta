import React, { useMemo } from 'react';
import { useQuery } from '@tanstack/react-query';
import { Link, useNavigate } from 'react-router-dom';
import { Alert } from '../components/ui/Alert';
import { Button } from '../components/ui/Button';
import { Card } from '../components/ui/Card';
import { LoadingState } from '../components/ui/LoadingState';
import { PageContainer, PageHeader } from '../components/ui/Page';
import { analyticsService } from '../services/analyticsService';

type CoverageItem = {
  suite_id: string;
  suite_name: string;
  advancement_algorithm_id: string;
  advancement_algorithm_name: string;
  investment_algorithm_id: string;
  investment_algorithm_name: string;
  optimizer_key: string;
  covered: number;
  total: number;
};

type CoverageResponse = {
  items: CoverageItem[];
};

export function LabEntriesPage() {
  const navigate = useNavigate();

  const coverageQuery = useQuery<CoverageResponse>({
    queryKey: ['lab', 'entries', 'coverage'],
    queryFn: () => analyticsService.listLabEntriesCoverage<CoverageResponse>(),
  });

  const items = coverageQuery.data?.items ?? [];

  const sorted = useMemo(() => {
    return items
      .slice()
      .sort((a, b) => {
        const aPct = a.total > 0 ? a.covered / a.total : 0;
        const bPct = b.total > 0 ? b.covered / b.total : 0;
        if (bPct !== aPct) return bPct - aPct;
        if (b.covered !== a.covered) return b.covered - a.covered;
        if (b.total !== a.total) return b.total - a.total;
        return a.suite_name.localeCompare(b.suite_name);
      });
  }, [items]);

  const formatCoverage = (covered: number, total: number) => {
    if (!Number.isFinite(total) || total <= 0) return String(covered);
    return `${covered}/${total}`;
  };

  return (
    <PageContainer className="max-w-none">
      <PageHeader
        title="Entries"
        subtitle="Generated entries by algorithm combo (advancement + investment + optimizer)."
        actions={
          <Link to="/lab" className="text-blue-600 hover:text-blue-800">
            ← Back to Lab
          </Link>
        }
      />

      <Card>
        <h2 className="text-xl font-semibold mb-4">Algorithm Combos</h2>

        {coverageQuery.isLoading ? <LoadingState label="Loading entries coverage..." layout="inline" /> : null}

        {coverageQuery.isError ? (
          <Alert variant="error" className="mt-3">
            <div className="font-semibold mb-1">Failed to load entries coverage</div>
            <div className="mb-3">{coverageQuery.error instanceof Error ? coverageQuery.error.message : 'An error occurred'}</div>
            <Button size="sm" onClick={() => coverageQuery.refetch()}>
              Retry
            </Button>
          </Alert>
        ) : null}

        {!coverageQuery.isLoading && !coverageQuery.isError && sorted.length === 0 ? (
          <Alert variant="info">No algorithm combos found.</Alert>
        ) : null}

        {!coverageQuery.isLoading && !coverageQuery.isError && sorted.length > 0 ? (
          <div className="overflow-x-auto">
            <table className="min-w-full divide-y divide-gray-200">
              <thead className="bg-gray-50">
                <tr>
                  <th className="px-3 py-2 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">Combo</th>
                  <th className="px-3 py-2 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">Advancement</th>
                  <th className="px-3 py-2 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">Investment</th>
                  <th className="px-3 py-2 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">Optimizer</th>
                  <th className="px-3 py-2 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">Calcuttas</th>
                </tr>
              </thead>
              <tbody className="bg-white divide-y divide-gray-200">
                {sorted.map((row) => (
                  <tr
                    key={row.suite_id}
                    className="hover:bg-gray-50 cursor-pointer"
                    onClick={() => navigate(`/lab/entries/suites/${encodeURIComponent(row.suite_id)}`)}
                  >
                    <td className="px-3 py-2 text-sm text-gray-900">
                      <div className="font-medium">{row.suite_name}</div>
                    </td>
                    <td className="px-3 py-2 text-sm text-gray-700">{row.advancement_algorithm_name}</td>
                    <td className="px-3 py-2 text-sm text-gray-700">{row.investment_algorithm_name}</td>
                    <td className="px-3 py-2 text-sm text-gray-700">{row.optimizer_key}</td>
                    <td className="px-3 py-2 text-sm text-gray-700">{formatCoverage(row.covered, row.total)}</td>
                  </tr>
                ))}
              </tbody>
            </table>
          </div>
        ) : null}
      </Card>
    </PageContainer>
  );
}
