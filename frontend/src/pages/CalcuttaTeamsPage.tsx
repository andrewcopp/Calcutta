import { useQuery } from '@tanstack/react-query';
import { Link, useParams } from 'react-router-dom';
import { CalcuttaEntryTeam, School, TournamentTeam } from '../types/calcutta';
import { calcuttaService } from '../services/calcuttaService';
import { queryKeys } from '../queryKeys';
import { Alert } from '../components/ui/Alert';

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

  const calcuttaTeamsQuery = useQuery({
    queryKey: queryKeys.calcuttas.teamsPage(calcuttaId),
    enabled: Boolean(calcuttaId),
    staleTime: 30_000,
    queryFn: async () => {
      if (!calcuttaId) throw new Error('Missing calcuttaId');

      const calcutta = await calcuttaService.getCalcutta(calcuttaId);

      const [entries, schools, tournamentTeams] = await Promise.all([
        calcuttaService.getCalcuttaEntries(calcuttaId),
        calcuttaService.getSchools(),
        calcuttaService.getTournamentTeams(calcutta.tournamentId),
      ]);

      const schoolMap = new Map(schools.map((school) => [school.id, school.name]));

      const teamStatsMap = new Map<string, TeamStats>();
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
          roi: 0,
        });
      }

      const entryTeamsByEntry = await Promise.all(
        entries.map((entry) => calcuttaService.getEntryTeams(entry.id, calcuttaId))
      );
      const allEntryTeams: CalcuttaEntryTeam[] = entryTeamsByEntry.flat();

      for (const entryTeam of allEntryTeams) {
        if (!entryTeam.team) continue;

        const existing = teamStatsMap.get(entryTeam.teamId);
        if (!existing) continue;

        existing.totalInvestment += entryTeam.bid || 0;
        teamStatsMap.set(entryTeam.teamId, existing);
      }

      for (const team of teamStatsMap.values()) {
        if (team.totalInvestment > 0) {
          team.roi = (team.points / team.totalInvestment) * 100;
        }
      }

      const teams = Array.from(teamStatsMap.values()).sort((a, b) => a.seed - b.seed);

      return { calcuttaName: calcutta.name, teams };
    },
  });

  if (!calcuttaId) {
    return <Alert variant="error">Missing required parameters</Alert>;
  }

  if (calcuttaTeamsQuery.isLoading) {
    return <div>Loading...</div>;
  }

  if (calcuttaTeamsQuery.isError) {
    const message = calcuttaTeamsQuery.error instanceof Error ? calcuttaTeamsQuery.error.message : 'Failed to fetch team data';
    return <Alert variant="error">{message}</Alert>;
  }

  const teams = calcuttaTeamsQuery.data?.teams || [];
  const calcuttaName = calcuttaTeamsQuery.data?.calcuttaName || '';

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