import React, { useMemo, useRef, useState } from 'react';
import { useNavigate, useParams } from 'react-router-dom';
import { useQuery, useMutation } from '@tanstack/react-query';
import { calcuttaService } from '../services/calcuttaService';
import { tournamentService } from '../services/tournamentService';
import { schoolService } from '../services/schoolService';
import { queryKeys } from '../queryKeys';
import type { School } from '../types/school';
import type { TournamentTeam } from '../types/calcutta';

const MIN_BID = 1;

export type SeedFilter = 'all' | '1-4' | '5-8' | '9-12' | '13-16';

export const SEED_FILTER_OPTIONS: SeedFilter[] = ['all', '1-4', '5-8', '9-12', '13-16'];

export type TeamWithSchool = TournamentTeam & { school?: School };

export interface PortfolioItem {
  teamId: string;
  name: string;
  seed: number;
  bid: number;
}

export function getSeedVariant(seed: number): 'default' | 'secondary' | 'outline' {
  if (seed <= 4) return 'default';
  if (seed <= 8) return 'secondary';
  return 'outline';
}

export function useBidding() {
  const { calcuttaId, entryId } = useParams<{ calcuttaId: string; entryId: string }>();
  const navigate = useNavigate();

  const [bidsByTeamId, setBidsByTeamId] = useState<Record<string, number>>({});
  const [seedFilter, setSeedFilter] = useState<SeedFilter>('all');
  const [unbidOnly, setUnbidOnly] = useState(false);
  const [collapsedRegions, setCollapsedRegions] = useState<Set<string>>(new Set());
  const [showConfirmModal, setShowConfirmModal] = useState(false);
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
  const BUDGET = calcutta?.budgetPoints ?? 100;
  const MIN_TEAMS = calcutta?.minTeams ?? 3;
  const MAX_TEAMS = calcutta?.maxTeams ?? 10;
  const MAX_BID = calcutta?.maxBid ?? 50;

  React.useEffect(() => {
    if (biddingQuery.data?.initialBids && !initializedRef.current) {
      setBidsByTeamId(biddingQuery.data.initialBids);
      initializedRef.current = true;
    }
  }, [biddingQuery.data?.initialBids]);

  const updateEntryMutation = useMutation({
    mutationFn: async (teams: Array<{ teamId: string; bid: number }>) => {
      if (!entryId) throw new Error('Missing entry ID');
      return calcuttaService.updateEntry(entryId, teams);
    },
    onSuccess: () => {
      navigate(`/calcuttas/${calcuttaId}/entries/${entryId}`);
    },
  });

  const handleBidChange = (teamId: string, bid: number) => {
    setBidsByTeamId((prev) => {
      if (bid === 0) {
        const next = { ...prev };
        delete next[teamId];
        return next;
      }
      return { ...prev, [teamId]: bid };
    });
  };

  const budgetRemaining = useMemo(() => {
    const spent = Object.values(bidsByTeamId).reduce((sum, bid) => sum + bid, 0);
    return BUDGET - spent;
  }, [bidsByTeamId, BUDGET]);

  const teamCount = useMemo(() => {
    return Object.keys(bidsByTeamId).length;
  }, [bidsByTeamId]);

  const validationErrors = useMemo(() => {
    const errors: string[] = [];

    if (teamCount < MIN_TEAMS) {
      errors.push(`Select at least ${MIN_TEAMS} teams`);
    }

    if (teamCount > MAX_TEAMS) {
      errors.push(`Select at most ${MAX_TEAMS} teams`);
    }

    if (budgetRemaining < 0) {
      errors.push(`Over budget by ${Math.abs(budgetRemaining).toFixed(2)} pts`);
    }

    Object.entries(bidsByTeamId).forEach(([teamId, bid]) => {
      if (bid < MIN_BID) {
        errors.push(`All bids must be at least ${MIN_BID} pts`);
      }
      if (bid > MAX_BID) {
        const team = biddingQuery.data?.teams.find((t) => t.id === teamId);
        errors.push(`Bid on ${team?.school?.name || 'team'} exceeds max ${MAX_BID} pts`);
      }
    });

    return errors;
  }, [bidsByTeamId, teamCount, budgetRemaining, biddingQuery.data?.teams, MIN_TEAMS, MAX_TEAMS, MAX_BID]);

  const isValid = validationErrors.length === 0 && teamCount >= MIN_TEAMS && teamCount <= MAX_TEAMS;

  const handleSubmit = () => {
    if (!isValid) return;
    setShowConfirmModal(true);
  };

  const handleConfirm = () => {
    const teams = Object.entries(bidsByTeamId).map(([teamId, bid]) => ({
      teamId,
      bid,
    }));

    updateEntryMutation.mutate(teams);
    setShowConfirmModal(false);
  };

  const sortedTeams = useMemo(() => {
    if (!biddingQuery.data?.teams) return [];
    return [...biddingQuery.data.teams].sort((a, b) => {
      if (a.region !== b.region) return a.region.localeCompare(b.region);
      return a.seed - b.seed;
    });
  }, [biddingQuery.data?.teams]);

  // Group teams by region
  const teamsByRegion = useMemo(() => {
    const groups = new Map<string, typeof sortedTeams>();
    for (const team of sortedTeams) {
      const existing = groups.get(team.region) || [];
      existing.push(team);
      groups.set(team.region, existing);
    }
    return groups;
  }, [sortedTeams]);

  // Apply filters
  const matchesSeedFilter = (seed: number) => {
    if (seedFilter === 'all') return true;
    const [min, max] = seedFilter.split('-').map(Number);
    return seed >= min && seed <= max;
  };

  const toggleRegion = (region: string) => {
    setCollapsedRegions((prev) => {
      const next = new Set(prev);
      if (next.has(region)) {
        next.delete(region);
      } else {
        next.add(region);
      }
      return next;
    });
  };

  // Portfolio summary - teams with bids
  const portfolioSummary = useMemo((): PortfolioItem[] => {
    if (!biddingQuery.data?.teams) return [];
    return Object.entries(bidsByTeamId)
      .filter(([, bid]) => bid > 0)
      .map(([teamId, bid]) => {
        const team = biddingQuery.data!.teams.find((t) => t.id === teamId);
        return {
          teamId,
          name: team?.school?.name || 'Unknown',
          seed: team?.seed || 0,
          bid,
        };
      })
      .sort((a, b) => b.bid - a.bid);
  }, [bidsByTeamId, biddingQuery.data]);

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

    // Bid state
    bidsByTeamId,
    budgetRemaining,
    teamCount,
    validationErrors,
    isValid,

    // Filters
    seedFilter,
    setSeedFilter,
    unbidOnly,
    setUnbidOnly,
    collapsedRegions,

    // Team data
    teamsByRegion,
    portfolioSummary,

    // Handlers
    handleBidChange,
    handleSubmit,
    handleConfirm,
    matchesSeedFilter,
    toggleRegion,

    // Modal
    showConfirmModal,
    setShowConfirmModal,
  };
}
