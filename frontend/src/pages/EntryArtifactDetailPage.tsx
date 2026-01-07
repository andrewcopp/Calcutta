import React, { useMemo } from 'react';
import { useQuery } from '@tanstack/react-query';
import { Link, useParams } from 'react-router-dom';

import { ApiError } from '../api/apiClient';
import { Alert } from '../components/ui/Alert';
import { Button } from '../components/ui/Button';
import { Card } from '../components/ui/Card';
import { LoadingState } from '../components/ui/LoadingState';
import { PageContainer, PageHeader } from '../components/ui/Page';
import { entryArtifactsService } from '../services/entryArtifactsService';

export function EntryArtifactDetailPage() {
  const { artifactId } = useParams<{ artifactId: string }>();

  const showError = (err: unknown) => {
    if (err instanceof ApiError) {
      if (err.status === 403) return 'You do not have permission to view entry artifacts (403).';
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

  const artifactQuery = useQuery({
    queryKey: ['entry-artifacts', 'get', artifactId],
    queryFn: () => entryArtifactsService.get(artifactId || ''),
    enabled: Boolean(artifactId),
  });

  const summaryText = useMemo(() => {
    const v = artifactQuery.data?.summary_json;
    if (v == null) return '';
    try {
      return JSON.stringify(v, null, 2);
    } catch {
      return String(v);
    }
  }, [artifactQuery.data?.summary_json]);

  return (
    <PageContainer className="max-w-none">
      <PageHeader
        title="Entry Artifact"
        subtitle={artifactId}
        leftActions={
          artifactQuery.data?.run_id ? (
            <Link to={`/lab/entry-runs/${encodeURIComponent(artifactQuery.data.run_id)}`} className="text-blue-600 hover:text-blue-800">
              ← Back to Entry Run
            </Link>
          ) : (
            <Link to="/lab" className="text-blue-600 hover:text-blue-800">
              ← Back to Lab
            </Link>
          )
        }
      />

      {!artifactId ? <Alert variant="error">Missing entry artifact ID.</Alert> : null}

      {artifactId && artifactQuery.isLoading ? <LoadingState label="Loading entry artifact..." /> : null}
      {artifactId && artifactQuery.isError ? (
        <Alert variant="error">
          <div className="font-semibold mb-1">Failed to load entry artifact</div>
          <div className="mb-3">{showError(artifactQuery.error)}</div>
          <Button size="sm" onClick={() => artifactQuery.refetch()}>
            Retry
          </Button>
        </Alert>
      ) : null}

      {artifactQuery.data ? (
        <div className="space-y-6">
          <Card>
            <h2 className="text-xl font-semibold mb-4">Artifact</h2>
            <div className="grid grid-cols-1 md:grid-cols-2 gap-4 text-sm">
              <div>
                <div className="text-gray-500">ID</div>
                <div className="text-gray-900 font-medium break-words">{artifactQuery.data.id}</div>
              </div>
              <div>
                <div className="text-gray-500">Kind</div>
                <div className="text-gray-900">{artifactQuery.data.artifact_kind}</div>
              </div>
              <div>
                <div className="text-gray-500">Run</div>
                <div className="text-gray-900 break-words">
                  <Link to={`/lab/entry-runs/${encodeURIComponent(artifactQuery.data.run_id)}`} className="text-blue-600 hover:text-blue-800">
                    {artifactQuery.data.run_id}
                  </Link>
                </div>
              </div>
              <div>
                <div className="text-gray-500">Schema version</div>
                <div className="text-gray-900">{artifactQuery.data.schema_version}</div>
              </div>
              <div>
                <div className="text-gray-500">Storage URI</div>
                <div className="text-gray-900 break-words">{artifactQuery.data.storage_uri || '—'}</div>
              </div>
              <div>
                <div className="text-gray-500">Created</div>
                <div className="text-gray-900">{formatDateTime(artifactQuery.data.created_at)}</div>
              </div>
              <div>
                <div className="text-gray-500">Updated</div>
                <div className="text-gray-900">{formatDateTime(artifactQuery.data.updated_at)}</div>
              </div>
              <div>
                <div className="text-gray-500">Lineage (inputs)</div>
                <div className="text-gray-900 text-xs break-words">
                  <div>market_share={artifactQuery.data.input_market_share_artifact_id || '—'}</div>
                  <div>advancement={artifactQuery.data.input_advancement_artifact_id || '—'}</div>
                </div>
              </div>
            </div>
          </Card>

          <Card>
            <h2 className="text-xl font-semibold mb-4">Summary</h2>
            {!summaryText ? <Alert variant="info">No summary_json available.</Alert> : null}
            {summaryText ? (
              <pre className="text-xs bg-gray-50 border border-gray-200 rounded p-3 overflow-auto max-h-[50vh]">
                {summaryText}
              </pre>
            ) : null}
          </Card>
        </div>
      ) : null}
    </PageContainer>
  );
}
