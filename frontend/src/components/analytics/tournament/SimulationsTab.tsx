import { useQuery } from '@tanstack/react-query';
import { analyticsService } from '../../../services/analyticsService';
import { Alert } from '../../ui/Alert';
import { Button } from '../../ui/Button';
import { LoadingState } from '../../ui/LoadingState';

interface SimulationStats {
  tournament_id: string;
  tournament_name: string;
  season: number;
  total_simulations: number;
  total_predictions: number;
  mean_wins: number;
  median_wins: number;
  max_wins: number;
  last_updated: string;
}

// Simulations Tab Component
export function SimulationsTab({ tournamentId }: { tournamentId: string }) {
  const simulationStatsQuery = useQuery<SimulationStats | null>({
    queryKey: ['analytics', 'simulations', tournamentId],
    queryFn: async () => {
      if (!tournamentId) return null;
      return analyticsService.getTournamentSimulationStats<SimulationStats>(tournamentId);
    },
    enabled: !!tournamentId,
  });

  const simulationStats = simulationStatsQuery.data;

  return (
    <div>
      <h2 className="text-xl font-semibold mb-4">Simulation Statistics</h2>

      {simulationStatsQuery.isLoading ? (
        <LoadingState label="Loading statistics..." layout="inline" />
      ) : simulationStatsQuery.isError ? (
        <Alert variant="error" className="mt-3">
          <div className="font-semibold mb-1">Failed to load statistics</div>
          <div className="mb-3">{simulationStatsQuery.error instanceof Error ? simulationStatsQuery.error.message : 'An error occurred'}</div>
          <Button size="sm" onClick={() => simulationStatsQuery.refetch()}>
            Retry
          </Button>
        </Alert>
      ) : simulationStats ? (
        <div className="space-y-6">
          <div className="grid grid-cols-1 md:grid-cols-3 gap-4">
            <div className="bg-gray-50 p-4 rounded-lg">
              <div className="text-sm text-gray-600 mb-1">Total Simulations</div>
              <div className="text-2xl font-bold">{simulationStats.total_simulations.toLocaleString()}</div>
            </div>

            <div className="bg-gray-50 p-4 rounded-lg">
              <div className="text-sm text-gray-600 mb-1">Predictions Generated</div>
              <div className="text-2xl font-bold">{simulationStats.total_predictions.toLocaleString()}</div>
            </div>

            <div className="bg-gray-50 p-4 rounded-lg">
              <div className="text-sm text-gray-600 mb-1">Last Updated</div>
              <div className="text-lg font-semibold">{new Date(simulationStats.last_updated).toLocaleDateString()}</div>
            </div>
          </div>

          <div>
            <h3 className="text-lg font-semibold mb-3">Win Distribution</h3>
            <div className="grid grid-cols-1 md:grid-cols-3 gap-4">
              <div className="border border-gray-200 p-4 rounded-lg">
                <div className="text-sm text-gray-600 mb-1">Mean Wins</div>
                <div className="text-xl font-bold">{simulationStats.mean_wins.toFixed(2)}</div>
              </div>

              <div className="border border-gray-200 p-4 rounded-lg">
                <div className="text-sm text-gray-600 mb-1">Median Wins</div>
                <div className="text-xl font-bold">{simulationStats.median_wins.toFixed(2)}</div>
              </div>

              <div className="border border-gray-200 p-4 rounded-lg">
                <div className="text-sm text-gray-600 mb-1">Max Wins</div>
                <div className="text-xl font-bold">{simulationStats.max_wins}</div>
              </div>
            </div>
          </div>
        </div>
      ) : (
        <Alert variant="info">
          <div>No simulation data available for this tournament.</div>
          <div className="mt-2 text-sm">Run simulations using the data science pipeline to generate analytics.</div>
        </Alert>
      )}
    </div>
  );
}
