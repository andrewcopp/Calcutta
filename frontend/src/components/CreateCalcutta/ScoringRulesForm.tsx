import { Input } from '../ui/Input';

const WIN_INDEX_LABELS: Record<number, string> = {
  1: 'First Four Win',
  2: 'Round of 64 Win',
  3: 'Round of 32 Win',
  4: 'Sweet 16 Win',
  5: 'Elite 8 Win',
  6: 'Final Four Win',
  7: 'Championship Win',
};

export interface ScoringRule {
  winIndex: number;
  pointsAwarded: number;
}

interface ScoringRulesFormProps {
  scoringRules: ScoringRule[];
  onPointsChange: (winIndex: number, value: string) => void;
}

export function ScoringRulesForm({ scoringRules, onPointsChange }: ScoringRulesFormProps) {
  if (scoringRules.length === 0) {
    return null;
  }

  return (
    <div>
      <label className="block text-sm font-medium text-gray-700 mb-3">Scoring Rules</label>
      <div className="space-y-3">
        {scoringRules.map((rule) => (
          <div key={rule.winIndex} className="flex items-center gap-3">
            <label htmlFor={`scoring-${rule.winIndex}`} className="text-sm text-gray-600 w-44 shrink-0">
              {WIN_INDEX_LABELS[rule.winIndex] ?? `Win ${rule.winIndex}`}
            </label>
            <Input
              type="number"
              id={`scoring-${rule.winIndex}`}
              min={0}
              value={rule.pointsAwarded}
              onChange={(e) => onPointsChange(rule.winIndex, e.target.value)}
              className="w-28"
            />
            <span className="text-sm text-gray-500">points</span>
          </div>
        ))}
      </div>
      <p className="mt-2 text-sm text-gray-500">Points teams earn for each win in the tournament</p>
    </div>
  );
}
