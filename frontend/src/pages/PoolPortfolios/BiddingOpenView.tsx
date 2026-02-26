import { Link } from 'react-router-dom';

import type { Portfolio } from '../../schemas/pool';
import { Alert } from '../../components/ui/Alert';
import { Card } from '../../components/ui/Card';
import { PageContainer, PageHeader } from '../../components/ui/Page';
import { Breadcrumb } from '../../components/ui/Breadcrumb';
import { Badge } from '../../components/ui/Badge';
import { Button } from '../../components/ui/Button';
import { IconClock, IconUsers } from '../../components/ui/Icons';
import { formatDate } from '../../utils/format';

interface BiddingOpenViewProps {
  poolId: string;
  poolName: string;
  currentUserPortfolio?: Portfolio;
  canEditSettings?: boolean;
  tournamentStartingAt?: string;
  totalPortfolios: number;
  isCreatingPortfolio: boolean;
  createPortfolioError: string | null;
  onCreatePortfolio: () => void;
}

const statusLabelMap: Record<string, string> = { draft: 'In Progress', submitted: 'Investments locked' };
const statusVariantMap: Record<string, string> = { draft: 'secondary', submitted: 'success' };

export function BiddingOpenView({
  poolId,
  poolName,
  currentUserPortfolio,
  canEditSettings,
  tournamentStartingAt,
  totalPortfolios,
  isCreatingPortfolio,
  createPortfolioError,
  onCreatePortfolio,
}: BiddingOpenViewProps) {
  const portfolioStatus = currentUserPortfolio?.status ?? '';
  const statusLabel = !currentUserPortfolio ? 'Not Started' : (statusLabelMap[portfolioStatus] ?? portfolioStatus);
  const statusVariant = !currentUserPortfolio ? 'secondary' : (statusVariantMap[portfolioStatus] ?? 'secondary');

  return (
    <PageContainer>
      <Breadcrumb items={[{ label: 'My Pools', href: '/pools' }, { label: poolName }]} />

      <PageHeader
        title={poolName}
        actions={
          canEditSettings ? (
            <Link to={`/pools/${poolId}/settings`}>
              <Button variant="outline" size="sm">
                Settings
              </Button>
            </Link>
          ) : undefined
        }
      />

      {createPortfolioError && (
        <Alert variant="error" className="mb-4">
          {createPortfolioError}
        </Alert>
      )}

      {!currentUserPortfolio ? (
        <Card variant="accent" padding="compact">
          <div className="flex items-center justify-between">
            <div className="flex items-center gap-3">
              <h3 className="text-lg font-semibold text-foreground">Your Portfolio</h3>
              <Badge variant={statusVariant as 'secondary' | 'success' | 'warning'}>{statusLabel}</Badge>
            </div>
            <Button onClick={onCreatePortfolio} disabled={isCreatingPortfolio} size="sm">
              {isCreatingPortfolio ? 'Creating...' : 'Start Portfolio'}
            </Button>
          </div>
        </Card>
      ) : (
        <Link to={`/pools/${poolId}/portfolios/${currentUserPortfolio.id}`} className="block">
          <Card variant="accent" padding="compact" className="hover:shadow-md transition-shadow">
            <div className="flex items-center justify-between">
              <div className="flex items-center gap-3">
                <h3 className="text-lg font-semibold text-foreground">{currentUserPortfolio.name}</h3>
                <Badge variant={statusVariant as 'secondary' | 'success' | 'warning'}>{statusLabel}</Badge>
              </div>
              <svg
                className="h-5 w-5 text-muted-foreground/60"
                fill="none"
                viewBox="0 0 24 24"
                strokeWidth="2"
                stroke="currentColor"
              >
                <path strokeLinecap="round" strokeLinejoin="round" d="M8.25 4.5l7.5 7.5-7.5 7.5" />
              </svg>
            </div>
          </Card>
        </Link>
      )}

      <div className="mt-6 p-6 border border-blue-200 rounded-lg bg-primary/10 text-center">
        <IconClock className="h-10 w-10 text-blue-400 mx-auto mb-3" />
        <p className="text-lg font-semibold text-blue-900 mb-2">Tip-off hasn't happened yet</p>
        <p className="text-primary">
          Come back after tip-off for the full leaderboard, ownership breakdowns, and live scoring.
        </p>
        {tournamentStartingAt && (
          <p className="mt-3 text-sm text-primary">Scouting reports revealed at tip-off — {formatDate(tournamentStartingAt, true)}</p>
        )}
      </div>

      <div className="mt-4 text-sm text-muted-foreground text-center flex items-center justify-center gap-1.5">
        <IconUsers className="h-4 w-4" />
        {totalPortfolios} {totalPortfolios === 1 ? 'investor' : 'investors'}
      </div>
    </PageContainer>
  );
}
