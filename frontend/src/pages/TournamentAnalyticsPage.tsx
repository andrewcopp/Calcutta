import React, { useState } from 'react';
import { useQuery } from '@tanstack/react-query';
import { Link } from 'react-router-dom';
import { Tournament } from '../types/calcutta';
import { tournamentService } from '../services/tournamentService';
import { queryKeys } from '../queryKeys';

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

  // Fetch all tournaments
  const { data: tournaments = [], isLoading: tournamentsLoading } = useQuery<Tournament[]>({
    queryKey: queryKeys.tournaments.all(),
    queryFn: tournamentService.getAllTournaments,
  });

  // Fetch simulation stats for selected tournament
  const { data: simulationStats, isLoading: statsLoading } = useQuery<SimulationStats>({
    queryKey: ['analytics', 'simulations', selectedTournamentId],
    queryFn: async () => {
      if (!selectedTournamentId) return null;
      const response = await fetch(`/api/analytics/tournaments/${selectedTournamentId}/simulations`);
      if (!response.ok) throw new Error('Failed to fetch simulation stats');
      return response.json();
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
        <h2 className="text-xl font-semibold mb-4">Select Tournament</h2>
        
        {tournamentsLoading ? (
          <div className="text-gray-500">Loading tournaments...</div>
        ) : tournaments.length === 0 ? (
          <div className="text-gray-500">No tournaments found</div>
        ) : (
          <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-4">
            {tournaments.map((tournament) => (
              <button
                key={tournament.id}
                onClick={() => setSelectedTournamentId(tournament.id)}
                className={`p-4 rounded-lg border-2 transition-colors text-left ${
                  selectedTournamentId === tournament.id
                    ? 'border-blue-500 bg-blue-50'
                    : 'border-gray-200 hover:border-blue-300'
                }`}
              >
                <div className="font-semibold">{tournament.name}</div>
                <div className="text-sm text-gray-600">{tournament.rounds} rounds</div>
              </button>
            ))}
          </div>
        )}
      </div>

      {/* Simulation Statistics */}
      {selectedTournamentId && (
        <div className="bg-white rounded-lg shadow p-6">
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

              {/* Actions */}
              <div className="flex gap-4 pt-4 border-t">
                <Link
                  to={`/admin/tournaments/${selectedTournamentId}`}
                  className="px-4 py-2 bg-blue-600 text-white rounded hover:bg-blue-700"
                >
                  View Tournament Details
                </Link>
                <button
                  onClick={() => {
                    // TODO: Add export functionality
                    alert('Export functionality coming soon');
                  }}
                  className="px-4 py-2 border border-gray-300 rounded hover:bg-gray-50"
                >
                  Export Data
                </button>
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
      )}

      {!selectedTournamentId && (
        <div className="bg-gray-50 rounded-lg p-8 text-center text-gray-500">
          Select a tournament above to view analytics
        </div>
      )}
    </div>
  );
}
