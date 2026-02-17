import React, { useMemo, useState } from 'react';
import { Link, useNavigate, useParams } from 'react-router-dom';
import { useQuery, useMutation } from '@tanstack/react-query';
import { calcuttaService } from '../services/calcuttaService';
import { tournamentService } from '../services/tournamentService';
import { schoolService } from '../services/schoolService';
import { PageContainer, PageHeader } from '../components/ui/Page';
import { Alert } from '../components/ui/Alert';
import { BiddingSkeleton } from '../components/skeletons/BiddingSkeleton';
import { Button } from '../components/ui/Button';
import { BudgetTracker } from '../components/Bidding/BudgetTracker';
import { TeamBidRow } from '../components/Bidding/TeamBidRow';
import { Badge } from '../components/ui/Badge';
import { Modal, ModalActions } from '../components/ui/Modal';
import { Breadcrumb } from '../components/ui/Breadcrumb';
import { cn } from '../lib/cn';
import { queryKeys } from '../queryKeys';

const MIN_BID = 1;

type SeedFilter = 'all' | '1-4' | '5-8' | '9-12' | '13-16';

function getSeedVariant(seed: number): 'default' | 'secondary' | 'outline' {
  if (seed <= 4) return 'default';
  if (seed <= 8) return 'secondary';
  return 'outline';
}

export function BiddingPage() {
  const { calcuttaId, entryId } = useParams<{ calcuttaId: string; entryId: string }>();
  const navigate = useNavigate();

  const [bidsByTeamId, setBidsByTeamId] = useState<Record<string, number>>({});
  const [seedFilter, setSeedFilter] = useState<SeedFilter>('all');
  const [unbidOnly, setUnbidOnly] = useState(false);
  const [collapsedRegions, setCollapsedRegions] = useState<Set<string>>(new Set());
  const [showConfirmModal, setShowConfirmModal] = useState(false);

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
    if (biddingQuery.data?.initialBids) {
      setBidsByTeamId(biddingQuery.data.initialBids);
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
  const portfolioSummary = useMemo(() => {
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

  if (!calcuttaId || !entryId) {
    return (
      <PageContainer>
        <Alert variant="error">Missing required parameters</Alert>
      </PageContainer>
    );
  }

  if (biddingQuery.isLoading) {
    return (
      <PageContainer>
        <BiddingSkeleton />
      </PageContainer>
    );
  }

  if (biddingQuery.isError) {
    const message = biddingQuery.error instanceof Error ? biddingQuery.error.message : 'Failed to fetch data';
    return (
      <PageContainer>
        <Alert variant="error">{message}</Alert>
      </PageContainer>
    );
  }

  return (
    <PageContainer>
      <Breadcrumb
        items={[
          { label: 'Calcuttas', href: '/calcuttas' },
          { label: calcutta?.name ?? 'Pool', href: `/calcuttas/${calcuttaId}` },
          { label: 'Bid' },
        ]}
      />

      <PageHeader
        title="Place Your Bids"
        subtitle={
          <>
            {`Budget: ${BUDGET} pts | Teams: ${MIN_TEAMS}-${MAX_TEAMS} | Max per team: ${MAX_BID} pts`}
            {' '}
            <Link to="/rules" className="text-blue-600 hover:text-blue-800 underline text-sm">Learn the rules</Link>
          </>
        }
        actions={
          <div className="flex gap-2">
            <Link to={`/calcuttas/${calcuttaId}/entries/${entryId}`}>
              <Button variant="secondary">Cancel</Button>
            </Link>
            <Button onClick={handleSubmit} disabled={!isValid || updateEntryMutation.isPending} loading={updateEntryMutation.isPending} title={!isValid && validationErrors.length > 0 ? validationErrors[0] : undefined}>
              {updateEntryMutation.isPending ? 'Saving...' : 'Save Bids'}
            </Button>
          </div>
        }
      />

      {updateEntryMutation.isError && (
        <Alert variant="error" className="mb-4">
          {updateEntryMutation.error instanceof Error ? updateEntryMutation.error.message : 'Failed to save bids'}
        </Alert>
      )}

      <BudgetTracker
        budgetRemaining={budgetRemaining}
        totalBudget={BUDGET}
        teamCount={teamCount}
        minTeams={MIN_TEAMS}
        maxTeams={MAX_TEAMS}
        isValid={isValid}
        validationErrors={validationErrors}
      />

      {/* Portfolio summary */}
      {portfolioSummary.length > 0 && (
        <div className="bg-white shadow rounded-lg p-4 mb-6">
          <h3 className="text-sm font-semibold text-gray-900 mb-2">Your Bids So Far</h3>
          <div className="flex flex-wrap gap-2">
            {portfolioSummary.map((item) => (
              <div key={item.teamId} className="flex items-center gap-1 bg-blue-50 rounded-md px-2 py-1">
                <Badge variant={getSeedVariant(item.seed)} className="text-xs">{item.seed}</Badge>
                <span className="text-sm text-gray-800">{item.name}</span>
                <span className="text-sm font-medium text-blue-700">{item.bid} pts</span>
              </div>
            ))}
          </div>
        </div>
      )}

      {/* Filters */}
      <div className="flex flex-wrap items-center gap-3 mb-4">
        <div className="flex items-center gap-2">
          <span className="text-sm text-gray-600">Seeds:</span>
          {(['all', '1-4', '5-8', '9-12', '13-16'] as SeedFilter[]).map((filter) => (
            <button
              key={filter}
              onClick={() => setSeedFilter(filter)}
              className={cn(
                'px-3 py-1 text-sm rounded-md transition-colors',
                seedFilter === filter
                  ? 'bg-blue-600 text-white'
                  : 'bg-gray-100 text-gray-700 hover:bg-gray-200'
              )}
            >
              {filter === 'all' ? 'All' : filter}
            </button>
          ))}
        </div>
        <label className="flex items-center gap-2 text-sm text-gray-600 cursor-pointer">
          <input
            type="checkbox"
            checked={unbidOnly}
            onChange={(e) => setUnbidOnly(e.target.checked)}
            className="h-4 w-4 rounded border-gray-300"
          />
          Unbid only
        </label>
      </div>

      <div className="bg-white shadow rounded-lg overflow-hidden">
        <div className="px-4 sm:px-6 py-4 border-b border-gray-200">
          <h2 className="text-lg font-semibold text-gray-900">Tournament Teams</h2>
          <p className="text-sm text-gray-600 mt-1">
            Select teams and enter your bid amounts. Leave bid at 0 to skip a team.
          </p>
        </div>

        <div className="px-4 sm:px-6 py-2">
          <div className="hidden sm:grid grid-cols-12 gap-4 py-2 text-sm font-medium text-gray-700 border-b-2 border-gray-300">
            <div className="col-span-5">School</div>
            <div className="col-span-2 text-center">Seed</div>
            <div className="col-span-2 text-center">Region</div>
            <div className="col-span-3">Bid Amount</div>
          </div>

          {Array.from(teamsByRegion.entries()).map(([region, teams]) => {
            const isCollapsed = collapsedRegions.has(region);
            const filteredTeams = teams.filter((team) => {
              if (!matchesSeedFilter(team.seed)) return false;
              if (unbidOnly && bidsByTeamId[team.id]) return false;
              return true;
            });

            if (filteredTeams.length === 0 && (seedFilter !== 'all' || unbidOnly)) return null;

            const regionBidCount = teams.filter((t) => bidsByTeamId[t.id]).length;

            return (
              <div key={region}>
                <button
                  type="button"
                  onClick={() => toggleRegion(region)}
                  className="w-full flex items-center justify-between py-3 px-2 bg-gray-50 hover:bg-gray-100 transition-colors border-b border-gray-200 mt-1"
                >
                  <div className="flex items-center gap-2">
                    <svg
                      className={cn('h-4 w-4 text-gray-500 transition-transform', !isCollapsed && 'rotate-90')}
                      fill="none"
                      viewBox="0 0 24 24"
                      strokeWidth="2"
                      stroke="currentColor"
                    >
                      <path strokeLinecap="round" strokeLinejoin="round" d="M8.25 4.5l7.5 7.5-7.5 7.5" />
                    </svg>
                    <span className="text-sm font-semibold text-gray-800">{region}</span>
                  </div>
                  {regionBidCount > 0 && (
                    <span className="text-xs text-blue-600 font-medium">{regionBidCount} bid{regionBidCount !== 1 ? 's' : ''}</span>
                  )}
                </button>
                {!isCollapsed && filteredTeams.map((team) => {
                  const bid = bidsByTeamId[team.id] || 0;
                  const validationError =
                    bid > 0 ? (bid < MIN_BID ? `Min ${MIN_BID} pts` : bid > MAX_BID ? `Max ${MAX_BID} pts` : undefined) : undefined;

                  return (
                    <TeamBidRow
                      key={team.id}
                      teamId={team.id}
                      schoolName={team.school?.name || 'Unknown'}
                      seed={team.seed}
                      region={team.region}
                      bidAmount={bid}
                      maxBid={MAX_BID}
                      onBidChange={handleBidChange}
                      validationError={validationError}
                    />
                  );
                })}
              </div>
            );
          })}
        </div>
      </div>

      <Modal open={showConfirmModal} onClose={() => setShowConfirmModal(false)} title="Confirm Your Bids">
        <div className="space-y-4">
          <div className="flex justify-between text-sm text-gray-600">
            <span>{portfolioSummary.length} teams selected</span>
            <span>Total spent: {BUDGET - budgetRemaining} pts</span>
          </div>
          <div className="text-sm text-gray-600">
            Budget remaining: {budgetRemaining} pts
          </div>
          <div className="max-h-60 overflow-y-auto space-y-2">
            {portfolioSummary.map((item) => (
              <div key={item.teamId} className="flex items-center justify-between py-1 border-b border-gray-100">
                <div className="flex items-center gap-2">
                  <Badge variant={getSeedVariant(item.seed)} className="text-xs">{item.seed}</Badge>
                  <span className="text-sm text-gray-800">{item.name}</span>
                </div>
                <span className="text-sm font-medium text-blue-700">{item.bid} pts</span>
              </div>
            ))}
          </div>
        </div>
        <ModalActions>
          <Button variant="secondary" onClick={() => setShowConfirmModal(false)}>Go Back</Button>
          <Button onClick={handleConfirm} loading={updateEntryMutation.isPending}>Confirm &amp; Submit</Button>
        </ModalActions>
      </Modal>
    </PageContainer>
  );
}
