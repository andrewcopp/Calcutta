import React from 'react';
import { Button } from '../ui/Button';
import type { ModelPipelineProgress } from '../../types/lab';
import { formatPayoutX } from '../../utils/labFormatters';

type PipelineSummaryProps = {
  progress: ModelPipelineProgress | null;
  isLoading: boolean;
  isPipelineRunning: boolean;
  onStartPipeline: () => void;
  onRerunAll: () => void;
  onCancelPipeline: () => void;
  isStarting: boolean;
  isRerunning: boolean;
  isCancelling: boolean;
};

export function PipelineSummary({
  progress,
  isLoading,
  isPipelineRunning,
  onStartPipeline,
  onRerunAll,
  onCancelPipeline,
  isStarting,
  isRerunning,
  isCancelling,
}: PipelineSummaryProps) {
  if (isLoading || !progress) {
    return (
      <div className="bg-white rounded-lg border border-gray-200 p-4 mb-6">
        <div className="animate-pulse">
          <div className="h-5 bg-gray-200 rounded w-1/3 mb-3"></div>
          <div className="h-4 bg-gray-200 rounded w-2/3"></div>
        </div>
      </div>
    );
  }

  const completedCount = progress.evaluationsCount;
  const totalCount = progress.totalCalcuttas;
  const progressPercent = totalCount > 0 ? (completedCount / totalCount) * 100 : 0;

  return (
    <div className="bg-white rounded-lg border border-gray-200 p-4 mb-6">
      <div className="flex items-center justify-between mb-4">
        <h2 className="text-lg font-semibold">Pipeline Progress</h2>
        {isPipelineRunning ? (
          <Button size="sm" variant="secondary" onClick={onCancelPipeline} disabled={isCancelling}>
            {isCancelling ? 'Cancelling...' : 'Cancel Pipeline'}
          </Button>
        ) : (
          <div className="flex gap-2">
            <Button
              size="sm"
              onClick={onStartPipeline}
              disabled={isStarting || isRerunning || completedCount === totalCount}
            >
              {isStarting ? 'Starting...' : completedCount === totalCount ? 'All Complete' : 'Run Missing'}
            </Button>
            <Button
              size="sm"
              variant="secondary"
              onClick={onRerunAll}
              disabled={isStarting || isRerunning || totalCount === 0}
            >
              {isRerunning ? 'Re-running...' : 'Re-run All'}
            </Button>
          </div>
        )}
      </div>

      <div className="mb-4">
        <div className="flex justify-between text-sm text-gray-600 mb-1">
          <span>Overall Progress</span>
          <span>
            {completedCount}/{totalCount} calcuttas evaluated
          </span>
        </div>
        <div className="w-full bg-gray-200 rounded-full h-2.5">
          <div
            className={`h-2.5 rounded-full transition-all duration-300 ${
              isPipelineRunning ? 'bg-blue-500 animate-pulse' : 'bg-green-500'
            }`}
            style={{ width: `${progressPercent}%` }}
          ></div>
        </div>
      </div>

      <div className="grid grid-cols-2 md:grid-cols-4 gap-4 text-sm">
        <div>
          <dt className="text-gray-500">Predictions</dt>
          <dd className="font-medium flex items-center">
            {progress.predictionsCount}/{totalCount}
            {progress.predictionsCount === totalCount && <span className="ml-1 text-green-500">&#10003;</span>}
          </dd>
        </div>
        <div>
          <dt className="text-gray-500">Entries</dt>
          <dd className="font-medium flex items-center">
            {progress.entriesCount}/{totalCount}
            {progress.entriesCount === totalCount && <span className="ml-1 text-green-500">&#10003;</span>}
          </dd>
        </div>
        <div>
          <dt className="text-gray-500">Evaluations</dt>
          <dd className="font-medium flex items-center">
            {progress.evaluationsCount}/{totalCount}
            {progress.evaluationsCount === totalCount && <span className="ml-1 text-green-500">&#10003;</span>}
          </dd>
        </div>
        <div>
          <dt className="text-gray-500">Avg Payout</dt>
          <dd className="font-medium">{formatPayoutX(progress.avgMeanPayout)}</dd>
        </div>
      </div>
    </div>
  );
}
