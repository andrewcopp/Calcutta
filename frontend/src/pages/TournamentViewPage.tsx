import React, { useState, useEffect } from 'react';
import { useParams, Link } from 'react-router-dom';
import { Tournament, TournamentTeam } from '../types/calcutta';
import { School } from '../types/school';
import { tournamentService } from '../services/tournamentService';
import { adminService } from '../services/adminService';

type SortField = 'seed' | 'school' | 'byes' | 'wins' | 'status';
type SortDirection = 'asc' | 'desc';

export const TournamentViewPage: React.FC = () => {
  const { id } = useParams<{ id: string }>();
  const [tournament, setTournament] = useState<Tournament | null>(null);
  const [teams, setTeams] = useState<TournamentTeam[]>([]);
  const [schools, setSchools] = useState<Record<string, School>>({});
  const [error, setError] = useState<string | null>(null);
  const [isLoading, setIsLoading] = useState(true);
  const [sortField, setSortField] = useState<SortField>('seed');
  const [sortDirection, setSortDirection] = useState<SortDirection>('asc');

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

  const handleSort = (field: SortField) => {
    if (field === sortField) {
      // Toggle direction if clicking the same field
      setSortDirection(sortDirection === 'asc' ? 'desc' : 'asc');
    } else {
      // Set new field and default to ascending
      setSortField(field);
      setSortDirection('asc');
    }
  };

  const getSortedTeams = () => {
    return [...teams].sort((a, b) => {
      let comparison = 0;
      
      switch (sortField) {
        case 'seed':
          comparison = a.seed - b.seed;
          break;
        case 'school':
          const schoolA = schools[a.schoolId]?.name || '';
          const schoolB = schools[b.schoolId]?.name || '';
          comparison = schoolA.localeCompare(schoolB);
          break;
        case 'byes':
          comparison = a.byes - b.byes;
          break;
        case 'wins':
          comparison = a.wins - b.wins;
          break;
        case 'status':
          // Sort eliminated teams to the bottom
          comparison = (a.eliminated ? 1 : 0) - (b.eliminated ? 1 : 0);
          break;
        default:
          comparison = 0;
      }
      
      return sortDirection === 'asc' ? comparison : -comparison;
    });
  };

  const getSortIcon = (field: SortField) => {
    if (field !== sortField) {
      return <span className="text-gray-400">↕</span>;
    }
    return sortDirection === 'asc' ? <span>↑</span> : <span>↓</span>;
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

  const sortedTeams = getSortedTeams();

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

      {teams.length > 0 && (
        <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-4 gap-4 mb-8">
          <div className="bg-white rounded-lg shadow p-4">
            <h3 className="text-lg font-semibold text-gray-700 mb-2">Total Teams</h3>
            <p className="text-3xl font-bold text-blue-600">{teams.length}</p>
          </div>
          <div className="bg-white rounded-lg shadow p-4">
            <h3 className="text-lg font-semibold text-gray-700 mb-2">Seed Distribution</h3>
            <div className="space-y-1">
              {Array.from({ length: 16 }, (_, i) => i + 1).map(seed => {
                const count = teams.filter(t => t.seed === seed).length;
                return count > 0 ? (
                  <div key={seed} className="flex justify-between text-sm">
                    <span>Seed {seed}:</span>
                    <span className="font-medium">{count}</span>
                  </div>
                ) : null;
              })}
            </div>
          </div>
          <div className="bg-white rounded-lg shadow p-4">
            <h3 className="text-lg font-semibold text-gray-700 mb-2">Bye Distribution</h3>
            <div className="space-y-1">
              {Array.from({ length: Math.max(...teams.map(t => t.byes)) + 1 }, (_, i) => i).map(byes => {
                const count = teams.filter(t => t.byes === byes).length;
                return count > 0 ? (
                  <div key={byes} className="flex justify-between text-sm">
                    <span>{byes} {byes === 1 ? 'Bye' : 'Byes'}:</span>
                    <span className="font-medium">{count}</span>
                  </div>
                ) : null;
              })}
            </div>
          </div>
          <div className="bg-white rounded-lg shadow p-4">
            <h3 className="text-lg font-semibold text-gray-700 mb-2">Win Distribution</h3>
            <div className="space-y-1">
              {Array.from({ length: Math.max(...teams.map(t => t.wins)) + 1 }, (_, i) => i).map(wins => {
                const count = teams.filter(t => t.wins === wins).length;
                return count > 0 ? (
                  <div key={wins} className="flex justify-between text-sm">
                    <span>{wins} {wins === 1 ? 'Win' : 'Wins'}:</span>
                    <span className="font-medium">{count}</span>
                  </div>
                ) : null;
              })}
              <div className="pt-2 mt-2 border-t border-gray-200">
                <div className="flex justify-between text-sm font-semibold">
                  <span>Total Wins:</span>
                  <span className="text-blue-600">{teams.reduce((sum, team) => sum + team.wins, 0)}</span>
                </div>
              </div>
            </div>
          </div>
        </div>
      )}

      {teams.length === 0 ? (
        <div className="bg-white rounded-lg shadow p-6 text-center">
          <p className="text-gray-500 mb-4">No teams have been added to this tournament yet.</p>
          <Link
            to={`/admin/tournaments/${id}/teams/add`}
            className="bg-blue-500 text-white px-4 py-2 rounded hover:bg-blue-600"
          >
            Add Teams
          </Link>
        </div>
      ) : (
        <div className="bg-white rounded-lg shadow overflow-hidden">
          <table className="min-w-full divide-y divide-gray-200">
            <thead className="bg-gray-50">
              <tr>
                <th 
                  className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider cursor-pointer"
                  onClick={() => handleSort('seed')}
                >
                  <div className="flex items-center">
                    Seed {getSortIcon('seed')}
                  </div>
                </th>
                <th 
                  className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider cursor-pointer"
                  onClick={() => handleSort('school')}
                >
                  <div className="flex items-center">
                    School {getSortIcon('school')}
                  </div>
                </th>
                <th 
                  className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider cursor-pointer"
                  onClick={() => handleSort('byes')}
                >
                  <div className="flex items-center">
                    Byes {getSortIcon('byes')}
                  </div>
                </th>
                <th 
                  className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider cursor-pointer"
                  onClick={() => handleSort('wins')}
                >
                  <div className="flex items-center">
                    Wins {getSortIcon('wins')}
                  </div>
                </th>
                <th 
                  className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider cursor-pointer"
                  onClick={() => handleSort('status')}
                >
                  <div className="flex items-center">
                    Status {getSortIcon('status')}
                  </div>
                </th>
              </tr>
            </thead>
            <tbody className="bg-white divide-y divide-gray-200">
              {sortedTeams.map(team => (
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
                      <span className="px-2 py-1 text-xs font-semibold rounded-full bg-red-100 text-red-800">
                        Eliminated
                      </span>
                    ) : (
                      <span className="px-2 py-1 text-xs font-semibold rounded-full bg-green-100 text-green-800">
                        Active
                      </span>
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