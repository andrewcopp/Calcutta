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
    <div className="bg-gray-50 rounded-lg border border-gray-200 p-3">
      <div className="text-xs text-gray-500 uppercase mb-2">Optimizer Configuration</div>
      <div className="flex flex-wrap gap-x-6 gap-y-1 text-sm">
        {optimizerKind && (
          <div>
            <span className="text-gray-500">Optimizer:</span>{' '}
            <span className="font-medium">{optimizerKind}</span>
          </div>
        )}
        {budgetPoints != null && (
          <div>
            <span className="text-gray-500">Budget:</span>{' '}
            <span className="font-medium">{budgetPoints} pts</span>
          </div>
        )}
        {maxPerTeam != null && (
          <div>
            <span className="text-gray-500">Max Per Team:</span>{' '}
            <span className="font-medium">{maxPerTeam} pts</span>
          </div>
        )}
        {minTeams != null && (
          <div>
            <span className="text-gray-500">Min Teams:</span>{' '}
            <span className="font-medium">{minTeams}</span>
          </div>
        )}
        {maxTeams != null && (
          <div>
            <span className="text-gray-500">Max Teams:</span>{' '}
            <span className="font-medium">{maxTeams}</span>
          </div>
        )}
        {minBid != null && (
          <div>
            <span className="text-gray-500">Min Bid:</span>{' '}
            <span className="font-medium">{minBid} pts</span>
          </div>
        )}
        {edgeMultiplier != null && (
          <div>
            <span className="text-gray-500">Edge Multiplier:</span>{' '}
            <span className="font-medium">{edgeMultiplier}x</span>
          </div>
        )}
      </div>
    </div>
  );
}
