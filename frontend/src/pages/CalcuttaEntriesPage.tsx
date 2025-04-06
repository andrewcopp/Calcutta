import { useEffect, useState } from 'react';
import { Link, useParams } from 'react-router-dom';
import { CalcuttaEntry, CalcuttaPortfolio, CalcuttaPortfolioTeam } from '../types/calcutta';
import { calcuttaService } from '../services/calcuttaService';

export function CalcuttaEntriesPage() {
  const { calcuttaId } = useParams<{ calcuttaId: string }>();
  const [entries, setEntries] = useState<CalcuttaEntry[]>([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);

  useEffect(() => {
    const fetchData = async () => {
      if (!calcuttaId) return;
      
      try {
        // Fetch entries
        const entriesData = await calcuttaService.getCalcuttaEntries(calcuttaId);
        
        // For each entry, fetch portfolios and portfolio teams
        const entriesWithPoints = await Promise.all(entriesData.map(async (entry) => {
          const portfolios = await calcuttaService.getPortfoliosByEntry(entry.id);
          
          // For each portfolio, fetch teams and sum up points
          const portfolioPoints = await Promise.all(portfolios.map(async (portfolio) => {
            const teams = await calcuttaService.getPortfolioTeams(portfolio.id);
            return teams.reduce((sum, team) => sum + team.actualPoints, 0);
          }));
          
          // Sum up points from all portfolios
          const totalPoints = portfolioPoints.reduce((sum, points) => sum + points, 0);
          
          return {
            ...entry,
            totalPoints: totalPoints
          };
        }));
        
        // Sort entries by total points in descending order
        const sortedEntries = entriesWithPoints.sort((a, b) => 
          (b.totalPoints || 0) - (a.totalPoints || 0)
        );
        
        setEntries(sortedEntries);
        setLoading(false);
      } catch (err) {
        setError('Failed to fetch entries');
        setLoading(false);
      }
    };

    fetchData();
  }, [calcuttaId]);

  if (loading) {
    return <div>Loading...</div>;
  }

  if (error) {
    return <div className="error">{error}</div>;
  }

  return (
    <div className="container mx-auto px-4 py-8">
      <div className="mb-6">
        <Link to="/" className="text-blue-600 hover:text-blue-800">← Back to Calcuttas</Link>
      </div>
      <h1 className="text-3xl font-bold mb-6">Leaderboard</h1>
      <div className="grid gap-4">
        {entries.map((entry) => (
          <Link
            key={entry.id}
            to={`/calcuttas/${calcuttaId}/entries/${entry.id}`}
            className="block p-4 bg-white rounded-lg shadow hover:shadow-md transition-shadow"
          >
            <div className="flex justify-between items-center">
              <div>
                <h2 className="text-xl font-semibold">{entry.name}</h2>
                <p className="text-gray-600">Created: {new Date(entry.created).toLocaleDateString()}</p>
              </div>
              <div className="text-right">
                <p className="text-2xl font-bold text-blue-600">
                  {entry.totalPoints ? entry.totalPoints.toFixed(2) : '0.00'} pts
                </p>
              </div>
            </div>
          </Link>
        ))}
      </div>
    </div>
  );
} 