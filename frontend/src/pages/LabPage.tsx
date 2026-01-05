import React, { useMemo, useState } from 'react';
import { useQuery } from '@tanstack/react-query';
import { TabsNav } from '../components/TabsNav';
import { Card } from '../components/ui/Card';
import { PageContainer, PageHeader } from '../components/ui/Page';
import { Select } from '../components/ui/Select';
import { tournamentService } from '../services/tournamentService';
import { Calcutta, Tournament } from '../types/calcutta';
import { calcuttaService } from '../services/calcuttaService';
import { analyticsService } from '../services/analyticsService';

type TabType = 'advancements' | 'investments';

type Algorithm = {
  id: string;
  kind: string;
  name: string;
  description?: string | null;
  params_json?: unknown;
  created_at?: string;
};

type GameOutcomeRun = {
  id: string;
  algorithm_id: string;
  tournament_id: string;
  params_json?: unknown;
  git_sha?: string | null;
  created_at: string;
};

type MarketShareRun = {
  id: string;
  algorithm_id: string;
  calcutta_id: string;
  params_json?: unknown;
  git_sha?: string | null;
  created_at: string;
};

type TeamPredictedAdvancement = {
  team_id: string;
  school_name: string;
  seed: number;
  region: string;
  prob_pi: number;
  reach_r64: number;
  reach_r32: number;
  reach_s16: number;
  reach_e8: number;
  reach_ff: number;
  reach_champ: number;
  win_champ: number;
};

type TeamPredictedMarketShare = {
  team_id: string;
  school_name: string;
  seed: number;
  region: string;
  rational_share: number;
  predicted_share: number;
  delta_percent: number;
};

export function LabPage() {
  const [activeTab, setActiveTab] = useState<TabType>('advancements');
  const [selectedAlgorithmId, setSelectedAlgorithmId] = useState<string>('');
  const [selectedTournamentId, setSelectedTournamentId] = useState<string | null>(null);
  const [selectedCalcuttaId, setSelectedCalcuttaId] = useState<string | null>(null);
  const [selectedGameOutcomeRunId, setSelectedGameOutcomeRunId] = useState<string | null>(null);
  const [selectedInvestmentsGameOutcomeRunId, setSelectedInvestmentsGameOutcomeRunId] = useState<string | null>(null);
  const [selectedMarketShareRunId, setSelectedMarketShareRunId] = useState<string | null>(null);

  const tabs = useMemo(
    () =>
      [
        { id: 'advancements' as const, label: 'Advancements' },
        { id: 'investments' as const, label: 'Investments' },
      ] as const,
    []
  );

  const { data: tournaments = [], isLoading: tournamentsLoading } = useQuery<Tournament[]>({
    queryKey: ['tournaments', 'all'],
    queryFn: tournamentService.getAllTournaments,
  });

  const algorithmsQuery = useQuery<{ items: Algorithm[] } | null>({
    queryKey: ['analytics', 'algorithms', activeTab],
    queryFn: async () => {
      const kind = activeTab === 'advancements' ? 'game_outcomes' : 'market_share';
      const filtered = await analyticsService.listAlgorithms<{ items: Algorithm[] }>(kind);
      if (filtered?.items?.length) return filtered;
      return analyticsService.listAlgorithms<{ items: Algorithm[] }>();
    },
  });

  const algorithms = algorithmsQuery.data?.items ?? [];
  const selectedAlgorithm = useMemo(
    () => algorithms.find((a) => a.id === selectedAlgorithmId) ?? null,
    [algorithms, selectedAlgorithmId]
  );

  const gameOutcomeRunsQuery = useQuery<{ runs: GameOutcomeRun[] } | null>({
    queryKey: ['analytics', 'game-outcome-runs', selectedTournamentId],
    queryFn: async () => {
      if (!selectedTournamentId) return null;
      return analyticsService.listGameOutcomeRunsForTournament<{ runs: GameOutcomeRun[] }>(selectedTournamentId);
    },
    enabled: Boolean(selectedTournamentId),
  });

  const filteredGameOutcomeRuns = useMemo(() => {
    const runs = gameOutcomeRunsQuery.data?.runs ?? [];
    if (!selectedAlgorithmId) return runs;
    return runs.filter((r) => r.algorithm_id === selectedAlgorithmId);
  }, [gameOutcomeRunsQuery.data, selectedAlgorithmId]);

  const { data: calcuttas = [], isLoading: calcuttasLoading } = useQuery<Calcutta[]>({
    queryKey: ['calcuttas', 'all'],
    queryFn: calcuttaService.getAllCalcuttas,
    enabled: Boolean(selectedTournamentId),
  });

  const calcuttasForTournament = useMemo(
    () => calcuttas.filter((c) => c.tournamentId === selectedTournamentId),
    [calcuttas, selectedTournamentId]
  );

  const marketShareRunsQuery = useQuery<{ runs: MarketShareRun[] } | null>({
    queryKey: ['analytics', 'market-share-runs', selectedCalcuttaId],
    queryFn: async () => {
      if (!selectedCalcuttaId) return null;
      return analyticsService.listMarketShareRunsForCalcutta<{ runs: MarketShareRun[] }>(selectedCalcuttaId);
    },
    enabled: Boolean(selectedCalcuttaId) && activeTab === 'investments',
  });

  const filteredMarketShareRuns = useMemo(() => {
    const runs = marketShareRunsQuery.data?.runs ?? [];
    if (!selectedAlgorithmId) return runs;
    return runs.filter((r) => r.algorithm_id === selectedAlgorithmId);
  }, [marketShareRunsQuery.data, selectedAlgorithmId]);

  const predictedAdvancementQuery = useQuery<{ teams: TeamPredictedAdvancement[] } | null>({
    queryKey: ['analytics', 'predicted-advancement', selectedTournamentId, selectedGameOutcomeRunId],
    queryFn: async () => {
      if (!selectedTournamentId) return null;
      return analyticsService.getTournamentPredictedAdvancement<{ teams: TeamPredictedAdvancement[] }>({
        tournamentId: selectedTournamentId,
        gameOutcomeRunId: selectedGameOutcomeRunId ?? undefined,
      });
    },
    enabled: Boolean(selectedTournamentId) && activeTab === 'advancements',
  });

  const predictedMarketShareQuery = useQuery<{ teams: TeamPredictedMarketShare[] } | null>({
    queryKey: [
      'analytics',
      'predicted-market-share',
      selectedCalcuttaId,
      selectedMarketShareRunId,
      selectedInvestmentsGameOutcomeRunId,
    ],
    queryFn: async () => {
      if (!selectedCalcuttaId) return null;
      return analyticsService.getCalcuttaPredictedMarketShare<{ teams: TeamPredictedMarketShare[] }>({
        calcuttaId: selectedCalcuttaId,
        marketShareRunId: selectedMarketShareRunId ?? undefined,
        gameOutcomeRunId: selectedInvestmentsGameOutcomeRunId ?? undefined,
      });
    },
    enabled: Boolean(selectedCalcuttaId) && activeTab === 'investments',
  });

  const formatProb = (p: number) => `${(p * 100).toFixed(1)}%`;
  const formatShare = (p: number) => `${(p * 100).toFixed(3)}%`;
  const formatDeltaPercent = (p: number) => {
    if (!Number.isFinite(p)) return '—';
    const formatted = p.toFixed(1);
    return p > 0 ? `+${formatted}%` : `${formatted}%`;
  };

  const deltaClass = (p: number) => {
    if (!Number.isFinite(p)) return 'text-gray-700';
    if (p < 0) return 'text-green-700 font-semibold';
    if (p > 0) return 'text-red-700 font-semibold';
    return 'text-gray-700';
  };

  return (
    <PageContainer>
      <PageHeader title="Lab" subtitle="Browse registered algorithms and their run outputs." />

      <Card className="mb-6">
        <TabsNav tabs={tabs} activeTab={activeTab} onTabChange={setActiveTab} />
      </Card>

      <Card className="mb-6">
        <h2 className="text-xl font-semibold mb-4">Algorithm</h2>

        {algorithmsQuery.isLoading ? <div className="text-gray-500">Loading algorithms...</div> : null}

        {!algorithmsQuery.isLoading ? (
          <div className="flex items-center gap-4">
            <label htmlFor="algorithm-select" className="text-lg font-semibold whitespace-nowrap">
              Select Algorithm:
            </label>
            <Select
              id="algorithm-select"
              value={selectedAlgorithmId}
              onChange={(e) => {
                const next = e.target.value;
                setSelectedAlgorithmId(next);
                // Clear run selections so we don't keep stale run IDs when switching algorithms.
                setSelectedGameOutcomeRunId(null);
                setSelectedMarketShareRunId(null);
              }}
              className="flex-1 max-w-2xl"
            >
              <option value="">All algorithms</option>
              {algorithms.map((a) => (
                <option key={a.id} value={a.id}>
                  {a.name}
                </option>
              ))}
            </Select>
          </div>
        ) : null}

        {selectedAlgorithm ? (
          <div className="mt-4 text-sm text-gray-700">
            <div className="font-medium text-gray-900">{selectedAlgorithm.name}</div>
            {selectedAlgorithm.description ? <div className="text-gray-600">{selectedAlgorithm.description}</div> : null}
          </div>
        ) : null}
      </Card>

      {activeTab === 'advancements' && (
        <Card className="mb-6">
          <div className="flex items-center gap-4">
            <label htmlFor="tournament-select" className="text-lg font-semibold whitespace-nowrap">
              Select Tournament:
            </label>

            {tournamentsLoading ? (
              <div className="text-gray-500">Loading tournaments...</div>
            ) : tournaments.length === 0 ? (
              <div className="text-gray-500">No tournaments found</div>
            ) : (
              <Select
                id="tournament-select"
                value={selectedTournamentId || ''}
                onChange={(e) => {
                  const nextTournamentId = e.target.value || null;
                  setSelectedTournamentId(nextTournamentId);
                  setSelectedCalcuttaId(null);
                  setSelectedGameOutcomeRunId(null);
                  setSelectedInvestmentsGameOutcomeRunId(null);
                  setSelectedMarketShareRunId(null);
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

          {selectedTournamentId && (
            <div className="mt-4 flex items-center gap-4">
              <label htmlFor="game-outcome-run-select" className="text-lg font-semibold whitespace-nowrap">
                Game Outcomes Run:
              </label>

              {gameOutcomeRunsQuery.isLoading ? (
                <div className="text-gray-500">Loading runs...</div>
              ) : filteredGameOutcomeRuns.length ? (
                <Select
                  id="game-outcome-run-select"
                  value={selectedGameOutcomeRunId || ''}
                  onChange={(e) => setSelectedGameOutcomeRunId(e.target.value || null)}
                  className="flex-1 max-w-xl"
                >
                  <option value="">-- Latest --</option>
                  {filteredGameOutcomeRuns.map((run) => (
                    <option key={run.id} value={run.id}>
                      {new Date(run.created_at).toLocaleString()} ({run.id.slice(0, 8)})
                    </option>
                  ))}
                </Select>
              ) : (
                <div className="text-gray-500">No game outcome runs found for this tournament.</div>
              )}
            </div>
          )}

          <div className="mt-4 text-sm text-gray-600">Per-team cumulative probability of reaching each round (no points).</div>
        </Card>
      )}

      {activeTab === 'advancements' && selectedTournamentId && (
        <div className="bg-white rounded-lg shadow p-6">
          {predictedAdvancementQuery.isLoading ? (
            <div className="text-gray-500">Loading predicted advancement...</div>
          ) : predictedAdvancementQuery.data?.teams ? (
            <div className="overflow-x-auto">
              <table className="min-w-full divide-y divide-gray-200">
                <thead className="bg-gray-50">
                  <tr>
                    <th className="px-4 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider sticky left-0 bg-gray-50">
                      Team
                    </th>
                    <th className="px-4 py-3 text-center text-xs font-medium text-gray-500 uppercase tracking-wider">Seed</th>
                    <th className="px-4 py-3 text-center text-xs font-medium text-gray-500 uppercase tracking-wider">Region</th>
                    <th className="px-4 py-3 text-center text-xs font-medium text-gray-500 uppercase tracking-wider">PI</th>
                    <th className="px-4 py-3 text-center text-xs font-medium text-gray-500 uppercase tracking-wider">R64</th>
                    <th className="px-4 py-3 text-center text-xs font-medium text-gray-500 uppercase tracking-wider">R32</th>
                    <th className="px-4 py-3 text-center text-xs font-medium text-gray-500 uppercase tracking-wider">S16</th>
                    <th className="px-4 py-3 text-center text-xs font-medium text-gray-500 uppercase tracking-wider">E8</th>
                    <th className="px-4 py-3 text-center text-xs font-medium text-gray-500 uppercase tracking-wider">FF</th>
                    <th className="px-4 py-3 text-center text-xs font-medium text-gray-500 uppercase tracking-wider">Champ</th>
                    <th className="px-4 py-3 text-center text-xs font-medium text-gray-500 uppercase tracking-wider bg-blue-50">
                      Win
                    </th>
                  </tr>
                </thead>
                <tbody className="bg-white divide-y divide-gray-200">
                  {predictedAdvancementQuery.data.teams.map((team) => (
                    <tr key={team.team_id} className="hover:bg-gray-50">
                      <td className="px-4 py-3 text-sm font-medium text-gray-900 sticky left-0 bg-white">
                        {team.school_name}
                      </td>
                      <td className="px-4 py-3 text-sm text-center text-gray-700">{team.seed}</td>
                      <td className="px-4 py-3 text-sm text-center text-gray-600">{team.region}</td>
                      <td className="px-4 py-3 text-sm text-center text-gray-700">{formatProb(team.prob_pi)}</td>
                      <td className="px-4 py-3 text-sm text-center text-gray-700">{formatProb(team.reach_r64)}</td>
                      <td className="px-4 py-3 text-sm text-center text-gray-700">{formatProb(team.reach_r32)}</td>
                      <td className="px-4 py-3 text-sm text-center text-gray-700">{formatProb(team.reach_s16)}</td>
                      <td className="px-4 py-3 text-sm text-center text-gray-700">{formatProb(team.reach_e8)}</td>
                      <td className="px-4 py-3 text-sm text-center text-gray-700">{formatProb(team.reach_ff)}</td>
                      <td className="px-4 py-3 text-sm text-center text-gray-700">{formatProb(team.reach_champ)}</td>
                      <td className="px-4 py-3 text-sm text-center font-semibold text-blue-700 bg-blue-50">
                        {formatProb(team.win_champ)}
                      </td>
                    </tr>
                  ))}
                  <tr className="bg-gray-100 font-bold border-t-2 border-gray-300">
                    <td className="px-4 py-3 text-sm text-gray-900 sticky left-0 bg-gray-100">TOTAL</td>
                    <td className="px-4 py-3 text-sm text-center" colSpan={2}></td>
                    {(() => {
                      const teams = predictedAdvancementQuery.data?.teams ?? [];
                      const sum = (key: keyof TeamPredictedAdvancement) => teams.reduce((acc, t) => acc + (t[key] as number), 0);
                      return (
                        <>
                          <td className="px-4 py-3 text-sm text-center text-gray-900">{formatProb(sum('prob_pi'))}</td>
                          <td className="px-4 py-3 text-sm text-center text-gray-900">{formatProb(sum('reach_r64'))}</td>
                          <td className="px-4 py-3 text-sm text-center text-gray-900">{formatProb(sum('reach_r32'))}</td>
                          <td className="px-4 py-3 text-sm text-center text-gray-900">{formatProb(sum('reach_s16'))}</td>
                          <td className="px-4 py-3 text-sm text-center text-gray-900">{formatProb(sum('reach_e8'))}</td>
                          <td className="px-4 py-3 text-sm text-center text-gray-900">{formatProb(sum('reach_ff'))}</td>
                          <td className="px-4 py-3 text-sm text-center text-gray-900">{formatProb(sum('reach_champ'))}</td>
                          <td className="px-4 py-3 text-sm text-center text-blue-700 bg-blue-100">{formatProb(sum('win_champ'))}</td>
                        </>
                      );
                    })()}
                  </tr>
                </tbody>
              </table>
            </div>
          ) : (
            <div className="text-gray-500">No predicted advancement data available.</div>
          )}
        </div>
      )}

      {activeTab === 'investments' && (
        <Card className="mb-6">
          <div className="flex items-center gap-4">
            <label htmlFor="tournament-select" className="text-lg font-semibold whitespace-nowrap">
              Select Tournament:
            </label>

            {tournamentsLoading ? (
              <div className="text-gray-500">Loading tournaments...</div>
            ) : tournaments.length === 0 ? (
              <div className="text-gray-500">No tournaments found</div>
            ) : (
              <Select
                id="tournament-select"
                value={selectedTournamentId || ''}
                onChange={(e) => {
                  const nextTournamentId = e.target.value || null;
                  setSelectedTournamentId(nextTournamentId);
                  setSelectedCalcuttaId(null);
                  setSelectedMarketShareRunId(null);
                  setSelectedInvestmentsGameOutcomeRunId(null);
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

          {selectedTournamentId && (
            <div className="mt-4 flex items-center gap-4">
              <label htmlFor="calcutta-select" className="text-lg font-semibold whitespace-nowrap">
                Select Calcutta:
              </label>

              {calcuttasLoading ? (
                <div className="text-gray-500">Loading calcuttas...</div>
              ) : calcuttasForTournament.length === 0 ? (
                <div className="text-gray-500">No calcuttas found for this tournament</div>
              ) : (
                <Select
                  id="calcutta-select"
                  value={selectedCalcuttaId || ''}
                  onChange={(e) => {
                    setSelectedCalcuttaId(e.target.value || null);
                    setSelectedMarketShareRunId(null);
                    setSelectedInvestmentsGameOutcomeRunId(null);
                  }}
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
          )}

          {selectedCalcuttaId && (
            <div className="mt-4 flex items-center gap-4">
              <label htmlFor="market-share-run-select" className="text-lg font-semibold whitespace-nowrap">
                Market Share Run:
              </label>

              {marketShareRunsQuery.isLoading ? (
                <div className="text-gray-500">Loading runs...</div>
              ) : filteredMarketShareRuns.length ? (
                <Select
                  id="market-share-run-select"
                  value={selectedMarketShareRunId || ''}
                  onChange={(e) => setSelectedMarketShareRunId(e.target.value || null)}
                  className="flex-1 max-w-xl"
                >
                  <option value="">-- Latest --</option>
                  {filteredMarketShareRuns.map((run) => (
                    <option key={run.id} value={run.id}>
                      {new Date(run.created_at).toLocaleString()} ({run.id.slice(0, 8)})
                    </option>
                  ))}
                </Select>
              ) : (
                <div className="text-gray-500">No market share runs found for this calcutta.</div>
              )}
            </div>
          )}

          {selectedCalcuttaId && (
            <div className="mt-4 flex items-center gap-4">
              <label htmlFor="investments-game-outcome-run-select" className="text-lg font-semibold whitespace-nowrap">
                Game Outcomes Run:
              </label>

              {gameOutcomeRunsQuery.isLoading ? (
                <div className="text-gray-500">Loading runs...</div>
              ) : gameOutcomeRunsQuery.data?.runs?.length ? (
                <Select
                  id="investments-game-outcome-run-select"
                  value={selectedInvestmentsGameOutcomeRunId || ''}
                  onChange={(e) => setSelectedInvestmentsGameOutcomeRunId(e.target.value || null)}
                  className="flex-1 max-w-xl"
                >
                  <option value="">-- Latest --</option>
                  {gameOutcomeRunsQuery.data.runs.map((run) => (
                    <option key={run.id} value={run.id}>
                      {new Date(run.created_at).toLocaleString()} ({run.id.slice(0, 8)})
                    </option>
                  ))}
                </Select>
              ) : (
                <div className="text-gray-500">No game outcome runs found for this tournament.</div>
              )}
            </div>
          )}

          <div className="mt-4 text-sm text-gray-600">
            Per-team predicted and rational market share (with delta percent).
          </div>
        </Card>
      )}

      {activeTab === 'investments' && selectedCalcuttaId && (
        <div className="bg-white rounded-lg shadow p-6">
          {predictedMarketShareQuery.isLoading ? (
            <div className="text-gray-500">Loading predicted market share...</div>
          ) : predictedMarketShareQuery.data?.teams ? (
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
                  {predictedMarketShareQuery.data.teams.map((team) => (
                    <tr key={team.team_id} className="hover:bg-gray-50">
                      <td className="px-4 py-3 text-sm font-medium text-gray-900 sticky left-0 bg-white">
                        {team.school_name}
                      </td>
                      <td className="px-4 py-3 text-sm text-center text-gray-700">{team.seed}</td>
                      <td className="px-4 py-3 text-sm text-center text-gray-600">{team.region}</td>
                      <td className="px-4 py-3 text-sm text-center text-gray-700">{formatShare(team.rational_share)}</td>
                      <td className="px-4 py-3 text-sm text-center font-semibold text-green-700 bg-green-50">
                        {formatShare(team.predicted_share)}
                      </td>
                      <td className={`px-4 py-3 text-sm text-center ${deltaClass(team.delta_percent)}`}>{
                        formatDeltaPercent(team.delta_percent)
                      }</td>
                    </tr>
                  ))}
                  <tr className="bg-gray-100 font-bold border-t-2 border-gray-300">
                    <td className="px-4 py-3 text-sm text-gray-900 sticky left-0 bg-gray-100">TOTAL</td>
                    <td className="px-4 py-3 text-sm text-center" colSpan={2}></td>
                    <td className="px-4 py-3 text-sm text-center text-gray-900">
                      {formatShare(predictedMarketShareQuery.data.teams.reduce((sum, t) => sum + t.rational_share, 0))}
                    </td>
                    <td className="px-4 py-3 text-sm text-center text-green-700 bg-green-100">
                      {formatShare(predictedMarketShareQuery.data.teams.reduce((sum, t) => sum + t.predicted_share, 0))}
                    </td>
                    <td className="px-4 py-3 text-sm text-center text-gray-500">—</td>
                  </tr>
                </tbody>
              </table>
            </div>
          ) : (
            <div className="text-gray-500">No predicted market share data available.</div>
          )}
        </div>
      )}
    </PageContainer>
  );
}
