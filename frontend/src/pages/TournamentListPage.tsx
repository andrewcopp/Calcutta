import React from 'react';
import { useQuery } from '@tanstack/react-query';
import { Link, useNavigate } from 'react-router-dom';
import { Tournament } from '../types/calcutta';
import { tournamentService } from '../services/tournamentService';
import { queryKeys } from '../queryKeys';
import { ErrorState } from '../components/ui/ErrorState';
import { Breadcrumb } from '../components/ui/Breadcrumb';
import { Button } from '../components/ui/Button';
import { Card } from '../components/ui/Card';
import { LoadingState } from '../components/ui/LoadingState';
import { PageContainer, PageHeader } from '../components/ui/Page';
import { formatDate } from '../utils/format';

type TournamentStatus = 'Complete' | 'In Progress';

interface TournamentWithStatus extends Tournament {
  status: TournamentStatus;
  totalTeams: number;
  eliminatedTeams: number;
}

export const TournamentListPage: React.FC = () => {
  const navigate = useNavigate();
  const tournamentsQuery = useQuery({
    queryKey: queryKeys.tournaments.all(),
    queryFn: async (): Promise<TournamentWithStatus[]> => {
      const data = await tournamentService.getAllTournaments();

      const tournamentsWithStatus = await Promise.all(
        data.map(async (tournament) => {
          const teams = await tournamentService.getTournamentTeams(tournament.id);
          const eliminatedTeams = teams.filter((team) => team.eliminated).length;
          const totalTeams = teams.length;

          const status: TournamentStatus = totalTeams - eliminatedTeams <= 1 ? 'Complete' : 'In Progress';

          return {
            ...tournament,
            status,
            totalTeams,
            eliminatedTeams,
          };
        })
      );

      return tournamentsWithStatus;
    },
  });

  const tournaments = tournamentsQuery.data || [];

  if (tournamentsQuery.isError) {
    return (
      <PageContainer>
        <ErrorState error={tournamentsQuery.error} onRetry={() => tournamentsQuery.refetch()} />
      </PageContainer>
    );
  }

  return (
    <PageContainer>
      <Breadcrumb
        items={[
          { label: 'Admin', href: '/admin' },
          { label: 'Tournaments' },
        ]}
      />
      <PageHeader
        title="Tournaments"
        actions={
          <Button onClick={() => navigate('/admin/tournaments/create')}>Create Tournament</Button>
        }
      />

      {tournamentsQuery.isLoading ? <LoadingState label="Loading tournaments..." /> : null}

      {!tournamentsQuery.isLoading ? (
        <div className="grid gap-4">
          {tournaments.map((tournament) => (
            <Link key={tournament.id} to={`/admin/tournaments/${tournament.id}`} className="block">
              <Card className="hover:shadow-md transition-shadow">
                <div className="flex justify-between items-start">
                  <div>
                    <h2 className="text-xl font-semibold mb-2">{tournament.name}</h2>
                    <p className="text-gray-600">
                      {tournament.rounds} rounds • Created {formatDate(tournament.created)}
                    </p>
                  </div>
                  <div className="flex items-center gap-4">
                    <div
                      className={`px-3 py-1 rounded-full text-sm font-medium ${
                        tournament.status === 'Complete'
                          ? 'bg-green-100 text-green-800'
                          : 'bg-yellow-100 text-yellow-800'
                      }`}
                    >
                      {tournament.status}
                    </div>
                    <Button
                      variant="ghost"
                      size="sm"
                      onClick={(e) => {
                        e.preventDefault();
                        e.stopPropagation();
                        navigate(`/admin/tournaments/${tournament.id}/teams/setup`);
                      }}
                    >
                      Edit
                    </Button>
                  </div>
                </div>
              </Card>
            </Link>
          ))}

          {tournaments.length === 0 ? (
            <div className="text-center py-8 text-gray-500">
              No tournaments found. Create your first tournament to get started.
            </div>
          ) : null}
        </div>
      ) : null}
    </PageContainer>
  );
 };