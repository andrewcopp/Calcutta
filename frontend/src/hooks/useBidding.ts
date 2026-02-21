import React, { useMemo, useRef, useState, useCallback } from 'react';
import { useNavigate, useParams } from 'react-router-dom';
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query';
import { calcuttaService } from '../services/calcuttaService';
import { tournamentService } from '../services/tournamentService';
import { schoolService } from '../services/schoolService';
import { queryKeys } from '../queryKeys';
import { computeBudgetRemaining, computeTeamCount, computeValidationErrors } from '../utils/bidValidation';
import type { School } from '../types/school';
import type { TournamentTeam } from '../types/tournament';
import type { ComboboxOption } from '../components/ui/Combobox';

const MIN_BID = 1;

export type TeamWithSchool = TournamentTeam & { school?: School };

export interface BidSlot {
  teamId: string;
  searchText: string;
  bidAmount: number;
}

export interface TeamComboboxOption extends ComboboxOption {
  seed: number;
  region: string;
}

function createEmptySlot(): BidSlot {
  return { teamId: '', searchText: '', bidAmount: 0 };
}

export function useBidding() {
  const { calcuttaId, entryId } = useParams<{ calcuttaId: string; entryId: string }>();
  const navigate = useNavigate();
  const queryClient = useQueryClient();

  const [slots, setSlots] = useState<BidSlot[]>([]);
  const initializedRef = useRef(false);

  const biddingQuery = useQuery({
    queryKey: queryKeys.bidding.page(calcuttaId, entryId),
    enabled: Boolean(calcuttaId && entryId),
    queryFn: async () => {
      if (!calcuttaId || !entryId) {
        throw new Error('Missing required parameters');
      }

      const [calcutta, entryTeams, schools] = await Promise.all([
        calcuttaService.getCalcutta(calcuttaId),
        calcuttaService.getEntryTeams(entryId, calcuttaId),
        schoolService.getSchools(),
      ]);

      const tournamentTeams = await tournamentService.getTournamentTeams(calcutta.tournamentId);

      const schoolMap = new Map(schools.map((school) => [school.id, school]));
      const teamsWithSchools = tournamentTeams.map((team) => ({
        ...team,
        school: schoolMap.get(team.schoolId),
      }));

      const initialBids: Record<string, number> = {};
      entryTeams.forEach((entryTeam) => {
        if (entryTeam.bid > 0) {
          initialBids[entryTeam.teamId] = entryTeam.bid;
        }
      });

      return {
        calcutta,
        teams: teamsWithSchools,
        initialBids,
      };
    },
  });

  const calcutta = biddingQuery.data?.calcutta;
  const teams = useMemo(() => biddingQuery.data?.teams ?? [], [biddingQuery.data?.teams]);
  const BUDGET = calcutta?.budgetPoints ?? 100;
  const MIN_TEAMS = calcutta?.minTeams ?? 3;
  const MAX_TEAMS = calcutta?.maxTeams ?? 10;
  const MAX_BID = calcutta?.maxBidPoints ?? 50;

  // Initialize slots from existing bids
  React.useEffect(() => {
    if (biddingQuery.data && !initializedRef.current) {
      const { initialBids, teams: loadedTeams } = biddingQuery.data;
      const maxTeams = biddingQuery.data.calcutta?.maxTeams ?? 10;
      const teamMap = new Map(loadedTeams.map((t) => [t.id, t]));

      const filledSlots: BidSlot[] = Object.entries(initialBids).map(([teamId, bid]) => {
        const team = teamMap.get(teamId);
        return {
          teamId,
          searchText: team?.school?.name ?? '',
          bidAmount: bid,
        };
      });

      const emptyCount = Math.max(0, maxTeams - filledSlots.length);
      const emptySlots = Array.from({ length: emptyCount }, createEmptySlot);
      setSlots([...filledSlots, ...emptySlots]);
      initializedRef.current = true;
    }
  }, [biddingQuery.data]);

  // Derive bidsByTeamId from slots for validation/budget computation
  const bidsByTeamId = useMemo(() => {
    const result: Record<string, number> = {};
    for (const slot of slots) {
      if (slot.teamId && slot.bidAmount > 0) {
        result[slot.teamId] = slot.bidAmount;
      }
    }
    return result;
  }, [slots]);

  // Team options for combobox
  const teamOptions = useMemo((): TeamComboboxOption[] => {
    return teams
      .map((team) => ({
        id: team.id,
        label: team.school?.name ?? 'Unknown',
        seed: team.seed,
        region: team.region,
      }))
      .sort((a, b) => a.seed - b.seed || a.label.localeCompare(b.label));
  }, [teams]);

  // Set of already-selected team IDs
  const usedTeamIds = useMemo(() => {
    return new Set(slots.filter((s) => s.teamId).map((s) => s.teamId));
  }, [slots]);

  const updateEntryMutation = useMutation({
    mutationFn: async (teamsPayload: Array<{ teamId: string; bid: number }>) => {
      if (!entryId) throw new Error('Missing entry ID');
      return calcuttaService.updateEntry(entryId, teamsPayload);
    },
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: queryKeys.calcuttas.dashboard(calcuttaId) });
      queryClient.invalidateQueries({ queryKey: queryKeys.calcuttas.entryTeams(calcuttaId, entryId) });
      navigate(`/calcuttas/${calcuttaId}`);
    },
  });

  // Slot handlers
  const handleSlotSelect = useCallback(
    (slotIndex: number, teamId: string) => {
      const team = teams.find((t) => t.id === teamId);
      setSlots((prev) => {
        const next = [...prev];
        next[slotIndex] = {
          teamId,
          searchText: team?.school?.name ?? '',
          bidAmount: prev[slotIndex].bidAmount,
        };
        return next;
      });
    },
    [teams]
  );

  const handleSlotClear = useCallback((slotIndex: number) => {
    setSlots((prev) => {
      const next = [...prev];
      next[slotIndex] = createEmptySlot();
      // Compact: move filled slots to top, empties to bottom
      const filled = next.filter((s) => s.teamId);
      const empty = next.filter((s) => !s.teamId);
      return [...filled, ...empty];
    });
  }, []);

  const handleSlotSearchChange = useCallback((slotIndex: number, text: string) => {
    setSlots((prev) => {
      const next = [...prev];
      next[slotIndex] = { ...prev[slotIndex], searchText: text };
      return next;
    });
  }, []);

  const handleSlotBidChange = useCallback((slotIndex: number, bid: number) => {
    setSlots((prev) => {
      const next = [...prev];
      next[slotIndex] = { ...prev[slotIndex], bidAmount: bid };
      return next;
    });
  }, []);

  // Budget & validation (same logic as before, derived from bidsByTeamId)
  const budgetRemaining = useMemo(() => {
    return computeBudgetRemaining(bidsByTeamId, BUDGET);
  }, [bidsByTeamId, BUDGET]);

  const teamCount = useMemo(() => {
    return computeTeamCount(bidsByTeamId);
  }, [bidsByTeamId]);

  const validationErrors = useMemo(() => {
    const teamLookups = (biddingQuery.data?.teams ?? []).map((t) => ({
      id: t.id,
      schoolName: t.school?.name ?? '',
    }));
    return computeValidationErrors(
      bidsByTeamId,
      { minTeams: MIN_TEAMS, maxTeams: MAX_TEAMS, maxBidPoints: MAX_BID, budget: BUDGET },
      teamLookups,
    );
  }, [bidsByTeamId, biddingQuery.data?.teams, MIN_TEAMS, MAX_TEAMS, MAX_BID, BUDGET]);

  const isValid = validationErrors.length === 0 && teamCount >= MIN_TEAMS && teamCount <= MAX_TEAMS;

  const handleSubmit = () => {
    if (!isValid) return;
    const teamsPayload = Object.entries(bidsByTeamId).map(([teamId, bid]) => ({
      teamId,
      bid,
    }));
    updateEntryMutation.mutate(teamsPayload);
  };

  return {
    // IDs
    calcuttaId,
    entryId,

    // Query state
    biddingQuery,
    updateEntryMutation,

    // Calcutta config
    calcutta,
    BUDGET,
    MIN_BID,
    MAX_BID,
    MIN_TEAMS,
    MAX_TEAMS,

    // Slot state
    slots,
    teamOptions,
    usedTeamIds,
    teams,

    // Derived bid state
    budgetRemaining,
    teamCount,
    validationErrors,
    isValid,

    // Slot handlers
    handleSlotSelect,
    handleSlotClear,
    handleSlotSearchChange,
    handleSlotBidChange,

    // Submit
    handleSubmit,
  };
}
