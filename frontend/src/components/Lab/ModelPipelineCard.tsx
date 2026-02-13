import React, { useState } from 'react';
import { useNavigate } from 'react-router-dom';
import { cn } from '../../lib/cn';
import { PipelineStatusCell, PipelineStatus } from './PipelineStatusCell';

export type CalcuttaPipelineRow = {
  calcuttaId: string;
  calcuttaName: string;
  entry: {
    id: string;
    status: PipelineStatus;
    optimizer?: string;
    startingState?: string;
  } | null;
  evaluation: {
    id: string;
    status: PipelineStatus;
    nSims?: number;
    meanPayout?: number | null;
  } | null;
};

export type ModelPipelineData = {
  modelId: string;
  modelName: string;
  modelKind: string;
  avgPayout: number | null;
  avgPTop1: number | null;
  rank: number;
  totalCalcuttas: number;
  completedEvaluations: number;
  rows: CalcuttaPipelineRow[];
};

type ModelPipelineCardProps = {
  data: ModelPipelineData;
  defaultExpanded?: boolean;
};

function formatPayoutX(val: number | null | undefined): string {
  if (val == null) return '−';
  return `${val.toFixed(2)}x`;
}

function getPayoutColor(val: number | null | undefined): string {
  if (val == null) return 'text-gray-400';
  if (val >= 1.2) return 'text-green-600';
  if (val < 0.9) return 'text-red-600';
  return 'text-gray-900';
}

export function ModelPipelineCard({ data, defaultExpanded = false }: ModelPipelineCardProps) {
  const [isExpanded, setIsExpanded] = useState(defaultExpanded);
  const navigate = useNavigate();

  const completionText = `${data.completedEvaluations}/${data.totalCalcuttas}`;
  const isFullyComplete = data.completedEvaluations === data.totalCalcuttas;

  return (
    <div className="bg-surface rounded-lg shadow-sm border border-gray-200 overflow-hidden">
      {/* Header - always visible */}
      <button
        type="button"
        onClick={() => setIsExpanded(!isExpanded)}
        className="w-full text-left px-4 py-3 hover:bg-gray-50 transition-colors focus:outline-none focus-visible:ring-2 focus-visible:ring-blue-500 focus-visible:ring-inset"
      >
        <div className="flex items-center justify-between">
          <div className="flex items-center gap-3">
            <span
              className={cn(
                'text-sm text-gray-400 transition-transform',
                isExpanded ? 'rotate-90' : ''
              )}
            >
              ▶
            </span>
            <div className="flex items-baseline gap-2">
              <h3 className="text-base font-semibold text-gray-900">{data.modelName}</h3>
              <span className="text-sm text-gray-400">{data.modelKind}</span>
            </div>
          </div>

          <div className="flex items-center gap-4">
            {/* Primary metric - payout */}
            <span className={cn('text-lg font-bold tabular-nums', getPayoutColor(data.avgPayout))}>
              {formatPayoutX(data.avgPayout)}
            </span>
            {/* Secondary metrics */}
            <span className="text-sm text-gray-500 tabular-nums">#{data.rank}</span>
            <span
              className={cn(
                'text-sm font-medium tabular-nums',
                isFullyComplete ? 'text-green-600' : 'text-amber-600'
              )}
            >
              {completionText}
            </span>
          </div>
        </div>
      </button>

      {/* Expanded content - pipeline matrix */}
      {isExpanded ? (
        <div className="border-t border-gray-200">
          <table className="min-w-full">
            <thead className="bg-gray-50 border-b border-gray-200">
              <tr>
                <th className="px-4 py-1.5 text-left text-xs font-medium text-gray-500 uppercase tracking-wider w-1/4 border-r border-gray-200">
                  Calcutta
                </th>
                <th className="px-4 py-1.5 text-left text-xs font-medium text-gray-500 uppercase tracking-wider w-3/8 bg-blue-50/50">
                  Entry
                </th>
                <th className="px-4 py-1.5 text-left text-xs font-medium text-gray-500 uppercase tracking-wider w-3/8 bg-blue-50/50">
                  Evaluation
                </th>
              </tr>
            </thead>
            <tbody className="divide-y divide-gray-100">
              {data.rows.map((row) => (
                <tr key={row.calcuttaId} className="hover:bg-gray-50">
                  <td className="px-4 py-1.5 text-sm text-gray-700 border-r border-gray-100">{row.calcuttaName}</td>
                  <td className="px-2 py-0.5 bg-blue-50/30">
                    {row.entry ? (
                      <PipelineStatusCell
                        status={row.entry.status}
                        label={row.entry.optimizer}
                        onClick={() => navigate(`/lab/models/${encodeURIComponent(data.modelName)}/calcutta/${encodeURIComponent(row.calcuttaId)}`)}
                      />
                    ) : (
                      <PipelineStatusCell status="missing" />
                    )}
                  </td>
                  <td className="px-2 py-0.5 bg-blue-50/30">
                    {row.evaluation ? (
                      <PipelineStatusCell
                        status={row.evaluation.status}
                        label={row.evaluation.nSims ? `${(row.evaluation.nSims / 1000).toFixed(0)}k` : undefined}
                        metric={row.evaluation.meanPayout}
                        metricFormat="payout"
                        onClick={() => navigate(`/lab/evaluations/${row.evaluation!.id}`)}
                      />
                    ) : (
                      <PipelineStatusCell status="missing" />
                    )}
                  </td>
                </tr>
              ))}
              {data.rows.length === 0 ? (
                <tr>
                  <td colSpan={3} className="px-4 py-3 text-sm text-gray-500 text-center">
                    No calcuttas found.
                  </td>
                </tr>
              ) : null}
            </tbody>
          </table>
        </div>
      ) : null}
    </div>
  );
}
