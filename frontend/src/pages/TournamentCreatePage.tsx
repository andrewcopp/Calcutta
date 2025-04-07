import React, { useState, useEffect } from 'react';
import { useNavigate } from 'react-router-dom';
import { School } from '../types/school';
import { adminService } from '../services/adminService';
import { tournamentService } from '../services/tournamentService';

interface TeamToAdd {
  schoolId: string;
  seed: number;
}

export const TournamentCreatePage: React.FC = () => {
  const navigate = useNavigate();
  const [name, setName] = useState('');
  const [rounds, setRounds] = useState(6);
  const [schools, setSchools] = useState<School[]>([]);
  const [selectedSchool, setSelectedSchool] = useState<string>('');
  const [selectedSeed, setSelectedSeed] = useState<number>(1);
  const [teamsToAdd, setTeamsToAdd] = useState<TeamToAdd[]>([]);
  const [searchTerm, setSearchTerm] = useState('');
  const [error, setError] = useState<string | null>(null);
  const [isSubmitting, setIsSubmitting] = useState(false);

  useEffect(() => {
    const loadSchools = async () => {
      try {
        const data = await adminService.getAllSchools();
        setSchools(data);
      } catch (err) {
        setError('Failed to load schools');
        console.error('Error loading schools:', err);
      }
    };

    loadSchools();
  }, []);

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

    setIsSubmitting(true);

    try {
      // Create the tournament first
      const tournament = await tournamentService.createTournament(name, rounds);

      // Then create all teams in parallel
      await Promise.all(
        teamsToAdd.map(team =>
          tournamentService.createTournamentTeam(tournament.id, team.schoolId, team.seed)
        )
      );

      navigate(`/admin/tournaments/${tournament.id}`);
    } catch (err) {
      setError('Failed to create tournament');
      console.error('Error creating tournament:', err);
    } finally {
      setIsSubmitting(false);
    }
  };

  return (
    <div className="container mx-auto px-4 py-8">
      <div className="flex justify-between items-center mb-8">
        <h1 className="text-3xl font-bold">Create New Tournament</h1>
        <button
          onClick={() => navigate('/admin/tournaments')}
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

      <form onSubmit={handleSubmit} className="space-y-8">
        <div className="bg-white p-6 rounded-lg shadow">
          <h2 className="text-xl font-semibold mb-4">Tournament Details</h2>
          <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
            <div>
              <label className="block text-sm font-medium text-gray-700 mb-1">
                Tournament Name
              </label>
              <input
                type="text"
                value={name}
                onChange={(e) => setName(e.target.value)}
                className="w-full p-2 border rounded"
                required
              />
            </div>
            <div>
              <label className="block text-sm font-medium text-gray-700 mb-1">
                Number of Rounds
              </label>
              <input
                type="number"
                value={rounds}
                onChange={(e) => setRounds(parseInt(e.target.value))}
                min="1"
                max="7"
                className="w-full p-2 border rounded"
                required
              />
            </div>
          </div>
        </div>

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
                max="68"
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

        <div className="flex justify-end">
          <button
            type="submit"
            disabled={isSubmitting || teamsToAdd.length === 0}
            className="bg-green-500 text-white px-4 py-2 rounded hover:bg-green-600 disabled:opacity-50"
          >
            {isSubmitting ? 'Creating Tournament...' : 'Create Tournament'}
          </button>
        </div>
      </form>
    </div>
  );
};

export default TournamentCreatePage; 