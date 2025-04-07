import React, { useState, useEffect } from 'react';
import { fetchTournaments, createTournament } from '../services/tournamentService';
import { Tournament } from '../types/tournament';
import { TournamentTeam } from '../types/calcutta';
import { fetchTournamentTeams, updateTournamentTeam, recalculatePortfolios } from '../services/adminService';
import { SchoolList } from '../components/SchoolList';

export const AdminPage: React.FC = () => {
  const [tournaments, setTournaments] = useState<Tournament[]>([]);
  const [selectedTournament, setSelectedTournament] = useState<string | null>(null);
  const [teams, setTeams] = useState<TournamentTeam[]>([]);
  const [loading, setLoading] = useState<boolean>(false);
  const [error, setError] = useState<string | null>(null);
  const [recalculating, setRecalculating] = useState(false);
  const [showSchools, setShowSchools] = useState(false);
  const [showAddTournament, setShowAddTournament] = useState(false);
  const [newTournamentName, setNewTournamentName] = useState('');
  const [newTournamentRounds, setNewTournamentRounds] = useState(7);

  useEffect(() => {
    const loadTournaments = async () => {
      try {
        setLoading(true);
        const data = await fetchTournaments();
        console.log('Loaded tournaments:', data);
        setTournaments(data);
      } catch (err) {
        setError('Failed to load tournaments');
        console.error(err);
      } finally {
        setLoading(false);
      }
    };

    loadTournaments();
  }, []);

  useEffect(() => {
    const loadTeams = async () => {
      if (!selectedTournament) {
        setTeams([]);
        return;
      }

      try {
        setLoading(true);
        const data = await fetchTournamentTeams(selectedTournament);
        setTeams(data);
      } catch (err) {
        setError('Failed to load teams');
        console.error(err);
      } finally {
        setLoading(false);
      }
    };

    loadTeams();
  }, [selectedTournament]);

  const handleTournamentChange = (e: React.ChangeEvent<HTMLSelectElement>) => {
    setSelectedTournament(e.target.value);
  };

  const handleTeamUpdate = async (teamId: string, field: 'wins' | 'byes', value: number) => {
    try {
      setLoading(true);
      await updateTournamentTeam(teamId, { [field]: value });
      
      // Update the local state
      setTeams(teams.map(team => 
        team.id === teamId ? { ...team, [field]: value } : team
      ));
    } catch (err) {
      setError(`Failed to update team ${field}`);
      console.error(err);
    } finally {
      setLoading(false);
    }
  };

  const handleRecalculatePortfolios = async () => {
    if (!selectedTournament) return;
    
    setRecalculating(true);
    setError(null);
    try {
      await recalculatePortfolios(selectedTournament);
      alert('Portfolios recalculated successfully!');
    } catch (err) {
      setError(err instanceof Error ? err.message : 'An error occurred');
    } finally {
      setRecalculating(false);
    }
  };

  const handleCreateTournament = async (e: React.FormEvent) => {
    e.preventDefault();
    try {
      setLoading(true);
      const tournament = await createTournament(newTournamentName, newTournamentRounds);
      setTournaments([...tournaments, tournament]);
      setNewTournamentName('');
      setNewTournamentRounds(7);
      setShowAddTournament(false);
    } catch (err) {
      setError('Failed to create tournament');
      console.error(err);
    } finally {
      setLoading(false);
    }
  };

  return (
    <div className="container mx-auto px-4 py-8">
      <h1 className="text-3xl font-bold mb-6">Admin Dashboard</h1>
      
      {error && (
        <div className="bg-red-100 border border-red-400 text-red-700 px-4 py-3 rounded mb-4">
          {error}
        </div>
      )}
      
      <div className="mb-6 flex space-x-4">
        <button
          onClick={() => setShowSchools(!showSchools)}
          className="bg-blue-500 hover:bg-blue-700 text-white font-bold py-2 px-4 rounded"
        >
          {showSchools ? 'Hide Schools' : 'Show Schools'}
        </button>
        <button
          onClick={() => setShowAddTournament(!showAddTournament)}
          className="bg-green-500 hover:bg-green-700 text-white font-bold py-2 px-4 rounded"
        >
          {showAddTournament ? 'Cancel' : 'Add Tournament'}
        </button>
      </div>

      {showAddTournament && (
        <div className="mb-6 p-4 border rounded-lg bg-white shadow">
          <h2 className="text-xl font-semibold mb-4">Create New Tournament</h2>
          <form onSubmit={handleCreateTournament} className="space-y-4">
            <div>
              <label htmlFor="tournamentName" className="block text-sm font-medium text-gray-700 mb-1">
                Tournament Name
              </label>
              <input
                type="text"
                id="tournamentName"
                value={newTournamentName}
                onChange={(e) => setNewTournamentName(e.target.value)}
                className="w-full p-2 border border-gray-300 rounded-md"
                required
              />
            </div>
            <div>
              <label htmlFor="tournamentRounds" className="block text-sm font-medium text-gray-700 mb-1">
                Number of Rounds
              </label>
              <input
                type="number"
                id="tournamentRounds"
                value={newTournamentRounds}
                onChange={(e) => setNewTournamentRounds(parseInt(e.target.value))}
                className="w-full p-2 border border-gray-300 rounded-md"
                min="1"
                required
              />
            </div>
            <button
              type="submit"
              className="bg-blue-500 hover:bg-blue-700 text-white font-bold py-2 px-4 rounded"
              disabled={loading}
            >
              {loading ? 'Creating...' : 'Create Tournament'}
            </button>
          </form>
        </div>
      )}

      {showSchools && (
        <div className="mb-6">
          <SchoolList />
        </div>
      )}
      
      <div className="mb-6">
        <label htmlFor="tournament" className="block text-sm font-medium text-gray-700 mb-2">
          Select Tournament
        </label>
        <select
          id="tournament"
          className="w-full p-2 border border-gray-300 rounded-md"
          value={selectedTournament || ''}
          onChange={handleTournamentChange}
        >
          <option value="">Select a tournament</option>
          {tournaments.map(tournament => (
            <option key={tournament.id} value={tournament.id}>
              {tournament.name}
            </option>
          ))}
        </select>
      </div>
      
      {loading && <div className="text-center py-4">Loading...</div>}
      
      {selectedTournament && teams.length > 0 && (
        <div className="bg-white shadow rounded-lg overflow-hidden">
          <table className="min-w-full divide-y divide-gray-200">
            <thead className="bg-gray-50">
              <tr>
                <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">
                  Team
                </th>
                <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">
                  Seed
                </th>
                <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">
                  Region
                </th>
                <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">
                  Byes
                </th>
                <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">
                  Wins
                </th>
                <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">
                  Actions
                </th>
              </tr>
            </thead>
            <tbody className="bg-white divide-y divide-gray-200">
              {teams.map(team => (
                <tr key={team.id}>
                  <td className="px-6 py-4 whitespace-nowrap">
                    {team.school?.name || 'Unknown School'}
                  </td>
                  <td className="px-6 py-4 whitespace-nowrap">
                    {team.seed}
                  </td>
                  <td className="px-6 py-4 whitespace-nowrap">
                    {team.region}
                  </td>
                  <td className="px-6 py-4 whitespace-nowrap">
                    <input
                      type="number"
                      min="0"
                      max="7"
                      className="w-16 p-1 border border-gray-300 rounded"
                      value={team.byes}
                      onChange={(e) => handleTeamUpdate(team.id, 'byes', parseInt(e.target.value) || 0)}
                    />
                  </td>
                  <td className="px-6 py-4 whitespace-nowrap">
                    <input
                      type="number"
                      min="0"
                      max="7"
                      className="w-16 p-1 border border-gray-300 rounded"
                      value={team.wins}
                      onChange={(e) => handleTeamUpdate(team.id, 'wins', parseInt(e.target.value) || 0)}
                    />
                  </td>
                  <td className="px-6 py-4 whitespace-nowrap">
                    <button
                      className="bg-blue-500 hover:bg-blue-700 text-white font-bold py-1 px-2 rounded text-sm"
                      onClick={() => handleTeamUpdate(team.id, 'wins', team.wins)}
                    >
                      Save
                    </button>
                  </td>
                </tr>
              ))}
            </tbody>
          </table>
        </div>
      )}
      
      {selectedTournament && teams.length === 0 && !loading && (
        <div className="text-center py-4 text-gray-500">
          No teams found for this tournament.
        </div>
      )}

      {selectedTournament && (
        <div className="mt-4">
          <button
            onClick={handleRecalculatePortfolios}
            disabled={recalculating}
            className="bg-blue-500 hover:bg-blue-700 text-white font-bold py-2 px-4 rounded disabled:opacity-50"
          >
            {recalculating ? 'Recalculating...' : 'Recalculate Portfolios'}
          </button>
        </div>
      )}
    </div>
  );
}; 