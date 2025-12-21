import { useState } from 'react';
import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query';
import { useNavigate } from 'react-router-dom';
import { Tournament } from '../types/calcutta';
import { calcuttaService } from '../services/calcuttaService';
import { tournamentService } from '../services/tournamentService';
import { queryKeys } from '../queryKeys';

export function CreateCalcuttaPage() {
  const navigate = useNavigate();
  const queryClient = useQueryClient();
  const [error, setError] = useState<string | null>(null);
  const [newCalcutta, setNewCalcutta] = useState({
    name: '',
    tournamentId: '',
  });

  const tournamentsQuery = useQuery({
    queryKey: queryKeys.tournaments.all(),
    staleTime: 30_000,
    queryFn: () => tournamentService.getAllTournaments(),
  });

  const createCalcuttaMutation = useMutation({
    mutationFn: async ({ name, tournamentId }: { name: string; tournamentId: string }) => {
      const sysAdminId = '090644de-1158-402e-a103-949b089d8cf9';
      return calcuttaService.createCalcutta(name, tournamentId, sysAdminId);
    },
    onSuccess: async () => {
      await queryClient.invalidateQueries({ queryKey: queryKeys.calcuttas.all() });
      navigate('/calcuttas');
    },
    onError: () => {
      setError('Failed to create calcutta');
    },
  });

  const handleCreateCalcutta = async (e: React.FormEvent) => {
    e.preventDefault();
    setError(null);

    createCalcuttaMutation.mutate({
      name: newCalcutta.name,
      tournamentId: newCalcutta.tournamentId,
    });
  };

  if (tournamentsQuery.isLoading) {
    return (
      <div className="container mx-auto px-4 py-8">
        <div className="text-center">Loading tournaments...</div>
      </div>
    );
  }

  if (tournamentsQuery.isError) {
    const message = tournamentsQuery.error instanceof Error ? tournamentsQuery.error.message : 'Failed to fetch tournaments';
    return (
      <div className="container mx-auto px-4 py-8">
        <div className="bg-red-100 border border-red-400 text-red-700 px-4 py-3 rounded">
          {message}
        </div>
      </div>
    );
  }

  const tournaments: Tournament[] = tournamentsQuery.data || [];

  return (
    <div className="container mx-auto px-4 py-8">
      <div className="max-w-2xl mx-auto">
        <div className="flex items-center mb-6">
          <button
            onClick={() => navigate('/calcuttas')}
            className="text-blue-500 hover:text-blue-700 mr-4"
          >
            ← Back to Calcuttas
          </button>
          <h1 className="text-3xl font-bold">Create New Calcutta</h1>
        </div>

        {error && (
          <div className="bg-red-100 border border-red-400 text-red-700 px-4 py-3 rounded mb-4">
            {error}
          </div>
        )}

        <form onSubmit={handleCreateCalcutta} className="bg-white p-6 rounded-lg shadow">
          <div className="space-y-6">
            <div>
              <label htmlFor="name" className="block text-sm font-medium text-gray-700">
                Calcutta Name
              </label>
              <input
                type="text"
                id="name"
                value={newCalcutta.name}
                onChange={(e) => setNewCalcutta({ ...newCalcutta, name: e.target.value })}
                className="mt-1 block w-full rounded-md border-gray-300 shadow-sm focus:border-blue-500 focus:ring-blue-500"
                placeholder="Enter a name for your Calcutta"
                required
              />
            </div>
            
            <div>
              <label htmlFor="tournament" className="block text-sm font-medium text-gray-700">
                Tournament
              </label>
              <select
                id="tournament"
                value={newCalcutta.tournamentId}
                onChange={(e) => setNewCalcutta({ ...newCalcutta, tournamentId: e.target.value })}
                className="mt-1 block w-full rounded-md border-gray-300 shadow-sm focus:border-blue-500 focus:ring-blue-500"
                required
              >
                <option value="">Select a tournament</option>
                {tournaments.map((tournament) => (
                  <option key={tournament.id} value={tournament.id}>
                    {tournament.name}
                  </option>
                ))}
              </select>
              <p className="mt-1 text-sm text-gray-500">
                Select the tournament this Calcutta will be based on
              </p>
            </div>
            
            <div className="pt-4">
              <button
                type="submit"
                disabled={createCalcuttaMutation.isPending}
                className="w-full bg-blue-500 text-white px-4 py-2 rounded hover:bg-blue-600 disabled:bg-blue-300"
              >
                {createCalcuttaMutation.isPending ? 'Creating...' : 'Create Calcutta'}
              </button>
            </div>
          </div>
        </form>
      </div>
    </div>
  );
} 