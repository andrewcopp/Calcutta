import { useMemo } from 'react';
import { Link, useParams } from 'react-router-dom';
import { Alert } from '../components/ui/Alert';
import { ErrorState } from '../components/ui/ErrorState';
import { Breadcrumb } from '../components/ui/Breadcrumb';
import { Card } from '../components/ui/Card';
import { CalcuttaTeamsSkeleton } from '../components/skeletons/CalcuttaTeamsSkeleton';
import { BiddingOverlay } from '../components/BiddingOverlay';
import { PageContainer, PageHeader } from '../components/ui/Page';
import { Table, TableBody, TableCell, TableHead, TableHeaderCell, TableRow } from '../components/ui/Table';
import { useCalcuttaDashboard } from '../hooks/useCalcuttaDashboard';

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

  const dashboardQuery = useCalcuttaDashboard(calcuttaId);

  const biddingOpen = dashboardQuery.data?.biddingOpen ?? false;
  const tournamentStartingAt = dashboardQuery.data?.tournamentStartingAt;

  const { calcuttaName, teams } = useMemo(() => {
    if (!dashboardQuery.data) {
      return { calcuttaName: '', teams: [] as TeamStats[] };
    }

    const { calcutta, entryTeams, tournamentTeams, schools } = dashboardQuery.data;
    const schoolMap = new Map(schools.map((s) => [s.id, s.name]));

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
        points: team.wins,
        roi: 0,
      });
    }

    for (const entryTeam of entryTeams) {
      if (!entryTeam.team) continue;
      const existing = teamStatsMap.get(entryTeam.teamId);
      if (!existing) continue;
      existing.totalInvestment += entryTeam.bid;
    }

    for (const team of teamStatsMap.values()) {
      if (team.totalInvestment > 0) {
        team.roi = (team.points / team.totalInvestment) * 100;
      }
    }

    const sortedTeams = Array.from(teamStatsMap.values()).sort((a, b) => a.seed - b.seed);

    return { calcuttaName: calcutta.name, teams: sortedTeams };
  }, [dashboardQuery.data]);

  if (!calcuttaId) {
    return (
      <PageContainer>
        <Alert variant="error">Missing required parameters</Alert>
      </PageContainer>
    );
  }

  if (dashboardQuery.isLoading) {
    return (
      <PageContainer>
        <CalcuttaTeamsSkeleton />
      </PageContainer>
    );
  }

  if (dashboardQuery.isError) {
    return (
      <PageContainer>
        <ErrorState error={dashboardQuery.error} onRetry={() => dashboardQuery.refetch()} />
      </PageContainer>
    );
  }

  return (
    <PageContainer>
      <Breadcrumb
        items={[
          { label: 'Calcuttas', href: '/calcuttas' },
          { label: calcuttaName || 'Pool', href: `/calcuttas/${calcuttaId}` },
          { label: 'Teams' },
        ]}
      />
      <PageHeader
        title={`${calcuttaName} - Teams`}
        actions={
          <Link to={`/calcuttas/${calcuttaId}`} className="text-blue-600 hover:text-blue-800">
            ← Back to Calcutta
          </Link>
        }
      />

      <BiddingOverlay tournamentStartingAt={tournamentStartingAt ?? ''} active={biddingOpen}>
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
                      {biddingOpen ? '0.00 pts' : `${team.totalInvestment.toFixed(2)} pts`}
                    </TableCell>
                    <TableCell className="text-right text-gray-700">{team.points}</TableCell>
                    <TableCell className="text-right">
                      {biddingOpen ? (
                        <span className="font-medium text-gray-400">--</span>
                      ) : (
                        <span className={`font-medium ${team.roi > 0 ? 'text-green-600' : 'text-red-600'}`}>
                          {team.roi.toFixed(2)}%
                        </span>
                      )}
                    </TableCell>
                  </TableRow>
                ))}
              </TableBody>
            </Table>
          </div>
        </Card>
      </BiddingOverlay>
    </PageContainer>
  );
}
