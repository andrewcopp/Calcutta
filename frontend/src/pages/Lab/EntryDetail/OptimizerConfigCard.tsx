interface OptimizerConfigCardProps {
  optimizerKind?: string;
  optimizerParams?: Record<string, unknown>;
}

export function OptimizerConfigCard({ optimizerKind, optimizerParams }: OptimizerConfigCardProps) {
  const budgetPoints = optimizerParams?.budget_points as number | undefined;
  const maxPerTeam = optimizerParams?.max_per_team as number | undefined;
  const minTeams = optimizerParams?.min_teams as number | undefined;
  const maxTeams = optimizerParams?.max_teams as number | undefined;
  const minBid = optimizerParams?.min_bid as number | undefined;
  const edgeMultiplier = optimizerParams?.edge_multiplier as number | undefined;

  if (!optimizerKind && budgetPoints == null) return null;

  return (
    <div className="bg-accent rounded-lg border border-border p-3">
      <div className="text-xs text-muted-foreground uppercase mb-2">Optimizer Configuration</div>
      <div className="flex flex-wrap gap-x-6 gap-y-1 text-sm">
        {optimizerKind && (
          <div>
            <span className="text-muted-foreground">Optimizer:</span>{' '}
            <span className="font-medium">{optimizerKind}</span>
          </div>
        )}
        {budgetPoints != null && (
          <div>
            <span className="text-muted-foreground">Budget:</span>{' '}
            <span className="font-medium">{budgetPoints} credits</span>
          </div>
        )}
        {maxPerTeam != null && (
          <div>
            <span className="text-muted-foreground">Max Per Team:</span>{' '}
            <span className="font-medium">{maxPerTeam} credits</span>
          </div>
        )}
        {minTeams != null && (
          <div>
            <span className="text-muted-foreground">Min Teams:</span> <span className="font-medium">{minTeams}</span>
          </div>
        )}
        {maxTeams != null && (
          <div>
            <span className="text-muted-foreground">Max Teams:</span> <span className="font-medium">{maxTeams}</span>
          </div>
        )}
        {minBid != null && (
          <div>
            <span className="text-muted-foreground">Min Bid:</span>{' '}
            <span className="font-medium">{minBid} credits</span>
          </div>
        )}
        {edgeMultiplier != null && (
          <div>
            <span className="text-muted-foreground">Edge Multiplier:</span>{' '}
            <span className="font-medium">{edgeMultiplier}x</span>
          </div>
        )}
      </div>
    </div>
  );
}
