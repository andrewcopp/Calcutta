import { useQuery } from '@tanstack/react-query';
import { analyticsService } from '../../../services/analyticsService';
import { Alert } from '../../ui/Alert';
import { Button } from '../../ui/Button';
import { LoadingState } from '../../ui/LoadingState';

interface TeamPredictedInvestment {
  team_id: string;
  school_name: string;
  seed: number;
  region: string;
  rational: number;
  predicted: number;
  delta: number;
}

// Predicted Investment Tab Component
export function PredictedInvestmentTab({ calcuttaId }: { calcuttaId: string | null }) {
  const predictedInvestmentQuery = useQuery<{ teams: TeamPredictedInvestment[] } | null>({
    queryKey: ['analytics', 'predicted-investment', calcuttaId],
    queryFn: async () => {
      if (!calcuttaId) return null;
      return analyticsService.getCalcuttaPredictedInvestment<{ teams: TeamPredictedInvestment[] }>(calcuttaId);
    },
    enabled: !!calcuttaId,
  });

  const predictedInvestment = predictedInvestmentQuery.data;

  const formatPoints = (points: number) => points.toFixed(1);
  const formatPercent = (percent: number) => {
    const formatted = percent.toFixed(1);
    return percent > 0 ? `+${formatted}%` : `${formatted}%`;
  };

  const getDeltaColor = (delta: number) => {
    if (delta < -5) return 'text-green-700 font-semibold';
    if (delta > 5) return 'text-red-700 font-semibold';
    return 'text-gray-700';
  };

  if (!calcuttaId) {
    return <Alert variant="info">Select a calcutta above to view points-based predicted investment.</Alert>;
  }

  return (
    <div>
      <h2 className="text-xl font-semibold mb-4">Predicted Investment</h2>
      <p className="text-gray-600 mb-6">
        Market inefficiency analysis comparing rational investment (equal ROI) vs. predicted market behavior.
      </p>
      <div className="mb-4 p-4 bg-blue-50 border border-blue-200 rounded-lg">
        <p className="text-sm text-blue-900 mb-2">
          <strong>Column Definitions:</strong>
        </p>
        <ul className="text-sm text-blue-800 space-y-1">
          <li>
            <strong>Rational:</strong> Efficient market baseline - proportional investment for equal ROI across all teams
          </li>
          <li>
            <strong>Predicted:</strong> ML model prediction of actual market behavior (ridge regression on historical data)
          </li>
          <li>
            <strong>Delta:</strong> Market inefficiency as % difference - positive means overvalued, negative means undervalued
          </li>
        </ul>
      </div>

      {predictedInvestmentQuery.isLoading ? (
        <LoadingState label="Loading predicted investment data..." layout="inline" />
      ) : predictedInvestmentQuery.isError ? (
        <Alert variant="error" className="mt-3">
          <div className="font-semibold mb-1">Failed to load predicted investment</div>
          <div className="mb-3">{predictedInvestmentQuery.error instanceof Error ? predictedInvestmentQuery.error.message : 'An error occurred'}</div>
          <Button size="sm" onClick={() => predictedInvestmentQuery.refetch()}>
            Retry
          </Button>
        </Alert>
      ) : predictedInvestment?.teams && predictedInvestment.teams.length > 0 ? (
        <div className="overflow-x-auto">
          <table className="min-w-full divide-y divide-gray-200">
            <thead className="bg-gray-50">
              <tr>
                <th className="px-4 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider sticky left-0 bg-gray-50">
                  Team
                </th>
                <th className="px-4 py-3 text-center text-xs font-medium text-gray-500 uppercase tracking-wider">Seed</th>
                <th className="px-4 py-3 text-center text-xs font-medium text-gray-500 uppercase tracking-wider">Region</th>
                <th className="px-4 py-3 text-center text-xs font-medium text-gray-500 uppercase tracking-wider">Rational</th>
                <th className="px-4 py-3 text-center text-xs font-medium text-gray-500 uppercase tracking-wider bg-green-50">
                  Predicted
                </th>
                <th className="px-4 py-3 text-center text-xs font-medium text-gray-500 uppercase tracking-wider">Delta</th>
              </tr>
            </thead>
            <tbody className="bg-white divide-y divide-gray-200">
              {predictedInvestment.teams.map((team) => (
                <tr key={team.team_id} className="hover:bg-gray-50">
                  <td className="px-4 py-3 text-sm font-medium text-gray-900 sticky left-0 bg-white">{team.school_name}</td>
                  <td className="px-4 py-3 text-sm text-center text-gray-700">{team.seed}</td>
                  <td className="px-4 py-3 text-sm text-center text-gray-600">{team.region}</td>
                  <td className="px-4 py-3 text-sm text-center text-gray-700">{formatPoints(team.rational)}</td>
                  <td className="px-4 py-3 text-sm text-center font-semibold text-green-700 bg-green-50">
                    {formatPoints(team.predicted)}
                  </td>
                  <td className={`px-4 py-3 text-sm text-center ${getDeltaColor(team.delta)}`}>{formatPercent(team.delta)}</td>
                </tr>
              ))}
              <tr className="bg-gray-100 font-bold border-t-2 border-gray-300">
                <td className="px-4 py-3 text-sm text-gray-900 sticky left-0 bg-gray-100">TOTAL</td>
                <td className="px-4 py-3 text-sm text-center" colSpan={2}></td>
                <td className="px-4 py-3 text-sm text-center text-gray-900">
                  {formatPoints(predictedInvestment.teams.reduce((sum, t) => sum + t.rational, 0))}
                </td>
                <td className="px-4 py-3 text-sm text-center text-green-700 bg-green-100">
                  {formatPoints(predictedInvestment.teams.reduce((sum, t) => sum + t.predicted, 0))}
                </td>
                <td className="px-4 py-3 text-sm text-center text-gray-500">-</td>
              </tr>
            </tbody>
          </table>
        </div>
      ) : (
        <Alert variant="info">No predicted investment data available for this calcutta.</Alert>
      )}
    </div>
  );
}
