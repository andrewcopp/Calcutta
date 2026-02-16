import { useQuery } from '@tanstack/react-query';
import { Link, useNavigate } from 'react-router-dom';
import { Alert } from '../components/ui/Alert';
import { Button } from '../components/ui/Button';
import { Card } from '../components/ui/Card';
import { EmptyState } from '../components/ui/EmptyState';
import { PageContainer, PageHeader } from '../components/ui/Page';
import { CalcuttaListSkeleton } from '../components/skeletons/CalcuttaListSkeleton';
import { CalcuttaWithRanking } from '../types/calcutta';
import { calcuttaService } from '../services/calcuttaService';
import { useUser } from '../contexts/useUser';
import { queryKeys } from '../queryKeys';

export function CalcuttaListPage() {
  const navigate = useNavigate();
  const { user } = useUser();

  const calcuttasQuery = useQuery({
    queryKey: queryKeys.calcuttas.listWithRankings(user?.id),
    staleTime: 30_000,
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
    const message = calcuttasQuery.error instanceof Error ? calcuttasQuery.error.message : 'Failed to fetch calcuttas';
    return (
      <PageContainer>
        <Alert variant="error">
          <h2 className="text-lg font-semibold mb-2">Error</h2>
          <p>{message}</p>
          <div className="mt-4">
            <Button onClick={() => calcuttasQuery.refetch()}>Retry</Button>
          </div>
        </Alert>
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
                    <p className="text-gray-600">Created: {new Date(calcutta.created).toLocaleDateString()}</p>
                  </div>
                  {ranking ? (
                    <div className="text-right">
                      <div className="text-lg font-semibold text-blue-600">
                        #{ranking.rank} of {ranking.totalEntries}
                      </div>
                      <div className="text-sm text-gray-500">{ranking.points} points</div>
                    </div>
                  ) : null}
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
