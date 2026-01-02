import React, { useState } from 'react';
import { useQuery } from '@tanstack/react-query';
import { Calcutta, Tournament } from '../types/calcutta';
import { tournamentService } from '../services/tournamentService';
import { calcuttaService } from '../services/calcuttaService';
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
  const [selectedCalcuttaId, setSelectedCalcuttaId] = useState<string | null>(null);
  const [activeTab, setActiveTab] = useState<TabType>('simulations');

  // Fetch all tournaments
  const { data: tournaments = [], isLoading: tournamentsLoading } = useQuery<Tournament[]>({
    queryKey: queryKeys.tournaments.all(),
    queryFn: tournamentService.getAllTournaments,
  });

  // Fetch all calcuttas (used to select scoring context for points-based analytics)
  const { data: calcuttas = [], isLoading: calcuttasLoading } = useQuery<Calcutta[]>({
    queryKey: ['calcuttas', 'all'],
    queryFn: calcuttaService.getAllCalcuttas,
    enabled: !!selectedTournamentId,
  });

  const calcuttasForTournament = calcuttas.filter((c) => c.tournamentId === selectedTournamentId);

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
              onChange={(e) => {
                const nextTournamentId = e.target.value || null;
                setSelectedTournamentId(nextTournamentId);
                setSelectedCalcuttaId(null);
              }}
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

      {/* Calcutta Selector */}
      {selectedTournamentId && (
        <div className="bg-white rounded-lg shadow p-6 mb-6">
          <div className="flex items-center gap-4">
            <label htmlFor="calcutta-select" className="text-lg font-semibold whitespace-nowrap">
              Select Calcutta:
            </label>

            {calcuttasLoading ? (
              <div className="text-gray-500">Loading calcuttas...</div>
            ) : calcuttasForTournament.length === 0 ? (
              <div className="text-gray-500">No calcuttas found for this tournament</div>
            ) : (
              <select
                id="calcutta-select"
                value={selectedCalcuttaId || ''}
                onChange={(e) => setSelectedCalcuttaId(e.target.value || null)}
                className="flex-1 max-w-md px-4 py-2 border border-gray-300 rounded-lg focus:ring-2 focus:ring-blue-500 focus:border-blue-500 outline-none"
              >
                <option value="">-- Select a calcutta --</option>
                {calcuttasForTournament.map((calcutta) => (
                  <option key={calcutta.id} value={calcutta.id}>
                    {calcutta.name}
                  </option>
                ))}
              </select>
            )}
          </div>
          <div className="mt-2 text-sm text-gray-600">
            Points-based analytics (returns, investment, simulated entries/calcuttas) are calcutta-scoped.
          </div>
        </div>
      )}

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
            {activeTab === 'predicted-returns' && <PredictedReturnsTab calcuttaId={selectedCalcuttaId} />}
            {activeTab === 'predicted-investment' && <PredictedInvestmentTab calcuttaId={selectedCalcuttaId} />}
            {activeTab === 'simulated-entries' && <SimulatedEntriesTab calcuttaId={selectedCalcuttaId} />}
            {activeTab === 'simulated-calcuttas' && <SimulatedCalcuttasTab calcuttaId={selectedCalcuttaId} />}
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

function PredictedReturnsTab({ calcuttaId }: { calcuttaId: string | null }) {
  const { data: predictedReturns, isLoading } = useQuery<{ teams: TeamPredictedReturns[] } | null>({
    queryKey: ['analytics', 'predicted-returns', calcuttaId],
    queryFn: async () => {
      if (!calcuttaId) return null;
      return apiClient.get<{ teams: TeamPredictedReturns[] }>(`/analytics/calcuttas/${calcuttaId}/predicted-returns`);
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
              <tr className="bg-gray-100 font-bold border-t-2 border-gray-300">
                <td className="px-4 py-3 text-sm text-gray-900 sticky left-0 bg-gray-100">
                  TOTAL
                </td>
                <td className="px-4 py-3 text-sm text-center" colSpan={2}></td>
                <td className="px-4 py-3 text-sm text-center" colSpan={6}></td>
                <td className="px-4 py-3 text-sm text-center text-blue-700 bg-blue-100">
                  {formatPoints(predictedReturns.teams.reduce((sum, team) => sum + team.expected_value, 0))}
                </td>
              </tr>
            </tbody>
          </table>
        </div>
      ) : (
        <div className="text-gray-500">
          No predicted returns data available for this calcutta.
        </div>
      )}
    </div>
  );
}

// Predicted Investment Tab Component
interface TeamPredictedInvestment {
  team_id: string;
  school_name: string;
  seed: number;
  region: string;
  rational: number;
  predicted: number;
  delta: number;
}

function PredictedInvestmentTab({ calcuttaId }: { calcuttaId: string | null }) {
  const { data: predictedInvestment, isLoading } = useQuery<{ teams: TeamPredictedInvestment[] } | null>({
    queryKey: ['analytics', 'predicted-investment', calcuttaId],
    queryFn: async () => {
      if (!calcuttaId) return null;
      return apiClient.get<{ teams: TeamPredictedInvestment[] }>(`/analytics/calcuttas/${calcuttaId}/predicted-investment`);
    },
    enabled: !!calcuttaId,
  });

  const formatPoints = (points: number) => points.toFixed(1);
  const formatPercent = (percent: number) => {
    const formatted = percent.toFixed(1);
    return percent > 0 ? `+${formatted}%` : `${formatted}%`;
  };

  const getDeltaColor = (delta: number) => {
    if (delta < -5) return 'text-green-700 font-semibold'; // Undervalued - opportunity
    if (delta > 5) return 'text-red-700 font-semibold'; // Overvalued - avoid
    return 'text-gray-700';
  };

  if (!calcuttaId) {
    return <div className="text-gray-500">Select a calcutta above to view points-based predicted investment.</div>;
  }

  return (
    <div>
      <h2 className="text-xl font-semibold mb-4">Predicted Investment</h2>
      <p className="text-gray-600 mb-6">
        Market inefficiency analysis comparing rational investment (equal ROI) vs. predicted market behavior.
      </p>
      <div className="mb-4 p-4 bg-blue-50 border border-blue-200 rounded-lg">
        <p className="text-sm text-blue-900 mb-2"><strong>Column Definitions:</strong></p>
        <ul className="text-sm text-blue-800 space-y-1">
          <li><strong>Rational:</strong> Efficient market baseline - proportional investment for equal ROI across all teams</li>
          <li><strong>Predicted:</strong> ML model prediction of actual market behavior (ridge regression on historical data)</li>
          <li><strong>Delta:</strong> Market inefficiency as % difference - positive means overvalued, negative means undervalued</li>
        </ul>
      </div>

      {isLoading ? (
        <div className="text-gray-500">Loading predicted investment data...</div>
      ) : predictedInvestment?.teams ? (
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
                  Rational
                </th>
                <th className="px-4 py-3 text-center text-xs font-medium text-gray-500 uppercase tracking-wider bg-green-50">
                  Predicted
                </th>
                <th className="px-4 py-3 text-center text-xs font-medium text-gray-500 uppercase tracking-wider">
                  Delta
                </th>
              </tr>
            </thead>
            <tbody className="bg-white divide-y divide-gray-200">
              {predictedInvestment.teams.map((team) => (
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
                    {formatPoints(team.rational)}
                  </td>
                  <td className="px-4 py-3 text-sm text-center font-semibold text-green-700 bg-green-50">
                    {formatPoints(team.predicted)}
                  </td>
                  <td className={`px-4 py-3 text-sm text-center ${getDeltaColor(team.delta)}`}>
                    {formatPercent(team.delta)}
                  </td>
                </tr>
              ))}
              <tr className="bg-gray-100 font-bold border-t-2 border-gray-300">
                <td className="px-4 py-3 text-sm text-gray-900 sticky left-0 bg-gray-100">
                  TOTAL
                </td>
                <td className="px-4 py-3 text-sm text-center" colSpan={2}></td>
                <td className="px-4 py-3 text-sm text-center text-gray-900">
                  {formatPoints(predictedInvestment.teams.reduce((sum, team) => sum + team.rational, 0))}
                </td>
                <td className="px-4 py-3 text-sm text-center text-green-700 bg-green-100">
                  {formatPoints(predictedInvestment.teams.reduce((sum, team) => sum + team.predicted, 0))}
                </td>
                <td className="px-4 py-3 text-sm text-center text-gray-500">
                  -
                </td>
              </tr>
            </tbody>
          </table>
        </div>
      ) : (
        <div className="text-gray-500">
          No predicted investment data available for this calcutta.
        </div>
      )}
    </div>
  );
}

// Simulated Entries Tab Component
interface TeamSimulatedEntry {
  team_id: string;
  school_name: string;
  seed: number;
  region: string;
  expected_points: number;
  expected_market: number;
  expected_roi: number;
  our_bid: number;
  our_roi: number;
}

function SimulatedEntriesTab({ calcuttaId }: { calcuttaId: string | null }) {
  const [sortColumn, setSortColumn] = React.useState<keyof TeamSimulatedEntry>('seed');
  const [sortDirection, setSortDirection] = React.useState<'asc' | 'desc'>('asc');

  const { data: simulatedEntry, isLoading } = useQuery<{ teams: TeamSimulatedEntry[] } | null>({
    queryKey: ['analytics', 'simulated-entry', calcuttaId],
    queryFn: async () => {
      if (!calcuttaId) return null;
      return apiClient.get<{ teams: TeamSimulatedEntry[] }>(`/analytics/calcuttas/${calcuttaId}/simulated-entry`);
    },
    enabled: !!calcuttaId,
  });

  const formatPoints = (points: number) => points.toFixed(1);
  const formatROI = (roi: number) => roi.toFixed(2);

  if (!calcuttaId) {
    return <div className="text-gray-500">Select a calcutta above to view simulated entries.</div>;
  }

  const handleSort = (column: keyof TeamSimulatedEntry) => {
    if (sortColumn === column) {
      setSortDirection(sortDirection === 'asc' ? 'desc' : 'asc');
    } else {
      setSortColumn(column);
      setSortDirection('asc');
    }
  };

  const sortedTeams = React.useMemo(() => {
    if (!simulatedEntry?.teams) return [];
    
    return [...simulatedEntry.teams].sort((a, b) => {
      let aVal = a[sortColumn];
      let bVal = b[sortColumn];
      
      // Handle string comparison for school_name and region
      if (typeof aVal === 'string' && typeof bVal === 'string') {
        return sortDirection === 'asc' 
          ? aVal.localeCompare(bVal)
          : bVal.localeCompare(aVal);
      }
      
      // Handle numeric comparison
      if (typeof aVal === 'number' && typeof bVal === 'number') {
        return sortDirection === 'asc' ? aVal - bVal : bVal - aVal;
      }
      
      return 0;
    });
  }, [simulatedEntry?.teams, sortColumn, sortDirection]);

  const SortIcon = ({ column }: { column: keyof TeamSimulatedEntry }) => {
    if (sortColumn !== column) {
      return <span className="ml-1 text-gray-400">⇅</span>;
    }
    return sortDirection === 'asc' ? (
      <span className="ml-1">↑</span>
    ) : (
      <span className="ml-1">↓</span>
    );
  };

  return (
    <div>
      <h2 className="text-xl font-semibold mb-4">Simulated Entry</h2>
      <p className="text-gray-600 mb-6">
        Detailed investment report showing expected performance, market predictions, and ROI analysis for all teams.
      </p>

      {isLoading ? (
        <div className="text-gray-500">Loading simulated entry data...</div>
      ) : simulatedEntry?.teams ? (
        <div className="overflow-x-auto">
          <table className="min-w-full divide-y divide-gray-200">
            <thead className="bg-gray-50">
              <tr>
                <th 
                  className="px-4 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider sticky left-0 bg-gray-50 cursor-pointer hover:bg-gray-100"
                  onClick={() => handleSort('school_name')}
                >
                  Team <SortIcon column="school_name" />
                </th>
                <th 
                  className="px-4 py-3 text-center text-xs font-medium text-gray-500 uppercase tracking-wider cursor-pointer hover:bg-gray-100"
                  onClick={() => handleSort('seed')}
                >
                  Seed <SortIcon column="seed" />
                </th>
                <th 
                  className="px-4 py-3 text-center text-xs font-medium text-gray-500 uppercase tracking-wider cursor-pointer hover:bg-gray-100"
                  onClick={() => handleSort('region')}
                >
                  Region <SortIcon column="region" />
                </th>
                <th 
                  className="px-4 py-3 text-center text-xs font-medium text-gray-500 uppercase tracking-wider cursor-pointer hover:bg-gray-100"
                  onClick={() => handleSort('expected_points')}
                >
                  Exp Pts <SortIcon column="expected_points" />
                </th>
                <th 
                  className="px-4 py-3 text-center text-xs font-medium text-gray-500 uppercase tracking-wider cursor-pointer hover:bg-gray-100"
                  onClick={() => handleSort('expected_market')}
                >
                  Exp Mkt <SortIcon column="expected_market" />
                </th>
                <th 
                  className="px-4 py-3 text-center text-xs font-medium text-gray-500 uppercase tracking-wider cursor-pointer hover:bg-gray-100"
                  onClick={() => handleSort('expected_roi')}
                >
                  Exp ROI <SortIcon column="expected_roi" />
                </th>
                <th 
                  className="px-4 py-3 text-center text-xs font-medium text-gray-500 uppercase tracking-wider bg-blue-50 cursor-pointer hover:bg-blue-100"
                  onClick={() => handleSort('our_bid')}
                >
                  Our Bid <SortIcon column="our_bid" />
                </th>
                <th 
                  className="px-4 py-3 text-center text-xs font-medium text-gray-500 uppercase tracking-wider bg-blue-50 cursor-pointer hover:bg-blue-100"
                  onClick={() => handleSort('our_roi')}
                >
                  Our ROI <SortIcon column="our_roi" />
                </th>
              </tr>
            </thead>
            <tbody className="bg-white divide-y divide-gray-200">
              {sortedTeams.map((team) => (
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
                    {formatPoints(team.expected_points)}
                  </td>
                  <td className="px-4 py-3 text-sm text-center text-gray-700">
                    {formatPoints(team.expected_market)}
                  </td>
                  <td className="px-4 py-3 text-sm text-center text-gray-700">
                    {formatROI(team.expected_roi)}
                  </td>
                  <td className="px-4 py-3 text-sm text-center font-semibold text-blue-700 bg-blue-50">
                    {team.our_bid > 0 ? formatPoints(team.our_bid) : '-'}
                  </td>
                  <td className="px-4 py-3 text-sm text-center font-semibold text-blue-700 bg-blue-50">
                    {team.our_bid > 0 ? formatROI(team.our_roi) : '-'}
                  </td>
                </tr>
              ))}
            </tbody>
          </table>
        </div>
      ) : (
        <div className="text-gray-500">
          No simulated entry data available for this tournament.
        </div>
      )}
      
      {simulatedEntry?.teams && (
        <div className="mt-4 text-sm text-gray-600">
          <p className="mb-2">Coming soon:</p>
          <ul className="text-sm text-gray-600 list-disc list-inside space-y-1">
            <li>Portfolio optimization (Our Bid column will show recommended allocations)</li>
            <li>Actual market data integration</li>
            <li>ROI degradation analysis</li>
          </ul>
        </div>
      )}
    </div>
  );
}

// Simulated Calcuttas Tab Component
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

function SimulatedCalcuttasTab({ calcuttaId }: { calcuttaId: string | null }) {
  const { data: simulatedCalcuttas, isLoading } = useQuery<{ entries: EntryRanking[] } | null>({
    queryKey: ['analytics', 'simulated-calcuttas', calcuttaId],
    queryFn: async () => {
      if (!calcuttaId) return null;
      return apiClient.get<{ entries: EntryRanking[] }>(`/analytics/calcuttas/${calcuttaId}/simulated-calcuttas`);
    },
    enabled: !!calcuttaId,
  });

  if (!calcuttaId) {
    return <div className="text-gray-500">Select a calcutta above to view simulated calcuttas.</div>;
  }

  const formatPayout = (value: number) => value.toFixed(3);
  const formatPercent = (value: number) => `${(value * 100).toFixed(1)}%`;

  return (
    <div>
      <h2 className="text-xl font-semibold mb-4">Simulated Calcuttas</h2>
      <p className="text-gray-600 mb-6">
        Entry rankings based on normalized payout across all simulations. Payouts are normalized by dividing by 1st place payout.
      </p>

      {isLoading ? (
        <div className="text-gray-500">Loading simulated calcutta data...</div>
      ) : simulatedCalcuttas?.entries ? (
        <div className="overflow-x-auto">
          <table className="min-w-full divide-y divide-gray-200">
            <thead className="bg-gray-50">
              <tr>
                <th className="px-4 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">
                  Rank
                </th>
                <th className="px-4 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">
                  Entry Name
                </th>
                <th className="px-4 py-3 text-center text-xs font-medium text-gray-500 uppercase tracking-wider">
                  Mean Payout
                </th>
                <th className="px-4 py-3 text-center text-xs font-medium text-gray-500 uppercase tracking-wider">
                  Median Payout
                </th>
                <th className="px-4 py-3 text-center text-xs font-medium text-gray-500 uppercase tracking-wider">
                  P(Top 1)
                </th>
                <th className="px-4 py-3 text-center text-xs font-medium text-gray-500 uppercase tracking-wider">
                  P(In Money)
                </th>
                <th className="px-4 py-3 text-center text-xs font-medium text-gray-500 uppercase tracking-wider">
                  Simulations
                </th>
              </tr>
            </thead>
            <tbody className="bg-white divide-y divide-gray-200">
              {simulatedCalcuttas.entries.map((entry) => (
                <tr 
                  key={entry.entry_name} 
                  className={entry.is_our_strategy ? 'bg-green-50 hover:bg-green-100' : 'hover:bg-gray-50'}
                >
                  <td className="px-4 py-3 text-sm font-medium text-gray-900">
                    {entry.rank}
                  </td>
                  <td className="px-4 py-3 text-sm font-medium text-gray-900">
                    {entry.entry_name}
                    {entry.is_our_strategy && (
                      <span className="ml-2 px-2 py-1 text-xs font-semibold text-green-800 bg-green-200 rounded">
                        Our Strategy
                      </span>
                    )}
                  </td>
                  <td className="px-4 py-3 text-sm text-center text-gray-700">
                    {formatPayout(entry.mean_payout)}
                  </td>
                  <td className="px-4 py-3 text-sm text-center text-gray-700">
                    {formatPayout(entry.median_payout)}
                  </td>
                  <td className="px-4 py-3 text-sm text-center text-gray-700">
                    {formatPercent(entry.p_top1)}
                  </td>
                  <td className="px-4 py-3 text-sm text-center text-gray-700">
                    {formatPercent(entry.p_in_money)}
                  </td>
                  <td className="px-4 py-3 text-sm text-center text-gray-500">
                    {entry.total_simulations.toLocaleString()}
                  </td>
                </tr>
              ))}
            </tbody>
          </table>
        </div>
      ) : (
        <div className="text-gray-500">
          No simulated calcutta data available for this tournament.
        </div>
      )}
    </div>
  );
}

