import React, { useState } from 'react';
import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query';
import { useParams, useNavigate, Link } from 'react-router-dom';
import { School } from '../types/school';
import { Tournament } from '../types/calcutta';
import { adminService } from '../services/adminService';
import { tournamentService } from '../services/tournamentService';
import { queryKeys } from '../queryKeys';
import { Alert } from '../components/ui/Alert';

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
  const [teamsToAdd, setTeamsToAdd] = useState<TeamToAdd[]>([]);
  const [searchTerm, setSearchTerm] = useState('');
  const [error, setError] = useState<string | null>(null);

  const tournamentQuery = useQuery({
    queryKey: queryKeys.tournaments.detail(id),
    enabled: Boolean(id),
    staleTime: 30_000,
    queryFn: () => tournamentService.getTournament(id!),
  });

  const schoolsQuery = useQuery({
    queryKey: queryKeys.schools.all(),
    staleTime: 30_000,
    queryFn: () => adminService.getAllSchools(),
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
    onError: () => {
      setError('Failed to create teams');
    },
  });

  const tournament: Tournament | null = tournamentQuery.data || null;
  const schools: School[] = schoolsQuery.data || [];

  const filteredSchools = schools.filter(school =>
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

  const handleSubmit = async () => {
    if (teamsToAdd.length === 0) {
      setError('Please add at least one team');
      return;
    }
    setError(null);

    createTeamsMutation.mutate({ teamsToAdd });
  };

  if (!id) {
    return <Alert variant="error">Missing required parameters</Alert>;
  }

  if (tournamentQuery.isLoading || schoolsQuery.isLoading) {
    return (
      <div className="container mx-auto px-4 py-8">
        <div className="text-center">Loading...</div>
      </div>
    );
  }

  if (tournamentQuery.isError || schoolsQuery.isError) {
    const message =
      (tournamentQuery.error instanceof Error ? tournamentQuery.error.message : null) ||
      (schoolsQuery.error instanceof Error ? schoolsQuery.error.message : null) ||
      'Failed to load data';

    return (
      <div className="container mx-auto px-4 py-8">
        <div className="bg-red-100 border border-red-400 text-red-700 px-4 py-3 rounded">
          {message}
        </div>
      </div>
    );
  }

  if (!tournament) {
    return (
      <div className="container mx-auto px-4 py-8">
        <div className="text-center">Loading...</div>
      </div>
    );
  }

  return (
    <div className="container mx-auto px-4 py-8">
      <div className="flex justify-between items-center mb-8">
        <div>
          <h1 className="text-3xl font-bold mb-2">Add Teams to {tournament.name}</h1>
          <p className="text-gray-600">
            {tournament.rounds} rounds • Created {new Date(tournament.created).toLocaleDateString()}
          </p>
        </div>
        <button
          onClick={() => navigate(`/admin/tournaments/${id}`)}
          className="px-4 py-2 border rounded hover:bg-gray-100"
        >
          Cancel
        </button>
      </div>

      {error && (
        <div className="bg-red-100 border border-red-400 text-red-700 px-4 py-3 rounded mb-4">
          {error}
        </div>
      )}

      <div className="bg-white p-6 rounded-lg shadow">
        <h2 className="text-xl font-semibold mb-4">Add Teams</h2>
        <div className="flex gap-4 mb-4">
          <div className="flex-1">
            <input
              type="text"
              placeholder="Search schools..."
              value={searchTerm}
              onChange={(e) => setSearchTerm(e.target.value)}
              className="w-full p-2 border rounded"
            />
          </div>
          <div className="flex-1">
            <select
              value={selectedSchool}
              onChange={(e) => setSelectedSchool(e.target.value)}
              className="w-full p-2 border rounded"
            >
              <option value="">Select a school</option>
              {filteredSchools.map(school => (
                <option key={school.id} value={school.id}>
                  {school.name}
                </option>
              ))}
            </select>
          </div>
          <div className="w-32">
            <input
              type="number"
              min="1"
              max="16"
              value={selectedSeed}
              onChange={(e) => setSelectedSeed(parseInt(e.target.value) || 1)}
              className="w-full p-2 border rounded"
              placeholder="Seed"
            />
          </div>
          <button
            type="button"
            onClick={handleAddTeam}
            className="bg-blue-500 text-white px-4 py-2 rounded hover:bg-blue-600"
          >
            Add Team
          </button>
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
                <button
                  type="button"
                  onClick={() => handleRemoveTeam(index)}
                  className="text-red-500 hover:text-red-700"
                >
                  Remove
                </button>
              </div>
            );
          })}
        </div>
      </div>

      <div className="flex justify-end mt-8">
        <button
          onClick={handleSubmit}
          disabled={createTeamsMutation.isPending || teamsToAdd.length === 0}
          className="bg-green-500 text-white px-4 py-2 rounded hover:bg-green-600 disabled:opacity-50"
        >
          {createTeamsMutation.isPending ? 'Creating Teams...' : 'Create Teams'}
        </button>
      </div>
    </div>
  );
};