import React, { useState, useEffect } from 'react';
import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query';
import { useParams } from 'react-router-dom';
import { BarChart, Bar, XAxis, YAxis, Tooltip, ResponsiveContainer, CartesianGrid } from 'recharts';

import { Alert } from '../../components/ui/Alert';
import { Breadcrumb } from '../../components/ui/Breadcrumb';
import { Button } from '../../components/ui/Button';
import { Card } from '../../components/ui/Card';
import { LoadingState } from '../../components/ui/LoadingState';
import { PageContainer, PageHeader } from '../../components/ui/Page';
import { PipelineSummary } from '../../components/Lab/PipelineSummary';
import { PipelineStatusTable } from '../../components/Lab/PipelineStatusTable';
import { labService } from '../../services/labService';
import type { InvestmentModel, ModelPipelineProgress } from '../../types/lab';
import { queryKeys } from '../../queryKeys';
import { formatDate } from '../../utils/labFormatters';

export function ModelDetailPage() {
  const { modelId } = useParams<{ modelId: string }>();
  const queryClient = useQueryClient();

  const [isPipelineRunning, setIsPipelineRunning] = useState(false);

  const modelQuery = useQuery<InvestmentModel | null>({
    queryKey: queryKeys.lab.models.detail(modelId),
    queryFn: () => (modelId ? labService.getModel(modelId) : Promise.resolve(null)),
    enabled: Boolean(modelId),
  });

  const pipelineProgressQuery = useQuery<ModelPipelineProgress | null>({
    queryKey: queryKeys.lab.models.pipelineProgress(modelId),
    queryFn: () => (modelId ? labService.getModelPipelineProgress(modelId) : Promise.resolve(null)),
    enabled: Boolean(modelId),
    refetchInterval: isPipelineRunning ? 2000 : false,
  });

  // Update pipeline running state based on query data
  useEffect(() => {
    if (pipelineProgressQuery.data?.active_pipeline_run_id) {
      setIsPipelineRunning(true);
    } else {
      setIsPipelineRunning(false);
    }
  }, [pipelineProgressQuery.data?.active_pipeline_run_id]);

  const startPipelineMutation = useMutation({
    mutationFn: () => labService.startPipeline(modelId!),
    onSuccess: () => {
      setIsPipelineRunning(true);
      queryClient.invalidateQueries({ queryKey: queryKeys.lab.models.pipelineProgress(modelId) });
    },
  });

  const rerunAllMutation = useMutation({
    mutationFn: () => labService.startPipeline(modelId!, { force_rerun: true }),
    onSuccess: () => {
      setIsPipelineRunning(true);
      queryClient.invalidateQueries({ queryKey: queryKeys.lab.models.pipelineProgress(modelId) });
    },
  });

  const cancelPipelineMutation = useMutation({
    mutationFn: () => {
      const runId = pipelineProgressQuery.data?.active_pipeline_run_id;
      if (!runId) throw new Error('No active pipeline to cancel');
      return labService.cancelPipeline(runId);
    },
    onSuccess: () => {
      setIsPipelineRunning(false);
      queryClient.invalidateQueries({ queryKey: queryKeys.lab.models.pipelineProgress(modelId) });
    },
  });

  const model = modelQuery.data;
  const pipelineProgress = pipelineProgressQuery.data;

  const [showParams, setShowParams] = useState(false);

  // Build cross-calcutta performance data for chart
  const performanceData = (pipelineProgress?.calcuttas ?? [])
    .filter((c) => c.has_evaluation && c.mean_payout != null)
    .sort((a, b) => a.calcutta_year - b.calcutta_year)
    .map((c) => ({
      name: String(c.calcutta_year),
      payout: c.mean_payout ?? 0,
      rank: c.our_rank ?? undefined,
    }));

  if (modelQuery.isLoading) {
    return (
      <PageContainer>
        <LoadingState label="Loading model..." />
      </PageContainer>
    );
  }

  if (modelQuery.isError || !model) {
    return (
      <PageContainer>
        <Alert variant="error">Failed to load model.</Alert>
      </PageContainer>
    );
  }

  return (
    <PageContainer>
      <Breadcrumb
        items={[
          { label: 'Lab', href: '/lab' },
          { label: 'Models', href: '/lab?tab=models' },
          { label: model.name },
        ]}
      />

      <PageHeader title={model.name} subtitle={`Kind: ${model.kind}`} />

      {model.notes && (
        <div className="bg-blue-50 border border-blue-200 rounded-lg p-4 mb-6">
          <div className="text-xs text-blue-600 uppercase font-semibold mb-1">Hypothesis</div>
          <p className="text-sm text-blue-900">{model.notes}</p>
        </div>
      )}

      <Card className="mb-6">
        <h2 className="text-lg font-semibold mb-3">Model Details</h2>
        <dl className="grid grid-cols-2 gap-4 text-sm">
          <div>
            <dt className="text-gray-500">Kind</dt>
            <dd className="font-medium">{model.kind}</dd>
          </div>
          <div>
            <dt className="text-gray-500">Created</dt>
            <dd className="font-medium">{formatDate(model.created_at)}</dd>
          </div>
          <div>
            <dt className="text-gray-500">Entries</dt>
            <dd className="font-medium">{model.n_entries}</dd>
          </div>
          <div>
            <dt className="text-gray-500">Evaluations</dt>
            <dd className="font-medium">{model.n_evaluations}</dd>
          </div>
        </dl>

        {model.params_json && Object.keys(model.params_json).length > 0 && (
          <div className="mt-4 pt-4 border-t border-gray-200">
            <Button
              variant="ghost"
              size="sm"
              onClick={() => setShowParams(!showParams)}
            >
              {showParams ? '▼' : '▶'} Model Parameters ({Object.keys(model.params_json).length})
            </Button>
            {showParams && (
              <dl className="grid grid-cols-2 gap-x-4 gap-y-2 text-sm mt-3">
                {Object.entries(model.params_json).map(([key, value]) => (
                  <div key={key}>
                    <dt className="text-gray-500 font-mono text-xs">{key}</dt>
                    <dd className="font-medium">
                      {typeof value === 'object' ? JSON.stringify(value) : String(value)}
                    </dd>
                  </div>
                ))}
              </dl>
            )}
          </div>
        )}
      </Card>

      {startPipelineMutation.isError && (
        <Alert variant="error" className="mb-4">
          Failed to start pipeline: {startPipelineMutation.error instanceof Error ? startPipelineMutation.error.message : 'Unknown error'}
        </Alert>
      )}

      {cancelPipelineMutation.isError && (
        <Alert variant="error" className="mb-4">
          Failed to cancel pipeline: {cancelPipelineMutation.error instanceof Error ? cancelPipelineMutation.error.message : 'Unknown error'}
        </Alert>
      )}

      {rerunAllMutation.isError && (
        <Alert variant="error" className="mb-4">
          Failed to re-run pipeline: {rerunAllMutation.error instanceof Error ? rerunAllMutation.error.message : 'Unknown error'}
        </Alert>
      )}

      <PipelineSummary
        progress={pipelineProgress ?? null}
        isLoading={pipelineProgressQuery.isLoading}
        isPipelineRunning={isPipelineRunning}
        onStartPipeline={() => startPipelineMutation.mutate()}
        onRerunAll={() => rerunAllMutation.mutate()}
        onCancelPipeline={() => cancelPipelineMutation.mutate()}
        isStarting={startPipelineMutation.isPending}
        isRerunning={rerunAllMutation.isPending}
        isCancelling={cancelPipelineMutation.isPending}
      />

      {performanceData.length > 1 && (
        <Card className="mb-6">
          <h2 className="text-lg font-semibold mb-3">Cross-Calcutta Performance</h2>
          <p className="text-sm text-gray-500 mb-4">Mean payout by calcutta year. 1.0x = break-even.</p>
          <ResponsiveContainer width="100%" height={250}>
            <BarChart data={performanceData}>
              <CartesianGrid strokeDasharray="3 3" />
              <XAxis dataKey="name" fontSize={12} />
              <YAxis domain={[0, 'auto']} fontSize={12} tickFormatter={(v: number) => `${v.toFixed(1)}x`} />
              <Tooltip
                formatter={(value: number) => [`${value.toFixed(2)}x`, 'Payout']}
                labelFormatter={(label: string) => `Year: ${label}`}
              />
              <Bar
                dataKey="payout"
                fill="#3b82f6"
                radius={[4, 4, 0, 0]}
              />
            </BarChart>
          </ResponsiveContainer>
        </Card>
      )}

      <h2 className="text-lg font-semibold mb-3">Historical Calcuttas</h2>
      <PipelineStatusTable
        calcuttas={pipelineProgress?.calcuttas ?? []}
        modelName={model.name}
        isLoading={pipelineProgressQuery.isLoading}
      />
    </PageContainer>
  );
}
