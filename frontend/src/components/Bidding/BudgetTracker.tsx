import React from 'react';
import { Card } from '../ui/Card';
import { cn } from '../../lib/cn';

type BudgetTrackerProps = {
  budgetRemaining: number;
  totalBudget: number;
  teamCount: number;
  minTeams: number;
  maxTeams: number;
  isValid: boolean;
  validationErrors: string[];
};

export function BudgetTracker({
  budgetRemaining,
  totalBudget,
  teamCount,
  minTeams,
  maxTeams,
  isValid,
  validationErrors,
}: BudgetTrackerProps) {
  const budgetUsed = totalBudget - budgetRemaining;
  const budgetPercent = (budgetUsed / totalBudget) * 100;

  const getBudgetColor = () => {
    if (budgetRemaining < 0) return 'text-red-600';
    if (budgetRemaining < 10) return 'text-yellow-600';
    return 'text-green-600';
  };

  const getTeamCountColor = () => {
    if (teamCount < minTeams || teamCount > maxTeams) return 'text-red-600';
    if (teamCount === maxTeams) return 'text-yellow-600';
    return 'text-green-600';
  };

  return (
    <div className="sticky top-16 z-10 mb-6">
      <Card className="bg-white shadow-md">
        <div className="grid grid-cols-1 md:grid-cols-3 gap-4">
          <div>
            <div className="text-sm text-gray-600 mb-1">Budget</div>
            <div className={cn('text-2xl font-bold', getBudgetColor())}>
              {budgetRemaining.toFixed(2)} / {totalBudget.toFixed(2)} pts
            </div>
            <div className="mt-2 bg-gray-200 rounded-full h-2 overflow-hidden">
              <div
                className={cn('h-full transition-all', {
                  'bg-green-500': budgetRemaining >= 10,
                  'bg-yellow-500': budgetRemaining < 10 && budgetRemaining >= 0,
                  'bg-red-500': budgetRemaining < 0,
                })}
                style={{ width: `${Math.min(budgetPercent, 100)}%` }}
              />
            </div>
          </div>

          <div>
            <div className="text-sm text-gray-600 mb-1">Teams Selected</div>
            <div className={cn('text-2xl font-bold', getTeamCountColor())}>
              {teamCount} / {minTeams}-{maxTeams}
            </div>
            {teamCount < minTeams && (
              <div className="text-xs text-red-600 mt-1">Need {minTeams - teamCount} more team(s)</div>
            )}
            {teamCount > maxTeams && (
              <div className="text-xs text-red-600 mt-1">Remove {teamCount - maxTeams} team(s)</div>
            )}
          </div>

          <div>
            <div className="text-sm text-gray-600 mb-1">Status</div>
            <div
              className={cn('text-2xl font-bold', {
                'text-green-600': isValid,
                'text-red-600': !isValid,
              })}
            >
              {isValid ? 'Ready to Submit' : 'Not Valid'}
            </div>
            {validationErrors.length > 0 && (
              <div className="text-xs text-red-600 mt-1 space-y-1">
                {validationErrors.map((error, index) => (
                  <div key={index}>{error}</div>
                ))}
              </div>
            )}
          </div>
        </div>
      </Card>
    </div>
  );
}
