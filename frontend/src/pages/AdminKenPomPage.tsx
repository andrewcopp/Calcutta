import { useState } from 'react';
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query';
import { tournamentService } from '../services/tournamentService';
import { queryKeys } from '../queryKeys';
import { TournamentTeam } from '../schemas/tournament';
import { Breadcrumb } from '../components/ui/Breadcrumb';
import { Button } from '../components/ui/Button';
import { Card } from '../components/ui/Card';
import { Input } from '../components/ui/Input';
import { Select } from '../components/ui/Select';
import { Alert } from '../components/ui/Alert';
import { LoadingState } from '../components/ui/LoadingState';
import { ErrorState } from '../components/ui/ErrorState';
import { PageContainer, PageHeader } from '../components/ui/Page';

type KenPomFormState = Record<string, { netRtg: number; oRtg: number; dRtg: number; adjT: number }>;

function buildFormState(teams: TournamentTeam[]): KenPomFormState {
  const state: KenPomFormState = {};
  for (const team of teams) {
    state[team.id] = {
      netRtg: team.kenPom?.netRtg ?? 0,
      oRtg: team.kenPom?.oRtg ?? 0,
      dRtg: team.kenPom?.dRtg ?? 0,
      adjT: team.kenPom?.adjT ?? 0,
    };
  }
  return state;
}

function sortTeams(teams: TournamentTeam[]): TournamentTeam[] {
  return [...teams].sort((a, b) => {
    if (a.region < b.region) return -1;
    if (a.region > b.region) return 1;
    return a.seed - b.seed;
  });
}

export function AdminKenPomPage() {
  const queryClient = useQueryClient();
  const [selectedTournamentId, setSelectedTournamentId] = useState<string>('');
  const [formState, setFormState] = useState<KenPomFormState>({});
  const [successMessage, setSuccessMessage] = useState('');

  const tournamentsQuery = useQuery({
    queryKey: queryKeys.tournaments.all(),
    queryFn: () => tournamentService.getAllTournaments(),
  });

  const teamsQuery = useQuery({
    queryKey: queryKeys.tournaments.teams(selectedTournamentId || undefined),
    queryFn: () => tournamentService.getTournamentTeams(selectedTournamentId),
    enabled: !!selectedTournamentId,
  });

  const mutation = useMutation({
    mutationFn: (stats: { teamId: string; netRtg: number; oRtg: number; dRtg: number; adjT: number }[]) =>
      tournamentService.updateKenPomStats(selectedTournamentId, stats),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: queryKeys.tournaments.teams(selectedTournamentId) });
      setSuccessMessage('KenPom ratings saved successfully.');
      setTimeout(() => setSuccessMessage(''), 3000);
    },
  });

  const handleTournamentChange = (id: string) => {
    setSelectedTournamentId(id);
    setFormState({});
    setSuccessMessage('');
  };

  // Initialize form state when teams load
  const teams = teamsQuery.data;
  if (teams && Object.keys(formState).length === 0 && teams.length > 0) {
    setFormState(buildFormState(teams));
  }

  const handleFieldChange = (teamId: string, field: keyof KenPomFormState[string], value: string) => {
    setFormState((prev) => ({
      ...prev,
      [teamId]: {
        ...prev[teamId],
        [field]: value === '' ? 0 : parseFloat(value),
      },
    }));
  };

  const handleSave = () => {
    const stats = Object.entries(formState).map(([teamId, values]) => ({
      teamId,
      netRtg: values.netRtg,
      oRtg: values.oRtg,
      dRtg: values.dRtg,
      adjT: values.adjT,
    }));
    mutation.mutate(stats);
  };

  const sortedTeams = teams ? sortTeams(teams) : [];

  return (
    <PageContainer>
      <Breadcrumb items={[{ label: 'Admin', href: '/admin' }, { label: 'KenPom Ratings' }]} />
      <PageHeader title="KenPom Ratings" subtitle="Enter or edit KenPom ratings for tournament teams" />

      <Card>
        <div className="mb-6">
          <label htmlFor="tournament-select" className="block text-sm font-medium text-foreground mb-2">
            Tournament
          </label>
          {tournamentsQuery.isLoading ? (
            <LoadingState />
          ) : tournamentsQuery.isError ? (
            <ErrorState error="Failed to load tournaments" onRetry={() => tournamentsQuery.refetch()} />
          ) : (
            <Select
              id="tournament-select"
              value={selectedTournamentId}
              onChange={(e) => handleTournamentChange(e.target.value)}
            >
              <option value="">Select a tournament...</option>
              {tournamentsQuery.data?.map((t) => (
                <option key={t.id} value={t.id}>
                  {t.name}
                </option>
              ))}
            </Select>
          )}
        </div>

        {selectedTournamentId && teamsQuery.isLoading && <LoadingState />}
        {selectedTournamentId && teamsQuery.isError && (
          <ErrorState error="Failed to load teams" onRetry={() => teamsQuery.refetch()} />
        )}

        {successMessage && (
          <Alert variant="success" className="mb-4">
            {successMessage}
          </Alert>
        )}
        {mutation.isError && (
          <Alert variant="error" className="mb-4">
            {mutation.error instanceof Error ? mutation.error.message : 'Failed to save KenPom ratings'}
          </Alert>
        )}

        {sortedTeams.length > 0 && Object.keys(formState).length > 0 && (
          <>
            <div className="overflow-x-auto">
              <table className="min-w-full text-sm">
                <thead>
                  <tr className="border-b text-left text-muted-foreground">
                    <th className="px-3 py-2 font-medium">Region</th>
                    <th className="px-3 py-2 font-medium">Seed</th>
                    <th className="px-3 py-2 font-medium">School</th>
                    <th className="px-3 py-2 font-medium">Net Rtg</th>
                    <th className="px-3 py-2 font-medium">O Rtg</th>
                    <th className="px-3 py-2 font-medium">D Rtg</th>
                    <th className="px-3 py-2 font-medium">Adj T</th>
                  </tr>
                </thead>
                <tbody>
                  {sortedTeams.map((team) => (
                    <tr key={team.id} className="border-b">
                      <td className="px-3 py-2">{team.region}</td>
                      <td className="px-3 py-2">{team.seed}</td>
                      <td className="px-3 py-2">{team.school?.name ?? team.schoolId}</td>
                      <td className="px-3 py-2">
                        <Input
                          type="number"
                          step="0.1"
                          className="w-24"
                          value={formState[team.id]?.netRtg ?? 0}
                          onChange={(e) => handleFieldChange(team.id, 'netRtg', e.target.value)}
                        />
                      </td>
                      <td className="px-3 py-2">
                        <Input
                          type="number"
                          step="0.1"
                          className="w-24"
                          value={formState[team.id]?.oRtg ?? 0}
                          onChange={(e) => handleFieldChange(team.id, 'oRtg', e.target.value)}
                        />
                      </td>
                      <td className="px-3 py-2">
                        <Input
                          type="number"
                          step="0.1"
                          className="w-24"
                          value={formState[team.id]?.dRtg ?? 0}
                          onChange={(e) => handleFieldChange(team.id, 'dRtg', e.target.value)}
                        />
                      </td>
                      <td className="px-3 py-2">
                        <Input
                          type="number"
                          step="0.1"
                          className="w-24"
                          value={formState[team.id]?.adjT ?? 0}
                          onChange={(e) => handleFieldChange(team.id, 'adjT', e.target.value)}
                        />
                      </td>
                    </tr>
                  ))}
                </tbody>
              </table>
            </div>

            <div className="mt-6 flex justify-end">
              <Button onClick={handleSave} loading={mutation.isPending}>
                Save
              </Button>
            </div>
          </>
        )}
      </Card>
    </PageContainer>
  );
}
