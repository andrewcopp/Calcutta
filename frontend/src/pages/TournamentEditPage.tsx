import React, { useMemo, useState } from 'react';
import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query';
import { useParams, useNavigate } from 'react-router-dom';
import { Tournament, TournamentTeam } from '../types/calcutta';
import { School } from '../types/school';
import { tournamentService } from '../services/tournamentService';
import { schoolService } from '../services/schoolService';
import { queryKeys } from '../queryKeys';
import { Alert } from '../components/ui/Alert';
import { Badge } from '../components/ui/Badge';
import { ErrorState } from '../components/ui/ErrorState';
import { Breadcrumb } from '../components/ui/Breadcrumb';
import { Button } from '../components/ui/Button';
import { Card } from '../components/ui/Card';
import { Input } from '../components/ui/Input';
import { Select } from '../components/ui/Select';
import { LoadingState } from '../components/ui/LoadingState';
import { PageContainer, PageHeader } from '../components/ui/Page';
import { formatDate } from '../utils/format';

export const TournamentEditPage: React.FC = () => {
  const { id } = useParams<{ id: string }>();
  const navigate = useNavigate();
  const queryClient = useQueryClient();
  const [error, setError] = useState<string | null>(null);
  const [isSaving, setIsSaving] = useState(false);

  const tournamentQuery = useQuery({
    queryKey: queryKeys.tournaments.detail(id),
    enabled: Boolean(id),
    queryFn: () => tournamentService.getTournament(id!),
  });

  const teamsQuery = useQuery({
    queryKey: queryKeys.tournaments.teams(id),
    enabled: Boolean(id),
    queryFn: () => tournamentService.getTournamentTeams(id!),
  });

  const schoolsQuery = useQuery({
    queryKey: queryKeys.schools.all(),
    queryFn: () => schoolService.getSchools(),
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
    onError: (error, _vars, context) => {
      if (context?.previousTeams) {
        queryClient.setQueryData(queryKeys.tournaments.teams(id), context.previousTeams);
      }
      setError(error instanceof Error ? error.message : 'Failed to update team');
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
    onError: (error) => {
      setError(error instanceof Error ? error.message : 'Failed to save changes');
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
      setError(err instanceof Error ? err.message : 'Failed to save changes');
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
    const firstError = tournamentQuery.error ?? teamsQuery.error ?? schoolsQuery.error;
    return (
      <PageContainer>
        <ErrorState
          error={firstError}
          onRetry={() => {
            if (tournamentQuery.isError) tournamentQuery.refetch();
            if (teamsQuery.isError) teamsQuery.refetch();
            if (schoolsQuery.isError) schoolsQuery.refetch();
          }}
        />
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
      <Breadcrumb
        items={[
          { label: 'Tournaments', href: '/admin/tournaments' },
          { label: tournament.name, href: `/admin/tournaments/${id}` },
          { label: 'Edit' },
        ]}
      />
      <PageHeader
        title={`Edit Tournament: ${tournament.name}`}
        subtitle={`${tournament.rounds} rounds • Created ${formatDate(tournament.created)}`}
        actions={
          <div className="flex gap-2">
            <Button variant="outline" onClick={() => navigate(`/admin/tournaments/${id}`)}>
              Cancel
            </Button>
            <Button onClick={handleSaveAll} disabled={isSaving} loading={isSaving}>
              {isSaving ? 'Saving...' : 'Save All Changes'}
            </Button>
          </div>
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
                Region
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
                    <Select
                      value={String(team.seed)}
                      onChange={(e) => handleTeamUpdate(team.id, 'seed', parseInt(e.target.value) || 1)}
                      className="w-20"
                    >
                      {Array.from({ length: 16 }, (_, i) => i + 1).map(seed => (
                        <option key={seed} value={seed}>
                          {seed}
                        </option>
                      ))}
                    </Select>
                  </td>
                  <td className="px-6 py-4 whitespace-nowrap text-sm text-gray-900">
                    {schools[team.schoolId]?.name || 'Unknown School'}
                  </td>
                  <td className="px-6 py-4 whitespace-nowrap text-sm text-gray-500">
                    <Badge variant="outline">{team.region || 'Unknown'}</Badge>
                  </td>
                  <td className="px-6 py-4 whitespace-nowrap">
                    <Input
                      type="number"
                      min="0"
                      max="7"
                      value={team.byes}
                      onChange={(e) => handleTeamUpdate(team.id, 'byes', parseInt(e.target.value) || 0)}
                      className="w-20"
                    />
                  </td>
                  <td className="px-6 py-4 whitespace-nowrap">
                    <Input
                      type="number"
                      min="0"
                      max="7"
                      value={team.wins}
                      onChange={(e) => handleTeamUpdate(team.id, 'wins', parseInt(e.target.value) || 0)}
                      className="w-20"
                    />
                  </td>
                  <td className="px-6 py-4 whitespace-nowrap">
                    <Select
                      value={team.eliminated ? 'eliminated' : 'active'}
                      onChange={(e) => handleTeamUpdate(team.id, 'eliminated', e.target.value === 'eliminated')}
                    >
                      <option value="active">Active</option>
                      <option value="eliminated">Eliminated</option>
                    </Select>
                  </td>
                </tr>
              ))}
          </tbody>
        </table>
      </Card>
    </PageContainer>
  );
};