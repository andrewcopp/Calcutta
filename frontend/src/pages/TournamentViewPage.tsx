import React, { useMemo, useState } from 'react';
import { useQuery } from '@tanstack/react-query';
import { useParams, Link } from 'react-router-dom';
import { TournamentTeam } from '../types/calcutta';
import { School } from '../types/school';
import { tournamentService } from '../services/tournamentService';
import { adminService } from '../services/adminService';
import { queryKeys } from '../queryKeys';
import { Alert } from '../components/ui/Alert';
import { Button } from '../components/ui/Button';
import { Card } from '../components/ui/Card';
import { LoadingState } from '../components/ui/LoadingState';
import { PageContainer, PageHeader } from '../components/ui/Page';

type SortField = 'seed' | 'school' | 'byes' | 'wins' | 'status';
type SortDirection = 'asc' | 'desc';

export const TournamentViewPage: React.FC = () => {
  const { id } = useParams<{ id: string }>();
  const [sortField, setSortField] = useState<SortField>('seed');
  const [sortDirection, setSortDirection] = useState<SortDirection>('asc');

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

  const handleSort = (field: SortField) => {
    if (field === sortField) {
      // Toggle direction if clicking the same field
      setSortDirection(sortDirection === 'asc' ? 'desc' : 'asc');
    } else {
      // Set new field and default to ascending
      setSortField(field);
      setSortDirection('asc');
    }
  };

  const getSortedTeams = () => {
    const teams = teamsQuery.data || [];
    return [...teams].sort((a, b) => {
      let comparison = 0;

      const schoolA = schools[a.schoolId]?.name || '';
      const schoolB = schools[b.schoolId]?.name || '';

      const comparisons: Record<SortField, number> = {
        seed: a.seed - b.seed,
        school: schoolA.localeCompare(schoolB),
        byes: a.byes - b.byes,
        wins: a.wins - b.wins,
        // Sort eliminated teams to the bottom
        status: (a.eliminated ? 1 : 0) - (b.eliminated ? 1 : 0),
      };

      comparison = comparisons[sortField] ?? 0;
      
      return sortDirection === 'asc' ? comparison : -comparison;
    });
  };

  const getSortIcon = (field: SortField) => {
    if (field !== sortField) {
      return <span className="text-gray-400">↕</span>;
    }
    return sortDirection === 'asc' ? <span>↑</span> : <span>↓</span>;
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

  const tournament = tournamentQuery.data || null;
  const teams: TournamentTeam[] = teamsQuery.data || [];

  if (!tournament) {
    return (
      <PageContainer>
        <Alert variant="error">Tournament not found</Alert>
      </PageContainer>
    );
  }

  const sortedTeams = getSortedTeams();

  return (
    <PageContainer>
      <PageHeader
        title={tournament.name}
        subtitle={`${tournament.rounds} rounds • Created ${new Date(tournament.created).toLocaleDateString()}`}
        actions={
          <>
            <Link to={`/admin/tournaments/${id}/bracket`} className="bg-purple-500 text-white px-4 py-2 rounded hover:bg-purple-600">
              Manage Bracket
            </Link>
            <Link to={`/admin/tournaments/${id}/edit`} className="bg-blue-500 text-white px-4 py-2 rounded hover:bg-blue-600">
              Edit Tournament
            </Link>
            <Link to={`/admin/tournaments/${id}/teams/add`} className="bg-green-500 text-white px-4 py-2 rounded hover:bg-green-600">
              Add Teams
            </Link>
          </>
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
          <Link
            to={`/admin/tournaments/${id}/teams/add`}
            className="bg-blue-500 text-white px-4 py-2 rounded hover:bg-blue-600"
          >
            Add Teams
          </Link>
        </Card>
      ) : (
        <Card className="p-0 overflow-hidden">
          <table className="min-w-full divide-y divide-gray-200">
            <thead className="bg-gray-50">
              <tr>
                <th 
                  className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider cursor-pointer"
                  onClick={() => handleSort('seed')}
                >
                  <div className="flex items-center">
                    Seed {getSortIcon('seed')}
                  </div>
                </th>
                <th 
                  className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider cursor-pointer"
                  onClick={() => handleSort('school')}
                >
                  <div className="flex items-center">
                    School {getSortIcon('school')}
                  </div>
                </th>
                <th 
                  className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider cursor-pointer"
                  onClick={() => handleSort('byes')}
                >
                  <div className="flex items-center">
                    Byes {getSortIcon('byes')}
                  </div>
                </th>
                <th 
                  className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider cursor-pointer"
                  onClick={() => handleSort('wins')}
                >
                  <div className="flex items-center">
                    Wins {getSortIcon('wins')}
                  </div>
                </th>
                <th 
                  className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider cursor-pointer"
                  onClick={() => handleSort('status')}
                >
                  <div className="flex items-center">
                    Status {getSortIcon('status')}
                  </div>
                </th>
              </tr>
            </thead>
            <tbody className="bg-white divide-y divide-gray-200">
              {sortedTeams.map(team => (
                <tr key={team.id}>
                  <td className="px-6 py-4 whitespace-nowrap text-sm font-medium text-gray-900">
                    {team.seed}
                  </td>
                  <td className="px-6 py-4 whitespace-nowrap text-sm text-gray-900">
                    {schools[team.schoolId]?.name || 'Unknown School'}
                  </td>
                  <td className="px-6 py-4 whitespace-nowrap text-sm text-gray-900">
                    {team.byes}
                  </td>
                  <td className="px-6 py-4 whitespace-nowrap text-sm text-gray-900">
                    {team.wins}
                  </td>
                  <td className="px-6 py-4 whitespace-nowrap text-sm text-gray-900">
                    {team.eliminated ? (
                      <span className="px-2 py-1 text-xs font-semibold rounded-full bg-red-100 text-red-800">
                        Eliminated
                      </span>
                    ) : (
                      <span className="px-2 py-1 text-xs font-semibold rounded-full bg-green-100 text-green-800">
                        Active
                      </span>
                    )}
                  </td>
                </tr>
              ))}
            </tbody>
          </table>
        </Card>
      )}
    </PageContainer>
  );
};