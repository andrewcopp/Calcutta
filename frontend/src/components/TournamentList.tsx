import React, { useEffect, useState } from 'react';

import { Card } from './ui/Card';
import { Alert } from './ui/Alert';
import { Button } from './ui/Button';
import { LoadingState } from './ui/LoadingState';
import { Table, TableBody, TableCell, TableHead, TableHeaderCell, TableRow } from './ui/Table';
import { Tournament } from '../types/calcutta';
import { tournamentService } from '../services/tournamentService';

const TournamentList: React.FC = () => {
  const [tournaments, setTournaments] = useState<Tournament[]>([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);

  const loadTournaments = async () => {
    setLoading(true);
    setError(null);
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

  useEffect(() => {
    loadTournaments();
  }, []);

  if (loading)
    return (
      <div className="max-w-3xl mx-auto">
        <Card>
          <LoadingState label="Loading tournaments..." />
        </Card>
      </div>
    );

  if (error)
    return (
      <div className="max-w-3xl mx-auto">
        <Alert variant="error">
          <div className="font-semibold mb-1">Failed to load tournaments</div>
          <div className="mb-3">{error}</div>
          <Button size="sm" onClick={loadTournaments}>
            Retry
          </Button>
        </Alert>
      </div>
    );

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