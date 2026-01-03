import React, { useMemo, useState } from 'react';
import { useQuery } from '@tanstack/react-query';
import { Calcutta, Tournament } from '../types/calcutta';
import { tournamentService } from '../services/tournamentService';
import { calcuttaService } from '../services/calcuttaService';
import { apiClient } from '../api/apiClient';
import { queryKeys } from '../queryKeys';
import { TabsNav } from '../components/TabsNav';

type TabType = 'returns' | 'investments' | 'entries';

export function PredictionsPage() {
  const [selectedTournamentId, setSelectedTournamentId] = useState<string | null>(null);
  const [selectedCalcuttaId, setSelectedCalcuttaId] = useState<string | null>(null);
  const [activeTab, setActiveTab] = useState<TabType>('returns');

  const { data: tournaments = [], isLoading: tournamentsLoading } = useQuery<Tournament[]>({
    queryKey: queryKeys.tournaments.all(),
    queryFn: tournamentService.getAllTournaments,
  });

  const { data: calcuttas = [], isLoading: calcuttasLoading } = useQuery<Calcutta[]>({
    queryKey: ['calcuttas', 'all'],
    queryFn: calcuttaService.getAllCalcuttas,
    enabled: !!selectedTournamentId,
  });

  const calcuttasForTournament = useMemo(
    () => calcuttas.filter((c) => c.tournamentId === selectedTournamentId),
    [calcuttas, selectedTournamentId]
  );

  const tabs = useMemo(
    () =>
      [
        { id: 'returns' as const, label: 'Returns' },
        { id: 'investments' as const, label: 'Investments' },
        { id: 'entries' as const, label: 'Entries' },
      ] as const,
    []
  );

  return (
    <div className="container mx-auto px-4 py-8">
      <div className="mb-8">
        <h1 className="text-3xl font-bold mb-2">Predictions</h1>
        <p className="text-gray-600">Model outputs for returns, investments, and recommended entries.</p>
      </div>

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
          <div className="mt-2 text-sm text-gray-600">Predictions are calcutta-scoped.</div>
        </div>
      )}

      {selectedTournamentId ? (
        <div className="bg-white rounded-lg shadow p-6">
          <TabsNav tabs={tabs} activeTab={activeTab} onTabChange={setActiveTab} />

          {activeTab === 'returns' && <PredictedReturnsTab calcuttaId={selectedCalcuttaId} />}
          {activeTab === 'investments' && <PredictedInvestmentTab calcuttaId={selectedCalcuttaId} />}
          {activeTab === 'entries' && <SimulatedCalcuttasTab calcuttaId={selectedCalcuttaId} />}
        </div>
      ) : (
        <div className="bg-gray-50 rounded-lg p-8 text-center text-gray-500">Select a tournament above to view predictions</div>
      )}
    </div>
  );
}

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
    queryKey: ['predictions', 'returns', calcuttaId],
    queryFn: async () => {
      if (!calcuttaId) return null;
      return apiClient.get<{ teams: TeamPredictedReturns[] }>(`/analytics/calcuttas/${calcuttaId}/predicted-returns`);
    },
    enabled: !!calcuttaId,
  });

  const formatPercent = (prob: number) => `${(prob * 100).toFixed(1)}%`;
  const formatPoints = (points: number) => points.toFixed(1);

  if (!calcuttaId) {
    return <div className="text-gray-500">Select a calcutta above to view predicted returns.</div>;
  }

  return (
    <div>
      <h2 className="text-xl font-semibold mb-4">Returns</h2>
      <p className="text-gray-600 mb-6">Probability of reaching each round and expected value for all teams.</p>

      {isLoading ? (
        <div className="text-gray-500">Loading predicted returns...</div>
      ) : predictedReturns?.teams ? (
        <div className="overflow-x-auto">
          <table className="min-w-full divide-y divide-gray-200">
            <thead className="bg-gray-50">
              <tr>
                <th className="px-4 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider sticky left-0 bg-gray-50">Team</th>
                <th className="px-4 py-3 text-center text-xs font-medium text-gray-500 uppercase tracking-wider">Seed</th>
                <th className="px-4 py-3 text-center text-xs font-medium text-gray-500 uppercase tracking-wider">Region</th>
                <th className="px-4 py-3 text-center text-xs font-medium text-gray-500 uppercase tracking-wider">R64</th>
                <th className="px-4 py-3 text-center text-xs font-medium text-gray-500 uppercase tracking-wider">R32</th>
                <th className="px-4 py-3 text-center text-xs font-medium text-gray-500 uppercase tracking-wider">S16</th>
                <th className="px-4 py-3 text-center text-xs font-medium text-gray-500 uppercase tracking-wider">E8</th>
                <th className="px-4 py-3 text-center text-xs font-medium text-gray-500 uppercase tracking-wider">FF</th>
                <th className="px-4 py-3 text-center text-xs font-medium text-gray-500 uppercase tracking-wider">Champ</th>
                <th className="px-4 py-3 text-center text-xs font-medium text-gray-500 uppercase tracking-wider bg-blue-50">EV (pts)</th>
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
                  <td className="px-4 py-3 text-sm text-center font-semibold text-blue-700 bg-blue-50">{formatPoints(team.expected_value)}</td>
                </tr>
              ))}
              <tr className="bg-gray-100 font-bold border-t-2 border-gray-300">
                <td className="px-4 py-3 text-sm text-gray-900 sticky left-0 bg-gray-100">TOTAL</td>
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
        <div className="text-gray-500">No predicted returns data available for this calcutta.</div>
      )}
    </div>
  );
}

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
    queryKey: ['predictions', 'investments', calcuttaId],
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
    if (delta < -5) return 'text-green-700 font-semibold';
    if (delta > 5) return 'text-red-700 font-semibold';
    return 'text-gray-700';
  };

  if (!calcuttaId) {
    return <div className="text-gray-500">Select a calcutta above to view predicted investment.</div>;
  }

  return (
    <div>
      <h2 className="text-xl font-semibold mb-4">Investments</h2>
      <p className="text-gray-600 mb-6">Market inefficiency analysis comparing rational vs. predicted market behavior.</p>

      {isLoading ? (
        <div className="text-gray-500">Loading predicted investment data...</div>
      ) : predictedInvestment?.teams ? (
        <div className="overflow-x-auto">
          <table className="min-w-full divide-y divide-gray-200">
            <thead className="bg-gray-50">
              <tr>
                <th className="px-4 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider sticky left-0 bg-gray-50">Team</th>
                <th className="px-4 py-3 text-center text-xs font-medium text-gray-500 uppercase tracking-wider">Seed</th>
                <th className="px-4 py-3 text-center text-xs font-medium text-gray-500 uppercase tracking-wider">Region</th>
                <th className="px-4 py-3 text-center text-xs font-medium text-gray-500 uppercase tracking-wider">Rational</th>
                <th className="px-4 py-3 text-center text-xs font-medium text-gray-500 uppercase tracking-wider bg-green-50">Predicted</th>
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
                  <td className="px-4 py-3 text-sm text-center font-semibold text-green-700 bg-green-50">{formatPoints(team.predicted)}</td>
                  <td className={`px-4 py-3 text-sm text-center ${getDeltaColor(team.delta)}`}>{formatPercent(team.delta)}</td>
                </tr>
              ))}
              <tr className="bg-gray-100 font-bold border-t-2 border-gray-300">
                <td className="px-4 py-3 text-sm text-gray-900 sticky left-0 bg-gray-100">TOTAL</td>
                <td className="px-4 py-3 text-sm text-center" colSpan={2}></td>
                <td className="px-4 py-3 text-sm text-center text-gray-900">
                  {formatPoints(predictedInvestment.teams.reduce((sum, team) => sum + team.rational, 0))}
                </td>
                <td className="px-4 py-3 text-sm text-center text-green-700 bg-green-100">
                  {formatPoints(predictedInvestment.teams.reduce((sum, team) => sum + team.predicted, 0))}
                </td>
                <td className="px-4 py-3 text-sm text-center text-gray-500">-</td>
              </tr>
            </tbody>
          </table>
        </div>
      ) : (
        <div className="text-gray-500">No predicted investment data available for this calcutta.</div>
      )}
    </div>
  );
}

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
    queryKey: ['predictions', 'simulated-calcuttas', calcuttaId],
    queryFn: async () => {
      if (!calcuttaId) return null;
      return apiClient.get<{ entries: EntryRanking[] }>(`/analytics/calcuttas/${calcuttaId}/simulated-calcuttas`);
    },
    enabled: !!calcuttaId,
  });

  if (!calcuttaId) {
    return <div className="text-gray-500">Select a calcutta above to view entries.</div>;
  }

  const formatPayout = (value: number) => value.toFixed(3);
  const formatPercent = (value: number) => `${(value * 100).toFixed(1)}%`;

  return (
    <div>
      <h2 className="text-xl font-semibold mb-4">Entries</h2>
      <p className="text-gray-600 mb-6">Entry rankings based on normalized payout across all simulations.</p>

      {isLoading ? (
        <div className="text-gray-500">Loading entries...</div>
      ) : simulatedCalcuttas?.entries ? (
        <div className="overflow-x-auto">
          <table className="min-w-full divide-y divide-gray-200">
            <thead className="bg-gray-50">
              <tr>
                <th className="px-4 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">Rank</th>
                <th className="px-4 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">Entry Name</th>
                <th className="px-4 py-3 text-center text-xs font-medium text-gray-500 uppercase tracking-wider">Mean Payout</th>
                <th className="px-4 py-3 text-center text-xs font-medium text-gray-500 uppercase tracking-wider">Median Payout</th>
                <th className="px-4 py-3 text-center text-xs font-medium text-gray-500 uppercase tracking-wider">P(Top 1)</th>
                <th className="px-4 py-3 text-center text-xs font-medium text-gray-500 uppercase tracking-wider">P(Payout)</th>
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
                      <span className="ml-2 px-2 py-1 text-xs font-semibold text-green-800 bg-green-200 rounded">Our Strategy</span>
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
        <div className="text-gray-500">No entries data available for this calcutta.</div>
      )}
    </div>
  );
}
