import React, { useState } from 'react';
import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query';
import { useNavigate } from 'react-router-dom';
import { tournamentService } from '../services/tournamentService';
import { queryKeys } from '../queryKeys';
import { Alert } from '../components/ui/Alert';
import { Breadcrumb } from '../components/ui/Breadcrumb';
import { Button } from '../components/ui/Button';
import { Card } from '../components/ui/Card';
import { Input } from '../components/ui/Input';
import { LoadingState } from '../components/ui/LoadingState';
import { PageContainer, PageHeader } from '../components/ui/Page';
import { Select } from '../components/ui/Select';

export const TournamentCreatePage: React.FC = () => {
  const navigate = useNavigate();
  const queryClient = useQueryClient();
  const [competition, setCompetition] = useState('');
  const [newCompetition, setNewCompetition] = useState('');
  const [year, setYear] = useState('');
  const [newYear, setNewYear] = useState('');
  const [rounds, setRounds] = useState(7);
  const [startingAt, setStartingAt] = useState('');
  const [regionNames, setRegionNames] = useState({
    topLeft: 'East',
    bottomLeft: 'West',
    topRight: 'South',
    bottomRight: 'Midwest',
  });
  const [error, setError] = useState<string | null>(null);

  const competitionsQuery = useQuery({
    queryKey: ['competitions'],
    queryFn: () => tournamentService.getCompetitions(),
  });

  const seasonsQuery = useQuery({
    queryKey: ['seasons'],
    queryFn: () => tournamentService.getSeasons(),
  });

  const competitions = competitionsQuery.data || [];
  const seasons = seasonsQuery.data || [];

  const isAddNewCompetition = competition === '__new__';
  const isAddNewYear = year === '__new__';

  const resolvedCompetition = isAddNewCompetition ? newCompetition.trim() : competitions.find(c => c.id === competition)?.name || '';
  const resolvedYear = isAddNewYear ? parseInt(newYear) : seasons.find(s => s.id === year)?.year || 0;

  const derivedName = resolvedCompetition && resolvedYear > 0 ? `${resolvedCompetition} (${resolvedYear})` : '';

  const createTournamentMutation = useMutation({
    mutationFn: async () => {
      const tournament = await tournamentService.createTournament(resolvedCompetition, resolvedYear, rounds);
      await tournamentService.updateTournament(tournament.id, {
        finalFourTopLeft: regionNames.topLeft,
        finalFourBottomLeft: regionNames.bottomLeft,
        finalFourTopRight: regionNames.topRight,
        finalFourBottomRight: regionNames.bottomRight,
        ...(startingAt ? { startingAt: new Date(startingAt).toISOString() } : {}),
      });
      return tournament;
    },
    onSuccess: async (tournament) => {
      await queryClient.invalidateQueries({ queryKey: queryKeys.tournaments.all() });
      navigate(`/admin/tournaments/${tournament.id}/teams/setup`);
    },
    onError: (err) => {
      setError(err instanceof Error ? err.message : 'Failed to create tournament');
    },
  });

  const handleSubmit = (e: React.FormEvent) => {
    e.preventDefault();
    setError(null);

    if (!resolvedCompetition) {
      setError('Please select or enter a competition');
      return;
    }
    if (!resolvedYear || resolvedYear <= 2000) {
      setError('Please select or enter a valid year (after 2000)');
      return;
    }
    if (rounds <= 0) {
      setError('Rounds must be at least 1');
      return;
    }

    createTournamentMutation.mutate();
  };

  const isLoading = competitionsQuery.isLoading || seasonsQuery.isLoading;

  return (
    <PageContainer>
      <Breadcrumb
        items={[
          { label: 'Tournaments', href: '/admin/tournaments' },
          { label: 'Create' },
        ]}
      />
      <PageHeader
        title="Create New Tournament"
        actions={
          <Button variant="outline" onClick={() => navigate('/admin/tournaments')}>
            Cancel
          </Button>
        }
      />

      {error && (
        <Alert variant="error" className="mb-4">
          {error}
        </Alert>
      )}

      {isLoading && <LoadingState label="Loading..." layout="inline" />}

      {!isLoading && (
        <form onSubmit={handleSubmit} className="space-y-8">
          <Card>
            <h2 className="text-xl font-semibold mb-4">Tournament Details</h2>
            <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
              <div>
                <label className="block text-sm font-medium text-gray-700 mb-1">
                  Competition
                </label>
                <Select
                  value={competition}
                  onChange={(e) => {
                    setCompetition(e.target.value);
                    if (e.target.value !== '__new__') setNewCompetition('');
                  }}
                >
                  <option value="">Select a competition</option>
                  {competitions.map(c => (
                    <option key={c.id} value={c.id}>{c.name}</option>
                  ))}
                  <option value="__new__">+ Add new...</option>
                </Select>
                {isAddNewCompetition && (
                  <Input
                    type="text"
                    placeholder="e.g. NCAA Men's"
                    value={newCompetition}
                    onChange={(e) => setNewCompetition(e.target.value)}
                    className="mt-2"
                  />
                )}
              </div>
              <div>
                <label className="block text-sm font-medium text-gray-700 mb-1">
                  Year
                </label>
                <Select
                  value={year}
                  onChange={(e) => {
                    setYear(e.target.value);
                    if (e.target.value !== '__new__') setNewYear('');
                  }}
                >
                  <option value="">Select a year</option>
                  {seasons.map(s => (
                    <option key={s.id} value={s.id}>{s.year}</option>
                  ))}
                  <option value="__new__">+ Add new...</option>
                </Select>
                {isAddNewYear && (
                  <Input
                    type="number"
                    placeholder="e.g. 2026"
                    value={newYear}
                    onChange={(e) => setNewYear(e.target.value)}
                    min={2001}
                    className="mt-2"
                  />
                )}
              </div>
              <div>
                <label className="block text-sm font-medium text-gray-700 mb-1">
                  Rounds
                </label>
                <Input
                  type="number"
                  value={rounds}
                  onChange={(e) => setRounds(parseInt(e.target.value) || 7)}
                  min={1}
                  max={7}
                  required
                />
              </div>
              <div>
                <label className="block text-sm font-medium text-gray-700 mb-1">
                  Start Time
                </label>
                <Input
                  type="datetime-local"
                  value={startingAt}
                  onChange={(e) => setStartingAt(e.target.value)}
                />
                <p className="mt-1 text-xs text-gray-500">
                  Used to auto-lock bidding when the tournament starts
                </p>
              </div>
            </div>

            {derivedName && (
              <div className="mt-4 p-3 bg-blue-50 rounded-lg">
                <span className="text-sm text-gray-600">Tournament name: </span>
                <span className="text-sm font-semibold text-blue-700">{derivedName}</span>
              </div>
            )}
          </Card>

          <Card>
            <h2 className="text-xl font-semibold mb-4">Region Names</h2>
            <div className="grid grid-cols-2 gap-4">
              {([
                ['topLeft', 'Top Left'],
                ['bottomLeft', 'Bottom Left'],
                ['topRight', 'Top Right'],
                ['bottomRight', 'Bottom Right'],
              ] as const).map(([key, label]) => (
                <div key={key}>
                  <label className="block text-sm font-medium text-gray-700 mb-1">
                    {label}
                  </label>
                  <Input
                    type="text"
                    value={regionNames[key]}
                    onChange={(e) => setRegionNames(prev => ({ ...prev, [key]: e.target.value }))}
                  />
                </div>
              ))}
            </div>
          </Card>

          <div className="flex justify-end">
            <Button
              type="submit"
              disabled={createTournamentMutation.isPending || !derivedName}
              loading={createTournamentMutation.isPending}
            >
              {createTournamentMutation.isPending ? 'Creating...' : 'Create Tournament'}
            </Button>
          </div>
        </form>
      )}
    </PageContainer>
  );
};
