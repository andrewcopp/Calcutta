import { EntryRosterCard } from '../../components/EntryRosterCard';
import type { CalcuttaEntryTeam } from '../../types/calcutta';

interface EntryTabProps {
  entryId: string;
  calcuttaId: string;
  entryStatus: string;
  entryTeams: CalcuttaEntryTeam[];
  budgetPoints: number;
  canEdit: boolean;
  title?: string;
}

export function EntryTab({
  entryId,
  calcuttaId,
  entryStatus,
  entryTeams,
  budgetPoints,
  canEdit,
  title,
}: EntryTabProps) {
  return (
    <EntryRosterCard
      entryId={entryId}
      calcuttaId={calcuttaId}
      entryStatus={entryStatus}
      entryTeams={entryTeams}
      budgetPoints={budgetPoints}
      canEdit={canEdit}
      title={title}
    />
  );
}
