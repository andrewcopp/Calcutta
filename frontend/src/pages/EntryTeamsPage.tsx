import { useEffect, useState } from 'react';
import { Link, useParams } from 'react-router-dom';
import { CalcuttaEntryTeam } from '../types/calcutta';
import { calcuttaService } from '../services/calcuttaService';

export function EntryTeamsPage() {
  const { entryId, calcuttaId } = useParams<{ entryId: string; calcuttaId: string }>();
  const [teams, setTeams] = useState<CalcuttaEntryTeam[]>([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);

  useEffect(() => {
    const fetchTeams = async () => {
      if (!entryId || !calcuttaId) {
        setError('Missing required parameters');
        setLoading(false);
        return;
      }
      
      try {
        const data = await calcuttaService.getEntryTeams(entryId, calcuttaId);
        setTeams(data);
        setLoading(false);
      } catch (err) {
        setError('Failed to fetch teams');
        setLoading(false);
      }
    };

    fetchTeams();
  }, [entryId, calcuttaId]);

  if (loading) {
    return <div>Loading...</div>;
  }

  if (error) {
    return <div className="error">{error}</div>;
  }

  return (
    <div className="container mx-auto px-4 py-8">
      <div className="mb-6">
        <Link to={`/calcuttas/${calcuttaId}`} className="text-blue-600 hover:text-blue-800">← Back to Entries</Link>
      </div>
      <h1 className="text-3xl font-bold mb-6">Teams and Bids</h1>
      <div className="grid gap-4">
        {teams.map((team) => (
          <div
            key={team.id}
            className="p-4 bg-white rounded-lg shadow"
          >
            <h2 className="text-xl font-semibold">
              {team.team?.school?.name || 'Unknown School'}
            </h2>
            <p className="text-gray-600">Bid Amount: ${team.bid}</p>
          </div>
        ))}
      </div>
    </div>
  );
} 