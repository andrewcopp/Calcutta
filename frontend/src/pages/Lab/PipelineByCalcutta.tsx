import React, { useMemo, useState } from 'react';
import { useQuery } from '@tanstack/react-query';
import { useNavigate } from 'react-router-dom';

import { Alert } from '../../components/ui/Alert';
import { LoadingState } from '../../components/ui/LoadingState';
import { PipelineStatusCell, PipelineStatus } from '../../components/Lab/PipelineStatusCell';
import {
  labService,
  LeaderboardResponse,
  ListEntriesResponse,
  ListEvaluationsResponse,
  EntryDetail,
  EvaluationDetail,
} from '../../services/labService';
import { calcuttaService } from '../../services/calcuttaService';
import { Calcutta } from '../../types/calcutta';
import { cn } from '../../lib/cn';

type ModelRow = {
  modelId: string;
  modelName: string;
  modelKind: string;
  entry: {
    id: string;
    status: PipelineStatus;
    optimizer: string;
  } | null;
  evaluation: {
    id: string;
    status: PipelineStatus;
    nSims: number;
    meanPayout: number | null;
  } | null;
};

type CalcuttaPipelineData = {
  calcuttaId: string;
  calcuttaName: string;
  modelsWithEntries: number;
  modelsWithEvaluations: number;
  totalModels: number;
  rows: ModelRow[];
};

function buildCalcuttaPipelineData(
  leaderboard: LeaderboardResponse,
  entries: EntryDetail[],
  evaluations: EvaluationDetail[],
  calcuttas: Calcutta[]
): CalcuttaPipelineData[] {
  // Group entries by calcutta
  const entriesByCalcutta = new Map<string, EntryDetail[]>();
  for (const entry of entries) {
    const list = entriesByCalcutta.get(entry.calcutta_id) || [];
    list.push(entry);
    entriesByCalcutta.set(entry.calcutta_id, list);
  }

  // Group evaluations by entry
  const evalsByEntry = new Map<string, EvaluationDetail[]>();
  for (const ev of evaluations) {
    const list = evalsByEntry.get(ev.entry_id) || [];
    list.push(ev);
    evalsByEntry.set(ev.entry_id, list);
  }

  const result: CalcuttaPipelineData[] = [];

  for (const calcutta of calcuttas) {
    const calcuttaEntries = entriesByCalcutta.get(calcutta.id) || [];

    // Map entries by model
    const entriesByModel = new Map<string, EntryDetail>();
    for (const entry of calcuttaEntries) {
      entriesByModel.set(entry.investment_model_id, entry);
    }

    // Build rows for each model
    const rows: ModelRow[] = [];
    for (const model of leaderboard.items) {
      const entry = entriesByModel.get(model.investment_model_id);
      const entryEvals = entry ? evalsByEntry.get(entry.id) || [] : [];
      const bestEval = entryEvals.length > 0 ? entryEvals[0] : null;

      rows.push({
        modelId: model.investment_model_id,
        modelName: model.model_name,
        modelKind: model.model_kind,
        entry: entry
          ? {
              id: entry.id,
              status: 'complete' as PipelineStatus,
              optimizer: entry.optimizer_kind,
            }
          : null,
        evaluation: bestEval
          ? {
              id: bestEval.id,
              status: 'complete' as PipelineStatus,
              nSims: bestEval.n_sims,
              meanPayout: bestEval.mean_normalized_payout ?? null,
            }
          : null,
      });
    }

    // Sort by evaluation payout (best first), then by model name
    rows.sort((a, b) => {
      const aPayout = a.evaluation?.meanPayout ?? -Infinity;
      const bPayout = b.evaluation?.meanPayout ?? -Infinity;
      if (bPayout !== aPayout) return bPayout - aPayout;
      return a.modelName.localeCompare(b.modelName);
    });

    result.push({
      calcuttaId: calcutta.id,
      calcuttaName: calcutta.name,
      modelsWithEntries: rows.filter((r) => r.entry != null).length,
      modelsWithEvaluations: rows.filter((r) => r.evaluation != null).length,
      totalModels: leaderboard.items.length,
      rows,
    });
  }

  // Sort calcuttas by name descending (most recent first)
  result.sort((a, b) => b.calcuttaName.localeCompare(a.calcuttaName));

  return result;
}

type CalcuttaPipelineCardProps = {
  data: CalcuttaPipelineData;
  defaultExpanded?: boolean;
};

function CalcuttaPipelineCard({ data, defaultExpanded = false }: CalcuttaPipelineCardProps) {
  const [isExpanded, setIsExpanded] = useState(defaultExpanded);
  const navigate = useNavigate();

  const completionText = `${data.modelsWithEvaluations}/${data.totalModels}`;
  const isFullyComplete = data.modelsWithEvaluations === data.totalModels;

  return (
    <div className="bg-surface rounded-lg shadow border border-gray-200 overflow-hidden">
      {/* Header */}
      <button
        type="button"
        onClick={() => setIsExpanded(!isExpanded)}
        className="w-full text-left p-4 hover:bg-gray-50 transition-colors focus:outline-none focus-visible:ring-2 focus-visible:ring-blue-500 focus-visible:ring-inset"
      >
        <div className="flex items-center justify-between">
          <div className="flex items-center gap-3">
            <span className={cn('text-lg transition-transform', isExpanded ? 'rotate-90' : '')}>
              ▶
            </span>
            <h3 className="text-lg font-semibold text-gray-900">{data.calcuttaName}</h3>
          </div>

          <div className="flex items-center gap-6 text-sm">
            <div className="text-center">
              <div className="text-gray-500">Entries</div>
              <div className="font-medium text-gray-700">{data.modelsWithEntries}</div>
            </div>
            <div className="text-center">
              <div className="text-gray-500">Evaluated</div>
              <div
                className={cn(
                  'font-medium',
                  isFullyComplete ? 'text-green-600' : 'text-amber-600'
                )}
              >
                {completionText}
              </div>
            </div>
          </div>
        </div>
      </button>

      {/* Expanded content */}
      {isExpanded ? (
        <div className="border-t border-gray-200">
          <table className="min-w-full divide-y divide-gray-200">
            <thead className="bg-gray-50">
              <tr>
                <th className="px-4 py-2 text-left text-xs font-medium text-gray-500 uppercase tracking-wider w-1/3">
                  Model
                </th>
                <th className="px-4 py-2 text-left text-xs font-medium text-gray-500 uppercase tracking-wider w-1/3">
                  Entry
                </th>
                <th className="px-4 py-2 text-left text-xs font-medium text-gray-500 uppercase tracking-wider w-1/3">
                  Evaluation
                </th>
              </tr>
            </thead>
            <tbody className="bg-white divide-y divide-gray-200">
              {data.rows.map((row) => (
                <tr key={row.modelId} className="hover:bg-gray-50">
                  <td className="px-4 py-2">
                    <button
                      type="button"
                      onClick={() => navigate(`/lab/models/${row.modelId}`)}
                      className="text-left hover:text-blue-600"
                    >
                      <div className="text-sm font-medium text-gray-900">{row.modelName}</div>
                      <div className="text-xs text-gray-500">{row.modelKind}</div>
                    </button>
                  </td>
                  <td className="px-4 py-1">
                    {row.entry ? (
                      <PipelineStatusCell
                        status={row.entry.status}
                        label={row.entry.optimizer}
                        onClick={() => navigate(`/lab/models/${encodeURIComponent(row.modelName)}/calcutta/${encodeURIComponent(data.calcuttaId)}`)}
                      />
                    ) : (
                      <PipelineStatusCell status="missing" />
                    )}
                  </td>
                  <td className="px-4 py-1">
                    {row.evaluation ? (
                      <PipelineStatusCell
                        status={row.evaluation.status}
                        label={`${(row.evaluation.nSims / 1000).toFixed(0)}k sims`}
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
                  <td colSpan={3} className="px-4 py-4 text-sm text-gray-500 text-center">
                    No models registered yet.
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

export function PipelineByCalcutta() {
  const leaderboardQuery = useQuery<LeaderboardResponse | null>({
    queryKey: ['lab', 'models', 'leaderboard'],
    queryFn: () => labService.getLeaderboard(),
  });

  const entriesQuery = useQuery<ListEntriesResponse | null>({
    queryKey: ['lab', 'entries', 'all'],
    queryFn: () => labService.listEntries({ limit: 500 }),
  });

  const evaluationsQuery = useQuery<ListEvaluationsResponse | null>({
    queryKey: ['lab', 'evaluations', 'all'],
    queryFn: () => labService.listEvaluations({ limit: 500 }),
  });

  const calcuttasQuery = useQuery<Calcutta[]>({
    queryKey: ['calcuttas'],
    queryFn: () => calcuttaService.getAllCalcuttas(),
  });

  const isLoading =
    leaderboardQuery.isLoading ||
    entriesQuery.isLoading ||
    evaluationsQuery.isLoading ||
    calcuttasQuery.isLoading;

  const isError =
    leaderboardQuery.isError ||
    entriesQuery.isError ||
    evaluationsQuery.isError ||
    calcuttasQuery.isError;

  const pipelineData = useMemo(() => {
    if (
      !leaderboardQuery.data ||
      !entriesQuery.data ||
      !evaluationsQuery.data ||
      !calcuttasQuery.data
    ) {
      return [];
    }

    return buildCalcuttaPipelineData(
      leaderboardQuery.data,
      entriesQuery.data.items,
      evaluationsQuery.data.items,
      calcuttasQuery.data
    );
  }, [leaderboardQuery.data, entriesQuery.data, evaluationsQuery.data, calcuttasQuery.data]);

  if (isLoading) {
    return <LoadingState label="Loading pipeline data..." layout="inline" />;
  }

  if (isError) {
    return <Alert variant="error">Failed to load pipeline data. Please try again.</Alert>;
  }

  if (pipelineData.length === 0) {
    return (
      <Alert variant="info">
        No calcuttas found. Create calcuttas to see them in the pipeline view.
      </Alert>
    );
  }

  return (
    <div className="space-y-3">
      {pipelineData.map((calcutta, index) => (
        <CalcuttaPipelineCard
          key={calcutta.calcuttaId}
          data={calcutta}
          defaultExpanded={index === 0}
        />
      ))}
    </div>
  );
}
