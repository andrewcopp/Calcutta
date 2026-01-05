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
import { Alert } from '../components/ui/Alert';
import { Button } from '../components/ui/Button';
import { Card } from '../components/ui/Card';
import { LoadingState } from '../components/ui/LoadingState';
import { PageContainer, PageHeader } from '../components/ui/Page';
import { Select } from '../components/ui/Select';

type TabType = 'simulations' | 'predicted-returns' | 'predicted-investment' | 'simulated-entries' | 'simulated-calcuttas';

export function TournamentAnalyticsPage() {
  const [selectedTournamentId, setSelectedTournamentId] = useState<string | null>(null);
  const [selectedCalcuttaId, setSelectedCalcuttaId] = useState<string | null>(null);
  const [activeTab, setActiveTab] = useState<TabType>('simulations');

  // Fetch all tournaments
  const tournamentsQuery = useQuery<Tournament[]>({
    queryKey: queryKeys.tournaments.all(),
    queryFn: tournamentService.getAllTournaments,
  });

  const tournaments = tournamentsQuery.data ?? [];

  // Fetch all calcuttas (used to select scoring context for points-based analytics)
  const calcuttasQuery = useQuery<Calcutta[]>({
    queryKey: ['calcuttas', 'all'],
    queryFn: calcuttaService.getAllCalcuttas,
    enabled: !!selectedTournamentId,
  });

  const calcuttas = calcuttasQuery.data ?? [];

  const calcuttasForTournament = calcuttas.filter((c) => c.tournamentId === selectedTournamentId);

  return (
    <PageContainer>
      <PageHeader title="Tournament Analytics" subtitle="View simulation data and predictions for tournaments" />

      {/* Tournament Selector */}
      <Card className="mb-6">
        <div className="flex items-center gap-4">
          <label htmlFor="tournament-select" className="text-lg font-semibold whitespace-nowrap">
            Select Tournament:
          </label>
          
          {tournamentsQuery.isLoading ? (
            <LoadingState label="Loading tournaments..." layout="inline" />
          ) : tournamentsQuery.isError ? (
            <Alert variant="error" className="flex-1">
              <div className="font-semibold mb-1">Failed to load tournaments</div>
              <div className="mb-3">{tournamentsQuery.error instanceof Error ? tournamentsQuery.error.message : 'An error occurred'}</div>
              <Button size="sm" onClick={() => tournamentsQuery.refetch()}>
                Retry
              </Button>
            </Alert>
          ) : tournaments.length === 0 ? (
            <Alert variant="info" className="flex-1">
              No tournaments found.
            </Alert>
          ) : (
            <Select
              id="tournament-select"
              value={selectedTournamentId || ''}
              onChange={(e) => {
                const nextTournamentId = e.target.value || null;
                setSelectedTournamentId(nextTournamentId);
                setSelectedCalcuttaId(null);
              }}
              className="flex-1 max-w-md"
            >
              <option value="">-- Select a tournament --</option>
              {tournaments.map((tournament) => (
                <option key={tournament.id} value={tournament.id}>
                  {tournament.name}
                </option>
              ))}
            </Select>
          )}
        </div>
      </Card>

      {/* Calcutta Selector */}
      {selectedTournamentId && (
        <Card className="mb-6">
          <div className="flex items-center gap-4">
            <label htmlFor="calcutta-select" className="text-lg font-semibold whitespace-nowrap">
              Select Calcutta:
            </label>

            {calcuttasQuery.isLoading ? (
              <LoadingState label="Loading calcuttas..." layout="inline" />
            ) : calcuttasQuery.isError ? (
              <Alert variant="error" className="flex-1">
                <div className="font-semibold mb-1">Failed to load calcuttas</div>
                <div className="mb-3">{calcuttasQuery.error instanceof Error ? calcuttasQuery.error.message : 'An error occurred'}</div>
                <Button size="sm" onClick={() => calcuttasQuery.refetch()}>
                  Retry
                </Button>
              </Alert>
            ) : calcuttasForTournament.length === 0 ? (
              <Alert variant="info" className="flex-1">
                No calcuttas found for this tournament.
              </Alert>
            ) : (
              <Select
                id="calcutta-select"
                value={selectedCalcuttaId || ''}
                onChange={(e) => setSelectedCalcuttaId(e.target.value || null)}
                className="flex-1 max-w-md"
              >
                <option value="">-- Select a calcutta --</option>
                {calcuttasForTournament.map((calcutta) => (
                  <option key={calcutta.id} value={calcutta.id}>
                    {calcutta.name}
                  </option>
                ))}
              </Select>
            )}
          </div>
          <div className="mt-2 text-sm text-gray-600">
            Points-based analytics (returns, investment, simulated entries/calcuttas) are calcutta-scoped.
          </div>
        </Card>
      )}

      {/* Analytics Tabs */}
      {selectedTournamentId && (
        <Card className="p-0 overflow-hidden">
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
        </Card>
      )}

      {!selectedTournamentId && (
        <Card className="bg-muted p-8 text-center text-muted-foreground">Select a tournament above to view analytics</Card>
      )}
    </PageContainer>
  );
}
