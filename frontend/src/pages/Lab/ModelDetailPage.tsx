import React, { useState, useEffect } from 'react';
import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query';
import { useParams } from 'react-router-dom';

import { Alert } from '../../components/ui/Alert';
import { Breadcrumb } from '../../components/ui/Breadcrumb';
import { Card } from '../../components/ui/Card';
import { LoadingState } from '../../components/ui/LoadingState';
import { PageContainer, PageHeader } from '../../components/ui/Page';
import { PipelineSummary } from '../../components/Lab/PipelineSummary';
import { PipelineStatusTable } from '../../components/Lab/PipelineStatusTable';
import { labService, InvestmentModel, ModelPipelineProgress } from '../../services/labService';

export function ModelDetailPage() {
  const { modelId } = useParams<{ modelId: string }>();
  const queryClient = useQueryClient();

  const [isPipelineRunning, setIsPipelineRunning] = useState(false);

  const modelQuery = useQuery<InvestmentModel | null>({
    queryKey: ['lab', 'models', modelId],
    queryFn: () => (modelId ? labService.getModel(modelId) : Promise.resolve(null)),
    enabled: Boolean(modelId),
  });

  const pipelineProgressQuery = useQuery<ModelPipelineProgress | null>({
    queryKey: ['lab', 'models', modelId, 'pipeline-progress'],
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
      queryClient.invalidateQueries({ queryKey: ['lab', 'models', modelId, 'pipeline-progress'] });
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
      queryClient.invalidateQueries({ queryKey: ['lab', 'models', modelId, 'pipeline-progress'] });
    },
  });

  const model = modelQuery.data;
  const pipelineProgress = pipelineProgressQuery.data;

  const formatDate = (dateStr: string) => {
    const d = new Date(dateStr);
    return d.toLocaleDateString('en-US', { month: 'short', day: 'numeric', year: 'numeric' });
  };

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
          {model.notes ? (
            <div className="col-span-2">
              <dt className="text-gray-500">Notes</dt>
              <dd className="font-medium">{model.notes}</dd>
            </div>
          ) : null}
        </dl>
      </Card>

      {startPipelineMutation.isError && (
        <Alert variant="error" className="mb-4">
          Failed to start pipeline: {(startPipelineMutation.error as Error)?.message || 'Unknown error'}
        </Alert>
      )}

      {cancelPipelineMutation.isError && (
        <Alert variant="error" className="mb-4">
          Failed to cancel pipeline: {(cancelPipelineMutation.error as Error)?.message || 'Unknown error'}
        </Alert>
      )}

      <PipelineSummary
        progress={pipelineProgress ?? null}
        isLoading={pipelineProgressQuery.isLoading}
        isPipelineRunning={isPipelineRunning}
        onStartPipeline={() => startPipelineMutation.mutate()}
        onCancelPipeline={() => cancelPipelineMutation.mutate()}
        isStarting={startPipelineMutation.isPending}
        isCancelling={cancelPipelineMutation.isPending}
      />

      <h2 className="text-lg font-semibold mb-3">Historical Calcuttas</h2>
      <PipelineStatusTable
        calcuttas={pipelineProgress?.calcuttas ?? []}
        modelName={model.name}
        isLoading={pipelineProgressQuery.isLoading}
      />
    </PageContainer>
  );
}
