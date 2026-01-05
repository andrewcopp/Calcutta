import React, { useMemo, useState } from 'react';
import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query';
import { useParams, useNavigate } from 'react-router-dom';
import { Tournament, TournamentTeam } from '../types/calcutta';
import { School } from '../types/school';
import { tournamentService } from '../services/tournamentService';
import { adminService } from '../services/adminService';
import { queryKeys } from '../queryKeys';
import { Alert } from '../components/ui/Alert';
import { Button } from '../components/ui/Button';
import { Card } from '../components/ui/Card';
import { LoadingState } from '../components/ui/LoadingState';
import { PageContainer, PageHeader } from '../components/ui/Page';

export const TournamentEditPage: React.FC = () => {
  const { id } = useParams<{ id: string }>();
  const navigate = useNavigate();
  const queryClient = useQueryClient();
  const [error, setError] = useState<string | null>(null);
  const [isSaving, setIsSaving] = useState(false);

  const tournamentQuery = useQuery({
    queryKey: queryKeys.tournaments.detail(id),
    enabled: Boolean(id),
    staleTime: 30_000,
    queryFn: () => tournamentService.getTournament(id!),
  });

  const teamsQuery = useQuery({
    queryKey: queryKeys.tournaments.teams(id),
    enabled: Boolean(id),
    staleTime: 30_000,
    queryFn: () => tournamentService.getTournamentTeams(id!),
  });

  const schoolsQuery = useQuery({
    queryKey: queryKeys.schools.all(),
    staleTime: 30_000,
    queryFn: () => adminService.getAllSchools(),
  });

  const schools = useMemo(() => {
    const schoolsData = schoolsQuery.data || [];
    return schoolsData.reduce((acc, school) => {
      acc[school.id] = school;
      return acc;
    }, {} as Record<string, School>);
  }, [schoolsQuery.data]);

  const updateTeamMutation = useMutation({
    mutationFn: async ({ teamId, updates }: { teamId: string; updates: Partial<TournamentTeam> }) => {
      return tournamentService.updateTournamentTeam(id!, teamId, updates);
    },
    onMutate: async ({ teamId, updates }) => {
      setError(null);

      await queryClient.cancelQueries({ queryKey: queryKeys.tournaments.teams(id) });

      const previousTeams = queryClient.getQueryData<TournamentTeam[]>(queryKeys.tournaments.teams(id));

      queryClient.setQueryData<TournamentTeam[]>(queryKeys.tournaments.teams(id), (current) => {
        const currentTeams = current || [];
        return currentTeams.map((team) => (team.id === teamId ? ({ ...team, ...updates } as TournamentTeam) : team));
      });

      return { previousTeams };
    },
    onError: (_err, _vars, context) => {
      if (context?.previousTeams) {
        queryClient.setQueryData(queryKeys.tournaments.teams(id), context.previousTeams);
      }
      setError('Failed to update team');
    },
    onSuccess: (updatedTeam) => {
      queryClient.setQueryData<TournamentTeam[]>(queryKeys.tournaments.teams(id), (current) => {
        const currentTeams = current || [];
        return currentTeams.map((team) => (team.id === updatedTeam.id ? updatedTeam : team));
      });
    },
  });

  const saveAllMutation = useMutation({
    mutationFn: async () => {
      const teams = queryClient.getQueryData<TournamentTeam[]>(queryKeys.tournaments.teams(id)) || [];

      await Promise.all(
        teams.map((team) =>
          tournamentService.updateTournamentTeam(id!, team.id, {
            seed: team.seed,
            byes: team.byes,
            wins: team.wins,
            eliminated: team.eliminated,
          })
        )
      );
    },
    onSuccess: async () => {
      await queryClient.invalidateQueries({ queryKey: queryKeys.tournaments.teams(id) });
      navigate(`/admin/tournaments/${id}`);
    },
    onError: () => {
      setError('Failed to save changes');
    },
  });

  const handleTeamUpdate = async (teamId: string, field: keyof TournamentTeam, value: TournamentTeam[keyof TournamentTeam]) => {
    updateTeamMutation.mutate({
      teamId,
      updates: { [field]: value } as Partial<TournamentTeam>,
    });
  };

  const handleSaveAll = async () => {
    setIsSaving(true);
    setError(null);

    try {
      await saveAllMutation.mutateAsync();
    } catch (err) {
      setError('Failed to save changes');
      console.error('Error saving changes:', err);
    } finally {
      setIsSaving(false);
    }
  };

  if (!id) {
    return (
      <PageContainer>
        <Alert variant="error">Missing tournament id</Alert>
      </PageContainer>
    );
  }

  if (tournamentQuery.isLoading || teamsQuery.isLoading || schoolsQuery.isLoading) {
    return (
      <PageContainer>
        <LoadingState label="Loading tournament..." />
      </PageContainer>
    );
  }

  if (tournamentQuery.isError || teamsQuery.isError || schoolsQuery.isError) {
    const tournamentMessage = tournamentQuery.error instanceof Error ? tournamentQuery.error.message : 'Failed to load tournament';
    const teamsMessage = teamsQuery.error instanceof Error ? teamsQuery.error.message : 'Failed to load teams';
    const schoolsMessage = schoolsQuery.error instanceof Error ? schoolsQuery.error.message : 'Failed to load schools';

    return (
      <PageContainer>
        <Alert variant="error">
          <div className="font-semibold mb-1">Failed to load tournament data</div>
          {tournamentQuery.isError ? <div className="text-sm">Tournament: {tournamentMessage}</div> : null}
          {teamsQuery.isError ? <div className="text-sm">Teams: {teamsMessage}</div> : null}
          {schoolsQuery.isError ? <div className="text-sm">Schools: {schoolsMessage}</div> : null}

          <div className="mt-4 flex flex-wrap gap-2">
            {tournamentQuery.isError ? (
              <Button size="sm" onClick={() => tournamentQuery.refetch()}>
                Retry tournament
              </Button>
            ) : null}
            {teamsQuery.isError ? (
              <Button size="sm" variant="secondary" onClick={() => teamsQuery.refetch()}>
                Retry teams
              </Button>
            ) : null}
            {schoolsQuery.isError ? (
              <Button size="sm" variant="secondary" onClick={() => schoolsQuery.refetch()}>
                Retry schools
              </Button>
            ) : null}
          </div>
        </Alert>
      </PageContainer>
    );
  }

  const tournament: Tournament | null = tournamentQuery.data || null;
  const teams: TournamentTeam[] = teamsQuery.data || [];

  if (!tournament) {
    return (
      <PageContainer>
        <Alert variant="error">Tournament not found</Alert>
      </PageContainer>
    );
  }

  return (
    <PageContainer>
      <PageHeader
        title={`Edit Tournament: ${tournament.name}`}
        subtitle={`${tournament.rounds} rounds • Created ${new Date(tournament.created).toLocaleDateString()}`}
        actions={
          <>
            <button onClick={() => navigate(`/admin/tournaments/${id}`)} className="px-4 py-2 border rounded hover:bg-gray-100">
              Cancel
            </button>
            <button
              onClick={handleSaveAll}
              disabled={isSaving}
              className="bg-blue-500 text-white px-4 py-2 rounded hover:bg-blue-600 disabled:opacity-50"
            >
              {isSaving ? 'Saving...' : 'Save All Changes'}
            </button>
          </>
        }
      />

      {error && (
        <Alert variant="error" className="mb-4">
          {error}
        </Alert>
      )}

      <Card className="p-0 overflow-hidden">
        <table className="min-w-full divide-y divide-gray-200">
          <thead className="bg-gray-50">
            <tr>
              <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">
                Seed
              </th>
              <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">
                School
              </th>
              <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">
                Byes
              </th>
              <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">
                Wins
              </th>
              <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">
                Status
              </th>
            </tr>
          </thead>
          <tbody className="bg-white divide-y divide-gray-200">
            {teams
              .sort((a, b) => a.seed - b.seed)
              .map(team => (
                <tr key={team.id}>
                  <td className="px-6 py-4 whitespace-nowrap text-sm font-medium text-gray-900">
                    <select
                      value={team.seed}
                      onChange={(e) => handleTeamUpdate(team.id, 'seed', parseInt(e.target.value) || 1)}
                      className="w-16 p-1 border rounded"
                    >
                      {Array.from({ length: 16 }, (_, i) => i + 1).map(seed => (
                        <option key={seed} value={seed}>
                          {seed}
                        </option>
                      ))}
                    </select>
                  </td>
                  <td className="px-6 py-4 whitespace-nowrap text-sm text-gray-900">
                    {schools[team.schoolId]?.name || 'Unknown School'}
                  </td>
                  <td className="px-6 py-4 whitespace-nowrap">
                    <input
                      type="number"
                      min="0"
                      max="7"
                      value={team.byes}
                      onChange={(e) => handleTeamUpdate(team.id, 'byes', parseInt(e.target.value) || 0)}
                      className="w-16 p-1 border rounded"
                    />
                  </td>
                  <td className="px-6 py-4 whitespace-nowrap">
                    <input
                      type="number"
                      min="0"
                      max="7"
                      value={team.wins}
                      onChange={(e) => handleTeamUpdate(team.id, 'wins', parseInt(e.target.value) || 0)}
                      className="w-16 p-1 border rounded"
                    />
                  </td>
                  <td className="px-6 py-4 whitespace-nowrap">
                    <select
                      value={team.eliminated ? 'eliminated' : 'active'}
                      onChange={(e) => handleTeamUpdate(team.id, 'eliminated', e.target.value === 'eliminated')}
                      className="p-1 border rounded"
                    >
                      <option value="active">Active</option>
                      <option value="eliminated">Eliminated</option>
                    </select>
                  </td>
                </tr>
              ))}
          </tbody>
        </table>
      </Card>
    </PageContainer>
  );
};