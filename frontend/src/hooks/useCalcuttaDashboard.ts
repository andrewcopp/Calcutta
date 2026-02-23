import { useQuery } from '@tanstack/react-query';
import { calcuttaService } from '../services/calcuttaService';
import { queryKeys } from '../queryKeys';
import { CalcuttaDashboard } from '../schemas/calcutta';

export function useCalcuttaDashboard(calcuttaId: string | undefined) {
  return useQuery<CalcuttaDashboard>({
    queryKey: queryKeys.calcuttas.dashboard(calcuttaId),
    enabled: Boolean(calcuttaId),
    refetchOnWindowFocus: true,
    queryFn: () => {
      if (!calcuttaId) throw new Error('Missing calcuttaId');
      return calcuttaService.getCalcuttaDashboard(calcuttaId);
    },
  });
}
