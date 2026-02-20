import { useQuery } from '@tanstack/react-query';
import { Link, useNavigate } from 'react-router-dom';
import { Button } from '../components/ui/Button';
import { Card } from '../components/ui/Card';
import { EmptyState } from '../components/ui/EmptyState';
import { ErrorState } from '../components/ui/ErrorState';
import { PageContainer, PageHeader } from '../components/ui/Page';
import { CalcuttaListSkeleton } from '../components/skeletons/CalcuttaListSkeleton';
import { CalcuttaWithRanking } from '../types/calcutta';
import { calcuttaService } from '../services/calcuttaService';
import { formatDate } from '../utils/format';
import { useUser } from '../contexts/useUser';
import { queryKeys } from '../queryKeys';

export function CalcuttaListPage() {
  const navigate = useNavigate();
  const { user } = useUser();

  const calcuttasQuery = useQuery({
    queryKey: queryKeys.calcuttas.listWithRankings(user?.id),
    queryFn: async () => {
      return calcuttaService.getCalcuttasWithRankings();
    },
  });

  if (calcuttasQuery.isLoading) {
    return (
      <PageContainer>
        <PageHeader
          title="Calcuttas"
          actions={
            <Button onClick={() => navigate('/calcuttas/create')}>Create New Calcutta</Button>
          }
        />
        <CalcuttaListSkeleton />
      </PageContainer>
    );
  }

  if (calcuttasQuery.isError) {
    return (
      <PageContainer>
        <ErrorState error={calcuttasQuery.error} onRetry={() => calcuttasQuery.refetch()} />
      </PageContainer>
    );
  }

  const calcuttas: CalcuttaWithRanking[] = calcuttasQuery.data || [];

  return (
    <PageContainer>
      <PageHeader
        title="Calcuttas"
        actions={
          <Button onClick={() => navigate('/calcuttas/create')}>Create New Calcutta</Button>
        }
      />

      <div className="grid gap-4">
        {calcuttas.map((calcutta) => {
          const ranking = calcutta.ranking;
          return (
            <Link
              key={calcutta.id}
              to={`/calcuttas/${calcutta.id}`}
              className="block"
            >
              <Card className="hover:shadow-md transition-shadow">
                <div className="flex items-center justify-between">
                  <div>
                    <h2 className="text-xl font-semibold">{calcutta.name}</h2>
                    <p className="text-gray-600">
                      {calcutta.tournamentStartingAt
                        ? `Tournament starts ${formatDate(calcutta.tournamentStartingAt)}`
                        : `Created ${formatDate(calcutta.created)}`}
                    </p>
                  </div>
                  {(() => {
                    const tournamentStarted = calcutta.tournamentStartingAt
                      ? new Date(calcutta.tournamentStartingAt) <= new Date()
                      : false;

                    if (!tournamentStarted) {
                      return (
                        <div className="text-right">
                          <div className={`text-sm font-medium ${calcutta.hasEntry ? 'text-green-600' : 'text-amber-600'}`}>
                            {calcutta.hasEntry ? 'Entry Submitted' : 'Awaiting Entry'}
                          </div>
                          {calcutta.tournamentStartingAt && (
                            <div className="text-sm text-gray-500">
                              Tournament starts {formatDate(calcutta.tournamentStartingAt, true)}
                            </div>
                          )}
                        </div>
                      );
                    }

                    if (ranking) {
                      return (
                        <div className="text-right">
                          <div className="text-lg font-semibold text-blue-600">
                            #{ranking.rank} of {ranking.totalEntries}
                          </div>
                          <div className="text-sm text-gray-500">
                            {ranking.points.toFixed(2)} points
                          </div>
                        </div>
                      );
                    }

                    return (
                      <div className="text-right">
                        <div className="text-sm font-medium text-gray-400">No Entry</div>
                        <div className="text-sm text-gray-500">0.00 points</div>
                      </div>
                    );
                  })()}
                </div>
              </Card>
            </Link>
          );
        })}

        {calcuttas.length === 0 ? (
          <EmptyState
            title="Welcome to Calcutta!"
            description="Create your first pool or wait for an invitation from a commissioner."
            action={
              <div className="flex gap-3">
                <Link to="/rules">
                  <Button variant="outline">Learn the Rules</Button>
                </Link>
                <Link to="/calcuttas/create">
                  <Button>Create a Pool</Button>
                </Link>
              </div>
            }
          />
        ) : null}
      </div>
    </PageContainer>
  );
}
