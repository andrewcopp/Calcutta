import React, { useEffect, useMemo, useRef, useState, useCallback } from 'react';
import { useNavigate, useParams } from 'react-router-dom';
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query';
import { poolService } from '../services/poolService';
import { tournamentService } from '../services/tournamentService';
import { schoolService } from '../services/schoolService';
import { queryKeys } from '../queryKeys';
import { toast } from '../lib/toast';
import { useUser } from '../contexts/UserContext';
import { computeBudgetRemaining, computeTeamCount, computeValidationErrors } from '../utils/investmentValidation';
import type { School } from '../schemas/school';
import type { TournamentTeam } from '../schemas/tournament';
import type { ComboboxOption } from '../components/ui/Combobox';

const MIN_INVESTMENT = 1;

export type TeamWithSchool = TournamentTeam & { school?: School };

export interface InvestmentSlot {
  teamId: string;
  searchText: string;
  investmentAmount: number;
}

export interface TeamComboboxOption extends ComboboxOption {
  seed: number;
  region: string;
}

export function createEmptySlot(): InvestmentSlot {
  return { teamId: '', searchText: '', investmentAmount: 0 };
}

export function initializeSlotsFromInvestments(
  initialInvestments: Record<string, number>,
  teams: TeamWithSchool[],
  maxTeams: number,
): InvestmentSlot[] {
  const teamMap = new Map(teams.map((t) => [t.id, t]));

  const filledSlots: InvestmentSlot[] = Object.entries(initialInvestments).map(([teamId, amount]) => {
    const team = teamMap.get(teamId);
    return {
      teamId,
      searchText: team?.school?.name ?? '',
      investmentAmount: amount,
    };
  });

  const emptyCount = Math.max(0, maxTeams - filledSlots.length);
  const emptySlots = Array.from({ length: emptyCount }, createEmptySlot);
  return [...filledSlots, ...emptySlots];
}

export function deriveInvestmentsByTeamId(slots: InvestmentSlot[]): Record<string, number> {
  const result: Record<string, number> = {};
  for (const slot of slots) {
    if (slot.teamId && slot.investmentAmount > 0) {
      result[slot.teamId] = slot.investmentAmount;
    }
  }
  return result;
}

export function compactSlots(slots: InvestmentSlot[]): InvestmentSlot[] {
  const filled = slots.filter((s) => s.teamId);
  const empty = slots.filter((s) => !s.teamId);
  return [...filled, ...empty];
}

export function deriveTeamOptions(teams: TeamWithSchool[]): TeamComboboxOption[] {
  return teams
    .map((team) => ({
      id: team.id,
      label: team.school?.name ?? 'Unknown',
      seed: team.seed,
      region: team.region,
    }))
    .sort((a, b) => a.seed - b.seed || a.label.localeCompare(b.label));
}

export function deriveUsedTeamIds(slots: InvestmentSlot[]): Set<string> {
  return new Set(slots.filter((s) => s.teamId).map((s) => s.teamId));
}

export function useInvesting() {
  const { poolId, portfolioId } = useParams<{ poolId: string; portfolioId?: string }>();
  const navigate = useNavigate();
  const queryClient = useQueryClient();
  const { user } = useUser();

  const isCreating = !portfolioId;

  const [slots, setSlots] = useState<InvestmentSlot[]>([]);
  const initializedRef = useRef(false);
  const initialInvestmentsRef = useRef<Record<string, number>>({});
  const [isDirty, setIsDirty] = useState(false);

  const investingQuery = useQuery({
    queryKey: queryKeys.investing.page(poolId, portfolioId ?? null),
    enabled: Boolean(poolId),
    queryFn: async () => {
      if (!poolId) {
        throw new Error('Missing required parameters');
      }

      const [pool, schools] = await Promise.all([
        poolService.getPool(poolId),
        schoolService.getSchools(),
      ]);

      const investments = portfolioId
        ? await poolService.getInvestments(portfolioId, poolId)
        : [];

      const tournamentTeams = await tournamentService.getTournamentTeams(pool.tournamentId);

      const schoolMap = new Map(schools.map((school) => [school.id, school]));
      const teamsWithSchools = tournamentTeams.map((team) => ({
        ...team,
        school: schoolMap.get(team.schoolId),
      }));

      const initialInvestments: Record<string, number> = {};
      investments.forEach((investment) => {
        if (investment.credits > 0) {
          initialInvestments[investment.teamId] = investment.credits;
        }
      });

      return {
        pool,
        teams: teamsWithSchools,
        initialInvestments,
      };
    },
  });

  const pool = investingQuery.data?.pool;
  const teams = useMemo(() => investingQuery.data?.teams ?? [], [investingQuery.data?.teams]);
  const BUDGET = pool?.budgetCredits ?? 100;
  const MIN_TEAMS = pool?.minTeams ?? 3;
  const MAX_TEAMS = pool?.maxTeams ?? 10;
  const MAX_INVESTMENT = pool?.maxInvestmentCredits ?? 50;

  // Initialize slots from existing investments
  React.useEffect(() => {
    if (investingQuery.data && !initializedRef.current) {
      const { initialInvestments, teams: loadedTeams } = investingQuery.data;
      const maxTeams = investingQuery.data.pool?.maxTeams ?? 10;
      setSlots(initializeSlotsFromInvestments(initialInvestments, loadedTeams, maxTeams));
      initialInvestmentsRef.current = { ...initialInvestments };
      initializedRef.current = true;
    }
  }, [investingQuery.data]);

  // Derive investmentsByTeamId from slots for validation/budget computation
  const investmentsByTeamId = useMemo(() => deriveInvestmentsByTeamId(slots), [slots]);

  // Track dirty state by comparing current investments to initial investments
  useEffect(() => {
    if (!initializedRef.current) return;
    const initial = initialInvestmentsRef.current;
    const currentKeys = Object.keys(investmentsByTeamId);
    const initialKeys = Object.keys(initial);
    if (currentKeys.length !== initialKeys.length) {
      setIsDirty(true);
      return;
    }
    const changed = currentKeys.some((k) => investmentsByTeamId[k] !== initial[k]);
    setIsDirty(changed);
  }, [investmentsByTeamId]);

  // Warn on browser navigation (close tab, refresh) when form is dirty
  useEffect(() => {
    if (!isDirty) return;
    const handleBeforeUnload = (e: BeforeUnloadEvent) => {
      e.preventDefault();
    };
    window.addEventListener('beforeunload', handleBeforeUnload);
    return () => window.removeEventListener('beforeunload', handleBeforeUnload);
  }, [isDirty]);

  // Team options for combobox
  const teamOptions = useMemo(() => deriveTeamOptions(teams), [teams]);

  // Set of already-selected team IDs
  const usedTeamIds = useMemo(() => deriveUsedTeamIds(slots), [slots]);

  const updatePortfolioMutation = useMutation({
    mutationFn: async (teamsPayload: Array<{ teamId: string; credits: number }>) => {
      if (!poolId || !portfolioId) throw new Error('Missing pool or portfolio ID');
      return poolService.updatePortfolio(poolId, portfolioId, teamsPayload);
    },
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: queryKeys.pools.dashboard(poolId) });
      queryClient.invalidateQueries({ queryKey: queryKeys.pools.investments(poolId, portfolioId) });
      toast.success('Investments saved successfully!');
      navigate(`/pools/${poolId}`);
    },
  });

  const createPortfolioMutation = useMutation({
    mutationFn: async (teamsPayload: Array<{ teamId: string; credits: number }>) => {
      if (!poolId) throw new Error('Missing pool ID');
      const name = user ? `${user.firstName} ${user.lastName}` : 'My Portfolio';
      return poolService.createPortfolio(poolId, name, teamsPayload);
    },
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: queryKeys.pools.dashboard(poolId) });
      toast.success('Portfolio created successfully!');
      navigate(`/pools/${poolId}`);
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
          investmentAmount: prev[slotIndex].investmentAmount,
        };
        return next;
      });
    },
    [teams],
  );

  const handleSlotClear = useCallback((slotIndex: number) => {
    setSlots((prev) => {
      const next = [...prev];
      next[slotIndex] = createEmptySlot();
      return compactSlots(next);
    });
  }, []);

  const handleSlotSearchChange = useCallback((slotIndex: number, text: string) => {
    setSlots((prev) => {
      const next = [...prev];
      next[slotIndex] = { ...prev[slotIndex], searchText: text };
      return next;
    });
  }, []);

  const handleSlotInvestmentChange = useCallback((slotIndex: number, amount: number) => {
    setSlots((prev) => {
      const next = [...prev];
      next[slotIndex] = { ...prev[slotIndex], investmentAmount: amount };
      return next;
    });
  }, []);

  // Budget & validation
  const budgetRemaining = useMemo(() => {
    return computeBudgetRemaining(investmentsByTeamId, BUDGET);
  }, [investmentsByTeamId, BUDGET]);

  const teamCount = useMemo(() => {
    return computeTeamCount(investmentsByTeamId);
  }, [investmentsByTeamId]);

  const validationErrors = useMemo(() => {
    const teamLookups = (investingQuery.data?.teams ?? []).map((t) => ({
      id: t.id,
      schoolName: t.school?.name ?? '',
    }));
    return computeValidationErrors(
      investmentsByTeamId,
      { minTeams: MIN_TEAMS, maxTeams: MAX_TEAMS, maxInvestmentCredits: MAX_INVESTMENT, budget: BUDGET },
      teamLookups,
    );
  }, [investmentsByTeamId, investingQuery.data?.teams, MIN_TEAMS, MAX_TEAMS, MAX_INVESTMENT, BUDGET]);

  const isValid = validationErrors.length === 0 && teamCount >= MIN_TEAMS && teamCount <= MAX_TEAMS;

  const activeMutation = isCreating ? createPortfolioMutation : updatePortfolioMutation;

  const handleSubmit = () => {
    if (!isValid) return;
    const teamsPayload = Object.entries(investmentsByTeamId).map(([teamId, credits]) => ({
      teamId,
      credits,
    }));
    activeMutation.mutate(teamsPayload);
  };

  const handleCancel = () => {
    if (isDirty) {
      const confirmed = window.confirm('You have unsaved changes. Are you sure you want to leave?');
      if (!confirmed) return;
    }
    navigate(`/pools/${poolId}`);
  };

  return {
    // IDs
    poolId,
    portfolioId,
    isCreating,

    // Query state
    investingQuery,
    updatePortfolioMutation,
    createPortfolioMutation,
    activeMutation,

    // Pool config
    pool,
    BUDGET,
    MIN_INVESTMENT,
    MAX_INVESTMENT,
    MIN_TEAMS,
    MAX_TEAMS,

    // Slot state
    slots,
    teamOptions,
    usedTeamIds,
    teams,

    // Derived investment state
    budgetRemaining,
    teamCount,
    validationErrors,
    isValid,
    isDirty,

    // Slot handlers
    handleSlotSelect,
    handleSlotClear,
    handleSlotSearchChange,
    handleSlotInvestmentChange,

    // Submit
    handleSubmit,
    handleCancel,
  };
}
