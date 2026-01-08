import React, { useMemo } from 'react';
import { useQuery } from '@tanstack/react-query';
import { Link, useNavigate, useSearchParams } from 'react-router-dom';

import { ApiError } from '../api/apiClient';
import { Alert } from '../components/ui/Alert';
import { Card } from '../components/ui/Card';
import { LoadingState } from '../components/ui/LoadingState';
import { PageContainer, PageHeader } from '../components/ui/Page';
import { Input } from '../components/ui/Input';
import { analyticsService } from '../services/analyticsService';

type CandidateListItem = {
  candidate_id: string;
  display_name: string;
  source_kind: string;
  source_entry_artifact_id?: string | null;
  calcutta_id: string;
  tournament_id: string;
  strategy_generation_run_id: string;
  market_share_run_id: string;
  market_share_artifact_id: string;
  advancement_run_id: string;
  optimizer_key: string;
  starting_state_key: string;
  excluded_entry_name?: string | null;
  git_sha?: string | null;
};

type ListCandidatesResponse = {
  items: CandidateListItem[];
};

export function LabCandidatesPage() {
  const navigate = useNavigate();
  const [searchParams, setSearchParams] = useSearchParams();

  const calcuttaId = searchParams.get('calcuttaId') || '';
  const optimizerKey = searchParams.get('optimizerKey') || '';

  const listQuery = useQuery<ListCandidatesResponse>({
    queryKey: ['lab', 'candidates', 'list', calcuttaId, optimizerKey],
    queryFn: async () => {
      return analyticsService.listLabCandidates<ListCandidatesResponse>({
        calcuttaId: calcuttaId || undefined,
        optimizerKey: optimizerKey || undefined,
        limit: 200,
      });
    },
  });

  const showError = (err: unknown) => {
    if (err instanceof ApiError) {
      if (err.status === 403) return 'You do not have permission to view Lab candidates (403).';
      return `Request failed (${err.status}): ${err.message}`;
    }
    return err instanceof Error ? err.message : 'Unknown error';
  };

  const items = useMemo(() => listQuery.data?.items ?? [], [listQuery.data?.items]);

  return (
    <PageContainer className="max-w-none">
      <PageHeader
        title="Candidates"
        subtitle="Lab candidates (one per calcutta per config)."
        leftActions={
          <Link to="/lab" className="text-blue-600 hover:text-blue-800">
            ← Back to Lab
          </Link>
        }
      />

      <Card className="mb-6">
        <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
          <div>
            <div className="text-xs text-gray-500 mb-1">Filter: calcuttaId</div>
            <Input
              value={calcuttaId}
              placeholder="calcutta UUID"
              onChange={(e) => {
                const next = new URLSearchParams(searchParams);
                const v = e.target.value;
                if (v) next.set('calcuttaId', v);
                else next.delete('calcuttaId');
                setSearchParams(next, { replace: true });
              }}
            />
          </div>
          <div>
            <div className="text-xs text-gray-500 mb-1">Filter: optimizerKey</div>
            <Input
              value={optimizerKey}
              placeholder="minlp_v1"
              onChange={(e) => {
                const next = new URLSearchParams(searchParams);
                const v = e.target.value;
                if (v) next.set('optimizerKey', v);
                else next.delete('optimizerKey');
                setSearchParams(next, { replace: true });
              }}
            />
          </div>
        </div>
      </Card>

      <Card>
        <div className="flex items-center justify-between mb-4">
          <h2 className="text-xl font-semibold">Candidates</h2>
        </div>

        {listQuery.isLoading ? <LoadingState label="Loading candidates..." layout="inline" /> : null}

        {listQuery.isError ? (
          <Alert variant="error" className="mt-3">
            <div className="font-semibold mb-1">Failed to load candidates</div>
            <div className="mb-3">{showError(listQuery.error)}</div>
          </Alert>
        ) : null}

        {!listQuery.isLoading && !listQuery.isError && items.length === 0 ? (
          <Alert variant="info">No candidates found for the current filters.</Alert>
        ) : null}

        {!listQuery.isLoading && !listQuery.isError && items.length > 0 ? (
          <div className="overflow-x-auto">
            <table className="min-w-full divide-y divide-gray-200">
              <thead className="bg-gray-50">
                <tr>
                  <th className="px-3 py-2 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">Candidate</th>
                  <th className="px-3 py-2 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">Calcutta</th>
                  <th className="px-3 py-2 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">Optimizer</th>
                  <th className="px-3 py-2 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">Starting</th>
                </tr>
              </thead>
              <tbody className="bg-white divide-y divide-gray-200">
                {items.map((c) => (
                  <tr
                    key={c.candidate_id}
                    className="hover:bg-gray-50 cursor-pointer"
                    onClick={() => navigate(`/lab/candidates/${encodeURIComponent(c.candidate_id)}`)}
                  >
                    <td className="px-3 py-2 text-sm text-gray-900">
                      <div className="font-medium">{c.display_name}</div>
                      <div className="text-xs text-gray-500 break-words">{c.candidate_id}</div>
                    </td>
                    <td className="px-3 py-2 text-sm text-gray-700 break-words">{c.calcutta_id || '—'}</td>
                    <td className="px-3 py-2 text-sm text-gray-700">{c.optimizer_key || '—'}</td>
                    <td className="px-3 py-2 text-sm text-gray-700">{c.starting_state_key || '—'}</td>
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
