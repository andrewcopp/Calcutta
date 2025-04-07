import React, { useState, useEffect } from 'react';
import { School } from '../types/school';
import { adminService } from '../services/adminService';
import { tournamentService } from '../services/tournamentService';

interface TournamentTeamFormProps {
  tournamentId: string;
  onComplete: () => void;
}

interface TeamToAdd {
  schoolId: string;
  seed: number;
}

export const TournamentTeamForm: React.FC<TournamentTeamFormProps> = ({
  tournamentId,
  onComplete,
}) => {
  const [schools, setSchools] = useState<School[]>([]);
  const [selectedSchool, setSelectedSchool] = useState<string>('');
  const [teamsToAdd, setTeamsToAdd] = useState<TeamToAdd[]>([]);
  const [error, setError] = useState<string | null>(null);
  const [isSubmitting, setIsSubmitting] = useState(false);
  const [searchTerm, setSearchTerm] = useState('');

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
      seed: teamsToAdd.length + 1,
    };

    setTeamsToAdd([...teamsToAdd, newTeam]);
    setSelectedSchool('');
    setSearchTerm('');
  };

  const handleRemoveTeam = (index: number) => {
    const newTeams = [...teamsToAdd];
    newTeams.splice(index, 1);
    // Update seeds for remaining teams
    newTeams.forEach((team, i) => {
      team.seed = i + 1;
    });
    setTeamsToAdd(newTeams);
  };

  const handleSubmit = async () => {
    if (teamsToAdd.length === 0) {
      setError('Please add at least one team');
      return;
    }

    setIsSubmitting(true);
    setError(null);

    try {
      // Create teams in parallel
      await Promise.all(
        teamsToAdd.map(team =>
          tournamentService.createTournamentTeam(tournamentId, team.schoolId, team.seed)
        )
      );
      onComplete();
    } catch (err) {
      setError('Failed to create teams');
      console.error('Error creating teams:', err);
    } finally {
      setIsSubmitting(false);
    }
  };

  return (
    <div className="bg-white p-6 rounded-lg shadow">
      <h2 className="text-2xl font-bold mb-4">Add Teams to Tournament</h2>
      
      {error && (
        <div className="bg-red-100 border border-red-400 text-red-700 px-4 py-3 rounded mb-4">
          {error}
        </div>
      )}

      <div className="mb-6">
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
          <button
            onClick={handleAddTeam}
            className="bg-blue-500 text-white px-4 py-2 rounded hover:bg-blue-600"
          >
            Add Team
          </button>
        </div>
      </div>

      <div className="mb-6">
        <h3 className="text-lg font-semibold mb-2">Teams to Add ({teamsToAdd.length}/68)</h3>
        <div className="space-y-2">
          {teamsToAdd.map((team, index) => {
            const school = schools.find(s => s.id === team.schoolId);
            return (
              <div
                key={index}
                className="flex items-center justify-between p-2 border rounded"
              >
                <span>
                  {index + 1}. {school?.name || 'Unknown School'} (Seed {team.seed})
                </span>
                <button
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

      <div className="flex justify-end gap-4">
        <button
          onClick={onComplete}
          className="px-4 py-2 border rounded hover:bg-gray-100"
        >
          Cancel
        </button>
        <button
          onClick={handleSubmit}
          disabled={isSubmitting || teamsToAdd.length === 0}
          className="bg-green-500 text-white px-4 py-2 rounded hover:bg-green-600 disabled:opacity-50"
        >
          {isSubmitting ? 'Creating Teams...' : 'Create Teams'}
        </button>
      </div>
    </div>
  );
}; 