import { useEffect, useState } from 'react';
import { Link, useParams } from 'react-router-dom';
import { CalcuttaEntry, CalcuttaPortfolio, CalcuttaPortfolioTeam, School, CalcuttaEntryTeam } from '../types/calcutta';
import { calcuttaService } from '../services/calcuttaService';
import { BarChart, Bar, XAxis, YAxis, CartesianGrid, Tooltip, ResponsiveContainer } from 'recharts';

interface TeamInvestment {
  teamId: string;
  teamName: string;
  totalBid: number;
}

interface SeedInvestment {
  seed: number;
  totalInvestment: number;
  teamCount: number;
}

export function CalcuttaEntriesPage() {
  const { calcuttaId } = useParams<{ calcuttaId: string }>();
  const [entries, setEntries] = useState<CalcuttaEntry[]>([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);
  const [totalEntries, setTotalEntries] = useState<number>(0);
  const [seedInvestments, setSeedInvestments] = useState<SeedInvestment[]>([]);
  const [schools, setSchools] = useState<School[]>([]);
  const [calcuttaName, setCalcuttaName] = useState<string>('');
  const [allEntryTeams, setAllEntryTeams] = useState<CalcuttaEntryTeam[]>([]);
  const [seedInvestmentData, setSeedInvestmentData] = useState<{ seed: number; totalInvestment: number }[]>([]);
  const [activeTab, setActiveTab] = useState<'leaderboard' | 'statistics'>('leaderboard');

  useEffect(() => {
    const fetchData = async () => {
      if (!calcuttaId) return;
      
      try {
        // Fetch calcutta details to get the name
        const calcutta = await calcuttaService.getCalcutta(calcuttaId);
        setCalcuttaName(calcutta.name);
        
        // Fetch all entries for this calcutta
        const entriesData = await calcuttaService.getCalcuttaEntries(calcuttaId);
        
        // For each entry, fetch portfolios and calculate total points
        const entriesWithPoints = await Promise.all(
          entriesData.map(async (entry) => {
            // Get portfolios for this entry
            const portfolios = await calcuttaService.getPortfoliosByEntry(entry.id);
            
            let totalPoints = 0;
            
            // For each portfolio, get portfolio teams and sum up actual points
            await Promise.all(
              portfolios.map(async (portfolio) => {
                const portfolioTeams = await calcuttaService.getPortfolioTeams(portfolio.id);
                const portfolioPoints = portfolioTeams.reduce((sum, team) => sum + team.actualPoints, 0);
                totalPoints += portfolioPoints;
              })
            );
            
            return {
              ...entry,
              totalPoints
            };
          })
        );
        
        // Sort entries by total points in descending order
        const sortedEntries = entriesWithPoints.sort((a, b) => (b.totalPoints || 0) - (a.totalPoints || 0));
        setEntries(sortedEntries);
        
        // Fetch all schools for team names
        const schoolsData = await calcuttaService.getSchools();
        setSchools(schoolsData);
        
        // Fetch all entry teams to calculate investments
        const allTeams: CalcuttaEntryTeam[] = [];
        for (const entry of entriesData) {
          const entryTeams = await calcuttaService.getEntryTeams(entry.id, calcuttaId);
          allTeams.push(...entryTeams);
        }
        setAllEntryTeams(allTeams);
        
        // Calculate investment by seed
        const seedMap = new Map<number, number>();
        for (const team of allTeams) {
          if (!team.team?.seed || !team.bid) continue;
          const seed = team.team.seed;
          const currentTotal = seedMap.get(seed) || 0;
          seedMap.set(seed, currentTotal + team.bid);
        }
        
        // Convert map to array and sort by seed
        const seedData = Array.from(seedMap.entries())
          .map(([seed, totalInvestment]) => ({ seed, totalInvestment }))
          .sort((a, b) => a.seed - b.seed);
        
        setSeedInvestmentData(seedData);
        setTotalEntries(entriesData.length);
        setLoading(false);
      } catch (err) {
        setError('Failed to fetch data');
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
        <Link to="/calcuttas" className="text-blue-600 hover:text-blue-800">← Back to Calcuttas</Link>
      </div>
      <h1 className="text-3xl font-bold mb-6">{calcuttaName}</h1>

      <div className="mb-8 flex gap-2 border-b border-gray-200">
        <button
          type="button"
          onClick={() => setActiveTab('leaderboard')}
          className={`px-4 py-2 -mb-px border-b-2 font-medium transition-colors ${
            activeTab === 'leaderboard'
              ? 'border-blue-600 text-blue-600'
              : 'border-transparent text-gray-600 hover:text-gray-900'
          }`}
        >
          Leaderboard
        </button>
        <button
          type="button"
          onClick={() => setActiveTab('statistics')}
          className={`px-4 py-2 -mb-px border-b-2 font-medium transition-colors ${
            activeTab === 'statistics'
              ? 'border-blue-600 text-blue-600'
              : 'border-transparent text-gray-600 hover:text-gray-900'
          }`}
        >
          Statistics
        </button>
      </div>
      

      {activeTab === 'statistics' && (
        <div className="grid grid-cols-1 md:grid-cols-2 gap-6 mb-8">
          <div className="bg-white rounded-lg shadow p-6">
            <h2 className="text-xl font-semibold mb-4">Tournament Statistics</h2>
            <p className="text-gray-600">Total Entries: {totalEntries}</p>
          </div>
          
          <div className="bg-white rounded-lg shadow p-6">
            <h2 className="text-xl font-semibold mb-4">Investment by Seed</h2>
            <div className="h-64">
              <ResponsiveContainer width="100%" height="100%">
                <BarChart data={seedInvestmentData}>
                  <CartesianGrid strokeDasharray="3 3" />
                  <XAxis dataKey="seed" label={{ value: 'Seed', position: 'insideBottom', offset: -5 }} />
                  <YAxis label={{ value: 'Total Investment ($)', angle: -90, position: 'insideLeft' }} />
                  <Tooltip formatter={(value: number) => [`$${value.toFixed(2)}`, 'Total Investment']} />
                  <Bar dataKey="totalInvestment" fill="#4F46E5" />
                </BarChart>
              </ResponsiveContainer>
            </div>
            <div className="mt-4 text-center">
              <Link 
                to={`/calcuttas/${calcuttaId}/teams`}
                className="text-blue-600 hover:text-blue-800 font-medium"
              >
                View All Teams →
              </Link>
            </div>
          </div>
        </div>
      )}

      {activeTab === 'leaderboard' && (
        <>
          <h2 className="text-2xl font-bold mb-6">Leaderboard</h2>
          <div className="grid gap-4">
            {entries.map((entry, index) => (
              <Link
                key={entry.id}
                to={`/calcuttas/${calcuttaId}/entries/${entry.id}`}
                className={`block p-4 rounded-lg shadow hover:shadow-md transition-shadow ${
                  index < 3 
                    ? 'bg-gradient-to-r from-yellow-50 to-yellow-100 border-2 border-yellow-400' 
                    : 'bg-white'
                }`}
              >
                <div className="flex justify-between items-center">
                  <div>
                    <h2 className="text-xl font-semibold">
                      {index + 1}. {entry.name}
                      {index < 3 && (
                        <span className="ml-2 text-yellow-600 text-sm">
                          {index === 0 ? '🥇' : index === 1 ? '🥈' : '🥉'}
                        </span>
                      )}
                    </h2>
                    <p className="text-gray-600">Created: {new Date(entry.created).toLocaleDateString()}</p>
                  </div>
                  <div className="text-right">
                    <p className={`text-2xl font-bold ${index < 3 ? 'text-yellow-600' : 'text-blue-600'}`}>
                      {entry.totalPoints ? entry.totalPoints.toFixed(2) : '0.00'} pts
                    </p>
                  </div>
                </div>
              </Link>
            ))}
          </div>
        </>
      )}
    </div>
  );
}
 