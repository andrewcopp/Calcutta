import { useQuery } from '@tanstack/react-query';
import { analyticsService } from '../../../services/analyticsService';
import { Alert } from '../../ui/Alert';
import { Button } from '../../ui/Button';
import { LoadingState } from '../../ui/LoadingState';

interface EntryRanking {
  rank: number;
  entry_name: string;
  is_our_strategy: boolean;
  mean_payout: number;
  median_payout: number;
  p_top1: number;
  p_in_money: number;
  total_simulations: number;
}

// Simulated Calcuttas Tab Component
export function SimulatedCalcuttasTab({ calcuttaId }: { calcuttaId: string | null }) {
  const simulatedCalcuttasQuery = useQuery<{ entries: EntryRanking[] } | null>({
    queryKey: ['analytics', 'simulated-calcuttas', calcuttaId],
    queryFn: async () => {
      if (!calcuttaId) return null;
      return analyticsService.getCalcuttaSimulatedCalcuttas<{ entries: EntryRanking[] }>(calcuttaId);
    },
    enabled: !!calcuttaId,
  });

  const simulatedCalcuttas = simulatedCalcuttasQuery.data;

  if (!calcuttaId) {
    return <Alert variant="info">Select a calcutta above to view simulated calcuttas.</Alert>;
  }

  const formatPayout = (value: number) => value.toFixed(3);
  const formatPercent = (value: number) => `${(value * 100).toFixed(1)}%`;

  return (
    <div>
      <h2 className="text-xl font-semibold mb-4">Simulated Calcuttas</h2>
      <p className="text-gray-600 mb-6">
        Entry rankings based on normalized payout across all simulations. Payouts are normalized by dividing by 1st place
        payout.
      </p>

      {simulatedCalcuttasQuery.isLoading ? (
        <LoadingState label="Loading simulated calcutta data..." layout="inline" />
      ) : simulatedCalcuttasQuery.isError ? (
        <Alert variant="error" className="mt-3">
          <div className="font-semibold mb-1">Failed to load simulated calcuttas</div>
          <div className="mb-3">{simulatedCalcuttasQuery.error instanceof Error ? simulatedCalcuttasQuery.error.message : 'An error occurred'}</div>
          <Button size="sm" onClick={() => simulatedCalcuttasQuery.refetch()}>
            Retry
          </Button>
        </Alert>
      ) : simulatedCalcuttas?.entries && simulatedCalcuttas.entries.length > 0 ? (
        <div className="overflow-x-auto">
          <table className="min-w-full divide-y divide-gray-200">
            <thead className="bg-gray-50">
              <tr>
                <th className="px-4 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">Rank</th>
                <th className="px-4 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">Entry Name</th>
                <th className="px-4 py-3 text-center text-xs font-medium text-gray-500 uppercase tracking-wider">Mean Payout</th>
                <th className="px-4 py-3 text-center text-xs font-medium text-gray-500 uppercase tracking-wider">Median Payout</th>
                <th className="px-4 py-3 text-center text-xs font-medium text-gray-500 uppercase tracking-wider">P(Top 1)</th>
                <th className="px-4 py-3 text-center text-xs font-medium text-gray-500 uppercase tracking-wider">P(In Money)</th>
                <th className="px-4 py-3 text-center text-xs font-medium text-gray-500 uppercase tracking-wider">Simulations</th>
              </tr>
            </thead>
            <tbody className="bg-white divide-y divide-gray-200">
              {simulatedCalcuttas.entries.map((entry) => (
                <tr
                  key={entry.entry_name}
                  className={entry.is_our_strategy ? 'bg-green-50 hover:bg-green-100' : 'hover:bg-gray-50'}
                >
                  <td className="px-4 py-3 text-sm font-medium text-gray-900">{entry.rank}</td>
                  <td className="px-4 py-3 text-sm font-medium text-gray-900">
                    {entry.entry_name}
                    {entry.is_our_strategy && (
                      <span className="ml-2 px-2 py-1 text-xs font-semibold text-green-800 bg-green-200 rounded">
                        Our Strategy
                      </span>
                    )}
                  </td>
                  <td className="px-4 py-3 text-sm text-center text-gray-700">{formatPayout(entry.mean_payout)}</td>
                  <td className="px-4 py-3 text-sm text-center text-gray-700">{formatPayout(entry.median_payout)}</td>
                  <td className="px-4 py-3 text-sm text-center text-gray-700">{formatPercent(entry.p_top1)}</td>
                  <td className="px-4 py-3 text-sm text-center text-gray-700">{formatPercent(entry.p_in_money)}</td>
                  <td className="px-4 py-3 text-sm text-center text-gray-500">{entry.total_simulations.toLocaleString()}</td>
                </tr>
              ))}
            </tbody>
          </table>
        </div>
      ) : (
        <Alert variant="info">No simulated calcutta data available for this tournament.</Alert>
      )}
    </div>
  );
}
