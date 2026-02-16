import React from 'react';
import { useNavigate } from 'react-router-dom';
import type { CalcuttaPipelineStatus } from '../../types/lab';

type PipelineStatusTableProps = {
  calcuttas: CalcuttaPipelineStatus[];
  modelName: string;
  isLoading: boolean;
};

function StatusIcon({ completed, running }: { completed: boolean; running: boolean }) {
  if (running) {
    return (
      <span className="inline-flex items-center justify-center w-5 h-5">
        <span className="animate-spin h-4 w-4 border-2 border-blue-500 border-t-transparent rounded-full"></span>
      </span>
    );
  }
  if (completed) {
    return <span className="text-green-500">&#10003;</span>;
  }
  return <span className="text-gray-300">&#9675;</span>;
}

function StageProgress({ stage }: { stage: string }) {
  const stageOrder = ['predictions', 'optimization', 'evaluation', 'completed'];
  const currentIndex = stageOrder.indexOf(stage);

  return (
    <div className="flex items-center space-x-1">
      {stageOrder.slice(0, 3).map((s, i) => {
        const isComplete = currentIndex > i || stage === 'completed';
        const isCurrent = s === stage;
        return (
          <div
            key={s}
            className={`h-1.5 w-6 rounded ${
              isComplete
                ? 'bg-green-500'
                : isCurrent
                ? 'bg-blue-500'
                : 'bg-gray-200'
            }`}
            title={s}
          ></div>
        );
      })}
      <span className="ml-2 text-xs text-gray-500 capitalize">{stage}</span>
    </div>
  );
}

export function PipelineStatusTable({ calcuttas, modelName, isLoading }: PipelineStatusTableProps) {
  const navigate = useNavigate();

  if (isLoading) {
    return (
      <div className="bg-white rounded-lg border border-gray-200 p-4">
        <div className="animate-pulse space-y-3">
          <div className="h-5 bg-gray-200 rounded w-1/4"></div>
          <div className="h-8 bg-gray-200 rounded"></div>
          <div className="h-8 bg-gray-200 rounded"></div>
          <div className="h-8 bg-gray-200 rounded"></div>
        </div>
      </div>
    );
  }

  if (calcuttas.length === 0) {
    return (
      <div className="bg-white rounded-lg border border-gray-200 p-4">
        <p className="text-gray-500 text-center py-4">No historical calcuttas available.</p>
      </div>
    );
  }

  const formatPayout = (val?: number | null) => {
    if (val == null) return '-';
    return `${val.toFixed(3)}x`;
  };

  return (
    <div className="bg-white rounded-lg border border-gray-200 overflow-hidden">
      <div className="overflow-x-auto">
        <table className="min-w-full divide-y divide-gray-200">
          <thead className="bg-gray-50">
            <tr>
              <th className="px-4 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">
                Calcutta
              </th>
              <th className="px-4 py-3 text-center text-xs font-medium text-gray-500 uppercase tracking-wider">
                Predictions
              </th>
              <th className="px-4 py-3 text-center text-xs font-medium text-gray-500 uppercase tracking-wider">
                Entry
              </th>
              <th className="px-4 py-3 text-center text-xs font-medium text-gray-500 uppercase tracking-wider">
                Evaluation
              </th>
              <th className="px-4 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">
                Progress
              </th>
              <th className="px-4 py-3 text-right text-xs font-medium text-gray-500 uppercase tracking-wider">
                Result
              </th>
              <th className="px-4 py-3 text-right text-xs font-medium text-gray-500 uppercase tracking-wider">
                Rank
              </th>
            </tr>
          </thead>
          <tbody className="bg-white divide-y divide-gray-200">
            {calcuttas.map((c) => {
              const isRunning = c.status === 'running';
              const isFailed = c.status === 'failed';
              const canNavigate = c.has_entry && c.entry_id;

              return (
                <tr
                  key={c.calcutta_id}
                  className={`${
                    canNavigate ? 'hover:bg-gray-50 cursor-pointer' : ''
                  } ${isFailed ? 'bg-red-50' : ''}`}
                  onClick={() => {
                    if (canNavigate) {
                      navigate(`/lab/models/${encodeURIComponent(modelName)}/calcutta/${encodeURIComponent(c.calcutta_id)}`);
                    }
                  }}
                >
                  <td className="px-4 py-3">
                    <div className="text-sm font-medium text-gray-900">{c.calcutta_name}</div>
                    <div className="text-xs text-gray-500">{c.calcutta_year}</div>
                  </td>
                  <td className="px-4 py-3 text-center">
                    <StatusIcon
                      completed={c.has_predictions}
                      running={isRunning && c.stage === 'predictions'}
                    />
                  </td>
                  <td className="px-4 py-3 text-center">
                    <StatusIcon
                      completed={c.has_entry}
                      running={isRunning && c.stage === 'optimization'}
                    />
                  </td>
                  <td className="px-4 py-3 text-center">
                    <StatusIcon
                      completed={c.has_evaluation}
                      running={isRunning && c.stage === 'evaluation'}
                    />
                  </td>
                  <td className="px-4 py-3">
                    {isFailed ? (
                      <span className="text-xs text-red-600" title={c.error_message || 'Failed'}>
                        Failed
                      </span>
                    ) : isRunning ? (
                      <StageProgress stage={c.stage} />
                    ) : c.has_evaluation ? (
                      <span className="text-xs text-green-600">Complete</span>
                    ) : (
                      <span className="text-xs text-gray-400">Pending</span>
                    )}
                  </td>
                  <td className="px-4 py-3 text-right">
                    {c.has_evaluation ? (
                      <span
                        className={`text-sm font-medium ${
                          c.mean_payout && c.mean_payout >= 1 ? 'text-green-600' : 'text-gray-900'
                        }`}
                      >
                        {formatPayout(c.mean_payout)}
                      </span>
                    ) : (
                      <span className="text-sm text-gray-400">-</span>
                    )}
                  </td>
                  <td className="px-4 py-3 text-right">
                    {c.has_evaluation && c.our_rank != null ? (
                      <span className="text-sm font-medium text-gray-900">
                        #{c.our_rank}
                      </span>
                    ) : (
                      <span className="text-sm text-gray-400">-</span>
                    )}
                  </td>
                </tr>
              );
            })}
          </tbody>
        </table>
      </div>
    </div>
  );
}
