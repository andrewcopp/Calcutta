import React, { useState } from 'react';
import { useQuery } from '@tanstack/react-query';
import { Calcutta, Tournament } from '../types/calcutta';
import { tournamentService } from '../services/tournamentService';
import { calcuttaService } from '../services/calcuttaService';
import { queryKeys } from '../queryKeys';
import { SimulationsTab } from '../components/analytics/tournament/SimulationsTab';
import { PredictedReturnsTab } from '../components/analytics/tournament/PredictedReturnsTab';
import { PredictedInvestmentTab } from '../components/analytics/tournament/PredictedInvestmentTab';
import { SimulatedEntriesTab } from '../components/analytics/tournament/SimulatedEntriesTab';
import { SimulatedCalcuttasTab } from '../components/analytics/tournament/SimulatedCalcuttasTab';

type TabType = 'simulations' | 'predicted-returns' | 'predicted-investment' | 'simulated-entries' | 'simulated-calcuttas';

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
