import React, { useState, useEffect } from 'react';
import { Link } from 'react-router-dom';
import { Tournament, TournamentTeam } from '../types/calcutta';
import { tournamentService } from '../services/tournamentService';

type TournamentStatus = 'Complete' | 'In Progress';

interface TournamentWithStatus extends Tournament {
  status: TournamentStatus;
  totalTeams: number;
  eliminatedTeams: number;
}

export const TournamentListPage: React.FC = () => {
  const [tournaments, setTournaments] = useState<TournamentWithStatus[]>([]);
  const [error, setError] = useState<string | null>(null);

  useEffect(() => {
    fetchTournaments();
  }, []);

  const fetchTournaments = async () => {
    try {
      const data = await tournamentService.getAllTournaments();
      
      // Fetch teams for each tournament to determine status
      const tournamentsWithStatus = await Promise.all(
        data.map(async (tournament) => {
          const teams = await tournamentService.getTournamentTeams(tournament.id);
          const eliminatedTeams = teams.filter(team => team.eliminated).length;
          const totalTeams = teams.length;
          
          const status: TournamentStatus = totalTeams - eliminatedTeams <= 1 ? 'Complete' : 'In Progress';
          
          return {
            ...tournament,
            status,
            totalTeams,
            eliminatedTeams
          };
        })
      );
      
      setTournaments(tournamentsWithStatus);
    } catch (err) {
      setError('Failed to load tournaments');
      console.error('Error loading tournaments:', err);
    }
  };

  return (
    <div className="container mx-auto px-4 py-8">
      <div className="flex justify-between items-center mb-8">
        <h1 className="text-3xl font-bold">Tournaments</h1>
        <Link
          to="/admin/tournaments/create"
          className="bg-blue-500 text-white px-4 py-2 rounded hover:bg-blue-600"
        >
          Create Tournament
        </Link>
      </div>

      {error && (
        <div className="bg-red-100 border border-red-400 text-red-700 px-4 py-3 rounded mb-4">
          {error}
        </div>
      )}

      <div className="grid gap-4">
        {tournaments.map(tournament => (
          <Link
            key={tournament.id}
            to={`/admin/tournaments/${tournament.id}`}
            className="block p-6 bg-white rounded-lg shadow hover:shadow-md transition-shadow"
          >
            <div className="flex justify-between items-start">
              <div>
                <h2 className="text-xl font-semibold mb-2">{tournament.name}</h2>
                <p className="text-gray-600">
                  {tournament.rounds} rounds • Created {new Date(tournament.created).toLocaleDateString()}
                </p>
              </div>
              <div className="flex items-center gap-4">
                <div className={`px-3 py-1 rounded-full text-sm font-medium ${
                  tournament.status === 'Complete' 
                    ? 'bg-green-100 text-green-800' 
                    : 'bg-yellow-100 text-yellow-800'
                }`}>
                  {tournament.status}
                </div>
                <Link
                  to={`/admin/tournaments/${tournament.id}/edit`}
                  className="text-blue-500 hover:text-blue-700"
                  onClick={(e) => e.stopPropagation()}
                >
                  Edit
                </Link>
              </div>
            </div>
          </Link>
        ))}

        {tournaments.length === 0 && !error && (
          <div className="text-center py-8 text-gray-500">
            No tournaments found. Create your first tournament to get started.
          </div>
        )}
      </div>
    </div>
  );
}; 