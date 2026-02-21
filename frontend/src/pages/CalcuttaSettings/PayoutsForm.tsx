import { useEffect, useState } from 'react';
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query';
import { calcuttaService } from '../../services/calcuttaService';
import { queryKeys } from '../../queryKeys';
import { Card } from '../../components/ui/Card';
import { Input } from '../../components/ui/Input';
import { Button } from '../../components/ui/Button';
import { Alert } from '../../components/ui/Alert';
import { LoadingState } from '../../components/ui/LoadingState';

interface PayoutRow {
  position: number;
  amountCents: number;
}

interface PayoutsFormProps {
  calcuttaId: string;
  onSuccess: () => void;
}

export function PayoutsForm({ calcuttaId, onSuccess }: PayoutsFormProps) {
  const queryClient = useQueryClient();

  const payoutsQuery = useQuery({
    queryKey: queryKeys.calcuttas.payouts(calcuttaId),
    enabled: Boolean(calcuttaId),
    queryFn: () => calcuttaService.getPayouts(calcuttaId),
  });

  const [payoutRows, setPayoutRows] = useState<PayoutRow[] | null>(null);

  const payoutsData = payoutsQuery.data;

  // Initialize payout form when data loads
  useEffect(() => {
    if (payoutsData && !payoutRows) {
      setPayoutRows(
        payoutsData.payouts.length > 0 ? [...payoutsData.payouts] : [{ position: 1, amountCents: 0 }]
      );
    }
  }, [payoutsData, payoutRows]);

  const payoutMutation = useMutation({
    mutationFn: (payouts: PayoutRow[]) => {
      return calcuttaService.replacePayouts(calcuttaId, payouts);
    },
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: queryKeys.calcuttas.payouts(calcuttaId) });
      onSuccess();
    },
  });

  return (
    <>
      {payoutMutation.isError && (
        <Alert variant="error" className="mb-4">
          {payoutMutation.error instanceof Error ? payoutMutation.error.message : 'Failed to save payouts'}
        </Alert>
      )}

      <Card>
        {payoutsQuery.isLoading ? (
          <LoadingState label="Loading payouts..." />
        ) : payoutRows ? (
          <div className="space-y-4">
            {payoutRows.map((row, idx) => (
              <div key={idx} className="flex items-center gap-3">
                <span className="text-sm font-medium text-gray-700 w-8">{row.position}.</span>
                <div className="flex-1">
                  <Input
                    type="number"
                    min={0}
                    step="0.01"
                    placeholder="Amount ($)"
                    value={row.amountCents ? (row.amountCents / 100).toFixed(2) : ''}
                    onChange={(e) => {
                      const updated = [...payoutRows];
                      const dollars = parseFloat(e.target.value);
                      updated[idx] = { ...updated[idx], amountCents: isNaN(dollars) ? 0 : Math.round(dollars * 100) };
                      setPayoutRows(updated);
                    }}
                  />
                </div>
                <Button
                  type="button"
                  variant="outline"
                  size="sm"
                  onClick={() => setPayoutRows(payoutRows.filter((_, i) => i !== idx))}
                >
                  Remove
                </Button>
              </div>
            ))}
            <div className="flex gap-2 pt-2">
              <Button
                type="button"
                variant="outline"
                onClick={() =>
                  setPayoutRows([...payoutRows, { position: payoutRows.length + 1, amountCents: 0 }])
                }
              >
                Add Position
              </Button>
              <Button
                type="button"
                loading={payoutMutation.isPending}
                onClick={() => payoutMutation.mutate(payoutRows)}
              >
                Save Payouts
              </Button>
            </div>
          </div>
        ) : null}
      </Card>
    </>
  );
}
