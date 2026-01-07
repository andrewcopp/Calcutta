import React from 'react';
import { useQuery } from '@tanstack/react-query';
import { useNavigate } from 'react-router-dom';

import { ApiError } from '../api/apiClient';
import { Alert } from '../components/ui/Alert';
import { Button } from '../components/ui/Button';
import { Card } from '../components/ui/Card';
import { LoadingState } from '../components/ui/LoadingState';
import { PageContainer, PageHeader } from '../components/ui/Page';
import { cohortsService, type CohortListItem } from '../services/cohortsService';

export function SandboxCohortsListPage() {
  const navigate = useNavigate();
  const listQuery = useQuery({
    queryKey: ['cohorts', 'list'],
    queryFn: () => cohortsService.list({ limit: 200, offset: 0 }),
  });

  const showError = (err: unknown) => {
    if (err instanceof ApiError) {
      if (err.status === 403) {
        return 'You do not have permission to view cohorts (403).';
      }
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

  const suites: CohortListItem[] = listQuery.data?.items ?? [];

  return (
    <PageContainer>
      <PageHeader title="Sandbox" subtitle="Cohorts (collections of synthetic calcuttas)" />

      <Card>
        <h2 className="text-xl font-semibold mb-4">Cohorts</h2>

        {listQuery.isLoading ? <LoadingState label="Loading cohorts..." layout="inline" /> : null}

        {listQuery.isError ? (
          <Alert variant="error" className="mt-3">
            <div className="font-semibold mb-1">Failed to load cohorts</div>
            <div className="mb-3">{showError(listQuery.error)}</div>
            <Button size="sm" onClick={() => listQuery.refetch()}>
              Retry
            </Button>
          </Alert>
        ) : null}

        {!listQuery.isLoading && !listQuery.isError && suites.length === 0 ? (
          <Alert variant="info" className="mt-3">
            No cohorts found.
          </Alert>
        ) : null}

        {!listQuery.isLoading && !listQuery.isError && suites.length > 0 ? (
          <div className="overflow-x-auto">
            <table className="min-w-full divide-y divide-gray-200">
              <thead className="bg-gray-50">
                <tr>
                  <th className="px-3 py-2 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">Cohort</th>
                  <th className="px-3 py-2 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">Latest simulation run batch</th>
                  <th className="px-3 py-2 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">Updated</th>
                </tr>
              </thead>
              <tbody className="bg-white divide-y divide-gray-200">
                {suites.map((s) => {
                  const exec = s.latest_execution_id
                    ? `${s.latest_execution_status ?? '—'} · ${s.latest_execution_id.slice(0, 8)}`
                    : '—';

                  const href = `/sandbox/cohorts/${encodeURIComponent(s.id)}`;

                  return (
                    <tr
                      key={s.id}
                      className="hover:bg-gray-50 cursor-pointer focus-visible:outline focus-visible:outline-2 focus-visible:outline-blue-500"
                      role="link"
                      tabIndex={0}
                      onClick={() => navigate(href)}
                      onKeyDown={(e) => {
                        if (e.key === 'Enter' || e.key === ' ') {
                          e.preventDefault();
                          navigate(href);
                        }
                      }}
                      aria-label={`Open cohort ${s.name || s.id}`}
                    >
                      <td className="px-3 py-2 text-sm text-gray-900">
                        <div className="font-medium">
                          <span className="text-blue-600 hover:text-blue-800">{s.name || s.id}</span>
                        </div>
                        <div className="text-xs text-gray-600">
                          {s.optimizer_key} · n={s.n_sims} · seed={s.seed}
                        </div>
                      </td>
                      <td className="px-3 py-2 text-sm text-gray-700">
                        <div>{exec}</div>
                        <div className="text-xs text-gray-600">{formatDateTime(s.latest_execution_created_at ?? null)}</div>
                      </td>
                      <td className="px-3 py-2 text-sm text-gray-700">{formatDateTime(s.updated_at)}</td>
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
