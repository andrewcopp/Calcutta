import { useQuery } from '@tanstack/react-query';
import { Link, useParams } from 'react-router-dom';
import { CalcuttaEntryTeam } from '../types/calcutta';
import { calcuttaService } from '../services/calcuttaService';
import { queryKeys } from '../queryKeys';
import { Alert } from '../components/ui/Alert';
import { Card } from '../components/ui/Card';
import { LoadingState } from '../components/ui/LoadingState';
import { PageContainer, PageHeader } from '../components/ui/Page';
import { Table, TableBody, TableCell, TableHead, TableHeaderCell, TableRow } from '../components/ui/Table';

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
    return (
      <PageContainer>
        <Alert variant="error">Missing required parameters</Alert>
      </PageContainer>
    );
  }

  if (calcuttaTeamsQuery.isLoading) {
    return (
      <PageContainer>
        <LoadingState label="Loading teams..." />
      </PageContainer>
    );
  }

  if (calcuttaTeamsQuery.isError) {
    const message = calcuttaTeamsQuery.error instanceof Error ? calcuttaTeamsQuery.error.message : 'Failed to fetch team data';
    return (
      <PageContainer>
        <Alert variant="error">{message}</Alert>
      </PageContainer>
    );
  }

  const teams = calcuttaTeamsQuery.data?.teams || [];
  const calcuttaName = calcuttaTeamsQuery.data?.calcuttaName || '';

  return (
    <PageContainer>
      <PageHeader
        title={`${calcuttaName} - Teams`}
        actions={
          <Link to={`/calcuttas/${calcuttaId}`} className="text-blue-600 hover:text-blue-800">
            ← Back to Calcutta
          </Link>
        }
      />

      <Card className="p-0 overflow-hidden">
        <div className="overflow-x-auto">
          <Table>
            <TableHead>
              <TableRow>
                <TableHeaderCell>Seed</TableHeaderCell>
                <TableHeaderCell>Team</TableHeaderCell>
                <TableHeaderCell>Region</TableHeaderCell>
                <TableHeaderCell className="text-right">Investment</TableHeaderCell>
                <TableHeaderCell className="text-right">Points</TableHeaderCell>
                <TableHeaderCell className="text-right">ROI</TableHeaderCell>
              </TableRow>
            </TableHead>
            <TableBody>
              {teams.map((team) => (
                <TableRow key={team.teamId} className="hover:bg-gray-50">
                  <TableCell className="font-medium text-gray-900">{team.seed}</TableCell>
                  <TableCell className="text-gray-700">{team.schoolName}</TableCell>
                  <TableCell className="text-gray-700">{team.region}</TableCell>
                  <TableCell className="text-right text-gray-700">
                    ${team.totalInvestment.toFixed(2)}
                  </TableCell>
                  <TableCell className="text-right text-gray-700">{team.points}</TableCell>
                  <TableCell className="text-right">
                    <span className={`font-medium ${team.roi > 0 ? 'text-green-600' : 'text-red-600'}`}>
                      {team.roi.toFixed(2)}%
                    </span>
                  </TableCell>
                </TableRow>
              ))}
            </TableBody>
          </Table>
        </div>
      </Card>
    </PageContainer>
  );
 } 