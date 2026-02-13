import React, { useMemo } from 'react';
import { useQuery } from '@tanstack/react-query';

import { Alert } from '../../components/ui/Alert';
import { LoadingState } from '../../components/ui/LoadingState';
import { ModelPipelineCard, ModelPipelineData, CalcuttaPipelineRow } from '../../components/Lab/ModelPipelineCard';
import { PipelineStatus } from '../../components/Lab/PipelineStatusCell';
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

type CalcuttaMap = Map<string, { id: string; name: string }>;

function buildPipelineData(
  leaderboard: LeaderboardResponse,
  entries: EntryDetail[],
  evaluations: EvaluationDetail[],
  calcuttas: CalcuttaMap
): ModelPipelineData[] {
  // Group entries by model
  const entriesByModel = new Map<string, EntryDetail[]>();
  for (const entry of entries) {
    const list = entriesByModel.get(entry.investment_model_id) || [];
    list.push(entry);
    entriesByModel.set(entry.investment_model_id, list);
  }

  // Group evaluations by entry
  const evalsByEntry = new Map<string, EvaluationDetail[]>();
  for (const ev of evaluations) {
    const list = evalsByEntry.get(ev.entry_id) || [];
    list.push(ev);
    evalsByEntry.set(ev.entry_id, list);
  }

  // Build pipeline data for each model
  const result: ModelPipelineData[] = [];

  leaderboard.items.forEach((model, index) => {
    const modelEntries = entriesByModel.get(model.investment_model_id) || [];

    // Create a row for each calcutta
    const rows: CalcuttaPipelineRow[] = [];
    const seenCalcuttas = new Set<string>();

    // First, add rows for calcuttas that have entries
    for (const entry of modelEntries) {
      seenCalcuttas.add(entry.calcutta_id);

      const entryEvals = evalsByEntry.get(entry.id) || [];
      // Get the best/most recent evaluation
      const bestEval = entryEvals.length > 0 ? entryEvals[0] : null;

      rows.push({
        calcuttaId: entry.calcutta_id,
        calcuttaName: entry.calcutta_name,
        entry: {
          id: entry.id,
          status: 'complete' as PipelineStatus,
          optimizer: entry.optimizer_kind,
          startingState: entry.starting_state_key,
        },
        evaluation: bestEval
          ? {
              id: bestEval.id,
              status: 'complete' as PipelineStatus,
              nSims: bestEval.n_sims,
              meanPayout: bestEval.mean_normalized_payout,
            }
          : null,
      });
    }

    // Then, add rows for calcuttas that don't have entries (missing)
    for (const [calcuttaId, calcutta] of calcuttas) {
      if (!seenCalcuttas.has(calcuttaId)) {
        rows.push({
          calcuttaId,
          calcuttaName: calcutta.name,
          entry: null,
          evaluation: null,
        });
      }
    }

    // Sort rows by calcutta name (most recent first based on year)
    rows.sort((a, b) => b.calcuttaName.localeCompare(a.calcuttaName));

    // Count completed evaluations
    const completedEvaluations = rows.filter((r) => r.evaluation?.status === 'complete').length;

    result.push({
      modelId: model.investment_model_id,
      modelName: model.model_name,
      modelKind: model.model_kind,
      avgPayout: model.avg_mean_payout ?? null,
      avgPTop1: model.avg_p_top1 ?? null,
      rank: index + 1,
      totalCalcuttas: calcuttas.size,
      completedEvaluations,
      rows,
    });
  });

  return result;
}

export function PipelineByModel() {
  // Fetch all required data
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

  // Build pipeline data from all queries
  const pipelineData = useMemo(() => {
    if (!leaderboardQuery.data || !entriesQuery.data || !evaluationsQuery.data || !calcuttasQuery.data) {
      return [];
    }

    const calcuttaMap: CalcuttaMap = new Map();
    for (const c of calcuttasQuery.data) {
      calcuttaMap.set(c.id, { id: c.id, name: c.name });
    }

    return buildPipelineData(
      leaderboardQuery.data,
      entriesQuery.data.items,
      evaluationsQuery.data.items,
      calcuttaMap
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
        No models found. Register investment models via Python to see them in the pipeline view.
      </Alert>
    );
  }

  return (
    <div className="space-y-3">
      {pipelineData.map((model, index) => (
        <ModelPipelineCard key={model.modelId} data={model} defaultExpanded={index === 0} />
      ))}
    </div>
  );
}
