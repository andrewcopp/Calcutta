import { useQuery } from '@tanstack/react-query';
import { Link, useNavigate } from 'react-router-dom';
import { Alert } from '../components/ui/Alert';
import { Button } from '../components/ui/Button';
import { Card } from '../components/ui/Card';
import { LoadingState } from '../components/ui/LoadingState';
import { PageContainer, PageHeader } from '../components/ui/Page';
import { Calcutta } from '../types/calcutta';
import { calcuttaService } from '../services/calcuttaService';
import { useUser } from '../contexts/useUser';
import { queryKeys } from '../queryKeys';

interface CalcuttaRanking {
  calcuttaId: string;
  rank: number;
  totalEntries: number;
  points: number;
}

export function CalcuttaListPage() {
  const navigate = useNavigate();
  const { user } = useUser();

  const calcuttasQuery = useQuery({
    queryKey: queryKeys.calcuttas.all(user?.id),
    staleTime: 30_000,
    queryFn: async () => {
      const calcuttas = await calcuttaService.getAllCalcuttas();

      const rankings: Record<string, CalcuttaRanking> = {};
      if (user) {
        for (const calcutta of calcuttas) {
          try {
            const entries = await calcuttaService.getCalcuttaEntries(calcutta.id);

            const userEntry = entries.find((entry) => entry.userId === user.id);
            if (!userEntry) continue;

            const sortedEntries = [...entries].sort((a, b) => (b.totalPoints || 0) - (a.totalPoints || 0));
            const userRank = sortedEntries.findIndex((entry) => entry.id === userEntry.id) + 1;

            rankings[calcutta.id] = {
              calcuttaId: calcutta.id,
              rank: userRank,
              totalEntries: entries.length,
              points: userEntry.totalPoints || 0,
            };
          } catch (entryError) {
            console.error(`Error fetching entries for calcutta ${calcutta.id}:`, entryError);
          }
        }
      }

      return { calcuttas, rankings };
    },
  });

  if (calcuttasQuery.isLoading) {
    return (
      <PageContainer>
        <LoadingState label="Loading calcuttas..." />
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

  const calcuttas: Calcutta[] = calcuttasQuery.data?.calcuttas || [];
  const rankings = calcuttasQuery.data?.rankings || {};

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
          const ranking = rankings[calcutta.id];
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
          <Card className="text-center py-8">
            <p className="text-gray-500 mb-4">No calcuttas found.</p>
            <Link to="/calcuttas/create" className="text-blue-500 hover:text-blue-700">
              Create your first Calcutta
            </Link>
          </Card>
        ) : null}
      </div>
    </PageContainer>
  );
}