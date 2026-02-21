import { Link } from 'react-router-dom';
import { Badge } from './ui/Badge';
import { Button } from './ui/Button';
import type { CalcuttaEntryTeam } from '../types/calcutta';

interface EntryRosterCardProps {
  entryId: string;
  calcuttaId: string;
  entryStatus: string;
  entryTeams: CalcuttaEntryTeam[];
  budgetPoints: number;
  canEdit?: boolean;
}

export function EntryRosterCard({
  entryId,
  calcuttaId,
  entryStatus,
  entryTeams,
  budgetPoints,
  canEdit = true,
}: EntryRosterCardProps) {
  const sortedTeams = [...entryTeams].sort((a, b) => b.bid - a.bid);
  const totalSpent = entryTeams.reduce((sum, et) => sum + et.bid, 0);

  return (
    <div className="bg-white border border-gray-200 rounded-lg shadow-sm">
      <div className="px-4 py-4 border-b border-gray-200 flex items-center justify-between">
        <div className="flex items-center gap-3">
          <h3 className="text-lg font-semibold text-gray-900">Your Entry</h3>
          <Badge variant={({ empty: 'secondary', invalid: 'warning', valid: 'default', accepted: 'success', rejected: 'destructive' } as const)[entryStatus] ?? 'secondary'}>
            {({ empty: 'Empty', invalid: 'Invalid', valid: 'Valid', accepted: 'Accepted', rejected: 'Rejected' } as Record<string, string>)[entryStatus] ?? entryStatus}
          </Badge>
        </div>
        {canEdit && (
          <Link to={`/calcuttas/${calcuttaId}/entries/${entryId}/bid`}>
            <Button size="sm">Edit</Button>
          </Link>
        )}
      </div>

      <div className="px-4 py-2 divide-y divide-gray-100">
        {sortedTeams.map((et) => (
          <div key={et.id} className="flex items-center justify-between py-2">
            <span className="text-sm text-gray-800">
              {et.team?.school?.name ?? 'Unknown'} ({et.team?.region ?? '?'} - {et.team?.seed ?? '?'})
            </span>
            <span className="text-sm font-medium text-blue-700">{et.bid} pts</span>
          </div>
        ))}
      </div>

      <div className="px-4 py-3 border-t border-gray-200 flex justify-between text-sm text-gray-600">
        <span>{sortedTeams.length} teams</span>
        <span>{totalSpent} / {budgetPoints} pts</span>
      </div>
    </div>
  );
}
