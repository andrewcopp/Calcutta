import { useQuery } from '@tanstack/react-query';
import { analyticsService } from '../../../services/analyticsService';

interface TeamPredictedReturns {
  team_id: string;
  school_name: string;
  seed: number;
  region: string;
  prob_pi: number;
  prob_r64: number;
  prob_r32: number;
  prob_s16: number;
  prob_e8: number;
  prob_ff: number;
  prob_champ: number;
  expected_value: number;
}

// Predicted Returns Tab Component
export function PredictedReturnsTab({ calcuttaId }: { calcuttaId: string | null }) {
  const { data: predictedReturns, isLoading } = useQuery<{ teams: TeamPredictedReturns[] } | null>({
    queryKey: ['analytics', 'predicted-returns', calcuttaId],
    queryFn: async () => {
      if (!calcuttaId) return null;
      return analyticsService.getCalcuttaPredictedReturns<{ teams: TeamPredictedReturns[] }>(calcuttaId);
    },
    enabled: !!calcuttaId,
  });

  const formatPercent = (prob: number) => `${(prob * 100).toFixed(1)}%`;
  const formatPoints = (points: number) => points.toFixed(1);

  if (!calcuttaId) {
    return <div className="text-gray-500">Select a calcutta above to view points-based predicted returns.</div>;
  }

  return (
    <div>
      <h2 className="text-xl font-semibold mb-4">Predicted Returns</h2>
      <p className="text-gray-600 mb-6">
        Probability of reaching each round and expected value for all teams based on {predictedReturns?.teams.length ? '5,000' : ''}{' '}
        simulations.
      </p>

      {isLoading ? (
        <div className="text-gray-500">Loading predicted returns...</div>
      ) : predictedReturns?.teams ? (
        <div className="overflow-x-auto">
          <table className="min-w-full divide-y divide-gray-200">
            <thead className="bg-gray-50">
              <tr>
                <th className="px-4 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider sticky left-0 bg-gray-50">
                  Team
                </th>
                <th className="px-4 py-3 text-center text-xs font-medium text-gray-500 uppercase tracking-wider">Seed</th>
                <th className="px-4 py-3 text-center text-xs font-medium text-gray-500 uppercase tracking-wider">Region</th>
                <th className="px-4 py-3 text-center text-xs font-medium text-gray-500 uppercase tracking-wider">R64</th>
                <th className="px-4 py-3 text-center text-xs font-medium text-gray-500 uppercase tracking-wider">R32</th>
                <th className="px-4 py-3 text-center text-xs font-medium text-gray-500 uppercase tracking-wider">S16</th>
                <th className="px-4 py-3 text-center text-xs font-medium text-gray-500 uppercase tracking-wider">E8</th>
                <th className="px-4 py-3 text-center text-xs font-medium text-gray-500 uppercase tracking-wider">FF</th>
                <th className="px-4 py-3 text-center text-xs font-medium text-gray-500 uppercase tracking-wider">Champ</th>
                <th className="px-4 py-3 text-center text-xs font-medium text-gray-500 uppercase tracking-wider bg-blue-50">
                  EV (pts)
                </th>
              </tr>
            </thead>
            <tbody className="bg-white divide-y divide-gray-200">
              {predictedReturns.teams.map((team) => (
                <tr key={team.team_id} className="hover:bg-gray-50">
                  <td className="px-4 py-3 text-sm font-medium text-gray-900 sticky left-0 bg-white">{team.school_name}</td>
                  <td className="px-4 py-3 text-sm text-center text-gray-700">{team.seed}</td>
                  <td className="px-4 py-3 text-sm text-center text-gray-600">{team.region}</td>
                  <td className="px-4 py-3 text-sm text-center text-gray-700">{formatPercent(team.prob_r64)}</td>
                  <td className="px-4 py-3 text-sm text-center text-gray-700">{formatPercent(team.prob_r32)}</td>
                  <td className="px-4 py-3 text-sm text-center text-gray-700">{formatPercent(team.prob_s16)}</td>
                  <td className="px-4 py-3 text-sm text-center text-gray-700">{formatPercent(team.prob_e8)}</td>
                  <td className="px-4 py-3 text-sm text-center text-gray-700">{formatPercent(team.prob_ff)}</td>
                  <td className="px-4 py-3 text-sm text-center text-gray-700">{formatPercent(team.prob_champ)}</td>
                  <td className="px-4 py-3 text-sm text-center font-semibold text-blue-700 bg-blue-50">
                    {formatPoints(team.expected_value)}
                  </td>
                </tr>
              ))}
              <tr className="bg-gray-100 font-bold border-t-2 border-gray-300">
                <td className="px-4 py-3 text-sm text-gray-900 sticky left-0 bg-gray-100">TOTAL</td>
                <td className="px-4 py-3 text-sm text-center" colSpan={2}></td>
                <td className="px-4 py-3 text-sm text-center" colSpan={6}></td>
                <td className="px-4 py-3 text-sm text-center text-blue-700 bg-blue-100">
                  {formatPoints(predictedReturns.teams.reduce((sum, t) => sum + t.expected_value, 0))}
                </td>
              </tr>
            </tbody>
          </table>
        </div>
      ) : (
        <div className="text-gray-500">No predicted returns data available for this calcutta.</div>
      )}
    </div>
  );
}
