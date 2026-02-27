import { Portfolio, OwnershipSummary, OwnershipDetail } from '../../schemas/pool';
import type { TournamentTeam } from '../../schemas/tournament';
import { Card } from '../../components/ui/Card';

interface DashboardSummaryProps {
  currentPortfolio: Portfolio;
  portfolios: Portfolio[];
  ownershipSummaries: OwnershipSummary[];
  ownershipDetails: OwnershipDetail[];
  tournamentTeams: TournamentTeam[];
}

export function DashboardSummary({
  currentPortfolio,
  portfolios,
  ownershipSummaries,
  ownershipDetails,
  tournamentTeams,
}: DashboardSummaryProps) {
  // Find user's ownership summary and details
  const userOwnershipSummary = ownershipSummaries.find((p) => p.portfolioId === currentPortfolio.id);
  const userOwnershipDetails = userOwnershipSummary ? ownershipDetails.filter((pt) => pt.portfolioId === userOwnershipSummary.id) : [];

  // Position: rank by totalReturns
  const sortedPortfolios = [...portfolios].sort((a, b) => (b.totalReturns ?? 0) - (a.totalReturns ?? 0));
  const rank = sortedPortfolios.findIndex((e) => e.id === currentPortfolio.id) + 1;
  const total = portfolios.length;

  // Teams alive
  const eliminatedSet = new Set(tournamentTeams.filter((tt) => tt.isEliminated).map((tt) => tt.id));
  const totalTeams = userOwnershipDetails.length;
  const aliveTeams = userOwnershipDetails.filter((pt) => !eliminatedSet.has(pt.teamId)).length;

  // Points
  const actualReturns = userOwnershipDetails.reduce((sum, pt) => sum + pt.actualReturns, 0);
  const remainingReturns = userOwnershipDetails
    .filter((pt) => !eliminatedSet.has(pt.teamId))
    .reduce((sum, pt) => sum + (pt.expectedReturns - pt.actualReturns), 0);

  // Payout info
  const payoutLabel =
    currentPortfolio.inTheMoney && currentPortfolio.payoutCents != null
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
        <p className="text-xs font-medium uppercase tracking-wider text-muted-foreground mb-1">Your Returns</p>
        <p className="text-2xl font-bold text-foreground">{actualReturns.toFixed(2)}</p>
        <p className="text-xs text-muted-foreground mt-1">{remainingReturns.toFixed(2)} possible remaining</p>
      </Card>
    </div>
  );
}
