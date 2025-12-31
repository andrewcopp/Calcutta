import React, { useState } from 'react';
import { useQuery } from '@tanstack/react-query';
import { Tournament } from '../types/calcutta';
import { tournamentService } from '../services/tournamentService';
import { apiClient } from '../api/apiClient';
import { queryKeys } from '../queryKeys';

type TabType = 'simulations' | 'predicted-returns' | 'predicted-investment' | 'simulated-entries' | 'simulated-calcuttas';

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

export function TournamentAnalyticsPage() {
  const [selectedTournamentId, setSelectedTournamentId] = useState<string | null>(null);
  const [activeTab, setActiveTab] = useState<TabType>('simulations');

  // Fetch all tournaments
  const { data: tournaments = [], isLoading: tournamentsLoading } = useQuery<Tournament[]>({
    queryKey: queryKeys.tournaments.all(),
    queryFn: tournamentService.getAllTournaments,
  });

  // Fetch simulation stats for selected tournament
  const { data: simulationStats, isLoading: statsLoading } = useQuery<SimulationStats | null>({
    queryKey: ['analytics', 'simulations', selectedTournamentId],
    queryFn: async () => {
      if (!selectedTournamentId) return null;
      return apiClient.get<SimulationStats>(`/analytics/tournaments/${selectedTournamentId}/simulations`);
    },
    enabled: !!selectedTournamentId,
  });

  return (
    <div className="container mx-auto px-4 py-8">
      <div className="mb-8">
        <h1 className="text-3xl font-bold mb-2">Tournament Analytics</h1>
        <p className="text-gray-600">
          View simulation data and predictions for tournaments
        </p>
      </div>

      {/* Tournament Selector */}
      <div className="bg-white rounded-lg shadow p-6 mb-6">
        <div className="flex items-center gap-4">
          <label htmlFor="tournament-select" className="text-lg font-semibold whitespace-nowrap">
            Select Tournament:
          </label>
          
          {tournamentsLoading ? (
            <div className="text-gray-500">Loading tournaments...</div>
          ) : tournaments.length === 0 ? (
            <div className="text-gray-500">No tournaments found</div>
          ) : (
            <select
              id="tournament-select"
              value={selectedTournamentId || ''}
              onChange={(e) => setSelectedTournamentId(e.target.value || null)}
              className="flex-1 max-w-md px-4 py-2 border border-gray-300 rounded-lg focus:ring-2 focus:ring-blue-500 focus:border-blue-500 outline-none"
            >
              <option value="">-- Select a tournament --</option>
              {tournaments.map((tournament) => (
                <option key={tournament.id} value={tournament.id}>
                  {tournament.name}
                </option>
              ))}
            </select>
          )}
        </div>
      </div>

      {/* Analytics Tabs */}
      {selectedTournamentId && (
        <div className="bg-white rounded-lg shadow">
          {/* Tab Navigation */}
          <div className="border-b border-gray-200">
            <nav className="flex -mb-px">
              <button
                onClick={() => setActiveTab('simulations')}
                className={`px-6 py-4 text-sm font-medium border-b-2 transition-colors ${
                  activeTab === 'simulations'
                    ? 'border-blue-500 text-blue-600'
                    : 'border-transparent text-gray-500 hover:text-gray-700 hover:border-gray-300'
                }`}
              >
                Simulations
              </button>
              <button
                onClick={() => setActiveTab('predicted-returns')}
                className={`px-6 py-4 text-sm font-medium border-b-2 transition-colors ${
                  activeTab === 'predicted-returns'
                    ? 'border-blue-500 text-blue-600'
                    : 'border-transparent text-gray-500 hover:text-gray-700 hover:border-gray-300'
                }`}
              >
                Predicted Returns
              </button>
              <button
                onClick={() => setActiveTab('predicted-investment')}
                className={`px-6 py-4 text-sm font-medium border-b-2 transition-colors ${
                  activeTab === 'predicted-investment'
                    ? 'border-blue-500 text-blue-600'
                    : 'border-transparent text-gray-500 hover:text-gray-700 hover:border-gray-300'
                }`}
              >
                Predicted Investment
              </button>
              <button
                onClick={() => setActiveTab('simulated-entries')}
                className={`px-6 py-4 text-sm font-medium border-b-2 transition-colors ${
                  activeTab === 'simulated-entries'
                    ? 'border-blue-500 text-blue-600'
                    : 'border-transparent text-gray-500 hover:text-gray-700 hover:border-gray-300'
                }`}
              >
                Simulated Entries
              </button>
              <button
                onClick={() => setActiveTab('simulated-calcuttas')}
                className={`px-6 py-4 text-sm font-medium border-b-2 transition-colors ${
                  activeTab === 'simulated-calcuttas'
                    ? 'border-blue-500 text-blue-600'
                    : 'border-transparent text-gray-500 hover:text-gray-700 hover:border-gray-300'
                }`}
              >
                Simulated Calcuttas
              </button>
            </nav>
          </div>

          {/* Tab Content */}
          <div className="p-6">
            {activeTab === 'simulations' && <SimulationsTab tournamentId={selectedTournamentId} />}
            {activeTab === 'predicted-returns' && <PredictedReturnsTab tournamentId={selectedTournamentId} />}
            {activeTab === 'predicted-investment' && <PredictedInvestmentTab tournamentId={selectedTournamentId} />}
            {activeTab === 'simulated-entries' && <SimulatedEntriesTab tournamentId={selectedTournamentId} />}
            {activeTab === 'simulated-calcuttas' && <SimulatedCalcuttasTab tournamentId={selectedTournamentId} />}
          </div>
        </div>
      )}

      {!selectedTournamentId && (
        <div className="bg-gray-50 rounded-lg p-8 text-center text-gray-500">
          Select a tournament above to view analytics
        </div>
      )}
    </div>
  );
}

// Simulations Tab Component
function SimulationsTab({ tournamentId }: { tournamentId: string }) {
  const { data: simulationStats, isLoading: statsLoading } = useQuery<SimulationStats | null>({
    queryKey: ['analytics', 'simulations', tournamentId],
    queryFn: async () => {
      if (!tournamentId) return null;
      return apiClient.get<SimulationStats>(`/analytics/tournaments/${tournamentId}/simulations`);
    },
    enabled: !!tournamentId,
  });

  return (
    <div>
      <h2 className="text-xl font-semibold mb-4">Simulation Statistics</h2>
          
      {statsLoading ? (
        <div className="text-gray-500">Loading statistics...</div>
      ) : simulationStats ? (
        <div className="space-y-6">
          {/* High-level stats */}
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
              <div className="text-lg font-semibold">
                {new Date(simulationStats.last_updated).toLocaleDateString()}
              </div>
            </div>
          </div>

          {/* Win statistics */}
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
        <div className="text-gray-500">
          No simulation data available for this tournament.
          <div className="mt-2 text-sm">
            Run simulations using the data science pipeline to generate analytics.
          </div>
        </div>
      )}
    </div>
  );
}

// Predicted Returns Tab Component
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

function PredictedReturnsTab({ tournamentId }: { tournamentId: string }) {
  const { data: predictedReturns, isLoading } = useQuery<{ teams: TeamPredictedReturns[] } | null>({
    queryKey: ['analytics', 'predicted-returns', tournamentId],
    queryFn: async () => {
      if (!tournamentId) return null;
      return apiClient.get<{ teams: TeamPredictedReturns[] }>(`/analytics/tournaments/${tournamentId}/predicted-returns`);
    },
    enabled: !!tournamentId,
  });

  const formatPercent = (prob: number) => `${(prob * 100).toFixed(1)}%`;
  const formatPoints = (points: number) => points.toFixed(1);

  return (
    <div>
      <h2 className="text-xl font-semibold mb-4">Predicted Returns</h2>
      <p className="text-gray-600 mb-6">
        Probability of reaching each round and expected value for all teams based on {predictedReturns?.teams.length ? '5,000' : ''} simulations.
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
                <th className="px-4 py-3 text-center text-xs font-medium text-gray-500 uppercase tracking-wider">
                  Seed
                </th>
                <th className="px-4 py-3 text-center text-xs font-medium text-gray-500 uppercase tracking-wider">
                  Region
                </th>
                <th className="px-4 py-3 text-center text-xs font-medium text-gray-500 uppercase tracking-wider">
                  R64
                </th>
                <th className="px-4 py-3 text-center text-xs font-medium text-gray-500 uppercase tracking-wider">
                  R32
                </th>
                <th className="px-4 py-3 text-center text-xs font-medium text-gray-500 uppercase tracking-wider">
                  S16
                </th>
                <th className="px-4 py-3 text-center text-xs font-medium text-gray-500 uppercase tracking-wider">
                  E8
                </th>
                <th className="px-4 py-3 text-center text-xs font-medium text-gray-500 uppercase tracking-wider">
                  FF
                </th>
                <th className="px-4 py-3 text-center text-xs font-medium text-gray-500 uppercase tracking-wider">
                  Champ
                </th>
                <th className="px-4 py-3 text-center text-xs font-medium text-gray-500 uppercase tracking-wider bg-blue-50">
                  EV (pts)
                </th>
              </tr>
            </thead>
            <tbody className="bg-white divide-y divide-gray-200">
              {predictedReturns.teams.map((team) => (
                <tr key={team.team_id} className="hover:bg-gray-50">
                  <td className="px-4 py-3 text-sm font-medium text-gray-900 sticky left-0 bg-white">
                    {team.school_name}
                  </td>
                  <td className="px-4 py-3 text-sm text-center text-gray-700">
                    {team.seed}
                  </td>
                  <td className="px-4 py-3 text-sm text-center text-gray-600">
                    {team.region}
                  </td>
                  <td className="px-4 py-3 text-sm text-center text-gray-700">
                    {formatPercent(team.prob_r64)}
                  </td>
                  <td className="px-4 py-3 text-sm text-center text-gray-700">
                    {formatPercent(team.prob_r32)}
                  </td>
                  <td className="px-4 py-3 text-sm text-center text-gray-700">
                    {formatPercent(team.prob_s16)}
                  </td>
                  <td className="px-4 py-3 text-sm text-center text-gray-700">
                    {formatPercent(team.prob_e8)}
                  </td>
                  <td className="px-4 py-3 text-sm text-center text-gray-700">
                    {formatPercent(team.prob_ff)}
                  </td>
                  <td className="px-4 py-3 text-sm text-center text-gray-700">
                    {formatPercent(team.prob_champ)}
                  </td>
                  <td className="px-4 py-3 text-sm text-center font-semibold text-blue-700 bg-blue-50">
                    {formatPoints(team.expected_value)}
                  </td>
                </tr>
              ))}
            </tbody>
          </table>
        </div>
      ) : (
        <div className="text-gray-500">
          No predicted returns data available for this tournament.
        </div>
      )}
    </div>
  );
}

// Predicted Investment Tab Component
function PredictedInvestmentTab({ tournamentId }: { tournamentId: string }) {
  const { data: predictedReturns, isLoading } = useQuery<{ teams: TeamPredictedReturns[] } | null>({
    queryKey: ['analytics', 'predicted-returns', tournamentId],
    queryFn: async () => {
      if (!tournamentId) return null;
      return apiClient.get<{ teams: TeamPredictedReturns[] }>(`/analytics/tournaments/${tournamentId}/predicted-returns`);
    },
    enabled: !!tournamentId,
  });

  const formatPoints = (points: number) => points.toFixed(1);

  return (
    <div>
      <h2 className="text-xl font-semibold mb-4">Predicted Investment</h2>
      <p className="text-gray-600 mb-6">
        Expected value for each team to guide investment decisions.
      </p>

      {isLoading ? (
        <div className="text-gray-500">Loading predicted investment data...</div>
      ) : predictedReturns?.teams ? (
        <div className="overflow-x-auto">
          <table className="min-w-full divide-y divide-gray-200">
            <thead className="bg-gray-50">
              <tr>
                <th className="px-4 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider sticky left-0 bg-gray-50">
                  Team
                </th>
                <th className="px-4 py-3 text-center text-xs font-medium text-gray-500 uppercase tracking-wider">
                  Seed
                </th>
                <th className="px-4 py-3 text-center text-xs font-medium text-gray-500 uppercase tracking-wider">
                  Region
                </th>
                <th className="px-4 py-3 text-center text-xs font-medium text-gray-500 uppercase tracking-wider bg-green-50">
                  EV (invst)
                </th>
              </tr>
            </thead>
            <tbody className="bg-white divide-y divide-gray-200">
              {predictedReturns.teams.map((team) => (
                <tr key={team.team_id} className="hover:bg-gray-50">
                  <td className="px-4 py-3 text-sm font-medium text-gray-900 sticky left-0 bg-white">
                    {team.school_name}
                  </td>
                  <td className="px-4 py-3 text-sm text-center text-gray-700">
                    {team.seed}
                  </td>
                  <td className="px-4 py-3 text-sm text-center text-gray-600">
                    {team.region}
                  </td>
                  <td className="px-4 py-3 text-sm text-center font-semibold text-green-700 bg-green-50">
                    {formatPoints(team.expected_value)}
                  </td>
                </tr>
              ))}
            </tbody>
          </table>
        </div>
      ) : (
        <div className="text-gray-500">
          No predicted investment data available for this tournament.
        </div>
      )}
    </div>
  );
}

// Simulated Entries Tab Component
function SimulatedEntriesTab({ tournamentId }: { tournamentId: string }) {
  return (
    <div>
      <h2 className="text-xl font-semibold mb-4">Simulated Entries</h2>
      <div className="text-gray-500">
        <p className="mb-4">Detailed investment report showing simulated entry performance.</p>
        <div className="bg-blue-50 border border-blue-200 rounded-lg p-4">
          <p className="text-sm text-blue-800 mb-2">Coming soon:</p>
          <ul className="text-sm text-blue-800 list-disc list-inside space-y-1">
            <li>Entry-level performance metrics</li>
            <li>Team ownership percentages</li>
            <li>Expected vs actual ROI</li>
            <li>Win probability distributions</li>
          </ul>
        </div>
      </div>
    </div>
  );
}

// Simulated Calcuttas Tab Component
function SimulatedCalcuttasTab({ tournamentId }: { tournamentId: string }) {
  return (
    <div>
      <h2 className="text-xl font-semibold mb-4">Simulated Calcuttas</h2>
      <div className="text-gray-500">
        <p className="mb-4">Simulated auction results with normalized payout analysis.</p>
        <div className="bg-blue-50 border border-blue-200 rounded-lg p-4">
          <p className="text-sm text-blue-800 mb-2">Coming soon:</p>
          <ul className="text-sm text-blue-800 list-disc list-inside space-y-1">
            <li>List of simulated calcuttas</li>
            <li>Normalized payout calculations</li>
            <li>Entry rankings by performance</li>
            <li>Payout distribution analysis</li>
          </ul>
        </div>
      </div>
    </div>
  );
}

