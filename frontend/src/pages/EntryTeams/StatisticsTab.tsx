import { CalcuttaEntryTeam, CalcuttaPortfolio, CalcuttaPortfolioTeam } from '../../schemas/calcutta';
import type { TournamentTeam } from '../../schemas/tournament';
import { Card } from '../../components/ui/Card';
import { PortfolioScores } from './PortfolioScores';

type StatisticsTabProps = {
  portfolios: CalcuttaPortfolio[];
  portfolioTeams: CalcuttaPortfolioTeam[];
  teams: CalcuttaEntryTeam[];
  tournamentTeams: TournamentTeam[];
  schools: { id: string; name: string }[];
};

export function StatisticsTab({ portfolios, portfolioTeams, teams, tournamentTeams, schools }: StatisticsTabProps) {
  if (portfolios.length === 0) {
    return (
      <p className="text-muted-foreground text-sm py-4">
        No portfolio statistics available yet. Statistics appear after the tournament begins.
      </p>
    );
  }

  const portfolio = portfolios[0];

  const totalInvestment = teams.reduce((sum, t) => sum + t.bid, 0);
  const actualPoints = portfolioTeams.reduce((sum, pt) => sum + pt.actualPoints, 0);
  const expectedPoints = portfolioTeams.reduce((sum, pt) => sum + pt.expectedPoints, 0);

  const eliminatedSet = new Set(tournamentTeams.filter((tt) => tt.isEliminated).map((tt) => tt.id));
  const teamsAlive = portfolioTeams.filter((pt) => !eliminatedSet.has(pt.teamId)).length;

  return (
    <div className="space-y-6">
      <div className="grid grid-cols-2 md:grid-cols-4 gap-4">
        <Card padding="compact">
          <p className="text-xs font-medium uppercase tracking-wider text-muted-foreground mb-1">Total Bids</p>
          <p className="text-2xl font-bold text-foreground">{totalInvestment}</p>
          <p className="text-xs text-muted-foreground mt-1">credits spent</p>
        </Card>
        <Card padding="compact">
          <p className="text-xs font-medium uppercase tracking-wider text-muted-foreground mb-1">Actual Points</p>
          <p className="text-2xl font-bold text-foreground">{actualPoints.toFixed(2)}</p>
          <p className="text-xs text-muted-foreground mt-1">earned so far</p>
        </Card>
        <Card padding="compact">
          <p className="text-xs font-medium uppercase tracking-wider text-muted-foreground mb-1">Expected Points</p>
          <p className="text-2xl font-bold text-foreground">{expectedPoints.toFixed(2)}</p>
          <p className="text-xs text-muted-foreground mt-1">if all teams win out</p>
        </Card>
        <Card padding="compact">
          <p className="text-xs font-medium uppercase tracking-wider text-muted-foreground mb-1">Teams Alive</p>
          <p className="text-2xl font-bold text-foreground">
            {teamsAlive} <span className="text-base font-normal text-muted-foreground">of {portfolioTeams.length}</span>
          </p>
          <p className="text-xs text-muted-foreground mt-1">still competing</p>
        </Card>
      </div>

      <PortfolioScores
        portfolio={portfolio}
        teams={portfolioTeams}
        tournamentTeams={tournamentTeams}
        schools={schools}
      />
    </div>
  );
}
