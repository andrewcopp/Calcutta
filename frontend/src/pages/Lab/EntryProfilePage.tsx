import React from 'react';
import { useQuery } from '@tanstack/react-query';
import { useParams } from 'react-router-dom';

import { Alert } from '../../components/ui/Alert';
import { Breadcrumb } from '../../components/ui/Breadcrumb';
import { Card } from '../../components/ui/Card';
import { LoadingState } from '../../components/ui/LoadingState';
import { labService } from '../../services/labService';
import type { EvaluationEntryProfile } from '../../types/lab';
import { cn } from '../../lib/cn';
import { queryKeys } from '../../queryKeys';
import { formatPayoutX, formatPct, getPayoutColor } from '../../utils/labFormatters';

export function EntryProfilePage() {
  const { entryResultId, modelName, calcuttaId } = useParams<{
    entryResultId: string;
    modelName?: string;
    calcuttaId?: string;
  }>();

  const profileQuery = useQuery<EvaluationEntryProfile | null>({
    queryKey: queryKeys.lab.entryResults.profile(entryResultId),
    queryFn: () =>
      entryResultId
        ? labService.getEvaluationEntryProfile(entryResultId)
        : Promise.resolve(null),
    enabled: Boolean(entryResultId),
  });

  const profile = profileQuery.data;

  if (profileQuery.isLoading) {
    return (
      <div className="container mx-auto px-4 py-4">
        <LoadingState label="Loading entry profile..." />
      </div>
    );
  }

  if (profileQuery.isError || !profile) {
    return (
      <div className="container mx-auto px-4 py-4">
        <Alert variant="error">Failed to load entry profile.</Alert>
      </div>
    );
  }

  const isOurStrategy = profile.entry_name === 'Our Strategy';

  return (
    <div className="container mx-auto px-4 py-4">
      <Breadcrumb
        items={[
          { label: 'Lab', href: '/lab' },
          ...(modelName && calcuttaId
            ? [
                { label: decodeURIComponent(modelName), href: `/lab/models/${modelName}/calcutta/${calcuttaId}?tab=evaluations` },
                { label: profile.entry_name },
              ]
            : [{ label: profile.entry_name }]),
        ]}
      />

      {/* Header with entry name and rank */}
      <div className="flex items-baseline gap-3 mb-4">
        <h1 className="text-xl font-bold text-gray-900">{profile.entry_name}</h1>
        <span className="text-gray-500">Rank #{profile.rank}</span>
        {isOurStrategy && (
          <span className="px-2 py-0.5 text-xs font-medium bg-blue-100 text-blue-800 rounded">Our Strategy</span>
        )}
      </div>

      {/* Compact metrics header */}
      <div className="bg-white rounded-lg border border-gray-200 p-4 mb-4">
        <div className="flex flex-wrap items-center gap-6">
          <div>
            <span className="text-xs text-gray-500 uppercase mr-2">Mean Payout:</span>
            <span className={cn('font-semibold', getPayoutColor(profile.mean_normalized_payout))}>
              {formatPayoutX(profile.mean_normalized_payout)}
            </span>
          </div>
          <div>
            <span className="text-xs text-gray-500 uppercase mr-2">P(Top 1):</span>
            <span className="font-medium">{formatPct(profile.p_top1)}</span>
          </div>
          <div>
            <span className="text-xs text-gray-500 uppercase mr-2">P(In Money):</span>
            <span className="font-medium">{formatPct(profile.p_in_money)}</span>
          </div>
          <div>
            <span className="text-xs text-gray-500 uppercase mr-2">Total Bid:</span>
            <span className="font-medium">{profile.total_bid_points.toLocaleString()} pts</span>
          </div>
        </div>
      </div>

      {/* Bids table */}
      <Card>
        <h2 className="text-lg font-semibold mb-3">Team Bids ({profile.bids.length})</h2>
        {profile.bids.length === 0 ? (
          <p className="text-gray-500 text-sm">No bids available for this entry.</p>
        ) : (
          <div className="overflow-x-auto">
            <table className="min-w-full divide-y divide-gray-200">
              <thead className="bg-gray-50">
                <tr>
                  <th className="px-3 py-2 text-left text-xs font-medium text-gray-500 uppercase">Team</th>
                  <th className="px-3 py-2 text-center text-xs font-medium text-gray-500 uppercase">Seed</th>
                  <th className="px-3 py-2 text-left text-xs font-medium text-gray-500 uppercase">Region</th>
                  <th className="px-3 py-2 text-right text-xs font-medium text-gray-500 uppercase">Bid</th>
                  <th className="px-3 py-2 text-right text-xs font-medium text-gray-500 uppercase">Ownership</th>
                </tr>
              </thead>
              <tbody className="bg-white divide-y divide-gray-200">
                {profile.bids.map((bid) => (
                  <tr key={bid.team_id}>
                    <td className="px-3 py-2 text-sm text-gray-900">{bid.school_name}</td>
                    <td className="px-3 py-2 text-sm text-gray-700 text-center">{bid.seed}</td>
                    <td className="px-3 py-2 text-sm text-gray-700">{bid.region}</td>
                    <td className="px-3 py-2 text-sm text-gray-700 text-right">{bid.bid_points.toLocaleString()}</td>
                    <td className="px-3 py-2 text-sm text-gray-700 text-right">{formatPct(bid.ownership)}</td>
                  </tr>
                ))}
              </tbody>
            </table>
          </div>
        )}
      </Card>
    </div>
  );
}
