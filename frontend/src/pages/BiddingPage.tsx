import React, { useMemo, useState } from 'react';
import { Link, useNavigate, useParams } from 'react-router-dom';
import { useQuery, useMutation } from '@tanstack/react-query';
import { calcuttaService } from '../services/calcuttaService';
import { PageContainer, PageHeader } from '../components/ui/Page';
import { Alert } from '../components/ui/Alert';
import { LoadingState } from '../components/ui/LoadingState';
import { Button } from '../components/ui/Button';
import { BudgetTracker } from '../components/Bidding/BudgetTracker';
import { TeamBidRow } from '../components/Bidding/TeamBidRow';

const BUDGET = 100;
const MIN_TEAMS = 3;
const MAX_TEAMS = 10;
const MIN_BID = 1;
const MAX_BID = 50;

export function BiddingPage() {
  const { calcuttaId, entryId } = useParams<{ calcuttaId: string; entryId: string }>();
  const navigate = useNavigate();

  const [bidsByTeamId, setBidsByTeamId] = useState<Record<string, number>>({});

  const biddingQuery = useQuery({
    queryKey: ['biddingPage', calcuttaId, entryId],
    enabled: Boolean(calcuttaId && entryId),
    staleTime: 30_000,
    queryFn: async () => {
      if (!calcuttaId || !entryId) {
        throw new Error('Missing required parameters');
      }

      const [calcutta, entryTeams, schools] = await Promise.all([
        calcuttaService.getCalcutta(calcuttaId),
        calcuttaService.getEntryTeams(entryId, calcuttaId),
        calcuttaService.getSchools(),
      ]);

      const tournamentTeams = await calcuttaService.getTournamentTeams(calcutta.tournamentId);

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
  }, [bidsByTeamId]);

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
      errors.push(`Over budget by $${Math.abs(budgetRemaining).toFixed(2)}`);
    }

    Object.entries(bidsByTeamId).forEach(([teamId, bid]) => {
      if (bid < MIN_BID) {
        errors.push(`All bids must be at least $${MIN_BID}`);
      }
      if (bid > MAX_BID) {
        const team = biddingQuery.data?.teams.find((t) => t.id === teamId);
        errors.push(`Bid on ${team?.school?.name || 'team'} exceeds max $${MAX_BID}`);
      }
    });

    return errors;
  }, [bidsByTeamId, teamCount, budgetRemaining, biddingQuery.data?.teams]);

  const isValid = validationErrors.length === 0 && teamCount >= MIN_TEAMS && teamCount <= MAX_TEAMS;

  const handleSubmit = () => {
    if (!isValid) return;

    const teams = Object.entries(bidsByTeamId).map(([teamId, bid]) => ({
      teamId,
      bid,
    }));

    updateEntryMutation.mutate(teams);
  };

  const sortedTeams = useMemo(() => {
    if (!biddingQuery.data?.teams) return [];
    return [...biddingQuery.data.teams].sort((a, b) => {
      if (a.region !== b.region) return a.region.localeCompare(b.region);
      return a.seed - b.seed;
    });
  }, [biddingQuery.data?.teams]);

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
        <LoadingState label="Loading bidding page..." />
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
      <PageHeader
        title="Place Your Bids"
        subtitle={`Budget: $${BUDGET} | Teams: ${MIN_TEAMS}-${MAX_TEAMS} | Max per team: $${MAX_BID}`}
        actions={
          <div className="flex gap-2">
            <Link to={`/calcuttas/${calcuttaId}/entries/${entryId}`}>
              <Button variant="secondary">Cancel</Button>
            </Link>
            <Button onClick={handleSubmit} disabled={!isValid || updateEntryMutation.isPending} loading={updateEntryMutation.isPending}>
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

      <div className="bg-white shadow rounded-lg overflow-hidden">
        <div className="px-6 py-4 border-b border-gray-200">
          <h2 className="text-lg font-semibold text-gray-900">Tournament Teams</h2>
          <p className="text-sm text-gray-600 mt-1">
            Select teams and enter your bid amounts. Leave bid at $0 to not select a team.
          </p>
        </div>

        <div className="px-6 py-2">
          <div className="grid grid-cols-12 gap-4 py-2 text-sm font-medium text-gray-700 border-b-2 border-gray-300">
            <div className="col-span-5">School</div>
            <div className="col-span-2 text-center">Seed</div>
            <div className="col-span-2 text-center">Region</div>
            <div className="col-span-3">Bid Amount</div>
          </div>

          {sortedTeams.map((team) => {
            const bid = bidsByTeamId[team.id] || 0;
            const validationError =
              bid > 0 && (bid < MIN_BID ? `Min $${MIN_BID}` : bid > MAX_BID ? `Max $${MAX_BID}` : undefined);

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
      </div>
    </PageContainer>
  );
}
