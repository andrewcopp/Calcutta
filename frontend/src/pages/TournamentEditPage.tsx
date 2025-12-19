import React, { useState, useEffect } from 'react';
import { useParams, useNavigate } from 'react-router-dom';
import { Tournament, TournamentTeam } from '../types/calcutta';
import { School } from '../types/school';
import { tournamentService } from '../services/tournamentService';
import { adminService } from '../services/adminService';

export const TournamentEditPage: React.FC = () => {
  const { id } = useParams<{ id: string }>();
  const navigate = useNavigate();
  const [tournament, setTournament] = useState<Tournament | null>(null);
  const [teams, setTeams] = useState<TournamentTeam[]>([]);
  const [schools, setSchools] = useState<Record<string, School>>({});
  const [error, setError] = useState<string | null>(null);
  const [isSaving, setIsSaving] = useState(false);

  useEffect(() => {
    if (id) {
      fetchTournamentData();
    }
  }, [id]);

  const fetchTournamentData = async () => {
    try {
      // Fetch tournament details
      const tournamentData = await tournamentService.getTournament(id!);
      setTournament(tournamentData);

      // Fetch teams
      const teamsData = await tournamentService.getTournamentTeams(id!);
      setTeams(teamsData);

      // Fetch schools for all teams
      const schoolsData = await adminService.getAllSchools();
      const schoolsMap = schoolsData.reduce((acc, school) => {
        acc[school.id] = school;
        return acc;
      }, {} as Record<string, School>);
      setSchools(schoolsMap);
    } catch (err) {
      setError('Failed to load tournament data');
      console.error('Error loading tournament data:', err);
    }
  };

  const handleTeamUpdate = async (teamId: string, field: keyof TournamentTeam, value: any) => {
    try {
      const updatedTeam = await tournamentService.updateTournamentTeam(id!, teamId, {
        [field]: value,
      });
      setTeams(teams.map(team => 
        team.id === teamId ? updatedTeam : team
      ));
    } catch (err) {
      setError(`Failed to update team ${field}`);
      console.error('Error updating team:', err);
    }
  };

  const handleSaveAll = async () => {
    setIsSaving(true);
    setError(null);

    try {
      // Save all team updates in parallel
      await Promise.all(
        teams.map(team =>
          tournamentService.updateTournamentTeam(id!, team.id, {
            seed: team.seed,
            byes: team.byes,
            wins: team.wins,
            eliminated: team.eliminated,
          })
        )
      );
      navigate(`/admin/tournaments/${id}`);
    } catch (err) {
      setError('Failed to save changes');
      console.error('Error saving changes:', err);
    } finally {
      setIsSaving(false);
    }
  };

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
          <h1 className="text-3xl font-bold mb-2">Edit Tournament: {tournament.name}</h1>
          <p className="text-gray-600">
            {tournament.rounds} rounds • Created {new Date(tournament.created).toLocaleDateString()}
          </p>
        </div>
        <div className="flex gap-4">
          <button
            onClick={() => navigate(`/tournaments/${id}`)}
            className="px-4 py-2 border rounded hover:bg-gray-100"
          >
            Cancel
          </button>
          <button
            onClick={handleSaveAll}
            disabled={isSaving}
            className="bg-blue-500 text-white px-4 py-2 rounded hover:bg-blue-600 disabled:opacity-50"
          >
            {isSaving ? 'Saving...' : 'Save All Changes'}
          </button>
        </div>
      </div>

      {error && (
        <div className="bg-red-100 border border-red-400 text-red-700 px-4 py-3 rounded mb-4">
          {error}
        </div>
      )}

      <div className="bg-white rounded-lg shadow overflow-hidden">
        <table className="min-w-full divide-y divide-gray-200">
          <thead className="bg-gray-50">
            <tr>
              <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">
                Seed
              </th>
              <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">
                School
              </th>
              <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">
                Byes
              </th>
              <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">
                Wins
              </th>
              <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">
                Status
              </th>
            </tr>
          </thead>
          <tbody className="bg-white divide-y divide-gray-200">
            {teams
              .sort((a, b) => a.seed - b.seed)
              .map(team => (
                <tr key={team.id}>
                  <td className="px-6 py-4 whitespace-nowrap text-sm font-medium text-gray-900">
                    <select
                      value={team.seed}
                      onChange={(e) => handleTeamUpdate(team.id, 'seed', parseInt(e.target.value) || 1)}
                      className="w-16 p-1 border rounded"
                    >
                      {Array.from({ length: 16 }, (_, i) => i + 1).map(seed => (
                        <option key={seed} value={seed}>
                          {seed}
                        </option>
                      ))}
                    </select>
                  </td>
                  <td className="px-6 py-4 whitespace-nowrap text-sm text-gray-900">
                    {schools[team.schoolId]?.name || 'Unknown School'}
                  </td>
                  <td className="px-6 py-4 whitespace-nowrap">
                    <input
                      type="number"
                      min="0"
                      max="7"
                      value={team.byes}
                      onChange={(e) => handleTeamUpdate(team.id, 'byes', parseInt(e.target.value) || 0)}
                      className="w-16 p-1 border rounded"
                    />
                  </td>
                  <td className="px-6 py-4 whitespace-nowrap">
                    <input
                      type="number"
                      min="0"
                      max="7"
                      value={team.wins}
                      onChange={(e) => handleTeamUpdate(team.id, 'wins', parseInt(e.target.value) || 0)}
                      className="w-16 p-1 border rounded"
                    />
                  </td>
                  <td className="px-6 py-4 whitespace-nowrap">
                    <select
                      value={team.eliminated ? 'eliminated' : 'active'}
                      onChange={(e) => handleTeamUpdate(team.id, 'eliminated', e.target.value === 'eliminated')}
                      className="p-1 border rounded"
                    >
                      <option value="active">Active</option>
                      <option value="eliminated">Eliminated</option>
                    </select>
                  </td>
                </tr>
              ))}
          </tbody>
        </table>
      </div>
    </div>
  );
}; 