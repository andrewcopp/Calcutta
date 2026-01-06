import React, { useMemo } from 'react';
import { useQuery } from '@tanstack/react-query';
import { Link, useParams, useSearchParams } from 'react-router-dom';

import { ApiError } from '../api/apiClient';
import { Alert } from '../components/ui/Alert';
import { Button } from '../components/ui/Button';
import { Card } from '../components/ui/Card';
import { LoadingState } from '../components/ui/LoadingState';
import { PageContainer, PageHeader } from '../components/ui/Page';
import {
  suiteCalcuttaEvaluationsService,
  type SuiteCalcuttaEvaluationSnapshotEntryResponse,
} from '../services/suiteCalcuttaEvaluationsService';

export function SuiteCalcuttaEvaluationEntryDetailPage() {
  const { id, snapshotEntryId } = useParams<{ id: string; snapshotEntryId: string }>();
  const [searchParams] = useSearchParams();

  const suiteId = searchParams.get('suiteId') || '';
  const executionId = searchParams.get('executionId') || '';
  const backUrl = id
    ? `/sandbox/evaluations/${encodeURIComponent(id)}${suiteId ? `?suiteId=${encodeURIComponent(suiteId)}${executionId ? `&executionId=${encodeURIComponent(executionId)}` : ''}` : ''}`
    : '/sandbox/suites';

  const showError = (err: unknown) => {
    if (err instanceof ApiError) {
      if (err.status === 403) {
        return 'You do not have permission to view simulation run entries (403).';
      }
      return `Request failed (${err.status}): ${err.message}`;
    }
    return err instanceof Error ? err.message : 'Unknown error';
  };

  const entryQuery = useQuery<SuiteCalcuttaEvaluationSnapshotEntryResponse>({
    queryKey: ['simulation-runs', 'snapshot-entry', id, snapshotEntryId],
    queryFn: () => suiteCalcuttaEvaluationsService.getSnapshotEntry(id!, snapshotEntryId!),
    enabled: Boolean(id && snapshotEntryId),
  });

  const sortedTeams = useMemo(() => {
    const teams = entryQuery.data?.teams ?? [];
    return teams.slice().sort((a, b) => (b.bid_points || 0) - (a.bid_points || 0));
  }, [entryQuery.data?.teams]);

  return (
    <PageContainer className="max-w-none">
      <PageHeader
        title="Entry"
        subtitle={entryQuery.data ? entryQuery.data.display_name : snapshotEntryId}
        leftActions={
          <Link to={backUrl} className="text-blue-600 hover:text-blue-800">
            ← Back
          </Link>
        }
      />

      {!id ? <Alert variant="error">Missing simulation run ID.</Alert> : null}
      {!snapshotEntryId ? <Alert variant="error">Missing snapshot entry ID.</Alert> : null}

      {entryQuery.isLoading ? <LoadingState label="Loading entry..." /> : null}

      {entryQuery.isError ? (
        <Alert variant="error">
          <div className="font-semibold mb-1">Failed to load entry</div>
          <div className="mb-3">{showError(entryQuery.error)}</div>
          <Button size="sm" onClick={() => entryQuery.refetch()}>
            Retry
          </Button>
        </Alert>
      ) : null}

      {entryQuery.data ? (
        <div className="space-y-6">
          <Card>
            <h2 className="text-xl font-semibold mb-4">Portfolio</h2>

            {sortedTeams.length === 0 ? (
              <Alert variant="info">No teams found for this entry snapshot.</Alert>
            ) : (
              <div className="overflow-x-auto">
                <table className="min-w-full divide-y divide-gray-200">
                  <thead className="bg-gray-50">
                    <tr>
                      <th className="px-3 py-2 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">Team</th>
                      <th className="px-3 py-2 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">Seed</th>
                      <th className="px-3 py-2 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">Region</th>
                      <th className="px-3 py-2 text-right text-xs font-medium text-gray-500 uppercase tracking-wider">Bid</th>
                    </tr>
                  </thead>
                  <tbody className="bg-white divide-y divide-gray-200">
                    {sortedTeams.map((t) => (
                      <tr key={t.team_id}>
                        <td className="px-3 py-2 text-sm text-gray-900">{t.school_name}</td>
                        <td className="px-3 py-2 text-sm text-gray-700">{t.seed || '—'}</td>
                        <td className="px-3 py-2 text-sm text-gray-700">{t.region || '—'}</td>
                        <td className="px-3 py-2 text-sm text-right text-gray-900 font-medium">{t.bid_points}</td>
                      </tr>
                    ))}
                  </tbody>
                </table>
              </div>
            )}
          </Card>
        </div>
      ) : null}
    </PageContainer>
  );
}
