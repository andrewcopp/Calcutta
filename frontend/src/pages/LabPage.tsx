import React from 'react';
import { useQuery } from '@tanstack/react-query';

import { Alert } from '../components/ui/Alert';
import { ErrorState } from '../components/ui/ErrorState';
import { LoadingState } from '../components/ui/LoadingState';
import { PageContainer, PageHeader } from '../components/ui/Page';
import { ModelLeaderboardCard } from '../components/Lab/ModelLeaderboardCard';
import { labService } from '../services/labService';
import type { LeaderboardResponse } from '../schemas/lab';
import { poolService } from '../services/poolService';
import type { Pool } from '../schemas/pool';
import { queryKeys } from '../queryKeys';

export function LabPage() {
  const leaderboardQuery = useQuery<LeaderboardResponse | null>({
    queryKey: queryKeys.lab.leaderboard(),
    queryFn: () => labService.getLeaderboard(),
  });

  const poolsQuery = useQuery<Pool[]>({
    queryKey: queryKeys.pools.all(),
    queryFn: () => poolService.getAllPools(),
  });

  const items = [...(leaderboardQuery.data?.items ?? [])].sort((a, b) => {
    const aTop1 = a.avgPTop1 ?? 0;
    const bTop1 = b.avgPTop1 ?? 0;
    return bTop1 - aTop1;
  });
  const totalPools = poolsQuery.data?.length ?? 0;

  const isLoading = leaderboardQuery.isLoading || poolsQuery.isLoading;
  const isError = leaderboardQuery.isError || poolsQuery.isError;

  if (isError) {
    const firstError = leaderboardQuery.error ?? poolsQuery.error;
    return (
      <PageContainer>
        <ErrorState
          error={firstError}
          onRetry={() => {
            if (leaderboardQuery.isError) leaderboardQuery.refetch();
            if (poolsQuery.isError) poolsQuery.refetch();
          }}
        />
      </PageContainer>
    );
  }

  return (
    <PageContainer>
      <PageHeader
        title="Lab"
        subtitle={`${items.length} model${items.length !== 1 ? 's' : ''} registered${totalPools > 0 ? ` across ${totalPools} pool${totalPools !== 1 ? 's' : ''}` : ''}`}
      />

      {isLoading ? <LoadingState label="Loading leaderboard..." layout="inline" /> : null}

      {!isLoading && items.length === 0 ? (
        <Alert variant="info">No models found. Register investment models via Python to see them here.</Alert>
      ) : null}

      {!isLoading && items.length > 0 ? (
        <div className="space-y-3">
          {items.map((entry, index) => (
            <ModelLeaderboardCard
              key={entry.investmentModelId}
              entry={entry}
              rank={index + 1}
              totalPools={totalPools}
            />
          ))}
        </div>
      ) : null}
    </PageContainer>
  );
}
