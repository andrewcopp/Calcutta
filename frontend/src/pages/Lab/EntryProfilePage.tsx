import React from 'react';
import { useQuery } from '@tanstack/react-query';
import { useParams } from 'react-router-dom';

import { ErrorState } from '../../components/ui/ErrorState';
import { Breadcrumb } from '../../components/ui/Breadcrumb';
import { Card } from '../../components/ui/Card';
import { LoadingState } from '../../components/ui/LoadingState';
import { PageContainer } from '../../components/ui/Page';
import { labService } from '../../services/labService';
import type { EvaluationEntryProfile } from '../../schemas/lab';
import { cn } from '../../lib/cn';
import { queryKeys } from '../../queryKeys';
import { formatPayoutX, formatPct, getPayoutColor } from '../../utils/labFormatters';

export function EntryProfilePage() {
  const { entryResultId, modelId, calcuttaId } = useParams<{
    entryResultId: string;
    modelId?: string;
    calcuttaId?: string;
  }>();

  const profileQuery = useQuery<EvaluationEntryProfile | null>({
    queryKey: queryKeys.lab.entryResults.profile(entryResultId),
    queryFn: () => (entryResultId ? labService.getEvaluationEntryProfile(entryResultId) : Promise.resolve(null)),
    enabled: Boolean(entryResultId),
  });

  const profile = profileQuery.data;

  if (profileQuery.isLoading) {
    return (
      <PageContainer>
        <LoadingState label="Loading entry profile..." />
      </PageContainer>
    );
  }

  if (profileQuery.isError || !profile) {
    return (
      <PageContainer>
        <ErrorState
          error={profileQuery.error ?? 'Failed to load entry profile.'}
          onRetry={() => profileQuery.refetch()}
        />
      </PageContainer>
    );
  }

  const isOurStrategy = profile.entryName === 'Our Strategy';

  return (
    <PageContainer>
      <Breadcrumb
        items={[
          { label: 'Lab', href: '/lab' },
          ...(modelId && calcuttaId
            ? [
                {
                  label: 'Entry Detail',
                  href: `/lab/models/${modelId}/calcutta/${calcuttaId}?tab=evaluations`,
                },
                { label: profile.entryName },
              ]
            : [{ label: profile.entryName }]),
        ]}
      />

      {/* Header with entry name and rank */}
      <div className="flex items-baseline gap-3 mb-4">
        <h1 className="text-xl font-bold text-foreground">{profile.entryName}</h1>
        <span className="text-muted-foreground">Rank #{profile.rank}</span>
        {isOurStrategy && (
          <span className="px-2 py-0.5 text-xs font-medium bg-blue-100 text-primary rounded">Our Strategy</span>
        )}
      </div>

      {/* Compact metrics header */}
      <div className="bg-card rounded-lg border border-border p-4 mb-4">
        <div className="flex flex-wrap items-center gap-6">
          <div>
            <span className="text-xs text-muted-foreground uppercase mr-2">Mean Payout:</span>
            <span className={cn('font-semibold', getPayoutColor(profile.meanNormalizedPayout))}>
              {formatPayoutX(profile.meanNormalizedPayout)}
            </span>
          </div>
          <div>
            <span className="text-xs text-muted-foreground uppercase mr-2">P(Top 1):</span>
            <span className="font-medium">{formatPct(profile.pTop1)}</span>
          </div>
          <div>
            <span className="text-xs text-muted-foreground uppercase mr-2">P(In Money):</span>
            <span className="font-medium">{formatPct(profile.pInMoney)}</span>
          </div>
          <div>
            <span className="text-xs text-muted-foreground uppercase mr-2">Total Bid:</span>
            <span className="font-medium">{profile.totalBidPoints.toLocaleString()} credits</span>
          </div>
        </div>
      </div>

      {/* Bids table */}
      <Card>
        <h2 className="text-lg font-semibold mb-3">Team Bids ({profile.bids.length})</h2>
        {profile.bids.length === 0 ? (
          <p className="text-muted-foreground text-sm">No bids available for this entry.</p>
        ) : (
          <div className="overflow-x-auto">
            <table className="min-w-full divide-y divide-border">
              <thead className="bg-accent">
                <tr>
                  <th className="px-3 py-2 text-left text-xs font-medium text-muted-foreground uppercase">Team</th>
                  <th className="px-3 py-2 text-center text-xs font-medium text-muted-foreground uppercase">Seed</th>
                  <th className="px-3 py-2 text-left text-xs font-medium text-muted-foreground uppercase">Region</th>
                  <th className="px-3 py-2 text-right text-xs font-medium text-muted-foreground uppercase">Bid</th>
                  <th className="px-3 py-2 text-right text-xs font-medium text-muted-foreground uppercase">
                    Ownership
                  </th>
                </tr>
              </thead>
              <tbody className="bg-card divide-y divide-border">
                {profile.bids.map((bid) => (
                  <tr key={bid.teamId}>
                    <td className="px-3 py-2 text-sm text-foreground">{bid.schoolName}</td>
                    <td className="px-3 py-2 text-sm text-foreground text-center">{bid.seed}</td>
                    <td className="px-3 py-2 text-sm text-foreground">{bid.region}</td>
                    <td className="px-3 py-2 text-sm text-foreground text-right">{bid.bidPoints.toLocaleString()}</td>
                    <td className="px-3 py-2 text-sm text-foreground text-right">{formatPct(bid.ownership)}</td>
                  </tr>
                ))}
              </tbody>
            </table>
          </div>
        )}
      </Card>
    </PageContainer>
  );
}
