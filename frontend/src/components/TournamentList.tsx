import React, { useEffect, useState } from 'react';
import { Tournament } from '../types/calcutta';
import { tournamentService } from '../services/tournamentService';

const TournamentList: React.FC = () => {
  const [tournaments, setTournaments] = useState<Tournament[]>([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);

  useEffect(() => {
    const loadTournaments = async () => {
      try {
        const data = await tournamentService.getAllTournaments();
        setTournaments(data);
      } catch (err) {
        setError('Failed to load tournaments');
        console.error('Error loading tournaments:', err);
      } finally {
        setLoading(false);
      }
    };

    loadTournaments();
  }, []);

  if (loading) return <div>Loading tournaments...</div>;
  if (error) return <div className="error">{error}</div>;

  return (
    <div className="tournament-list">
      <h2>NCAA Tournament Winners</h2>
      <table className="tournament-table">
        <thead>
          <tr>
            <th>Name</th>
            <th>Rounds</th>
          </tr>
        </thead>
        <tbody>
          {tournaments.map((tournament) => (
            <tr key={tournament.id}>
              <td>{tournament.name}</td>
              <td>{tournament.rounds}</td>
            </tr>
          ))}
        </tbody>
      </table>
    </div>
  );
};

export default TournamentList; 