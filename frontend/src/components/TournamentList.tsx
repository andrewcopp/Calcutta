import React, { useEffect, useState } from 'react';

import { Card } from './ui/Card';
import { Table, TableBody, TableCell, TableHead, TableHeaderCell, TableRow } from './ui/Table';
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

  if (loading) return <div className="text-gray-500">Loading tournaments...</div>;
  if (error) return <div className="bg-red-100 border border-red-400 text-red-700 px-4 py-3 rounded">{error}</div>;

  return (
    <div className="max-w-3xl mx-auto">
      <Card>
        <h2 className="text-xl font-semibold mb-4">NCAA Tournament Winners</h2>
        <Table>
          <TableHead>
            <tr>
              <TableHeaderCell>Name</TableHeaderCell>
              <TableHeaderCell>Rounds</TableHeaderCell>
            </tr>
          </TableHead>
          <TableBody>
            {tournaments.map((tournament) => (
              <TableRow key={tournament.id}>
                <TableCell className="font-medium text-gray-900">{tournament.name}</TableCell>
                <TableCell>{tournament.rounds}</TableCell>
              </TableRow>
            ))}
          </TableBody>
        </Table>
      </Card>
    </div>
  );
};

export default TournamentList;