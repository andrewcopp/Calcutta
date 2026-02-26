import { useMemo } from 'react';
import { Link, useParams } from 'react-router-dom';
import { Alert } from '../components/ui/Alert';
import { ErrorState } from '../components/ui/ErrorState';
import { Breadcrumb } from '../components/ui/Breadcrumb';
import { Card } from '../components/ui/Card';
import { PoolTeamsSkeleton } from '../components/skeletons/PoolTeamsSkeleton';
import { PageContainer, PageHeader } from '../components/ui/Page';
import { Table, TableBody, TableCell, TableHead, TableHeaderCell, TableRow } from '../components/ui/Table';
import { usePoolDashboard } from '../hooks/usePoolDashboard';

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

export function PoolTeamsPage() {
  const { poolId } = useParams<{ poolId: string }>();

  const dashboardQuery = usePoolDashboard(poolId);

  const investingOpen = dashboardQuery.data?.investingOpen ?? false;

  const { poolName, teams } = useMemo(() => {
    if (!dashboardQuery.data) {
      return { poolName: '', teams: [] as TeamStats[] };
    }

    const { pool, investments, tournamentTeams, schools } = dashboardQuery.data;
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

    for (const investment of investments) {
      if (!investment.team) continue;
      const existing = teamStatsMap.get(investment.teamId);
      if (!existing) continue;
      existing.totalInvestment += investment.credits;
    }

    for (const team of teamStatsMap.values()) {
      if (team.totalInvestment > 0) {
        team.roi = (team.points / team.totalInvestment) * 100;
      }
    }

    const sortedTeams = Array.from(teamStatsMap.values()).sort((a, b) => a.seed - b.seed);

    return { poolName: pool.name, teams: sortedTeams };
  }, [dashboardQuery.data]);

  if (!poolId) {
    return (
      <PageContainer>
        <Alert variant="error">Missing required parameters</Alert>
      </PageContainer>
    );
  }

  if (dashboardQuery.isLoading) {
    return (
      <PageContainer>
        <PoolTeamsSkeleton />
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
          { label: 'My Pools', href: '/pools' },
          { label: poolName || 'Pool', href: `/pools/${poolId}` },
          { label: 'Teams' },
        ]}
      />
      <PageHeader
        title={`${poolName} - Teams`}
        actions={
          <Link to={`/pools/${poolId}`} className="text-primary hover:text-primary">
            ← Back to Pool
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
                <TableRow key={team.teamId} className="hover:bg-accent">
                  <TableCell className="font-medium text-foreground">{team.seed}</TableCell>
                  <TableCell className="text-foreground">{team.schoolName}</TableCell>
                  <TableCell className="text-foreground">{team.region}</TableCell>
                  <TableCell className="text-right text-foreground">
                    {investingOpen ? '0.00 credits' : `${team.totalInvestment.toFixed(2)} credits`}
                  </TableCell>
                  <TableCell className="text-right text-foreground">{team.points}</TableCell>
                  <TableCell className="text-right">
                    {investingOpen ? (
                      <span className="font-medium text-muted-foreground/60">--</span>
                    ) : (
                      <span className={`font-medium ${team.roi > 0 ? 'text-success' : 'text-destructive'}`}>
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
    </PageContainer>
  );
}
