import React, { useMemo } from 'react';
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

export function SimulatedCalcuttaDetailPage() {
  const { id } = useParams<{ id: string }>();
  const [searchParams] = useSearchParams();
  const cohortId = searchParams.get('cohortId') || '';

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
                      <th className="px-3 py-2 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">Source</th>
                      <th className="px-3 py-2 text-right text-xs font-medium text-gray-500 uppercase tracking-wider">Teams</th>
                      <th className="px-3 py-2 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">Created</th>
                    </tr>
                  </thead>
                  <tbody className="bg-white divide-y divide-gray-200">
                    {entries.map((e) => (
                      <tr key={e.id} className="hover:bg-gray-50">
                        <td className="px-3 py-2 text-sm text-gray-900">
                          <div className="font-medium">{e.display_name}</div>
                          <div className="text-xs text-gray-500">{e.id.slice(0, 8)}</div>
                        </td>
                        <td className="px-3 py-2 text-sm text-gray-700">{e.source_kind}</td>
                        <td className="px-3 py-2 text-sm text-right text-gray-700">{e.teams?.length ?? 0}</td>
                        <td className="px-3 py-2 text-sm text-gray-700">{formatDateTime(e.created_at)}</td>
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
