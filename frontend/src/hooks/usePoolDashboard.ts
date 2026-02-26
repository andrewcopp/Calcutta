import { useQuery } from '@tanstack/react-query';
import { poolService } from '../services/poolService';
import { queryKeys } from '../queryKeys';
import { PoolDashboard } from '../schemas/pool';

export function usePoolDashboard(poolId: string | undefined) {
  return useQuery<PoolDashboard>({
    queryKey: queryKeys.pools.dashboard(poolId),
    enabled: Boolean(poolId),
    refetchOnWindowFocus: true,
    queryFn: () => {
      if (!poolId) throw new Error('Missing poolId');
      return poolService.getPoolDashboard(poolId);
    },
  });
}
