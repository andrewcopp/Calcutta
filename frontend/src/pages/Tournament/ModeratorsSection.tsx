import React, { useState } from 'react';
import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query';
import { TournamentModerator } from '../../types/calcutta';
import { tournamentService } from '../../services/tournamentService';
import { queryKeys } from '../../queryKeys';
import { Alert } from '../../components/ui/Alert';
import { Button } from '../../components/ui/Button';
import { Card } from '../../components/ui/Card';
import { Input } from '../../components/ui/Input';
import { LoadingState } from '../../components/ui/LoadingState';
import { Modal, ModalActions } from '../../components/ui/Modal';
import { useFlashMessage } from '../../hooks/useFlashMessage';

interface ModeratorsSectionProps {
  tournamentId: string;
}

export const ModeratorsSection: React.FC<ModeratorsSectionProps> = ({ tournamentId }) => {
  const queryClient = useQueryClient();
  const [showAddModal, setShowAddModal] = useState(false);
  const [addEmail, setAddEmail] = useState('');
  const [modSuccess, flashModSuccess] = useFlashMessage();
  const [modError, setModError] = useState<string | null>(null);

  const moderatorsQuery = useQuery({
    queryKey: queryKeys.tournaments.moderators(tournamentId),
    enabled: Boolean(tournamentId),
    queryFn: () => tournamentService.getTournamentModerators(tournamentId),
  });

  const grantMutation = useMutation({
    mutationFn: (email: string) => tournamentService.grantTournamentModerator(tournamentId, email),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: queryKeys.tournaments.moderators(tournamentId) });
      setShowAddModal(false);
      setAddEmail('');
      setModError(null);
      flashModSuccess('Moderator added successfully.');
    },
    onError: (err) => {
      setModError(err instanceof Error ? err.message : 'Failed to add moderator');
    },
  });

  const revokeMutation = useMutation({
    mutationFn: (userId: string) => tournamentService.revokeTournamentModerator(tournamentId, userId),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: queryKeys.tournaments.moderators(tournamentId) });
      flashModSuccess('Moderator removed successfully.');
    },
    onError: (err) => {
      setModError(err instanceof Error ? err.message : 'Failed to remove moderator');
    },
  });

  return (
    <>
      <Card className="mt-8 mb-8">
        <div className="flex items-center justify-between mb-4">
          <h3 className="text-lg font-semibold text-gray-900">Moderators</h3>
          <Button size="sm" onClick={() => { setShowAddModal(true); setModError(null); }}>
            Add Moderator
          </Button>
        </div>

        {modSuccess && <Alert variant="success" className="mb-4">{modSuccess}</Alert>}
        {modError && !showAddModal && <Alert variant="error" className="mb-4">{modError}</Alert>}

        {moderatorsQuery.isLoading ? (
          <LoadingState label="Loading moderators..." />
        ) : moderatorsQuery.isError ? (
          <Alert variant="error">
            Failed to load moderators.
            <Button size="sm" className="ml-2" onClick={() => moderatorsQuery.refetch()}>Retry</Button>
          </Alert>
        ) : (moderatorsQuery.data ?? []).length === 0 ? (
          <p className="text-gray-500 text-sm">No moderators have been added yet. Moderators can enter game results on your behalf.</p>
        ) : (
          <table className="w-full text-sm">
            <thead>
              <tr className="border-b border-gray-200 text-left text-gray-500">
                <th className="pb-2 font-medium">Name</th>
                <th className="pb-2 font-medium">Email</th>
                <th className="pb-2 font-medium text-right">Actions</th>
              </tr>
            </thead>
            <tbody>
              {(moderatorsQuery.data ?? []).map((mod: TournamentModerator) => (
                <tr key={mod.id} className="border-b border-gray-100 last:border-0">
                  <td className="py-2">{mod.firstName} {mod.lastName}</td>
                  <td className="py-2 text-gray-600">{mod.email}</td>
                  <td className="py-2 text-right">
                    <Button
                      size="sm"
                      variant="destructive"
                      disabled={revokeMutation.isPending}
                      onClick={() => revokeMutation.mutate(mod.id)}
                    >
                      Remove
                    </Button>
                  </td>
                </tr>
              ))}
            </tbody>
          </table>
        )}
      </Card>

      <Modal open={showAddModal} onClose={() => { setShowAddModal(false); setAddEmail(''); setModError(null); }} title="Add Moderator">
        <p className="text-sm text-gray-600 mb-4">Enter the email address of the user you want to add as a moderator.</p>
        {modError && <Alert variant="error" className="mb-4">{modError}</Alert>}
        <form onSubmit={(e) => { e.preventDefault(); if (addEmail.trim()) grantMutation.mutate(addEmail.trim()); }}>
          <Input
            type="email"
            placeholder="user@example.com"
            value={addEmail}
            onChange={(e) => setAddEmail(e.target.value)}
            required
          />
          <ModalActions>
            <Button variant="secondary" type="button" onClick={() => { setShowAddModal(false); setAddEmail(''); setModError(null); }}>
              Cancel
            </Button>
            <Button type="submit" disabled={grantMutation.isPending || !addEmail.trim()}>
              {grantMutation.isPending ? 'Adding...' : 'Add'}
            </Button>
          </ModalActions>
        </form>
      </Modal>
    </>
  );
};
