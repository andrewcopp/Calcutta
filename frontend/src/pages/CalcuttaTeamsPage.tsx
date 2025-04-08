import { useEffect, useState } from 'react';
import { Link, useParams } from 'react-router-dom';
import { CalcuttaEntryTeam, School, TournamentTeam } from '../types/calcutta';
import { calcuttaService } from '../services/calcuttaService';

interface TeamStats {
  teamId: string;
  schoolId: string;
  schoolName: string;
  seed: number;
  region: string;
  totalInvestment: number;
  points: number;
  roi: number;
}

export function CalcuttaTeamsPage() {
  const { calcuttaId } = useParams<{ calcuttaId: string }>();
  const [teams, setTeams] = useState<TeamStats[]>([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);
  const [calcuttaName, setCalcuttaName] = useState<string>('');

  useEffect(() => {
    const fetchData = async () => {
      if (!calcuttaId) return;
      
      try {
        // Fetch calcutta details to get the name
        const calcutta = await calcuttaService.getCalcutta(calcuttaId);
        setCalcuttaName(calcutta.name);
        
        // Fetch all entries for this calcutta
        const entries = await calcuttaService.getCalcuttaEntries(calcuttaId);
        
        // Fetch all schools for team names
        const schools = await calcuttaService.getSchools();
        const schoolMap = new Map(schools.map(school => [school.id, school.name]));
        
        // Fetch all tournament teams for this calcutta's tournament
        const tournamentTeams = await calcuttaService.getTournamentTeams(calcutta.tournamentId);
        
        // Create a map of team stats
        const teamStatsMap = new Map<string, TeamStats>();
        
        // Initialize the map with all tournament teams
        for (const team of tournamentTeams) {
          const schoolName = schoolMap.get(team.schoolId) || 'Unknown School';
          teamStatsMap.set(team.id, {
            teamId: team.id,
            schoolId: team.schoolId,
            schoolName,
            seed: team.seed,
            region: team.region,
            totalInvestment: 0,
            points: team.wins || 0,
            roi: 0
          });
        }
        
        // Fetch all entry teams and calculate investments
        for (const entry of entries) {
          const entryTeams = await calcuttaService.getEntryTeams(entry.id, calcuttaId);
          
          for (const entryTeam of entryTeams) {
            if (!entryTeam.team) continue;
            
            const teamId = entryTeam.teamId;
            const existing = teamStatsMap.get(teamId);
            
            if (existing) {
              existing.totalInvestment += entryTeam.bid;
              teamStatsMap.set(teamId, existing);
            }
          }
        }
        
        // Calculate ROI for each team
        for (const team of teamStatsMap.values()) {
          if (team.totalInvestment > 0) {
            team.roi = (team.points / team.totalInvestment) * 100;
          }
        }
        
        // Convert map to array and sort by seed
        const sortedTeams = Array.from(teamStatsMap.values())
          .sort((a, b) => a.seed - b.seed);
        
        setTeams(sortedTeams);
        setLoading(false);
      } catch (err) {
        setError('Failed to fetch team data');
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
        <Link to={`/calcuttas/${calcuttaId}`} className="text-blue-600 hover:text-blue-800">← Back to Calcutta</Link>
      </div>
      <h1 className="text-3xl font-bold mb-6">{calcuttaName} - Teams</h1>
      
      <div className="bg-white rounded-lg shadow overflow-hidden">
        <div className="overflow-x-auto">
          <table className="min-w-full divide-y divide-gray-200">
            <thead className="bg-gray-50">
              <tr>
                <th scope="col" className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">Seed</th>
                <th scope="col" className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">Team</th>
                <th scope="col" className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">Region</th>
                <th scope="col" className="px-6 py-3 text-right text-xs font-medium text-gray-500 uppercase tracking-wider">Investment</th>
                <th scope="col" className="px-6 py-3 text-right text-xs font-medium text-gray-500 uppercase tracking-wider">Points</th>
                <th scope="col" className="px-6 py-3 text-right text-xs font-medium text-gray-500 uppercase tracking-wider">ROI</th>
              </tr>
            </thead>
            <tbody className="bg-white divide-y divide-gray-200">
              {teams.map((team) => (
                <tr key={team.teamId} className="hover:bg-gray-50">
                  <td className="px-6 py-4 whitespace-nowrap text-sm font-medium text-gray-900">{team.seed}</td>
                  <td className="px-6 py-4 whitespace-nowrap text-sm text-gray-500">{team.schoolName}</td>
                  <td className="px-6 py-4 whitespace-nowrap text-sm text-gray-500">{team.region}</td>
                  <td className="px-6 py-4 whitespace-nowrap text-sm text-gray-500 text-right">
                    ${team.totalInvestment.toFixed(2)}
                  </td>
                  <td className="px-6 py-4 whitespace-nowrap text-sm text-gray-500 text-right">{team.points}</td>
                  <td className="px-6 py-4 whitespace-nowrap text-sm text-right">
                    <span className={`font-medium ${team.roi > 0 ? 'text-green-600' : 'text-red-600'}`}>
                      {team.roi.toFixed(2)}%
                    </span>
                  </td>
                </tr>
              ))}
            </tbody>
          </table>
        </div>
      </div>
    </div>
  );
} 