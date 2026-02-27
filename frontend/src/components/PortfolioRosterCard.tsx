import { Link } from 'react-router-dom';
import { Button } from './ui/Button';
import { Card } from './ui/Card';
import type { Investment } from '../schemas/pool';

interface PortfolioRosterCardProps {
  portfolioId: string;
  poolId: string;
  investments: Investment[];
  budgetCredits: number;
  canEdit?: boolean;
  title?: string;
}

export function PortfolioRosterCard({
  portfolioId,
  poolId,
  investments,
  budgetCredits,
  canEdit = true,
  title = 'Your Portfolio',
}: PortfolioRosterCardProps) {
  const sortedInvestments = [...investments].sort((a, b) => b.credits - a.credits);
  const totalSpent = investments.reduce((sum, et) => sum + et.credits, 0);

  return (
    <Card variant="default" padding="none">
      <div className="px-4 py-4 border-b border-gray-200 flex items-center justify-between">
        <h3 className="text-lg font-semibold text-gray-900">{title}</h3>
        {canEdit && (
          <Link to={`/pools/${poolId}/portfolios/${portfolioId}/invest`}>
            <Button size="sm">Edit</Button>
          </Link>
        )}
      </div>

      <div className="px-4 py-2 divide-y divide-gray-100">
        {sortedInvestments.map((et) => (
          <div key={et.id} className="flex items-center justify-between py-2">
            <span className="text-sm text-gray-800">
              {et.team?.school?.name ?? 'Unknown'} ({et.team?.region ?? '?'} - {et.team?.seed ?? '?'})
            </span>
            <span className="text-sm font-medium text-blue-700">{et.credits} credits</span>
          </div>
        ))}
      </div>

      <div className="px-4 py-3 border-t border-gray-200 flex justify-between text-sm text-gray-600">
        <span>{sortedInvestments.length} teams</span>
        <span>
          {totalSpent} / {budgetCredits} credits
        </span>
      </div>
    </Card>
  );
}
