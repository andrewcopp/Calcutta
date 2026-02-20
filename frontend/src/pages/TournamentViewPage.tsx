import React, { useMemo } from 'react';
import { useQuery } from '@tanstack/react-query';
import { useParams, Link } from 'react-router-dom';
import { type ColumnDef } from '@tanstack/react-table';
import { TournamentTeam } from '../types/calcutta';
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
import { LoadingState } from '../components/ui/LoadingState';
import { PageContainer, PageHeader } from '../components/ui/Page';
import { formatDate } from '../utils/format';
import { TeamStatsCards } from './Tournament/TeamStatsCards';
import { ModeratorsSection } from './Tournament/ModeratorsSection';

export const TournamentViewPage: React.FC = () => {
  const { id } = useParams<{ id: string }>();

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
            <Link to={`/admin/tournaments/${id}/teams/setup`}>
              <Button variant="secondary">Setup Teams</Button>
            </Link>
            <Link to={`/admin/tournaments/${id}/teams/add`}>
              <Button>Add Teams</Button>
            </Link>
          </div>
        }
      />

      {teams.length > 0 && <TeamStatsCards teams={teams} />}

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

      <ModeratorsSection tournamentId={id} />
    </PageContainer>
  );
};
