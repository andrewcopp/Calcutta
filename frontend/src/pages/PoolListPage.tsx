import { useQuery } from '@tanstack/react-query';
import { Link, useNavigate } from 'react-router-dom';
import { Button } from '../components/ui/Button';
import { Card } from '../components/ui/Card';
import { Collapsible } from '../components/ui/Collapsible';
import { EmptyState } from '../components/ui/EmptyState';
import { ErrorState } from '../components/ui/ErrorState';
import { PageContainer, PageHeader } from '../components/ui/Page';
import { PoolListSkeleton } from '../components/skeletons/PoolListSkeleton';
import { PoolWithRanking } from '../schemas/pool';
import { poolService } from '../services/poolService';
import { formatRelativeTime } from '../utils/format';
import { groupPools } from '../utils/poolListHelpers';
import { useUser } from '../contexts/UserContext';
import { queryKeys } from '../queryKeys';
import { IconBriefcase } from '../components/ui/Icons';

export function PoolListPage() {
  const navigate = useNavigate();
  const { user } = useUser();

  const poolsQuery = useQuery({
    queryKey: queryKeys.pools.listWithRankings(user?.id),
    queryFn: async () => {
      return poolService.getPoolsWithRankings();
    },
  });

  if (poolsQuery.isLoading) {
    return (
      <PageContainer>
        <PageHeader
          title="My Pools"
          actions={<Button onClick={() => navigate('/pools/create')}>Start a New Pool</Button>}
        />
        <PoolListSkeleton />
      </PageContainer>
    );
  }

  if (poolsQuery.isError) {
    return (
      <PageContainer>
        <ErrorState error={poolsQuery.error} onRetry={() => poolsQuery.refetch()} />
      </PageContainer>
    );
  }

  const pools: PoolWithRanking[] = poolsQuery.data || [];
  const { current, historical } = groupPools(pools);

  return (
    <PageContainer>
      <PageHeader
        title="My Pools"
        actions={<Button onClick={() => navigate('/pools/create')}>Start a New Pool</Button>}
      />

      <div className="grid gap-4">
        {current.map((pool) => (
          <PoolCard key={pool.id} pool={pool} />
        ))}

        {historical.length > 0 && (
          <Collapsible title="Historical" count={historical.length}>
            <div className="grid gap-4">
              {historical.map((pool) => (
                <PoolCard key={pool.id} pool={pool} />
              ))}
            </div>
          </Collapsible>
        )}

        {pools.length === 0 ? (
          <EmptyState
            icon={<IconBriefcase />}
            title="Welcome!"
            description="Create your first pool or wait for an invite."
            action={
              <div className="flex gap-3">
                <Link to="/rules">
                  <Button variant="outline">Learn the Rules</Button>
                </Link>
                <Link to="/pools/create">
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

function PoolCard({ pool }: { pool: PoolWithRanking }) {
  const ranking = pool.ranking;
  const tournamentStarted = pool.tournamentStartingAt
    ? new Date(pool.tournamentStartingAt) <= new Date()
    : false;

  // State C: Tournament started + has ranking
  if (tournamentStarted && ranking) {
    return (
      <Link to={`/pools/${pool.id}`} className="block">
        <Card variant="elevated" padding="compact" className="hover:shadow-lg transition-shadow">
          <div className="flex items-center justify-between">
            <span className="text-2xl font-bold text-primary">
              #{ranking.rank} of {ranking.totalPortfolios}
            </span>
            <span className="text-lg font-semibold">{ranking.returns.toFixed(2)} returns</span>
          </div>
          <p className="text-sm text-muted-foreground mt-1">{pool.name}</p>
        </Card>
      </Link>
    );
  }

  // State D: Tournament started + no portfolio
  if (tournamentStarted) {
    return (
      <Link to={`/pools/${pool.id}`} className="block">
        <Card padding="compact" className="hover:shadow-md transition-shadow">
          <div className="flex items-center justify-between">
            <h2 className="text-xl font-semibold">{pool.name}</h2>
            <span className="text-sm font-medium text-muted-foreground/60">On the bench</span>
          </div>
        </Card>
      </Link>
    );
  }

  // States A/B: Investing open
  const countdown = pool.tournamentStartingAt ? formatRelativeTime(pool.tournamentStartingAt) : null;

  return (
    <Link to={`/pools/${pool.id}`} className="block">
      <Card variant="accent" padding="compact" className="hover:shadow-md transition-shadow">
        <div className="flex items-center justify-between">
          <div>
            <div className="flex items-center gap-2">
              <h2 className="text-xl font-semibold">{pool.name}</h2>
              <span
                className={`text-xs font-medium px-2 py-0.5 rounded-full ${
                  pool.hasPortfolio ? 'bg-success/10 text-success' : 'bg-amber-100 text-amber-700'
                }`}
              >
                {pool.hasPortfolio ? "You're in" : 'Build your portfolio'}
              </span>
            </div>
            {countdown && (
              <p
                className={`text-sm mt-1 ${countdown.urgent ? 'text-destructive font-medium' : 'text-muted-foreground'}`}
              >
                {countdown.urgent ? `${countdown.text} remaining` : `${countdown.text} until tip-off`}
              </p>
            )}
          </div>
        </div>
      </Card>
    </Link>
  );
}
