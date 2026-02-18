import React, { useMemo, useState } from 'react';
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query';
import { useParams, Link } from 'react-router-dom';
import { type ColumnDef } from '@tanstack/react-table';
import { TournamentTeam, TournamentModerator } from '../types/calcutta';
import { School } from '../types/school';
import { tournamentService } from '../services/tournamentService';
import { schoolService } from '../services/schoolService';
import { queryKeys } from '../queryKeys';
import { Alert } from '../components/ui/Alert';
import { Badge } from '../components/ui/Badge';
import { Breadcrumb } from '../components/ui/Breadcrumb';
import { Button } from '../components/ui/Button';
import { Card } from '../components/ui/Card';
import { DataTable } from '../components/ui/DataTable';
import { Input } from '../components/ui/Input';
import { LoadingState } from '../components/ui/LoadingState';
import { Modal, ModalActions } from '../components/ui/Modal';
import { PageContainer, PageHeader } from '../components/ui/Page';
import { useFlashMessage } from '../hooks/useFlashMessage';
import { formatDate } from '../utils/format';

export const TournamentViewPage: React.FC = () => {
  const { id } = useParams<{ id: string }>();
  const queryClient = useQueryClient();
  const [showAddModal, setShowAddModal] = useState(false);
  const [addEmail, setAddEmail] = useState('');
  const [modSuccess, flashModSuccess] = useFlashMessage();
  const [modError, setModError] = useState<string | null>(null);

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

  const moderatorsQuery = useQuery({
    queryKey: queryKeys.tournaments.moderators(id),
    enabled: Boolean(id),
    queryFn: () => tournamentService.getTournamentModerators(id!),
  });

  const grantMutation = useMutation({
    mutationFn: (email: string) => tournamentService.grantTournamentModerator(id!, email),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: queryKeys.tournaments.moderators(id) });
      setShowAddModal(false);
      setAddEmail('');
      setModError(null);
      flashModSuccess('Moderator added successfully.');
    },
    onError: (err) => {
      setModError(err instanceof Error ? err.message : 'Failed to add moderator');
    },
  });

  const revokeMutation = useMutation({
    mutationFn: (userId: string) => tournamentService.revokeTournamentModerator(id!, userId),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: queryKeys.tournaments.moderators(id) });
      flashModSuccess('Moderator removed successfully.');
    },
    onError: (err) => {
      setModError(err instanceof Error ? err.message : 'Failed to remove moderator');
    },
  });

  const teamColumns = useMemo<ColumnDef<TournamentTeam, unknown>[]>(() => [
    { accessorKey: 'seed', header: 'Seed' },
    {
      id: 'school',
      header: 'School',
      accessorFn: (row) => schools[row.schoolId]?.name || 'Unknown School',
    },
    { accessorKey: 'byes', header: 'Byes' },
    { accessorKey: 'wins', header: 'Wins' },
    {
      id: 'status',
      header: 'Status',
      accessorFn: (row) => (row.eliminated ? 1 : 0),
      cell: ({ row }) => row.original.eliminated ? (
        <Badge variant="destructive">Eliminated</Badge>
      ) : (
        <Badge variant="success">Active</Badge>
      ),
    },
  ], [schools]);

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

  const tournament = tournamentQuery.data || null;
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
          { label: tournament.name },
        ]}
      />
      <PageHeader
        title={tournament.name}
        subtitle={`${tournament.rounds} rounds • Created ${formatDate(tournament.created)}`}
        actions={
          <div className="flex gap-2">
            <Link to={`/admin/tournaments/${id}/bracket`}>
              <Button variant="outline">Manage Bracket</Button>
            </Link>
            <Link to={`/admin/tournaments/${id}/edit`}>
              <Button variant="secondary">Edit Tournament</Button>
            </Link>
            <Link to={`/admin/tournaments/${id}/teams/add`}>
              <Button>Add Teams</Button>
            </Link>
          </div>
        }
      />

      {teams.length > 0 && (
        <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-4 gap-4 mb-8">
          <Card className="p-4">
            <h3 className="text-lg font-semibold text-gray-700 mb-2">Total Teams</h3>
            <p className="text-3xl font-bold text-blue-600">{teams.length}</p>
          </Card>
          <Card className="p-4">
            <h3 className="text-lg font-semibold text-gray-700 mb-2">Seed Distribution</h3>
            <div className="space-y-1">
              {Array.from({ length: 16 }, (_, i) => i + 1).map(seed => {
                const count = teams.filter(t => t.seed === seed).length;
                return count > 0 ? (
                  <div key={seed} className="flex justify-between text-sm">
                    <span>Seed {seed}:</span>
                    <span className="font-medium">{count}</span>
                  </div>
                ) : null;
              })}
            </div>
          </Card>
          <Card className="p-4">
            <h3 className="text-lg font-semibold text-gray-700 mb-2">Bye Distribution</h3>
            <div className="space-y-1">
              {Array.from({ length: Math.max(...teams.map(t => t.byes)) + 1 }, (_, i) => i).map(byes => {
                const count = teams.filter(t => t.byes === byes).length;
                return count > 0 ? (
                  <div key={byes} className="flex justify-between text-sm">
                    <span>{byes} {byes === 1 ? 'Bye' : 'Byes'}:</span>
                    <span className="font-medium">{count}</span>
                  </div>
                ) : null;
              })}
            </div>
          </Card>
          <Card className="p-4">
            <h3 className="text-lg font-semibold text-gray-700 mb-2">Win Distribution</h3>
            <div className="space-y-1">
              {Array.from({ length: Math.max(...teams.map(t => t.wins)) + 1 }, (_, i) => i).map(wins => {
                const count = teams.filter(t => t.wins === wins).length;
                return count > 0 ? (
                  <div key={wins} className="flex justify-between text-sm">
                    <span>{wins} {wins === 1 ? 'Win' : 'Wins'}:</span>
                    <span className="font-medium">{count}</span>
                  </div>
                ) : null;
              })}
              <div className="pt-2 mt-2 border-t border-gray-200">
                <div className="flex justify-between text-sm font-semibold">
                  <span>Total Wins:</span>
                  <span className="text-blue-600">{teams.reduce((sum, team) => sum + team.wins, 0)}</span>
                </div>
              </div>
            </div>
          </Card>
        </div>
      )}

      {teams.length === 0 ? (
        <Card className="text-center">
          <p className="text-gray-500 mb-4">No teams have been added to this tournament yet.</p>
          <Link to={`/admin/tournaments/${id}/teams/add`}>
            <Button>Add Teams</Button>
          </Link>
        </Card>
      ) : (
        <Card className="p-0 overflow-hidden">
          <DataTable
            columns={teamColumns}
            data={teams}
            initialSorting={[{ id: 'seed', desc: false }]}
          />
        </Card>
      )}

      <Card className="mt-8">
        <div className="flex items-center justify-between mb-4">
          <h3 className="text-lg font-semibold text-gray-900">Moderators</h3>
          <Button size="sm" onClick={() => { setShowAddModal(true); setModError(null); }}>
            Add Moderator
          </Button>
        </div>

        {modSuccess && <Alert variant="success" className="mb-4">{modSuccess}</Alert>}
        {modError && !showAddModal && <Alert variant="error" className="mb-4">{modError}</Alert>}

        {moderatorsQuery.isLoading ? (
          <LoadingState label="Loading moderators..." />
        ) : moderatorsQuery.isError ? (
          <Alert variant="error">
            Failed to load moderators.
            <Button size="sm" className="ml-2" onClick={() => moderatorsQuery.refetch()}>Retry</Button>
          </Alert>
        ) : (moderatorsQuery.data ?? []).length === 0 ? (
          <p className="text-gray-500 text-sm">No moderators have been added yet. Moderators can enter game results on your behalf.</p>
        ) : (
          <table className="w-full text-sm">
            <thead>
              <tr className="border-b border-gray-200 text-left text-gray-500">
                <th className="pb-2 font-medium">Name</th>
                <th className="pb-2 font-medium">Email</th>
                <th className="pb-2 font-medium text-right">Actions</th>
              </tr>
            </thead>
            <tbody>
              {(moderatorsQuery.data ?? []).map((mod: TournamentModerator) => (
                <tr key={mod.id} className="border-b border-gray-100 last:border-0">
                  <td className="py-2">{mod.firstName} {mod.lastName}</td>
                  <td className="py-2 text-gray-600">{mod.email}</td>
                  <td className="py-2 text-right">
                    <Button
                      size="sm"
                      variant="destructive"
                      disabled={revokeMutation.isPending}
                      onClick={() => revokeMutation.mutate(mod.id)}
                    >
                      Remove
                    </Button>
                  </td>
                </tr>
              ))}
            </tbody>
          </table>
        )}
      </Card>

      <Modal open={showAddModal} onClose={() => { setShowAddModal(false); setAddEmail(''); setModError(null); }} title="Add Moderator">
        <p className="text-sm text-gray-600 mb-4">Enter the email address of the user you want to add as a moderator.</p>
        {modError && <Alert variant="error" className="mb-4">{modError}</Alert>}
        <form onSubmit={(e) => { e.preventDefault(); if (addEmail.trim()) grantMutation.mutate(addEmail.trim()); }}>
          <Input
            type="email"
            placeholder="user@example.com"
            value={addEmail}
            onChange={(e) => setAddEmail(e.target.value)}
            required
          />
          <ModalActions>
            <Button variant="secondary" type="button" onClick={() => { setShowAddModal(false); setAddEmail(''); setModError(null); }}>
              Cancel
            </Button>
            <Button type="submit" disabled={grantMutation.isPending || !addEmail.trim()}>
              {grantMutation.isPending ? 'Adding...' : 'Add'}
            </Button>
          </ModalActions>
        </form>
      </Modal>
    </PageContainer>
  );
};
