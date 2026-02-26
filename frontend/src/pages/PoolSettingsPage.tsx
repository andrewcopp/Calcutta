import { useParams } from 'react-router-dom';
import { useQuery } from '@tanstack/react-query';
import { poolService } from '../services/poolService';
import { queryKeys } from '../queryKeys';
import { PageContainer, PageHeader } from '../components/ui/Page';
import { Breadcrumb } from '../components/ui/Breadcrumb';
import { Alert } from '../components/ui/Alert';
import { ErrorState } from '../components/ui/ErrorState';
import { SettingsSkeleton } from '../components/skeletons/SettingsSkeleton';
import { toast } from '../lib/toast';
import { SettingsForm } from './PoolSettings/SettingsForm';
import { PayoutsForm } from './PoolSettings/PayoutsForm';

export function PoolSettingsPage() {
  const { poolId } = useParams<{ poolId: string }>();

  const poolQuery = useQuery({
    queryKey: queryKeys.pools.settings(poolId),
    enabled: Boolean(poolId),
    queryFn: () => poolService.getPool(poolId!),
  });

  if (!poolId) {
    return (
      <PageContainer>
        <Alert variant="error">Missing pool ID</Alert>
      </PageContainer>
    );
  }

  if (poolQuery.isLoading) {
    return (
      <PageContainer>
        <SettingsSkeleton />
      </PageContainer>
    );
  }

  if (poolQuery.isError) {
    return (
      <PageContainer>
        <ErrorState error={poolQuery.error} onRetry={() => poolQuery.refetch()} />
      </PageContainer>
    );
  }

  const pool = poolQuery.data;

  if (pool && !pool.abilities?.canEditSettings) {
    return (
      <PageContainer>
        <Alert variant="error">You do not have permission to access settings.</Alert>
      </PageContainer>
    );
  }

  if (!pool) return null;

  return (
    <PageContainer>
      <Breadcrumb
        items={[
          { label: 'My Pools', href: '/pools' },
          { label: pool.name, href: `/pools/${poolId}` },
          { label: 'Settings' },
        ]}
      />

      <PageHeader title="Pool Rules" />

      <SettingsForm
        poolId={poolId}
        pool={pool}
        onSuccess={() => toast.success('Settings saved successfully.')}
      />

      <PageHeader title="Payout Structure" className="mt-8" />

      <PayoutsForm poolId={poolId} onSuccess={() => toast.success('Payouts saved successfully.')} />
    </PageContainer>
  );
}
