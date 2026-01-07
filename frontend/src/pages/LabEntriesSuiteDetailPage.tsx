import React, { useMemo } from 'react';
import { useQuery } from '@tanstack/react-query';
import { Link, useNavigate, useParams } from 'react-router-dom';
import { Alert } from '../components/ui/Alert';
import { Button } from '../components/ui/Button';
import { Card } from '../components/ui/Card';
import { LoadingState } from '../components/ui/LoadingState';
import { PageContainer, PageHeader } from '../components/ui/Page';
import { analyticsService } from '../services/analyticsService';

type SuiteDetailResponse = {
  cohort: {
    id: string;
    name: string;
    advancement_algorithm: { id: string; name: string };
    investment_algorithm: { id: string; name: string };
    optimizer_key: string;
    starting_state_key: string;
    excluded_entry_name?: string | null;
  };
  items: {
    scenario_id: string;
    calcutta_id: string;
    calcutta_name: string;
    tournament_name: string;
    season: string;
    team_count: number;
    entry_created_at?: string | null;
    scenario_created_at: string;
    strategy_generation_run_id?: string | null;
  }[];
};

type CreateSandboxExecutionResponse = {
  executionId: string;
  evaluationCount: number;
};

export function LabEntriesSuiteDetailPage() {
  const { cohortId } = useParams<{ cohortId: string }>();
  const navigate = useNavigate();

  const detailQuery = useQuery<SuiteDetailResponse | null>({
    queryKey: ['lab', 'entries', 'cohort', cohortId],
    queryFn: async () => {
      if (!cohortId) return null;
      return analyticsService.getLabEntriesCohortDetail<SuiteDetailResponse>(cohortId);
    },
    enabled: Boolean(cohortId),
  });

  const items = detailQuery.data?.items ?? [];

  const canRunInSandbox = useMemo(() => {
    return items.some((it) => Boolean(it.strategy_generation_run_id));
  }, [items]);

  const sorted = useMemo(() => {
    return items
      .slice()
      .sort((a, b) => {
        // Prefer season desc; fallback to created_at desc.
        const ay = Number(a.season);
        const by = Number(b.season);
        if (Number.isFinite(ay) && Number.isFinite(by) && by !== ay) return by - ay;
        return (b.entry_created_at ?? b.scenario_created_at).localeCompare(a.entry_created_at ?? a.scenario_created_at);
      });
  }, [items]);

  const suite = detailQuery.data?.cohort ?? null;

  const fmtDateTime = (iso?: string | null) => {
    if (!iso) return '—';
    const d = new Date(iso);
    if (Number.isNaN(d.getTime())) return '—';
    return d.toLocaleString();
  };

  return (
    <PageContainer className="max-w-none">
      <PageHeader
        title="Entries"
        subtitle={suite ? suite.name : cohortId}
        leftActions={
          <Link to="/lab/entries" className="text-blue-600 hover:text-blue-800">
            ← Back to Entries
          </Link>
        }
        actions={
				<Button
					size="sm"
					disabled={!cohortId || detailQuery.isLoading || !canRunInSandbox}
					onClick={async () => {
						if (!cohortId) return;
					const res = await analyticsService.createLabCohortSandboxExecution<CreateSandboxExecutionResponse>(cohortId);
					navigate(`/sandbox/cohorts/${encodeURIComponent(cohortId)}?executionId=${encodeURIComponent(res.executionId)}`);
				}}
			>
					Run in Sandbox
				</Button>
        }
      />

      <div className="space-y-6">
        <Card>
          <h2 className="text-xl font-semibold mb-2">Algorithm Combo</h2>
          {detailQuery.isLoading ? <LoadingState label="Loading suite..." layout="inline" /> : null}

          {suite ? (
            <div className="text-sm text-gray-700">
              <div>
                <span className="font-medium text-gray-900">Advancement:</span> {suite.advancement_algorithm.name}
              </div>
              <div>
                <span className="font-medium text-gray-900">Investment:</span> {suite.investment_algorithm.name}
              </div>
              <div>
                <span className="font-medium text-gray-900">Optimizer:</span> {suite.optimizer_key}
              </div>
              <div className="text-gray-600 mt-1">
                starting_state_key={suite.starting_state_key}
                {suite.excluded_entry_name ? ` | excluded_entry_name=${suite.excluded_entry_name}` : ''}
              </div>
            </div>
          ) : null}

          {detailQuery.isError ? (
            <Alert variant="error" className="mt-3">
              <div className="font-semibold mb-1">Failed to load suite</div>
              <div className="mb-3">{detailQuery.error instanceof Error ? detailQuery.error.message : 'An error occurred'}</div>
              <Button size="sm" onClick={() => detailQuery.refetch()}>
                Retry
              </Button>
            </Alert>
          ) : null}
        </Card>

        <Card>
          <h2 className="text-xl font-semibold mb-4">Calcuttas</h2>

          {!detailQuery.isLoading && !detailQuery.isError && sorted.length === 0 ? (
            <Alert variant="info">No calcuttas found for this algorithm combo.</Alert>
          ) : null}

          {!detailQuery.isLoading && !detailQuery.isError && sorted.length > 0 ? (
            <div className="overflow-x-auto">
              <table className="min-w-full divide-y divide-gray-200">
                <thead className="bg-gray-50">
                  <tr>
                    <th className="px-3 py-2 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">Season</th>
                    <th className="px-3 py-2 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">Calcutta</th>
                    <th className="px-3 py-2 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">Teams</th>
                    <th className="px-3 py-2 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">Entry Created</th>
                  </tr>
                </thead>
                <tbody className="bg-white divide-y divide-gray-200">
                  {sorted.map((row) => (
                    <tr
                      key={row.scenario_id}
                      className="hover:bg-gray-50 cursor-pointer"
                      onClick={() => navigate(`/lab/entries/scenarios/${encodeURIComponent(row.scenario_id)}`)}
                    >
                      <td className="px-3 py-2 text-sm text-gray-700">{row.season}</td>
                      <td className="px-3 py-2 text-sm text-gray-900">
                        <div className="font-medium">{row.calcutta_name}</div>
                        <div className="text-xs text-gray-600">{row.tournament_name}</div>
                      </td>
                      <td className="px-3 py-2 text-sm text-gray-700">{row.team_count}</td>
                      <td className="px-3 py-2 text-sm text-gray-700">{fmtDateTime(row.entry_created_at)}</td>
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
