import React, { useState } from 'react';
import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query';
import { useNavigate } from 'react-router-dom';
import { School } from '../types/school';
import { tournamentService } from '../services/tournamentService';
import { schoolService } from '../services/schoolService';
import { queryKeys } from '../queryKeys';
import { Alert } from '../components/ui/Alert';
import { Breadcrumb } from '../components/ui/Breadcrumb';
import { Button } from '../components/ui/Button';
import { Card } from '../components/ui/Card';
import { ErrorState } from '../components/ui/ErrorState';
import { Input } from '../components/ui/Input';
import { LoadingState } from '../components/ui/LoadingState';
import { PageContainer, PageHeader } from '../components/ui/Page';
import { Select } from '../components/ui/Select';

interface TeamToAdd {
  schoolId: string;
  seed: number;
  region: string;
}

export const TournamentCreatePage: React.FC = () => {
  const navigate = useNavigate();
  const queryClient = useQueryClient();
  const [name, setName] = useState('');
  const [rounds, setRounds] = useState(6);
  const [selectedSchool, setSelectedSchool] = useState<string>('');
  const [selectedSeed, setSelectedSeed] = useState<number>(1);
  const [teamsToAdd, setTeamsToAdd] = useState<TeamToAdd[]>([]);
  const [searchTerm, setSearchTerm] = useState('');
  const [error, setError] = useState<string | null>(null);

  const schoolsQuery = useQuery({
    queryKey: queryKeys.schools.all(),
    queryFn: () => schoolService.getSchools(),
  });

  const createTournamentMutation = useMutation({
    mutationFn: async ({ name, rounds, teamsToAdd }: { name: string; rounds: number; teamsToAdd: TeamToAdd[] }) => {
      const tournament = await tournamentService.createTournament(name, rounds);

      await Promise.all(
        teamsToAdd.map((team) => tournamentService.createTournamentTeam(tournament.id, team.schoolId, team.seed, team.region))
      );

      return tournament;
    },
    onSuccess: async (tournament) => {
      await queryClient.invalidateQueries({ queryKey: queryKeys.tournaments.all() });
      await queryClient.invalidateQueries({ queryKey: queryKeys.tournaments.teams(tournament.id) });
      navigate(`/admin/tournaments/${tournament.id}`);
    },
    onError: () => {
      setError('Failed to create tournament');
    },
  });

  const schools: School[] = schoolsQuery.data || [];

  const filteredSchools = schools.filter((school) =>
    school.name.toLowerCase().includes(searchTerm.toLowerCase())
  );

  const handleAddTeam = () => {
    if (!selectedSchool) {
      setError('Please select a school');
      return;
    }

    if (teamsToAdd.length >= 68) {
      setError('Maximum number of teams (68) reached');
      return;
    }

    const newTeam: TeamToAdd = {
      schoolId: selectedSchool,
      seed: selectedSeed,
      region: 'Unknown', // Default region
    };

    setTeamsToAdd([...teamsToAdd, newTeam]);
    setSelectedSchool('');
    setSearchTerm('');
    setSelectedSeed(1);
  };

  const handleRemoveTeam = (index: number) => {
    const newTeams = [...teamsToAdd];
    newTeams.splice(index, 1);
    setTeamsToAdd(newTeams);
  };

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    setError(null);

    if (!name.trim()) {
      setError('Please enter a tournament name');
      return;
    }

    if (teamsToAdd.length === 0) {
      setError('Please add at least one team');
      return;
    }

    createTournamentMutation.mutate({ name, rounds, teamsToAdd });
  };

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

      {schoolsQuery.isError && (
        <ErrorState
          error={schoolsQuery.error instanceof Error ? schoolsQuery.error.message : 'Failed to load schools'}
          onRetry={() => schoolsQuery.refetch()}
        />
      )}

      {schoolsQuery.isLoading && <LoadingState label="Loading schools..." layout="inline" />}

      {!schoolsQuery.isLoading && !schoolsQuery.isError && schools.length === 0 ? (
        <Alert variant="info" className="mb-4">
          No schools are available.
        </Alert>
      ) : null}

      <form onSubmit={handleSubmit} className="space-y-8">
        <Card>
          <h2 className="text-xl font-semibold mb-4">Tournament Details</h2>
          <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
            <div>
              <label className="block text-sm font-medium text-gray-700 mb-1">
                Tournament Name
              </label>
              <Input
                type="text"
                value={name}
                onChange={(e) => setName(e.target.value)}
                required
              />
            </div>
            <div>
              <label className="block text-sm font-medium text-gray-700 mb-1">
                Number of Rounds
              </label>
              <Input
                type="number"
                value={rounds}
                onChange={(e) => setRounds(parseInt(e.target.value))}
                min={1}
                max={7}
                required
              />
            </div>
          </div>
        </Card>

        <Card>
          <h2 className="text-xl font-semibold mb-4">Add Teams</h2>
          <div className="flex gap-4 mb-4">
            <div className="flex-1">
              <Input
                type="text"
                placeholder="Search schools..."
                value={searchTerm}
                onChange={(e) => setSearchTerm(e.target.value)}
              />
            </div>
            <div className="flex-1">
              <Select
                value={selectedSchool}
                onChange={(e) => setSelectedSchool(e.target.value)}
              >
                <option value="">Select a school</option>
                {filteredSchools.map(school => (
                  <option key={school.id} value={school.id}>
                    {school.name}
                  </option>
                ))}
              </Select>
            </div>
            <div className="w-32">
              <Input
                type="number"
                min={1}
                max={68}
                value={selectedSeed}
                onChange={(e) => setSelectedSeed(parseInt(e.target.value) || 1)}
                placeholder="Seed"
              />
            </div>
            <Button
              type="button"
              onClick={handleAddTeam}
            >
              Add Team
            </Button>
          </div>

          <div className="space-y-2">
            <h3 className="text-lg font-semibold mb-2">Teams to Add ({teamsToAdd.length}/68)</h3>
            {teamsToAdd.map((team, index) => {
              const school = schools.find(s => s.id === team.schoolId);
              return (
                <div
                  key={index}
                  className="flex items-center justify-between p-2 border rounded"
                >
                  <span>
                    {school?.name || 'Unknown School'} (Seed {team.seed})
                  </span>
                  <Button
                    type="button"
                    variant="ghost"
                    size="sm"
                    onClick={() => handleRemoveTeam(index)}
                    className="text-red-500 hover:text-red-700 hover:bg-red-50"
                  >
                    Remove
                  </Button>
                </div>
              );
            })}
          </div>
        </Card>

        <div className="flex justify-end">
          <Button
            type="submit"
            disabled={createTournamentMutation.isPending || schoolsQuery.isLoading || teamsToAdd.length === 0}
            loading={createTournamentMutation.isPending}
          >
            {createTournamentMutation.isPending ? 'Creating Tournament...' : 'Create Tournament'}
          </Button>
        </div>
      </form>
    </PageContainer>
  );
};