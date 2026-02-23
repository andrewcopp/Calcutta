import { CalcuttaEntry, CalcuttaPortfolio, CalcuttaPortfolioTeam } from '../../schemas/calcutta';
import type { TournamentTeam } from '../../schemas/tournament';
import { Card } from '../../components/ui/Card';

interface DashboardSummaryProps {
  currentEntry: CalcuttaEntry;
  entries: CalcuttaEntry[];
  portfolios: CalcuttaPortfolio[];
  portfolioTeams: CalcuttaPortfolioTeam[];
  tournamentTeams: TournamentTeam[];
}

export function DashboardSummary({
  currentEntry,
  entries,
  portfolios,
  portfolioTeams,
  tournamentTeams,
}: DashboardSummaryProps) {
  // Find user's portfolio and teams
  const userPortfolio = portfolios.find((p) => p.entryId === currentEntry.id);
  const userPortfolioTeams = userPortfolio ? portfolioTeams.filter((pt) => pt.portfolioId === userPortfolio.id) : [];

  // Position: rank by totalPoints
  const sortedEntries = [...entries].sort((a, b) => (b.totalPoints ?? 0) - (a.totalPoints ?? 0));
  const rank = sortedEntries.findIndex((e) => e.id === currentEntry.id) + 1;
  const total = entries.length;

  // Teams alive
  const eliminatedSet = new Set(tournamentTeams.filter((tt) => tt.isEliminated).map((tt) => tt.id));
  const totalTeams = userPortfolioTeams.length;
  const aliveTeams = userPortfolioTeams.filter((pt) => !eliminatedSet.has(pt.teamId)).length;

  // Points
  const actualPoints = userPortfolioTeams.reduce((sum, pt) => sum + pt.actualPoints, 0);
  const remainingPoints = userPortfolioTeams
    .filter((pt) => !eliminatedSet.has(pt.teamId))
    .reduce((sum, pt) => sum + (pt.expectedPoints - pt.actualPoints), 0);

  // Payout info
  const payoutLabel =
    currentEntry.inTheMoney && currentEntry.payoutCents != null
      ? `In the money`
      : rank <= Math.ceil(total * 0.25)
        ? 'Top quartile'
        : null;

  // Color by percentile
  const percentile = total > 0 ? rank / total : 1;
  const rankColor =
    percentile <= 0.25
      ? 'text-success'
      : percentile <= 0.5
        ? 'text-primary'
        : percentile <= 0.75
          ? 'text-amber-600'
          : 'text-muted-foreground';

  return (
    <div className="grid grid-cols-1 md:grid-cols-3 gap-4 mb-6">
      <Card variant="elevated" padding="compact">
        <p className="text-xs font-medium uppercase tracking-wider text-muted-foreground mb-1">Your Position</p>
        <p className={`text-2xl font-bold ${rankColor}`}>
          #{rank} <span className="text-base font-normal text-muted-foreground">of {total}</span>
        </p>
        {payoutLabel && <p className="text-xs text-muted-foreground mt-1">{payoutLabel}</p>}
      </Card>

      <Card padding="compact">
        <p className="text-xs font-medium uppercase tracking-wider text-muted-foreground mb-1">Your Teams</p>
        <p className="text-2xl font-bold text-foreground">{totalTeams} teams</p>
        <p className="text-xs text-muted-foreground mt-1">{aliveTeams} still alive</p>
      </Card>

      <Card padding="compact">
        <p className="text-xs font-medium uppercase tracking-wider text-muted-foreground mb-1">Your Points</p>
        <p className="text-2xl font-bold text-foreground">{actualPoints.toFixed(2)} points</p>
        <p className="text-xs text-muted-foreground mt-1">{remainingPoints.toFixed(2)} possible remaining</p>
      </Card>
    </div>
  );
}
