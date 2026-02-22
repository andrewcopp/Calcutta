import { Link } from 'react-router-dom';

import type { CalcuttaEntry } from '../../types/calcutta';
import { Alert } from '../../components/ui/Alert';
import { Card } from '../../components/ui/Card';
import { PageContainer, PageHeader } from '../../components/ui/Page';
import { Breadcrumb } from '../../components/ui/Breadcrumb';
import { Badge } from '../../components/ui/Badge';
import { Button } from '../../components/ui/Button';
import { IconClock, IconUsers } from '../../components/ui/Icons';
import { formatDate } from '../../utils/format';

interface BiddingOpenViewProps {
  calcuttaId: string;
  calcuttaName: string;
  currentUserEntry?: CalcuttaEntry;
  canEditSettings?: boolean;
  tournamentStartingAt?: string;
  totalEntries: number;
  isCreatingEntry: boolean;
  createEntryError: string | null;
  onCreateEntry: () => void;
}

const statusLabelMap: Record<string, string> = { incomplete: 'In Progress', accepted: 'Portfolio locked' };
const statusVariantMap: Record<string, string> = { incomplete: 'secondary', accepted: 'success' };

export function BiddingOpenView({
  calcuttaId,
  calcuttaName,
  currentUserEntry,
  canEditSettings,
  tournamentStartingAt,
  totalEntries,
  isCreatingEntry,
  createEntryError,
  onCreateEntry,
}: BiddingOpenViewProps) {
  const entryStatus = currentUserEntry?.status ?? '';
  const entryStatusLabel = !currentUserEntry
    ? 'Not Started'
    : statusLabelMap[entryStatus] ?? entryStatus;
  const entryStatusVariant = !currentUserEntry
    ? 'secondary'
    : statusVariantMap[entryStatus] ?? 'secondary';

  return (
    <PageContainer>
      <Breadcrumb
        items={[
          { label: 'My Pools', href: '/calcuttas' },
          { label: calcuttaName },
        ]}
      />

      <PageHeader
        title={calcuttaName}
        actions={
          canEditSettings ? (
            <Link to={`/calcuttas/${calcuttaId}/settings`}>
              <Button variant="outline" size="sm">Settings</Button>
            </Link>
          ) : undefined
        }
      />

      {createEntryError && (
        <Alert variant="error" className="mb-4">{createEntryError}</Alert>
      )}

      {!currentUserEntry ? (
        <Card variant="accent" padding="compact">
          <div className="flex items-center justify-between">
            <div className="flex items-center gap-3">
              <h3 className="text-lg font-semibold text-gray-900">Your Portfolio</h3>
              <Badge variant={entryStatusVariant as 'secondary' | 'success' | 'warning'}>{entryStatusLabel}</Badge>
            </div>
            <Button onClick={onCreateEntry} disabled={isCreatingEntry} size="sm">
              {isCreatingEntry ? 'Creating...' : 'Start Portfolio'}
            </Button>
          </div>
        </Card>
      ) : (
        <Link
          to={`/calcuttas/${calcuttaId}/entries/${currentUserEntry.id}`}
          className="block"
        >
          <Card variant="accent" padding="compact" className="hover:shadow-md transition-shadow">
            <div className="flex items-center justify-between">
              <div className="flex items-center gap-3">
                <h3 className="text-lg font-semibold text-gray-900">{currentUserEntry.name}</h3>
                <Badge variant={entryStatusVariant as 'secondary' | 'success' | 'warning'}>{entryStatusLabel}</Badge>
              </div>
              <svg className="h-5 w-5 text-gray-400" fill="none" viewBox="0 0 24 24" strokeWidth="2" stroke="currentColor">
                <path strokeLinecap="round" strokeLinejoin="round" d="M8.25 4.5l7.5 7.5-7.5 7.5" />
              </svg>
            </div>
          </Card>
        </Link>
      )}

      <div className="mt-6 p-6 border border-blue-200 rounded-lg bg-blue-50 text-center">
        <IconClock className="h-10 w-10 text-blue-400 mx-auto mb-3" />
        <p className="text-lg font-semibold text-blue-900 mb-2">
          Tournament hasn't started yet
        </p>
        <p className="text-blue-700">
          Come back once the tournament starts for the full leaderboard, ownership breakdowns, and live scoring.
        </p>
        {tournamentStartingAt && (
          <p className="mt-3 text-sm text-blue-600">
            Portfolios revealed {formatDate(tournamentStartingAt, true)}
          </p>
        )}
      </div>

      <div className="mt-4 text-sm text-gray-500 text-center flex items-center justify-center gap-1.5">
        <IconUsers className="h-4 w-4" />
        {totalEntries} {totalEntries === 1 ? 'portfolio' : 'portfolios'} submitted
      </div>
    </PageContainer>
  );
}
