import { useState, useCallback } from 'react';
import { useMutation, useQueryClient } from '@tanstack/react-query';
import { useNavigate } from 'react-router-dom';
import { tournamentService } from '../services/tournamentService';
import { queryKeys } from '../queryKeys';
import type { RegionState } from './useTeamSetupForm';
import { collectTeamsForSubmission } from './useTeamSetupForm';

interface UseTeamSetupMutationsParams {
  tournamentId: string;
  regionList: string[];
  regions: Record<string, RegionState>;
}

export interface TeamSetupMutationsState {
  errors: string[];
  handleSubmit: () => void;
  isPending: boolean;
}

export function useTeamSetupMutations({
  tournamentId,
  regionList,
  regions,
}: UseTeamSetupMutationsParams): TeamSetupMutationsState {
  const navigate = useNavigate();
  const queryClient = useQueryClient();
  const [errors, setErrors] = useState<string[]>([]);

  const replaceTeamsMutation = useMutation({
    mutationFn: async () => {
      const teams = collectTeamsForSubmission(regionList, regions);
      return tournamentService.replaceTeams(tournamentId, teams);
    },
    onSuccess: async () => {
      await queryClient.invalidateQueries({ queryKey: queryKeys.tournaments.teams(tournamentId) });
      navigate(`/admin/tournaments/${tournamentId}`);
    },
    onError: (err: unknown) => {
      const apiErr = err as { body?: { errors?: string[] } };
      if (apiErr?.body?.errors) {
        setErrors(apiErr.body.errors);
      } else {
        setErrors([err instanceof Error ? err.message : 'Failed to save teams']);
      }
    },
  });

  const handleSubmit = useCallback(() => {
    setErrors([]);
    replaceTeamsMutation.mutate();
  }, [replaceTeamsMutation]);

  return {
    errors,
    handleSubmit,
    isPending: replaceTeamsMutation.isPending,
  };
}
