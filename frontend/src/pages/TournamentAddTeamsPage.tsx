import React, { useState } from 'react';
import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query';
import { useParams, useNavigate } from 'react-router-dom';
import { School } from '../types/school';
import { Tournament } from '../types/calcutta';
import { tournamentService } from '../services/tournamentService';
import { schoolService } from '../services/schoolService';
import { queryKeys } from '../queryKeys';
import { Alert } from '../components/ui/Alert';
import { Breadcrumb } from '../components/ui/Breadcrumb';
import { Button } from '../components/ui/Button';
import { Card } from '../components/ui/Card';
import { Input } from '../components/ui/Input';
import { LoadingState } from '../components/ui/LoadingState';
import { PageContainer, PageHeader } from '../components/ui/Page';
import { Select } from '../components/ui/Select';
import { formatDate } from '../utils/format';

interface TeamToAdd {
  schoolId: string;
  seed: number;
  region: string;
}

export const TournamentAddTeamsPage: React.FC = () => {
  const { id } = useParams<{ id: string }>();
  const navigate = useNavigate();
  const queryClient = useQueryClient();
  const [selectedSchool, setSelectedSchool] = useState<string>('');
  const [selectedSeed, setSelectedSeed] = useState<number>(1);
  const [selectedRegion, setSelectedRegion] = useState<string>('East');
  const [teamsToAdd, setTeamsToAdd] = useState<TeamToAdd[]>([]);
  const [searchTerm, setSearchTerm] = useState('');
  const [error, setError] = useState<string | null>(null);

  const tournamentQuery = useQuery({
    queryKey: queryKeys.tournaments.detail(id),
    enabled: Boolean(id),
    queryFn: () => tournamentService.getTournament(id!),
  });

  const schoolsQuery = useQuery({
    queryKey: queryKeys.schools.all(),
    queryFn: () => schoolService.getSchools(),
  });

  const createTeamsMutation = useMutation({
    mutationFn: async ({ teamsToAdd }: { teamsToAdd: TeamToAdd[] }) => {
      await Promise.all(
        teamsToAdd.map((team) => tournamentService.createTournamentTeam(id!, team.schoolId, team.seed, team.region))
      );
    },
    onSuccess: async () => {
      await queryClient.invalidateQueries({ queryKey: queryKeys.tournaments.teams(id) });
      navigate(`/admin/tournaments/${id}`);
    },
    onError: (error) => {
      setError(error instanceof Error ? error.message : 'Failed to create teams');
    },
  });

  const tournament: Tournament | null = tournamentQuery.data || null;
  const schools: School[] = schoolsQuery.data || [];

  const filteredSchools = schools.filter((school) => school.name.toLowerCase().includes(searchTerm.toLowerCase()));

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
      region: selectedRegion,
    };

    setTeamsToAdd([...teamsToAdd, newTeam]);
    setSelectedSchool('');
    setSearchTerm('');
    setSelectedSeed(1);
    setSelectedRegion('East');
  };

  const handleRemoveTeam = (index: number) => {
    const newTeams = [...teamsToAdd];
    newTeams.splice(index, 1);
    setTeamsToAdd(newTeams);
  };

  const handleSubmit = async () => {
    if (teamsToAdd.length === 0) {
      setError('Please add at least one team');
      return;
    }
    setError(null);

    createTeamsMutation.mutate({ teamsToAdd });
  };

  if (!id) {
    return (
      <PageContainer>
        <Alert variant="error">Missing required parameters</Alert>
      </PageContainer>
    );
  }

  if (tournamentQuery.isLoading || schoolsQuery.isLoading) {
    return (
      <PageContainer>
        <LoadingState label="Loading..." />
      </PageContainer>
    );
  }

  if (tournamentQuery.isError || schoolsQuery.isError) {
    const message =
      (tournamentQuery.error instanceof Error ? tournamentQuery.error.message : null) ||
      (schoolsQuery.error instanceof Error ? schoolsQuery.error.message : null) ||
      'Failed to load data';

    return (
      <PageContainer>
        <Alert variant="error">{message}</Alert>
      </PageContainer>
    );
  }

  if (!tournament) {
    return (
      <PageContainer>
        <LoadingState label="Loading..." />
      </PageContainer>
    );
  }

  return (
    <PageContainer>
      <Breadcrumb
        items={[
          { label: 'Tournaments', href: '/admin/tournaments' },
          { label: tournament.name, href: `/admin/tournaments/${id}` },
          { label: 'Add Teams' },
        ]}
      />
      <PageHeader
        title={
          <span>
            Add Teams to {tournament.name}
          </span>
        }
        subtitle={`${tournament.rounds} rounds • Created ${formatDate(tournament.created)}`}
        actions={
          <Button variant="ghost" onClick={() => navigate(`/admin/tournaments/${id}`)}>
            Cancel
          </Button>
        }
      />

      {error ? <Alert variant="error">{error}</Alert> : null}

      <Card>
        <h2 className="text-xl font-semibold mb-4">Add Teams</h2>
        <div className="flex flex-col gap-4 mb-4 md:flex-row md:items-end">
          <div className="flex-1">
            <Input
              type="text"
              placeholder="Search schools..."
              value={searchTerm}
              onChange={(e) => setSearchTerm(e.target.value)}
            />
          </div>
          <div className="flex-1">
            <Select value={selectedSchool} onChange={(e) => setSelectedSchool(e.target.value)}>
              <option value="">Select a school</option>
              {filteredSchools.map((school) => (
                <option key={school.id} value={school.id}>
                  {school.name}
                </option>
              ))}
            </Select>
          </div>
          <div className="w-full md:w-32">
            <Input
              type="number"
              min={1}
              max={16}
              value={selectedSeed}
              onChange={(e) => setSelectedSeed(parseInt(e.target.value) || 1)}
              placeholder="Seed"
            />
          </div>
          <div className="w-full md:w-36">
            <Select value={selectedRegion} onChange={(e) => setSelectedRegion(e.target.value)}>
              <option value="East">East</option>
              <option value="West">West</option>
              <option value="South">South</option>
              <option value="Midwest">Midwest</option>
            </Select>
          </div>
          <Button type="button" onClick={handleAddTeam}>
            Add Team
          </Button>
        </div>

        <div className="space-y-2">
          <h3 className="text-lg font-semibold mb-2">Teams to Add ({teamsToAdd.length}/68)</h3>
          {teamsToAdd.map((team, index) => {
            const school = schools.find((s) => s.id === team.schoolId);
            return (
              <div key={index} className="flex items-center justify-between gap-4 p-2 border rounded-lg">
                <span>
                  {school?.name || 'Unknown School'} (Seed {team.seed}, {team.region})
                </span>
                <Button type="button" variant="ghost" size="sm" onClick={() => handleRemoveTeam(index)}>
                  Remove
                </Button>
              </div>
            );
          })}
        </div>
      </Card>

      <div className="flex justify-end mt-8">
        <Button
          onClick={handleSubmit}
          loading={createTeamsMutation.isPending}
          disabled={teamsToAdd.length === 0}
        >
          Create Teams
        </Button>
      </div>
    </PageContainer>
  );
 };