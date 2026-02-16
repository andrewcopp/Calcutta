import { useEffect, useState } from 'react';
import { useParams } from 'react-router-dom';
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query';
import { calcuttaService } from '../services/calcuttaService';
import { useUser } from '../contexts/useUser';
import { queryKeys } from '../queryKeys';
import { PageContainer, PageHeader } from '../components/ui/Page';
import { Breadcrumb } from '../components/ui/Breadcrumb';
import { Card } from '../components/ui/Card';
import { Input } from '../components/ui/Input';
import { Button } from '../components/ui/Button';
import { Alert } from '../components/ui/Alert';
import { LoadingState } from '../components/ui/LoadingState';
import { SettingsSkeleton } from '../components/skeletons/SettingsSkeleton';

export function CalcuttaSettingsPage() {
  const { calcuttaId } = useParams<{ calcuttaId: string }>();
  const { user } = useUser();
  const queryClient = useQueryClient();
  const [successMessage, setSuccessMessage] = useState<string | null>(null);

  const calcuttaQuery = useQuery({
    queryKey: queryKeys.calcuttas.settings(calcuttaId),
    enabled: Boolean(calcuttaId),
    staleTime: 30_000,
    queryFn: () => calcuttaService.getCalcutta(calcuttaId!),
  });

  const [form, setForm] = useState<{
    name: string;
    minTeams: number;
    maxTeams: number;
    maxBid: number;
    biddingOpen: boolean;
  } | null>(null);

  const calcutta = calcuttaQuery.data;

  // Initialize form when data loads
  useEffect(() => {
    if (calcutta && !form) {
      setForm({
        name: calcutta.name,
        minTeams: calcutta.minTeams,
        maxTeams: calcutta.maxTeams,
        maxBid: calcutta.maxBid,
        biddingOpen: calcutta.biddingOpen,
      });
    }
  }, [calcutta, form]);

  const payoutsQuery = useQuery({
    queryKey: queryKeys.calcuttas.payouts(calcuttaId),
    enabled: Boolean(calcuttaId),
    staleTime: 30_000,
    queryFn: () => calcuttaService.getPayouts(calcuttaId!),
  });

  const [payoutRows, setPayoutRows] = useState<Array<{ position: number; amountCents: number }> | null>(null);

  const payoutsData = payoutsQuery.data;

  // Initialize payout form when data loads
  useEffect(() => {
    if (payoutsData && !payoutRows) {
      setPayoutRows(payoutsData.payouts.length > 0 ? [...payoutsData.payouts] : [{ position: 1, amountCents: 0 }]);
    }
  }, [payoutsData, payoutRows]);

  const payoutMutation = useMutation({
    mutationFn: (payouts: Array<{ position: number; amountCents: number }>) => {
      return calcuttaService.replacePayouts(calcuttaId!, payouts);
    },
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: queryKeys.calcuttas.payouts(calcuttaId) });
      queryClient.invalidateQueries({ queryKey: queryKeys.calcuttas.entriesPage(calcuttaId) });
      setSuccessMessage('Payouts saved successfully.');
      setTimeout(() => setSuccessMessage(null), 3000);
    },
  });

  const updateMutation = useMutation({
    mutationFn: (updates: Parameters<typeof calcuttaService.updateCalcutta>[1]) => {
      return calcuttaService.updateCalcutta(calcuttaId!, updates);
    },
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: queryKeys.calcuttas.settings(calcuttaId) });
      queryClient.invalidateQueries({ queryKey: queryKeys.calcuttas.entriesPage(calcuttaId) });
      setSuccessMessage('Settings saved successfully.');
      setTimeout(() => setSuccessMessage(null), 3000);
    },
  });

  const handleSubmit = (e: React.FormEvent) => {
    e.preventDefault();
    if (!form) return;
    updateMutation.mutate(form);
  };

  if (!calcuttaId) {
    return (
      <PageContainer>
        <Alert variant="error">Missing calcutta ID</Alert>
      </PageContainer>
    );
  }

  if (calcuttaQuery.isLoading) {
    return (
      <PageContainer>
        <SettingsSkeleton />
      </PageContainer>
    );
  }

  if (calcuttaQuery.isError) {
    const message = calcuttaQuery.error instanceof Error ? calcuttaQuery.error.message : 'Failed to load settings';
    return (
      <PageContainer>
        <Alert variant="error">{message}</Alert>
      </PageContainer>
    );
  }

  if (calcutta && user?.id !== calcutta.ownerId) {
    return (
      <PageContainer>
        <Alert variant="error">Only the pool owner can access settings.</Alert>
      </PageContainer>
    );
  }

  if (!form) return null;

  return (
    <PageContainer>
      <Breadcrumb
        items={[
          { label: 'Calcuttas', href: '/calcuttas' },
          { label: calcutta?.name ?? 'Pool', href: `/calcuttas/${calcuttaId}` },
          { label: 'Settings' },
        ]}
      />

      <PageHeader title="Pool Settings" />

      {successMessage && <Alert variant="success" className="mb-4">{successMessage}</Alert>}
      {updateMutation.isError && (
        <Alert variant="error" className="mb-4">
          {updateMutation.error instanceof Error ? updateMutation.error.message : 'Failed to save settings'}
        </Alert>
      )}

      <Card>
        <form onSubmit={handleSubmit} className="space-y-6">
          <div>
            <label htmlFor="name" className="block text-sm font-medium text-gray-700 mb-1">
              Pool Name
            </label>
            <Input
              id="name"
              type="text"
              value={form.name}
              onChange={(e) => setForm({ ...form, name: e.target.value })}
              required
            />
          </div>

          <div className="grid grid-cols-1 sm:grid-cols-3 gap-4">
            <div>
              <label htmlFor="minTeams" className="block text-sm font-medium text-gray-700 mb-1">
                Min Teams
              </label>
              <Input
                id="minTeams"
                type="number"
                min={1}
                value={form.minTeams}
                onChange={(e) => setForm({ ...form, minTeams: parseInt(e.target.value) || 0 })}
                required
              />
            </div>

            <div>
              <label htmlFor="maxTeams" className="block text-sm font-medium text-gray-700 mb-1">
                Max Teams
              </label>
              <Input
                id="maxTeams"
                type="number"
                min={1}
                value={form.maxTeams}
                onChange={(e) => setForm({ ...form, maxTeams: parseInt(e.target.value) || 0 })}
                required
              />
            </div>

            <div>
              <label htmlFor="maxBid" className="block text-sm font-medium text-gray-700 mb-1">
                Max Bid per Team
              </label>
              <Input
                id="maxBid"
                type="number"
                min={1}
                value={form.maxBid}
                onChange={(e) => setForm({ ...form, maxBid: parseInt(e.target.value) || 0 })}
                required
              />
            </div>
          </div>

          <div className="flex items-center gap-3">
            <input
              id="biddingOpen"
              type="checkbox"
              checked={form.biddingOpen}
              onChange={(e) => setForm({ ...form, biddingOpen: e.target.checked })}
              className="h-4 w-4 rounded border-gray-300"
            />
            <label htmlFor="biddingOpen" className="text-sm font-medium text-gray-700">
              Bidding Open
            </label>
          </div>

          {calcutta?.biddingLockedAt && (
            <p className="text-sm text-gray-500">
              Bidding locked at: {new Date(calcutta.biddingLockedAt).toLocaleString()}
            </p>
          )}

          <div className="pt-2">
            <Button type="submit" loading={updateMutation.isPending}>
              Save Changes
            </Button>
          </div>
        </form>
      </Card>

      <PageHeader title="Payout Structure" className="mt-8" />

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
                    placeholder="Amount (cents)"
                    value={row.amountCents}
                    onChange={(e) => {
                      const updated = [...payoutRows];
                      updated[idx] = { ...updated[idx], amountCents: parseInt(e.target.value) || 0 };
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
    </PageContainer>
  );
}
