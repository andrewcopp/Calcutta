import { useState } from 'react';
import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query';
import { useNavigate } from 'react-router-dom';
import { Tournament } from '../types/calcutta';
import { calcuttaService } from '../services/calcuttaService';
import { tournamentService } from '../services/tournamentService';
import { queryKeys } from '../queryKeys';
import { Alert } from '../components/ui/Alert';
import { Button } from '../components/ui/Button';
import { Card } from '../components/ui/Card';
import { Input } from '../components/ui/Input';
import { LoadingState } from '../components/ui/LoadingState';
import { PageContainer, PageHeader } from '../components/ui/Page';
import { Select } from '../components/ui/Select';

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
      <PageContainer>
        <LoadingState label="Loading tournaments..." />
      </PageContainer>
    );
  }

  if (tournamentsQuery.isError) {
    const message = tournamentsQuery.error instanceof Error ? tournamentsQuery.error.message : 'Failed to fetch tournaments';
    return (
      <PageContainer>
        <Alert variant="error">{message}</Alert>
      </PageContainer>
    );
  }

  const tournaments: Tournament[] = tournamentsQuery.data || [];

  return (
    <PageContainer>
      <div className="max-w-2xl mx-auto">
        <PageHeader
          title="Create New Calcutta"
          actions={
            <Button variant="ghost" onClick={() => navigate('/calcuttas')}>
              ← Back to Calcuttas
            </Button>
          }
        />

        {error ? <Alert variant="error" className="mb-4">{error}</Alert> : null}

        <Card>
          <form onSubmit={handleCreateCalcutta}>
            <div className="space-y-6">
              <div>
                <label htmlFor="name" className="block text-sm font-medium text-gray-700 mb-1">
                  Calcutta Name
                </label>
                <Input
                  type="text"
                  id="name"
                  value={newCalcutta.name}
                  onChange={(e) => setNewCalcutta({ ...newCalcutta, name: e.target.value })}
                  placeholder="Enter a name for your Calcutta"
                  required
                />
              </div>

              <div>
                <label htmlFor="tournament" className="block text-sm font-medium text-gray-700 mb-1">
                  Tournament
                </label>
                <Select
                  id="tournament"
                  value={newCalcutta.tournamentId}
                  onChange={(e) => setNewCalcutta({ ...newCalcutta, tournamentId: e.target.value })}
                  required
                >
                  <option value="">Select a tournament</option>
                  {tournaments.map((tournament) => (
                    <option key={tournament.id} value={tournament.id}>
                      {tournament.name}
                    </option>
                  ))}
                </Select>
                <p className="mt-1 text-sm text-gray-500">Select the tournament this Calcutta will be based on</p>
              </div>

              <div className="pt-2">
                <Button type="submit" className="w-full" loading={createCalcuttaMutation.isPending}>
                  Create Calcutta
                </Button>
              </div>
            </div>
          </form>
        </Card>
      </div>
    </PageContainer>
  );
 }
 