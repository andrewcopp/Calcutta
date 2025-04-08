import { useEffect, useState } from 'react';
import { Link } from 'react-router-dom';
import { Calcutta, CalcuttaEntry } from '../types/calcutta';
import { calcuttaService } from '../services/calcuttaService';
import { useUser } from '../contexts/UserContext';

interface CalcuttaRanking {
  calcuttaId: string;
  rank: number;
  totalEntries: number;
  points: number;
}

export function CalcuttaListPage() {
  const [calcuttas, setCalcuttas] = useState<Calcutta[]>([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);
  const [userRankings, setUserRankings] = useState<Map<string, CalcuttaRanking>>(new Map());
  const { user } = useUser();

  useEffect(() => {
    const fetchCalcuttas = async () => {
      try {
        console.log('Fetching calcuttas...');
        const data = await calcuttaService.getAllCalcuttas();
        console.log('Fetched calcuttas:', data);
        setCalcuttas(data);
        
        // If user is logged in, fetch entries and calculate rankings for each calcutta
        if (user) {
          console.log('User is logged in, fetching entries...');
          const rankingsMap = new Map<string, CalcuttaRanking>();
          
          // Fetch entries sequentially to avoid overwhelming the server
          for (const calcutta of data) {
            try {
              console.log(`Fetching entries for calcutta ${calcutta.id}...`);
              const entries = await calcuttaService.getCalcuttaEntries(calcutta.id);
              console.log(`Fetched ${entries.length} entries for calcutta ${calcutta.id}`);
              
              // Find user's entry and calculate ranking
              const userEntry = entries.find(entry => entry.userId === user.id);
              if (userEntry) {
                console.log(`User is a participant in calcutta ${calcutta.id}`);
                
                // Sort entries by total points to determine ranking
                const sortedEntries = [...entries].sort((a, b) => (b.totalPoints || 0) - (a.totalPoints || 0));
                const userRank = sortedEntries.findIndex(entry => entry.id === userEntry.id) + 1;
                
                rankingsMap.set(calcutta.id, {
                  calcuttaId: calcutta.id,
                  rank: userRank,
                  totalEntries: entries.length,
                  points: userEntry.totalPoints || 0
                });
              }
            } catch (entryError) {
              console.error(`Error fetching entries for calcutta ${calcutta.id}:`, entryError);
              // Continue with other calcuttas even if one fails
            }
          }
          setUserRankings(rankingsMap);
        }
        
        setLoading(false);
      } catch (err) {
        console.error('Error fetching calcuttas:', err);
        setError(err instanceof Error ? err.message : 'Failed to fetch calcuttas');
        setLoading(false);
      }
    };

    fetchCalcuttas();
  }, [user]);

  if (loading) {
    return <div>Loading...</div>;
  }

  if (error) {
    return (
      <div className="error p-4 bg-red-100 text-red-700 rounded">
        <h2 className="text-lg font-semibold mb-2">Error</h2>
        <p>{error}</p>
        <button 
          onClick={() => window.location.reload()} 
          className="mt-4 px-4 py-2 bg-red-600 text-white rounded hover:bg-red-700"
        >
          Retry
        </button>
      </div>
    );
  }

  return (
    <div className="container mx-auto px-4 py-8">
      <div className="flex justify-between items-center mb-6">
        <h1 className="text-3xl font-bold">Calcuttas</h1>
        <Link
          to="/calcuttas/create"
          className="bg-blue-500 text-white px-4 py-2 rounded hover:bg-blue-600"
        >
          Create New Calcutta
        </Link>
      </div>

      <div className="grid gap-4">
        {calcuttas.map((calcutta) => {
          const ranking = userRankings.get(calcutta.id);
          return (
            <Link
              key={calcutta.id}
              to={`/calcuttas/${calcutta.id}`}
              className="block p-4 bg-white rounded-lg shadow hover:shadow-md transition-shadow"
            >
              <div className="flex items-center justify-between">
                <div>
                  <h2 className="text-xl font-semibold">{calcutta.name}</h2>
                  <p className="text-gray-600">Created: {new Date(calcutta.created).toLocaleDateString()}</p>
                </div>
                {ranking && (
                  <div className="text-right">
                    <div className="text-lg font-semibold text-blue-600">
                      #{ranking.rank} of {ranking.totalEntries}
                    </div>
                    <div className="text-sm text-gray-500">
                      {ranking.points} points
                    </div>
                  </div>
                )}
              </div>
            </Link>
          );
        })}
        
        {calcuttas.length === 0 && (
          <div className="text-center py-8 bg-white rounded-lg shadow">
            <p className="text-gray-500 mb-4">No calcuttas found.</p>
            <Link
              to="/calcuttas/create"
              className="text-blue-500 hover:text-blue-700"
            >
              Create your first Calcutta
            </Link>
          </div>
        )}
      </div>
    </div>
  );
} 