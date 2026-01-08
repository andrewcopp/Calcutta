import React, { useMemo } from 'react';
import { useQuery } from '@tanstack/react-query';
import { Link, useParams } from 'react-router-dom';

import { ApiError } from '../api/apiClient';
import { Alert } from '../components/ui/Alert';
import { Button } from '../components/ui/Button';
import { Card } from '../components/ui/Card';
import { LoadingState } from '../components/ui/LoadingState';
import { PageContainer, PageHeader } from '../components/ui/Page';
import { analyticsService } from '../services/analyticsService';

type CandidateTeam = {
  team_id: string;
  bid_points: number;
};

type CandidateDetailResponse = {
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
  teams: CandidateTeam[];
};

export function LabCandidateDetailPage() {
  const { candidateId } = useParams<{ candidateId: string }>();

  const showError = (err: unknown) => {
    if (err instanceof ApiError) {
      if (err.status === 403) return 'You do not have permission to view Lab candidates (403).';
      return `Request failed (${err.status}): ${err.message}`;
    }
    return err instanceof Error ? err.message : 'Unknown error';
  };

  const candidateQuery = useQuery<CandidateDetailResponse | null>({
    queryKey: ['lab', 'candidates', 'get', candidateId],
    queryFn: async () => {
      if (!candidateId) return null;
      return analyticsService.getLabCandidate<CandidateDetailResponse>(candidateId);
    },
    enabled: Boolean(candidateId),
  });

  const teams = useMemo(() => candidateQuery.data?.teams ?? [], [candidateQuery.data?.teams]);

  return (
    <PageContainer className="max-w-none">
      <PageHeader
        title="Candidate"
        subtitle={candidateId}
        leftActions={
          <Link to="/lab/candidates" className="text-blue-600 hover:text-blue-800">
            ← Back to Candidates
          </Link>
        }
      />

      {!candidateId ? <Alert variant="error">Missing candidateId.</Alert> : null}

      {candidateId && candidateQuery.isLoading ? <LoadingState label="Loading candidate..." /> : null}

      {candidateId && candidateQuery.isError ? (
        <Alert variant="error">
          <div className="font-semibold mb-1">Failed to load candidate</div>
          <div className="mb-3">{showError(candidateQuery.error)}</div>
          <Button size="sm" onClick={() => candidateQuery.refetch()}>
            Retry
          </Button>
        </Alert>
      ) : null}

      {candidateQuery.data ? (
        <div className="space-y-6">
          <Card>
            <h2 className="text-xl font-semibold mb-4">Provenance</h2>
            <div className="grid grid-cols-1 md:grid-cols-2 gap-4 text-sm">
              <div>
                <div className="text-gray-500">Display name</div>
                <div className="text-gray-900 font-medium break-words">{candidateQuery.data.display_name}</div>
              </div>
              <div>
                <div className="text-gray-500">Candidate ID</div>
                <div className="text-gray-900 break-words">{candidateQuery.data.candidate_id}</div>
              </div>
              <div>
                <div className="text-gray-500">Calcutta</div>
                <div className="text-gray-900 break-words">{candidateQuery.data.calcutta_id || '—'}</div>
              </div>
              <div>
                <div className="text-gray-500">Tournament</div>
                <div className="text-gray-900 break-words">{candidateQuery.data.tournament_id || '—'}</div>
              </div>
              <div>
                <div className="text-gray-500">Optimizer</div>
                <div className="text-gray-900">{candidateQuery.data.optimizer_key}</div>
              </div>
              <div>
                <div className="text-gray-500">Starting state</div>
                <div className="text-gray-900">{candidateQuery.data.starting_state_key}</div>
              </div>
              <div>
                <div className="text-gray-500">Excluded entry</div>
                <div className="text-gray-900 break-words">{candidateQuery.data.excluded_entry_name || '—'}</div>
              </div>
              <div>
                <div className="text-gray-500">Strategy generation run</div>
                <div className="text-gray-900 break-words">
                  {candidateQuery.data.strategy_generation_run_id ? (
                    <Link
                      to={`/lab/entry-runs/${encodeURIComponent(candidateQuery.data.strategy_generation_run_id)}`}
                      className="text-blue-600 hover:text-blue-800"
                    >
                      {candidateQuery.data.strategy_generation_run_id}
                    </Link>
                  ) : (
                    '—'
                  )}
                </div>
              </div>
              <div>
                <div className="text-gray-500">Metrics artifact</div>
                <div className="text-gray-900 break-words">
                  {candidateQuery.data.source_entry_artifact_id ? (
                    <Link
                      to={`/lab/entry-artifacts/${encodeURIComponent(candidateQuery.data.source_entry_artifact_id)}`}
                      className="text-blue-600 hover:text-blue-800"
                    >
                      {candidateQuery.data.source_entry_artifact_id}
                    </Link>
                  ) : (
                    '—'
                  )}
                </div>
              </div>
            </div>
          </Card>

          <Card>
            <h2 className="text-xl font-semibold mb-4">Bids</h2>
            {teams.length === 0 ? <Alert variant="info">No bids available yet.</Alert> : null}
            {teams.length > 0 ? (
              <div className="overflow-x-auto">
                <table className="min-w-full divide-y divide-gray-200">
                  <thead className="bg-gray-50">
                    <tr>
                      <th className="px-3 py-2 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">Team</th>
                      <th className="px-3 py-2 text-right text-xs font-medium text-gray-500 uppercase tracking-wider">Bid</th>
                    </tr>
                  </thead>
                  <tbody className="bg-white divide-y divide-gray-200">
                    {teams.map((t) => (
                      <tr key={t.team_id}>
                        <td className="px-3 py-2 text-sm text-gray-700 break-words">{t.team_id}</td>
                        <td className="px-3 py-2 text-sm text-gray-900 text-right font-medium">{t.bid_points}</td>
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
