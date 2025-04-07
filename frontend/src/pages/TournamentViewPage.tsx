import React, { useState, useEffect } from 'react';
import { useParams, Link } from 'react-router-dom';
import { Tournament, TournamentTeam } from '../types/calcutta';
import { School } from '../types/school';
import { tournamentService } from '../services/tournamentService';
import { adminService } from '../services/adminService';

export const TournamentViewPage: React.FC = () => {
  const { id } = useParams<{ id: string }>();
  const [tournament, setTournament] = useState<Tournament | null>(null);
  const [teams, setTeams] = useState<TournamentTeam[]>([]);
  const [schools, setSchools] = useState<Record<string, School>>({});
  const [error, setError] = useState<string | null>(null);
  const [isLoading, setIsLoading] = useState(true);

  useEffect(() => {
    if (id) {
      fetchTournamentData();
    }
  }, [id]);

  const fetchTournamentData = async () => {
    try {
      setIsLoading(true);
      setError(null);
      
      // Fetch tournament details
      const tournamentData = await tournamentService.getTournament(id!);
      setTournament(tournamentData);

      // Fetch teams
      const teamsData = await tournamentService.getTournamentTeams(id!);
      setTeams(teamsData || []);

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
      setTeams([]);
    } finally {
      setIsLoading(false);
    }
  };

  if (isLoading) {
    return (
      <div className="container mx-auto px-4 py-8">
        <div className="text-center">Loading...</div>
      </div>
    );
  }

  if (!tournament) {
    return (
      <div className="container mx-auto px-4 py-8">
        <div className="text-center text-red-600">Tournament not found</div>
      </div>
    );
  }

  return (
    <div className="container mx-auto px-4 py-8">
      <div className="flex justify-between items-center mb-8">
        <div>
          <h1 className="text-3xl font-bold mb-2">{tournament.name}</h1>
          <p className="text-gray-600">
            {tournament.rounds} rounds • Created {new Date(tournament.created).toLocaleDateString()}
          </p>
        </div>
        <div className="flex gap-4">
          <Link
            to={`/admin/tournaments/${id}/edit`}
            className="bg-blue-500 text-white px-4 py-2 rounded hover:bg-blue-600"
          >
            Edit Tournament
          </Link>
          <Link
            to={`/admin/tournaments/${id}/teams/add`}
            className="bg-green-500 text-white px-4 py-2 rounded hover:bg-green-600"
          >
            Add Teams
          </Link>
        </div>
      </div>

      {error && (
        <div className="bg-red-100 border border-red-400 text-red-700 px-4 py-3 rounded mb-4">
          {error}
        </div>
      )}

      {teams.length === 0 ? (
        <div className="bg-white rounded-lg shadow p-8 text-center">
          <h2 className="text-xl font-semibold mb-4">No Teams Added Yet</h2>
          <p className="text-gray-600 mb-6">This tournament has been created but no teams have been added yet.</p>
          <Link
            to={`/admin/tournaments/${id}/teams/add`}
            className="bg-green-500 text-white px-6 py-3 rounded hover:bg-green-600 inline-block"
          >
            Add Teams to Tournament
          </Link>
        </div>
      ) : (
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
                      {team.seed}
                    </td>
                    <td className="px-6 py-4 whitespace-nowrap text-sm text-gray-900">
                      {schools[team.schoolId]?.name || 'Unknown School'}
                    </td>
                    <td className="px-6 py-4 whitespace-nowrap text-sm text-gray-900">
                      {team.byes}
                    </td>
                    <td className="px-6 py-4 whitespace-nowrap text-sm text-gray-900">
                      {team.wins}
                    </td>
                    <td className="px-6 py-4 whitespace-nowrap text-sm text-gray-900">
                      {team.eliminated ? (
                        <span className="text-red-600">Eliminated</span>
                      ) : (
                        <span className="text-green-600">Active</span>
                      )}
                    </td>
                  </tr>
                ))}
            </tbody>
          </table>
        </div>
      )}
    </div>
  );
}; 