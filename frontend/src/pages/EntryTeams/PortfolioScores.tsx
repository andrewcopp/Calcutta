import { CalcuttaPortfolio, CalcuttaPortfolioTeam } from '../../types/calcutta';
import type { TournamentTeam } from '../../types/tournament';
import { Card } from '../../components/ui/Card';

interface PortfolioScoresProps {
  portfolio: CalcuttaPortfolio;
  teams: CalcuttaPortfolioTeam[];
  tournamentTeams: TournamentTeam[];
  schools: { id: string; name: string }[];
}

export function PortfolioScores({ portfolio, teams, tournamentTeams, schools }: PortfolioScoresProps) {
  const tournamentTeamMap = new Map(tournamentTeams.map((tt) => [tt.id, tt]));
  const schoolMap = new Map(schools.map((s) => [s.id, s]));

  const enrichedTeams = teams
    .map((pt) => {
      const tt = tournamentTeamMap.get(pt.teamId);
      const school = pt.team?.school ?? (tt ? schoolMap.get(tt.schoolId) : undefined);
      return {
        ...pt,
        seed: tt?.seed,
        region: tt?.region,
        isEliminated: tt?.isEliminated === true,
        schoolName: school?.name ?? 'Unknown Team',
      };
    })
    .sort((a, b) => b.actualPoints - a.actualPoints);

  return (
    <Card>
      <div className="flex items-center justify-between mb-4">
        <h3 className="text-lg font-semibold">Team Scores</h3>
        <span className="text-sm text-muted-foreground">Max possible: {portfolio.maximumPoints.toFixed(2)}</span>
      </div>
      <div className="overflow-x-auto">
        <table className="min-w-full table-fixed border-separate border-spacing-y-1">
          <thead>
            <tr className="text-left text-xs uppercase tracking-wide text-muted-foreground">
              <th className="px-2 py-2 w-14">Seed</th>
              <th className="px-2 py-2 w-20">Region</th>
              <th className="px-2 py-2">Team</th>
              <th className="px-2 py-2 w-28 text-right">Actual</th>
              <th className="px-2 py-2 w-28 text-right">Expected</th>
              <th className="px-2 py-2 w-24 text-right">Ownership</th>
            </tr>
          </thead>
          <tbody>
            {enrichedTeams.map((team) => (
              <tr key={team.id} className={team.isEliminated ? 'bg-accent text-muted-foreground/60' : 'bg-accent'}>
                <td className="px-2 py-2 font-medium rounded-l-md whitespace-nowrap">{team.seed ?? '—'}</td>
                <td className="px-2 py-2 whitespace-nowrap">{team.region ?? '—'}</td>
                <td
                  className={`px-2 py-2 font-medium whitespace-nowrap truncate ${team.isEliminated ? 'line-through' : ''}`}
                >
                  {team.schoolName}
                </td>
                <td className="px-2 py-2 text-right font-medium whitespace-nowrap">{team.actualPoints.toFixed(2)}</td>
                <td className="px-2 py-2 text-right whitespace-nowrap">{team.expectedPoints.toFixed(2)}</td>
                <td className="px-2 py-2 text-right rounded-r-md whitespace-nowrap">
                  {(team.ownershipPercentage * 100).toFixed(1)}%
                </td>
              </tr>
            ))}
          </tbody>
        </table>
      </div>
    </Card>
  );
}
