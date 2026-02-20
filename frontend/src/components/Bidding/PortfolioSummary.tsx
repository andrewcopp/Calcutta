import { Badge } from '../ui/Badge';
import { getSeedVariant } from '../../hooks/useBidding';
import type { PortfolioItem } from '../../hooks/useBidding';

interface PortfolioSummaryProps {
  portfolioSummary: PortfolioItem[];
}

export function PortfolioSummary({ portfolioSummary }: PortfolioSummaryProps) {
  if (portfolioSummary.length === 0) return null;

  return (
    <div className="bg-white shadow rounded-lg p-4 mb-6">
      <h3 className="text-sm font-semibold text-gray-900 mb-2">Your Bids So Far</h3>
      <div className="flex flex-wrap gap-2">
        {portfolioSummary.map((item) => (
          <div key={item.teamId} className="flex items-center gap-1 bg-blue-50 rounded-md px-2 py-1">
            <Badge variant={getSeedVariant(item.seed)} className="text-xs">{item.seed}</Badge>
            <span className="text-sm text-gray-800">{item.name}</span>
            <span className="text-sm font-medium text-blue-700">{item.bid} pts</span>
          </div>
        ))}
      </div>
    </div>
  );
}
