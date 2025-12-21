import { useQuery } from '@tanstack/react-query';
import { Link } from 'react-router-dom';
import { Calcutta } from '../types/calcutta';
import { calcuttaService } from '../services/calcuttaService';
import { useUser } from '../contexts/UserContext';
import { queryKeys } from '../queryKeys';

interface CalcuttaRanking {
  calcuttaId: string;
  rank: number;
  totalEntries: number;
  points: number;
}

export function CalcuttaListPage() {
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
    return <div>Loading...</div>;
  }

  if (calcuttasQuery.isError) {
    const message = calcuttasQuery.error instanceof Error ? calcuttasQuery.error.message : 'Failed to fetch calcuttas';
    return (
      <div className="error p-4 bg-red-100 text-red-700 rounded">
        <h2 className="text-lg font-semibold mb-2">Error</h2>
        <p>{message}</p>
        <button 
          onClick={() => calcuttasQuery.refetch()} 
          className="mt-4 px-4 py-2 bg-red-600 text-white rounded hover:bg-red-700"
        >
          Retry
        </button>
      </div>
    );
  }

  const calcuttas: Calcutta[] = calcuttasQuery.data?.calcuttas || [];
  const rankings = calcuttasQuery.data?.rankings || {};

  return (
    <div className="container mx-auto px-4 py-8">
      <div className="flex justify-between items-center mb-6">
        <h1 className="text-3xl font-bold">Calcuttas</h1>
        <Link
          to="/calcuttas/create"
          className="bg-blue-500 text-white px-4 py-2 rounded hover:bg-blue-600"
        >
          Create New Calcutta
        </Link>
      </div>

      <div className="grid gap-4">
        {calcuttas.map((calcutta) => {
          const ranking = rankings[calcutta.id];
          return (
            <Link
              key={calcutta.id}
              to={`/calcuttas/${calcutta.id}`}
              className="block p-4 bg-white rounded-lg shadow hover:shadow-md transition-shadow"
            >
              <div className="flex items-center justify-between">
                <div>
                  <h2 className="text-xl font-semibold">{calcutta.name}</h2>
                  <p className="text-gray-600">Created: {new Date(calcutta.created).toLocaleDateString()}</p>
                </div>
                {ranking && (
                  <div className="text-right">
                    <div className="text-lg font-semibold text-blue-600">
                      #{ranking.rank} of {ranking.totalEntries}
                    </div>
                    <div className="text-sm text-gray-500">
                      {ranking.points} points
                    </div>
                  </div>
                )}
              </div>
            </Link>
          );
        })}
        
        {calcuttas.length === 0 && (
          <div className="text-center py-8 bg-white rounded-lg shadow">
            <p className="text-gray-500 mb-4">No calcuttas found.</p>
            <Link
              to="/calcuttas/create"
              className="text-blue-500 hover:text-blue-700"
            >
              Create your first Calcutta
            </Link>
          </div>
        )}
      </div>
    </div>
  );
} 