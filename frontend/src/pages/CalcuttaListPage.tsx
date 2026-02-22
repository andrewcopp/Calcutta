import { useQuery } from '@tanstack/react-query';
import { Link, useNavigate } from 'react-router-dom';
import { Button } from '../components/ui/Button';
import { Card } from '../components/ui/Card';
import { Collapsible } from '../components/ui/Collapsible';
import { EmptyState } from '../components/ui/EmptyState';
import { ErrorState } from '../components/ui/ErrorState';
import { PageContainer, PageHeader } from '../components/ui/Page';
import { CalcuttaListSkeleton } from '../components/skeletons/CalcuttaListSkeleton';
import { CalcuttaWithRanking } from '../types/calcutta';
import { calcuttaService } from '../services/calcuttaService';
import { formatRelativeTime } from '../utils/format';
import { groupCalcuttas } from '../utils/calcuttaListHelpers';
import { useUser } from '../contexts/UserContext';
import { queryKeys } from '../queryKeys';
import { IconBriefcase } from '../components/ui/Icons';

export function CalcuttaListPage() {
  const navigate = useNavigate();
  const { user } = useUser();

  const calcuttasQuery = useQuery({
    queryKey: queryKeys.calcuttas.listWithRankings(user?.id),
    queryFn: async () => {
      return calcuttaService.getCalcuttasWithRankings();
    },
  });

  if (calcuttasQuery.isLoading) {
    return (
      <PageContainer>
        <PageHeader
          title="Calcuttas"
          actions={
            <Button onClick={() => navigate('/calcuttas/create')}>Create New Calcutta</Button>
          }
        />
        <CalcuttaListSkeleton />
      </PageContainer>
    );
  }

  if (calcuttasQuery.isError) {
    return (
      <PageContainer>
        <ErrorState error={calcuttasQuery.error} onRetry={() => calcuttasQuery.refetch()} />
      </PageContainer>
    );
  }

  const calcuttas: CalcuttaWithRanking[] = calcuttasQuery.data || [];
  const { current, historical } = groupCalcuttas(calcuttas);

  return (
    <PageContainer>
      <PageHeader
        title="Calcuttas"
        actions={
          <Button onClick={() => navigate('/calcuttas/create')}>Create New Calcutta</Button>
        }
      />

      <div className="grid gap-4">
        {current.map((calcutta) => (
          <CalcuttaCard key={calcutta.id} calcutta={calcutta} />
        ))}

        {historical.length > 0 && (
          <Collapsible title="Historical" count={historical.length}>
            <div className="grid gap-4">
              {historical.map((calcutta) => (
                <CalcuttaCard key={calcutta.id} calcutta={calcutta} />
              ))}
            </div>
          </Collapsible>
        )}

        {calcuttas.length === 0 ? (
          <EmptyState
            icon={<IconBriefcase />}
            title="Welcome to Calcutta!"
            description="Create your first pool or wait for an invitation from a commissioner."
            action={
              <div className="flex gap-3">
                <Link to="/rules">
                  <Button variant="outline">Learn the Rules</Button>
                </Link>
                <Link to="/calcuttas/create">
                  <Button>Create a Pool</Button>
                </Link>
              </div>
            }
          />
        ) : null}
      </div>
    </PageContainer>
  );
}

function CalcuttaCard({ calcutta }: { calcutta: CalcuttaWithRanking }) {
  const ranking = calcutta.ranking;
  const tournamentStarted = calcutta.tournamentStartingAt
    ? new Date(calcutta.tournamentStartingAt) <= new Date()
    : false;

  // State C: Tournament started + has ranking → rank is the hero
  if (tournamentStarted && ranking) {
    return (
      <Link to={`/calcuttas/${calcutta.id}`} className="block">
        <Card variant="elevated" padding="compact" className="hover:shadow-lg transition-shadow">
          <div className="flex items-center justify-between">
            <span className="text-2xl font-bold text-blue-600">
              #{ranking.rank} of {ranking.totalEntries}
            </span>
            <span className="text-lg font-semibold">
              {ranking.points.toFixed(2)} pts
            </span>
          </div>
          <p className="text-sm text-gray-600 mt-1">{calcutta.name}</p>
        </Card>
      </Link>
    );
  }

  // State D: Tournament started + no entry
  if (tournamentStarted) {
    return (
      <Link to={`/calcuttas/${calcutta.id}`} className="block">
        <Card padding="compact" className="hover:shadow-md transition-shadow">
          <div className="flex items-center justify-between">
            <h2 className="text-xl font-semibold">{calcutta.name}</h2>
            <span className="text-sm font-medium text-gray-400">No Entry</span>
          </div>
        </Card>
      </Link>
    );
  }

  // States A/B: Bidding open
  const countdown = calcutta.tournamentStartingAt
    ? formatRelativeTime(calcutta.tournamentStartingAt)
    : null;

  return (
    <Link to={`/calcuttas/${calcutta.id}`} className="block">
      <Card variant="accent" padding="compact" className="hover:shadow-md transition-shadow">
        <div className="flex items-center justify-between">
          <div>
            <div className="flex items-center gap-2">
              <h2 className="text-xl font-semibold">{calcutta.name}</h2>
              <span className={`text-xs font-medium px-2 py-0.5 rounded-full ${
                calcutta.hasEntry
                  ? 'bg-green-100 text-green-700'
                  : 'bg-amber-100 text-amber-700'
              }`}>
                {calcutta.hasEntry ? 'Entry Submitted' : 'Awaiting Entry'}
              </span>
            </div>
            {countdown && (
              <p className={`text-sm mt-1 ${countdown.urgent ? 'text-red-600 font-medium' : 'text-gray-500'}`}>
                {countdown.urgent
                  ? `${countdown.text} remaining`
                  : `${countdown.text} until portfolios reveal`}
              </p>
            )}
          </div>
        </div>
      </Card>
    </Link>
  );
}
