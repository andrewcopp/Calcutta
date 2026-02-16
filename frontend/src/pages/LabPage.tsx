import React from 'react';
import { useQuery } from '@tanstack/react-query';

import { Alert } from '../components/ui/Alert';
import { LoadingState } from '../components/ui/LoadingState';
import { ModelLeaderboardCard } from '../components/Lab/ModelLeaderboardCard';
import { labService } from '../services/labService';
import type { LeaderboardResponse } from '../types/lab';
import { calcuttaService } from '../services/calcuttaService';
import { Calcutta } from '../types/calcutta';
import { queryKeys } from '../queryKeys';

export function LabPage() {
  const leaderboardQuery = useQuery<LeaderboardResponse | null>({
    queryKey: queryKeys.lab.leaderboard(),
    queryFn: () => labService.getLeaderboard(),
  });

  const calcuttasQuery = useQuery<Calcutta[]>({
    queryKey: queryKeys.calcuttas.all(),
    queryFn: () => calcuttaService.getAllCalcuttas(),
  });

  const items = [...(leaderboardQuery.data?.items ?? [])].sort((a, b) => {
    const aTop1 = a.avg_p_top1 ?? 0;
    const bTop1 = b.avg_p_top1 ?? 0;
    return bTop1 - aTop1;
  });
  const totalCalcuttas = calcuttasQuery.data?.length ?? 0;

  const isLoading = leaderboardQuery.isLoading || calcuttasQuery.isLoading;
  const isError = leaderboardQuery.isError || calcuttasQuery.isError;

  return (
    <div className="container mx-auto px-4 py-4">
      <div className="flex items-center justify-between mb-4 border-b border-gray-200 pb-3">
        <div>
          <h1 className="text-xl font-bold text-gray-900">Lab</h1>
          <p className="text-sm text-gray-500">
            {items.length} model{items.length !== 1 ? 's' : ''} registered
            {totalCalcuttas > 0 ? ` across ${totalCalcuttas} calcutta${totalCalcuttas !== 1 ? 's' : ''}` : ''}
          </p>
        </div>
      </div>

      {isLoading ? <LoadingState label="Loading leaderboard..." layout="inline" /> : null}

      {isError ? <Alert variant="error">Failed to load leaderboard.</Alert> : null}

      {!isLoading && !isError && items.length === 0 ? (
        <Alert variant="info">
          No models found. Register investment models via Python to see them here.
        </Alert>
      ) : null}

      {!isLoading && !isError && items.length > 0 ? (
        <div className="space-y-3">
          {items.map((entry, index) => (
            <ModelLeaderboardCard
              key={entry.investment_model_id}
              entry={entry}
              rank={index + 1}
              totalCalcuttas={totalCalcuttas}
            />
          ))}
        </div>
      ) : null}
    </div>
  );
}
