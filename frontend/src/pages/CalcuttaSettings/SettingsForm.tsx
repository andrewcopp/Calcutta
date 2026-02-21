import { useEffect, useState } from 'react';
import { useMutation, useQueryClient } from '@tanstack/react-query';
import { calcuttaService } from '../../services/calcuttaService';
import { queryKeys } from '../../queryKeys';
import { Card } from '../../components/ui/Card';
import { Input } from '../../components/ui/Input';
import { Button } from '../../components/ui/Button';
import { Alert } from '../../components/ui/Alert';
import type { Calcutta } from '../../types/calcutta';

interface SettingsFormValues {
  name: string;
  minTeams: number;
  maxTeams: number;
  maxBid: number;
}

interface SettingsFormProps {
  calcuttaId: string;
  calcutta: Calcutta;
  onSuccess: () => void;
}

export function SettingsForm({ calcuttaId, calcutta, onSuccess }: SettingsFormProps) {
  const queryClient = useQueryClient();

  const [form, setForm] = useState<SettingsFormValues | null>(null);

  // Initialize form when data loads
  useEffect(() => {
    if (calcutta && !form) {
      setForm({
        name: calcutta.name,
        minTeams: calcutta.minTeams,
        maxTeams: calcutta.maxTeams,
        maxBid: calcutta.maxBid,
      });
    }
  }, [calcutta, form]);

  const updateMutation = useMutation({
    mutationFn: (updates: Parameters<typeof calcuttaService.updateCalcutta>[1]) => {
      return calcuttaService.updateCalcutta(calcuttaId, updates);
    },
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: queryKeys.calcuttas.settings(calcuttaId) });
      onSuccess();
    },
  });

  const handleSubmit = (e: React.FormEvent) => {
    e.preventDefault();
    if (!form) return;
    updateMutation.mutate(form);
  };

  if (!form) return null;

  return (
    <>
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

          <div className="pt-2">
            <Button type="submit" loading={updateMutation.isPending}>
              Save Changes
            </Button>
          </div>
        </form>
      </Card>
    </>
  );
}
